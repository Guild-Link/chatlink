package hypixel

import (
	"slices"
	"strings"
)

func (c *Client) ChatCommands(channel Channel, sender, content string) []string {
	prefix := "/" + string(channel) + " "
	if sender != "" {
		prefix += sender + ": "
	}

	head := []rune(prefix)
	head = head[:min(len(head), maxLength-1)]

	var commands []string
	for line := range strings.SplitSeq(content, "\n") {
		for chunk := range slices.Chunk([]rune(strings.Join(strings.Fields(line), " ")), maxLength-len(head)) {
			if len(commands) == c.maxCommands {
				return commands
			}

			commands = append(commands, string(head)+string(chunk))
		}
	}

	return commands
}
