package cmp

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

type UserMentionSource struct{}

func (UserMentionSource) ExtractContext(input string, cursorPos tea.Position) Context {
	if input == "" {
		return Context{}
	}

	lines := strings.Split(input, "\n")
	if cursorPos.Y > len(lines) {
		return Context{}
	}

	line := lines[cursorPos.Y]
	runes := []rune(line)

	if cursorPos.X < 0 {
		cursorPos.X = 0
	}
	if cursorPos.X > len(runes) {
		cursorPos.X = len(runes)
	}

	userStart := 0
	// If the curosr is on position X - the user types on position X
	// This means the last thing he typed was on X-1
	for i := cursorPos.X - 1; i >= 0; i-- {
		if isWordBoundary(runes[i]) {
			userStart = i + 1
			break
		}
		userStart = i
	}

	if userStart >= len(runes) || runes[userStart] != '@' {
		return Context{}
	}

	userEnd := len(runes)
	for i := cursorPos.X; i < len(runes); i++ {
		if isWordBoundary(runes[i]) {
			userEnd = i
			break
		}
	}

	return Context{
		Start:   tea.Position{X: userStart, Y: cursorPos.Y},
		End:     tea.Position{X: userEnd, Y: cursorPos.Y},
		Content: string(runes[userStart+1 : userEnd]),
	}
}

func (UserMentionSource) InsertSuggestion(
	input string,
	suggestion string,
	contextStart tea.Position,
	contextEnd tea.Position,
) (newInput string, newCursorPos tea.Position) {
	lines := lines(input)
	runes := []rune(lines[contextStart.Y])
	replacement := "@" + suggestion + " "
	newLine := string(runes[:contextStart.X]) + replacement + string(runes[contextEnd.X:])

	before := joinLines(lines[:contextStart.Y])
	if before != "" {
		before += string('\n')
	}
	newValue := before + newLine + joinLines(lines[contextEnd.Y+1:])
	newCursorPos.X = contextStart.X + len([]rune(replacement))
	return newValue, newCursorPos
}

func (UserMentionSource) ItemsToExclude(input string, cursorPos tea.Position) []string {
	return nil
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

func isWordBoundary(r rune) bool {
	return isWhitespace(r) || r == ',' || r == '.' || r == '!' || r == '?' || r == ';' ||
		r == ':' ||
		r == '(' ||
		r == ')' ||
		r == '[' ||
		r == ']' ||
		r == '{' ||
		r == '}' ||
		r == '<' ||
		r == '>' ||
		r == '"' ||
		r == '\'' ||
		r == '`'
}
