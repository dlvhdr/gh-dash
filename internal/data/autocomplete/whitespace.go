package autocomplete

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

type WordInfo struct {
	Word     string
	StartIdx tea.Position
	EndIdx   tea.Position
	IsFirst  bool
	IsLast   bool
}

type WhitespaceSource struct{}

func (WhitespaceSource) ExtractContext(input string, cursorPos tea.Position) Context {
	info := ExtractWordAtCursor(input, cursorPos)
	return Context{
		Start:   info.StartIdx,
		End:     info.EndIdx,
		Content: info.Word,
	}
}

func (WhitespaceSource) InsertSuggestion(
	input string,
	suggestion string,
	contextStart tea.Position,
	contextEnd tea.Position,
) (newInput string, newCursorPos tea.Position) {
	lines := lines(input)
	runes := []rune(lines[contextStart.Y])
	replacement := suggestion + " "
	newLine := string(runes[:contextStart.X]) + replacement + string(runes[contextEnd.X:])
	newValue := joinLines(lines[:contextStart.Y]) + newLine + joinLines(lines[contextEnd.Y+1:])
	newCursorPos.X = contextStart.X + len([]rune(replacement))
	return newValue, newCursorPos
}

func (WhitespaceSource) ItemsToExclude(input string, cursorPos tea.Position) []string {
	if strings.TrimSpace(input) == "" {
		return nil
	}

	wordInfo := ExtractWordAtCursor(input, cursorPos)
	allWords := AllWords(input)
	if allWords == nil {
		return nil
	}

	excluded := make([]string, 0, len(allWords))
	for _, word := range allWords {
		if word != wordInfo.Word {
			excluded = append(excluded, word)
		}
	}

	return excluded
}

func ExtractWordAtCursor(input string, cursorPos tea.Position) WordInfo {
	if input == "" {
		return WordInfo{
			Word:     "",
			StartIdx: tea.Position{X: 0},
			EndIdx:   tea.Position{X: 0},
			IsFirst:  true,
			IsLast:   true,
		}
	}

	lines := strings.Split(input, "\n")
	if cursorPos.Y > len(lines) {
		return WordInfo{
			Word:     "",
			StartIdx: tea.Position{X: 0},
			EndIdx:   tea.Position{X: 0},
			IsFirst:  true,
			IsLast:   true,
		}
	}

	line := lines[cursorPos.Y]
	runes := []rune(line)

	if cursorPos.X < 0 {
		cursorPos.X = 0
	}
	if cursorPos.X > len(runes) {
		cursorPos.X = len(runes)
	}

	wordStart := 0
	for i := cursorPos.X - 1; i >= 0; i-- {
		if isWhitespace(runes[i]) {
			wordStart = i + 1
			break
		}
		wordStart = i
	}

	wordEnd := len(runes)
	for i := cursorPos.X; i < len(runes); i++ {
		if isWhitespace(runes[i]) {
			wordEnd = i
			break
		}
	}

	wordText := strings.TrimSpace(string(runes[wordStart:wordEnd]))
	isFirst := wordStart == 0
	isLast := wordEnd == len(runes)

	return WordInfo{
		Word:     wordText,
		StartIdx: tea.Position{X: wordStart, Y: cursorPos.Y},
		EndIdx:   tea.Position{X: wordEnd, Y: cursorPos.Y},
		IsFirst:  isFirst,
		IsLast:   isLast,
	}
}

// AllWords splits the input by whitespace and returns all trimmed, non-empty words.
func AllWords(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Fields(value)
	words := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			words = append(words, trimmed)
		}
	}
	return words
}
