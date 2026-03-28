package autocomplete

import (
	"strings"
)

func lines(input string) []string {
	return strings.Split(input, string('\n'))
}

func joinLines(lines []string) string {
	return strings.Join(lines, string('\n'))
}
