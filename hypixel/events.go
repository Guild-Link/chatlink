package hypixel

import (
	"regexp"
	"strings"

	"github.com/guild-link/chatlink/bots"
)

type eventFormat struct {
	kind Kind
	text string
	expr *regexp.Regexp
}

var chatExpr = regexp.MustCompile(`(?i)^(Guild|Officer) > (?:\[[^]]+] )?(\S+)(?: \[[^]]+])?: (.+)$`)
var guildPrefixExpr = regexp.MustCompile(`(?i)^Guild > `)
var events = []eventFormat{
	format(Invited, "You invited {player} to your guild{rest}"),
	format(Invited, "You sent an offline invite to {player}! They will have 5 minutes to accept once they come online!"),
	format(JoinRequest, "{player} has requested to join the Guild!"),
	format(Joined, "{player} joined the guild!"),
	format(Left, "{player} left the guild!"),
	format(Kicked, "{player} was kicked from the guild by {by}!"),
	format(Promoted, "{player} was promoted from {from} to {to}"),
	format(Demoted, "{player} was demoted from {from} to {to}"),
	format(Muted, "{by} has muted {player} for {duration}"),
	format(Unmuted, "{by} has unmuted {player}"),
	format(ChatMuted, "{by} has muted the guild chat for {duration}"),
	format(ChatUnmuted, "{by} has unmuted the guild chat!"),
	format(ThrottleEnabled, "{by} enabled the chat throttle! You can only send messages every {duration}"),
	format(ThrottleDisabled, "{by} disabled the chat throttle!"),
}

func format(kind Kind, text string) eventFormat {
	return eventFormat{kind: kind, text: text, expr: template(text)}
}

func parse(message string) (Event, bool) {
	for line := range strings.SplitSeq(bots.StripColors(message), "\n") {
		if match := chatExpr.FindStringSubmatch(line); match != nil {
			channel := GuildChat
			if strings.EqualFold(match[1], "Officer") {
				channel = OfficerChat
			}

			return Event{Kind: Chat, Channel: channel, Text: match[2] + ": " + match[3], Player: match[2], Message: match[3]}, true
		}
	}

	for _, line := range lines(message) {
		if event, ok := parseLine(line); ok {
			return event, true
		}
	}

	return Event{}, false
}

func parseLine(line string) (Event, bool) {
	line = guildPrefixExpr.ReplaceAllString(line, "")
	for _, e := range events {
		if values, ok := match(e.expr, line); ok {
			return Event{
				Kind:     e.kind,
				Player:   values["player"],
				By:       values["by"],
				To:       values["to"],
				From:     values["from"],
				Duration: values["duration"],
				Text:     render(e.text, values),
			}, true
		}
	}

	return Event{}, false
}
