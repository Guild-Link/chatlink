package bots

import (
	"time"

	"github.com/valkey-io/valkey-go"
)

const cacheTTL = 5 * time.Minute

type Handler func(bot *Bot, message string)

type Tags map[string]string

type Client struct {
	valkey valkey.Client
}

type Bot struct {
	Tags     Tags
	Username string
	UUID     string
	Enabled  bool
}
