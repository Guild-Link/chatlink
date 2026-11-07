package hypixel

import (
	"regexp"
	"strings"

	"github.com/guild-link/chatlink/bots"
)

const player = `(?:\[[^]]+] )?(\w{1,16})`

var chatExpr = regexp.MustCompile(`^(Guild|Officer) > (?:\[[^]]+] )?(\S+)(?: \[[^]]+])?: (.+)$`)
var eventExprs = []struct {
	expr  *regexp.Regexp
	event func(match []string) Event
}{
	{regexp.MustCompile(`(?m)^` + player + ` has requested to join the Guild!`), func(m []string) Event {
		return Event{Kind: JoinRequest, Player: m[1]}
	}},
	{regexp.MustCompile(`(?m)^` + player + ` joined the guild!`), func(m []string) Event {
		return Event{Kind: Joined, Player: m[1]}
	}},
	{regexp.MustCompile(`(?m)^` + player + ` left the guild!`), func(m []string) Event {
		return Event{Kind: Left, Player: m[1]}
	}},
	{regexp.MustCompile(`(?m)^` + player + ` was kicked from the guild by ` + player + `!`), func(m []string) Event {
		return Event{Kind: Kicked, Player: m[1], By: m[2]}
	}},
	{regexp.MustCompile(`(?m)^` + player + ` was promoted from (.+?) to (.+?)\s*$`), func(m []string) Event {
		return Event{Kind: Promoted, Player: m[1], From: m[2], To: m[3]}
	}},
	{regexp.MustCompile(`(?m)^` + player + ` was demoted from (.+?) to (.+?)\s*$`), func(m []string) Event {
		return Event{Kind: Demoted, Player: m[1], From: m[2], To: m[3]}
	}},
	{regexp.MustCompile(`(?m)^` + player + ` has muted ` + player + ` for (\S+)`), func(m []string) Event {
		return Event{Kind: Muted, Player: m[2], By: m[1], Duration: m[3]}
	}},
	{regexp.MustCompile(`(?m)^` + player + ` has unmuted ` + player + `\s*$`), func(m []string) Event {
		return Event{Kind: Unmuted, Player: m[2], By: m[1]}
	}},
}

func parse(message string) (Event, bool) {
	plain := bots.StripColors(message)
	for _, e := range eventExprs {
		if match := e.expr.FindStringSubmatch(plain); match != nil {
			return e.event(match), true
		}
	}

	line, _, _ := strings.Cut(plain, "\n")
	match := chatExpr.FindStringSubmatch(line)
	if match == nil {
		return Event{}, false
	}

	channel := GuildChat
	if match[1] == "Officer" {
		channel = OfficerChat
	}

	return Event{Kind: Chat, Channel: channel, Player: match[2], Message: match[3]}, true
}
