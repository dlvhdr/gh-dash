package prview

import "strings"

func stripANSIReset(value string) string {
	return strings.ReplaceAll(value, "\x1b[0m", "")
}
