package hypixel

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/guild-link/chatlink/bots"
)

var replies = map[Action]string{
	Invite:  `you (?:invited|sent an offline invite to) %s\b`,
	Accept:  `%s joined the guild!`,
	Mute:    `has muted %s for`,
	Unmute:  `has unmuted %s\b`,
	Promote: `%s was promoted from`,
	Demote:  `%s was demoted from`,
	Kick:    `%s was kicked from the guild`,
	SetRank: `%s was (?:promoted|demoted) from`,
}

var errorExprs = []*regexp.Regexp{
	regexp.MustCompile(`^Can't find a player by the name of '.+'$`),
	regexp.MustCompile(`^You (?:do not|don't) have permission`),
	regexp.MustCompile(`^You must be `),
	regexp.MustCompile(`^You cannot `),
	regexp.MustCompile(`^You can only `),
	regexp.MustCompile(` is (?:not|already) in (?:your|another) guild!$`),
	regexp.MustCompile(`^Already in a guild!$`),
	regexp.MustCompile(`^Your guild is full`),
}

var separatorExpr = regexp.MustCompile(`(?m)^\s*-{10,}\s*$`)

func New(bots *bots.Client) *Client {
	return &Client{bots: bots}
}

func (c *Client) Listen(ctx context.Context, group string, handler Handler) {
	c.bots.OnMessage(ctx, group, func(bot *bots.Bot, message string) {
		if event, ok := parse(message); ok {
			handler(bot, event)
		}
	})
}

func (c *Client) Command(ctx context.Context, uuid string, where bots.Tags, action Action, player string, args ...string) (string, error) {
	command := strings.TrimSpace(strings.Join(append([]string{"/g", string(action), player}, args...), " "))
	reply := regexp.MustCompile("(?i)" + fmt.Sprintf(replies[action], `(?:\[[^]]+] )?`+regexp.QuoteMeta(player)))

	var result string
	err := c.request(ctx, uuid, where, command, func(line string) (bool, error) {
		text := separatorExpr.ReplaceAllString(bots.StripColors(line), "")
		text = strings.Join(strings.Fields(text), " ")
		if reply.MatchString(text) {
			result = text
			return true, nil
		}

		if slices.ContainsFunc(errorExprs, func(e *regexp.Regexp) bool { return e.MatchString(text) }) {
			return true, errors.New(text)
		}

		return false, nil
	})

	return result, err
}

func (c *Client) request(ctx context.Context, uuid string, where bots.Tags, command string, read func(line string) (bool, error)) error {
	err := c.bots.Request(ctx, uuid, command, where, read)
	if errors.Is(err, context.DeadlineExceeded) {
		return errors.New("no response from Hypixel")
	}

	return err
}
