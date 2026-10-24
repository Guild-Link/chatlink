package bots

import (
	"strconv"
	"strings"
	"time"
)

func isRecent(id string) bool {
	timestamp, _, valid := strings.Cut(id, "-")
	created, err := strconv.ParseInt(timestamp, 10, 64)
	return valid && err == nil && time.Since(time.UnixMilli(created)) < 2*time.Minute
}

func parseBot(uuid string, data map[string]string) Bot {
	tags := Tags{}
	for field, value := range data {
		if key, ok := strings.CutPrefix(field, "tag:"); ok {
			tags[key] = value
		}
	}

	return Bot{
		Enabled:  data["enabled"] == "true",
		Username: data["username"],
		UUID:     uuid,
		Tags:     tags,
	}
}

func (b Bot) has(tags Tags) bool {
	for key, value := range tags {
		if b.Tags[key] != value {
			return false
		}
	}
	return true
}

func withTags(args []string, tags Tags) []string {
	for key, value := range tags {
		args = append(args, key, value)
	}
	return args
}
