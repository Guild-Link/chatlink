package hypixel

import (
	"regexp"
	"strings"

	"github.com/guild-link/chatlink/bots"
)

var placeholderExpr = regexp.MustCompile(`\\\{(\w+)\\\}`)
var separatorExpr = regexp.MustCompile(`-{10,}`)

func template(text string) *regexp.Regexp {
	expr := placeholderExpr.ReplaceAllStringFunc(regexp.QuoteMeta(text), func(placeholder string) string {
		name := placeholder[2 : len(placeholder)-2]
		if name == "player" || name == "by" {
			return `(?:\[[^]]+] )?(?P<` + name + `>\w{1,16})`
		}

		return `(?P<` + name + `>.+?)`
	})

	return regexp.MustCompile("(?i)^" + expr + "$")
}

func templates(texts ...string) []*regexp.Regexp {
	exprs := make([]*regexp.Regexp, len(texts))
	for i, text := range texts {
		exprs[i] = template(text)
	}

	return exprs
}

func match(expr *regexp.Regexp, line string) (map[string]string, bool) {
	found := expr.FindStringSubmatch(line)
	if found == nil {
		return nil, false
	}

	values := map[string]string{}
	for i, name := range expr.SubexpNames() {
		if name != "" {
			values[name] = found[i]
		}
	}

	return values, true
}

func lines(message string) []string {
	var result []string
	for line := range strings.SplitSeq(separatorExpr.ReplaceAllString(bots.StripColors(message), "\n"), "\n") {
		if line = strings.Join(strings.Fields(line), " "); line != "" {
			result = append(result, line)
		}
	}

	return result
}

func render(text string, values map[string]string) string {
	pairs := make([]string, 0, 2*len(values))
	for name, value := range values {
		pairs = append(pairs, "{"+name+"}", value)
	}

	return strings.NewReplacer(pairs...).Replace(text)
}
