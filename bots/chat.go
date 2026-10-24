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
