package hypixel

import (
	"context"
	"regexp"
	"strings"

	"github.com/guild-link/chatlink/bots"
)

var guildNameExpr = regexp.MustCompile(`^Guild Name: (.+)$`)
var rankExpr = regexp.MustCompile(`^\s*-- (.+) --\s*$`)
var memberLineExpr = regexp.MustCompile(`^(?:\s*(?:\[[^]]+] )?\w{1,16} ●)+\s*$`)
var onlineExpr = regexp.MustCompile(`^Online Members: \d+$`)
var rgbExpr = regexp.MustCompile(`\x1b\[38;2;([0-9;]+)m`)

func (c *Client) Members(ctx context.Context, uuid string, where bots.Tags) (Guild, error) {
	var guild Guild
	err := c.request(ctx, uuid, where, "/g list", func(line string) (bool, error) {
		plain := bots.StripColors(line)
		if match := guildNameExpr.FindStringSubmatch(plain); match != nil {
			guild.Name = match[1]
		}

		if match := rankExpr.FindStringSubmatch(plain); match != nil {
			guild.Ranks = append(guild.Ranks, Rank{Name: match[1]})
		}

		if memberLineExpr.MatchString(plain) && len(guild.Ranks) > 0 {
			r := &guild.Ranks[len(guild.Ranks)-1]
			r.Members = append(r.Members, parseMembers(line)...)
		}

		return onlineExpr.MatchString(plain), nil
	})

	return guild, err
}

func parseMembers(line string) []Member {
	var members []Member
	parts := strings.Split(line, "●")
	color := ""
	for _, part := range parts[:len(parts)-1] {
		if codes := rgbExpr.FindAllStringSubmatch(part, -1); len(codes) > 0 {
			color = codes[len(codes)-1][1]
		}

		fields := strings.Fields(bots.StripColors(part))
		if len(fields) == 0 {
			continue
		}

		members = append(members, Member{Name: fields[len(fields)-1], Online: color == onlineColor})
	}

	return members
}
