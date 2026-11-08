package hypixel

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/guild-link/chatlink/bots"
)

var actions = map[Action]struct {
	replies []Kind
	errors  []*regexp.Regexp
}{
	Invite: {
		replies: []Kind{Invited},
		errors:  templates("{player} is already in another guild!", "{player} is already in your guild!", "You've already invited {player} to your guild! Wait for them to accept!"),
	},
	Accept: {
		replies: []Kind{Joined},
		errors:  templates("{player} is already in another guild!", "{player} is already in your guild!", "Already in a guild!"),
	},
	Mute: {
		replies: []Kind{Muted, ChatMuted},
		errors:  templates("{player} is not in your guild!", "You cannot mute someone for less than a minute"),
	},
	Unmute: {
		replies: []Kind{Unmuted, ChatUnmuted},
		errors:  templates("{player} is not in your guild!", "This player is not muted!"),
	},
	Promote: {
		replies: []Kind{Promoted},
		errors:  templates("You can only promote up to your own rank!"),
	},
	Demote: {
		replies: []Kind{Demoted},
		errors:  templates("{player} is already the lowest rank you've created!"),
	},
	Kick: {
		replies: []Kind{Kicked},
		errors:  templates("{player} is not in your guild!"),
	},
	SetRank: {
		replies: []Kind{Promoted, Demoted},
		errors:  templates("I couldn't find a rank by the name of '{rank}'!", "You can only promote up to your own rank!"),
	},
}

var nameExpr = regexp.MustCompile(`^\w{1,16}$`)

var commonErrors = templates(
	"You must be the Guild Master to use that command!",
	"Can't find a player by the name of '{player}'",
	"Invalid usage! {usage}",
)

func New(bots *bots.Client, timeout time.Duration, maxCommands int) *Client {
	return &Client{bots: bots, timeout: timeout, maxCommands: maxCommands}
}

func (c *Client) Listen(ctx context.Context, group string, handler Handler) {
	c.bots.OnMessage(ctx, group, func(bot *bots.Bot, message string) {
		event, ok := parse(message)
		if !ok || event.Kind == Chat && c.isBot(ctx, event.Player) {
			return
		}

		handler(bot, event)
	})
}

func (c *Client) Command(ctx context.Context, uuid string, where bots.Tags, action Action, player string, args ...string) (string, error) {
	if !nameExpr.MatchString(player) {
		return "", fmt.Errorf("%q isn't a valid Minecraft name", player)
	}

	command := strings.TrimSpace(strings.Join(append([]string{"/g", string(action), player}, args...), " "))
	errs := slices.Concat(actions[action].errors, commonErrors)

	var result string
	err := c.request(ctx, uuid, where, command, func(message string) (bool, error) {
		for _, line := range lines(message) {
			if event, ok := parseLine(line); ok && slices.Contains(actions[action].replies, event.Kind) && strings.EqualFold(cmp.Or(event.Player, "everyone"), player) {
				result = line
				return true, nil
			}

			if matches(errs, line, player) {
				return true, errors.New(line)
			}
		}

		return false, nil
	})

	return result, err
}

func (c *Client) request(ctx context.Context, uuid string, where bots.Tags, command string, read func(message string) (bool, error)) error {
	lock, _ := c.locks.LoadOrStore(uuid, &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
	defer lock.(*sync.Mutex).Unlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	err := c.bots.Request(ctx, uuid, command, where, read)
	if errors.Is(err, context.DeadlineExceeded) {
		return errors.New("no response from Hypixel")
	}

	return err
}

func (c *Client) isBot(ctx context.Context, player string) bool {
	list, err := c.bots.Find(ctx, nil)
	return err != nil || slices.ContainsFunc(list, func(b bots.Bot) bool { return strings.EqualFold(b.Username, player) })
}

func matches(exprs []*regexp.Regexp, line, player string) bool {
	return slices.ContainsFunc(exprs, func(expr *regexp.Regexp) bool {
		values, ok := match(expr, line)
		return ok && (values["player"] == "" || strings.EqualFold(values["player"], player))
	})
}
