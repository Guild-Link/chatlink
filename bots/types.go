package bots

import (
	"regexp"
	"time"

	"github.com/valkey-io/valkey-go"
)

const cacheTTL = 5 * time.Minute

var colorExpr = regexp.MustCompile(`\x1b\[[0-9;]*m|§.`)

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
