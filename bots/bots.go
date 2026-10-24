package bots

import (
	"context"
	"slices"
	"strconv"

	"github.com/chickeniq/chatlink/auth"

	"github.com/valkey-io/valkey-go"
)

func New(valkey valkey.Client) *Client {
	return &Client{valkey: valkey}
}

func (c *Client) Find(ctx context.Context, where Tags) ([]Bot, error) {
	uuids, err := c.valkey.DoCache(ctx, c.valkey.B().Smembers().Key("bots").Cache(), cacheTTL).AsStrSlice()
	if err != nil {
		return nil, err
	}

	bots, err := c.Get(ctx, uuids...)
	if err != nil {
		return nil, err
	}

	return slices.DeleteFunc(bots, func(bot Bot) bool { return !bot.has(where) }), nil
}

func (c *Client) Get(ctx context.Context, uuids ...string) ([]Bot, error) {
	commands := make([]valkey.CacheableTTL, len(uuids))
	for i, uuid := range uuids {
		commands[i] = valkey.CT(c.valkey.B().Hgetall().Key("bots:"+uuid).Cache(), cacheTTL)
	}

	bots := []Bot{}
	for i, result := range c.valkey.DoMultiCache(ctx, commands...) {
		data, err := result.AsStrMap()
		if err != nil {
			return nil, err
		}

		if len(data) != 0 {
			bots = append(bots, parseBot(uuids[i], data))
		}
	}

	return bots, nil
}

func (c *Client) Add(ctx context.Context, account *auth.Account, tags Tags) error {
	return addScript.Exec(ctx, c.valkey, nil, withTags([]string{account.UUID, account.Username, account.Token}, tags)).Error()
}

func (c *Client) SetEnabled(ctx context.Context, uuid string, enabled bool, where Tags) error {
	return setEnabledScript.Exec(ctx, c.valkey, nil, withTags([]string{uuid, strconv.FormatBool(enabled)}, where)).Error()
}

func (c *Client) SetTag(ctx context.Context, uuid, key, value string, where Tags) error {
	return setTagScript.Exec(ctx, c.valkey, nil, withTags([]string{uuid, key, value}, where)).Error()
}

func (c *Client) Remove(ctx context.Context, uuid string, where Tags) error {
	return removeScript.Exec(ctx, c.valkey, nil, withTags([]string{uuid}, where)).Error()
}
