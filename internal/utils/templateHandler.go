package utils

import (
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/go-sprout/sprout"
)

type TemplateRegistry struct {
	handler sprout.Handler
}

func NewRegistry() *TemplateRegistry {
	return &TemplateRegistry{}
}

func (or *TemplateRegistry) UID() string {
	return "dlvhdr/gh-dash.registry"
}

func (or *TemplateRegistry) LinkHandler(fh sprout.Handler) error {
	or.handler = fh
	return nil
}

func (or *TemplateRegistry) NowModify(input string) (string, error) {
	now := time.Now()
	duration, err := ParseDuration(input)
	if err != nil {
		log.Error("failed parsing duration", "input", input)
		return "", err
	}

	return now.Add(duration).Format("2006-01-02"), nil
}

func (or *TemplateRegistry) RegisterFunctions(funcsMap sprout.FunctionMap) error {
	sprout.AddFunction(funcsMap, "nowModify", or.NowModify)
	return nil
}

func (or *TemplateRegistry) RegisterAliases(aliasMap sprout.FunctionAliasMap) error {
	return nil
}

// ParseDuration parses a duration string.
// examples: "10d", "-1.5w" or "3Y4M5d".
// Add time units are "d"="D", "w"="W", "mo=M", "y"="Y".
func ParseDuration(s string) (time.Duration, error) {
	neg := false
	if len(s) > 0 && s[0] == '-' {
		neg = true
		s = s[1:]
	}

	re := regexp.MustCompile(`(\d*\.\d+|\d+)[^\d]*`)
	unitMap := map[string]time.Duration{
		"d":  24,
		"D":  24,
		"w":  7 * 24,
		"W":  7 * 24,
		"mo": 30 * 24,
		"M":  30 * 24,
		"y":  365 * 24,
		"Y":  365 * 24,
	}

	strs := re.FindAllString(s, -1)
	var sumDur time.Duration
	for _, str := range strs {
		var _hours time.Duration = 1
		for unit, hours := range unitMap {
			if strings.Contains(str, unit) {
				str = strings.ReplaceAll(str, unit, "h")
				_hours = hours
				break
			}
		}

		dur, err := time.ParseDuration(str)
		if err != nil {
			return 0, err
		}

		sumDur += dur * _hours
	}

	if neg {
		sumDur = -sumDur
	}
	return sumDur, nil
}
