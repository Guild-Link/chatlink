package bots

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/valkey-io/valkey-go"
)

func (c *Client) Send(ctx context.Context, uuid, message string, where Tags) error {
	if message == "" {
		return fmt.Errorf("message is required")
	}

	return sendScript.Exec(ctx, c.valkey, nil, withTags([]string{uuid, message}, where)).Error()
}

func (c *Client) Request(ctx context.Context, uuid, message string, where Tags, read func(line string) (bool, error)) error {
	key := "bots:" + uuid + ":outbox"
	entries, err := c.valkey.Do(ctx, c.valkey.B().Xrevrange().Key(key).End("+").Start("-").Count(1).Build()).AsXRange()
	if err != nil {
		return err
	}

	last := "0-0"
	if len(entries) > 0 {
		last = entries[0].ID
	}

	if err := c.Send(ctx, uuid, message, where); err != nil {
		return err
	}

	for {
		streams, err := c.valkey.Do(ctx, c.valkey.B().Xread().Count(100).Block(5000).Streams().Key(key).Id(last).Build()).AsXRead()
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if valkey.IsValkeyNil(err) {
			continue
		}

		if err != nil {
			return err
		}

		for _, entry := range streams[key] {
			last = entry.ID
			if line := entry.FieldValues["message"]; line != "" {
				if done, err := read(line); done || err != nil {
					return err
				}
			}
		}
	}
}

func StripColors(line string) string {
	return colorExpr.ReplaceAllString(line, "")
}

func (c *Client) Broadcast(ctx context.Context, message, exclude string) error {
	if message == "" {
		return fmt.Errorf("message is required")
	}

	return broadcastScript.Exec(ctx, c.valkey, nil, []string{message, exclude}).Error()
}

func (c *Client) OnMessage(ctx context.Context, group string, handler Handler) {
	go func() {
		for ctx.Err() == nil {
			bots, err := c.Find(ctx, nil)
			if err != nil || len(bots) == 0 {
				time.Sleep(5 * time.Second)
				continue
			}

			keys, ids := make([]string, len(bots)), make([]string, len(bots))
			botMap := make(map[string]*Bot, len(bots))

			for i := range bots {
				keys[i], ids[i] = "bots:"+bots[i].UUID+":outbox", ">"
				botMap[keys[i]] = &bots[i]
			}

			read := c.valkey.B().Xreadgroup().Group(group, group).Count(1).Block(5000).Noack().Streams().Key(keys...).Id(ids...).Build()
			streams, err := c.valkey.Do(ctx, read).AsXRead()
			if valkey.IsValkeyNil(err) || ctx.Err() != nil {
				continue
			}

			if err != nil && strings.HasPrefix(err.Error(), "NOGROUP") {
				for _, key := range keys {
					c.valkey.Do(ctx, c.valkey.B().XgroupCreate().Key(key).Group(group).Id("0").Mkstream().Build())
				}
				continue
			}

			if err != nil {
				log.Printf("read Minecraft messages: %v", err)
				time.Sleep(time.Second)
				continue
			}

			for key, entries := range streams {
				for _, entry := range entries {
					if message := entry.FieldValues["message"]; message != "" && isRecent(entry.ID) {
						handler(botMap[key], message)
					}
				}
			}
		}
	}()
}
