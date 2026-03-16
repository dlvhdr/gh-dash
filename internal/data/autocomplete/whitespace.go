package autocomplete

import "strings"

type WordInfo struct {
	Word     string
	StartIdx int
	EndIdx   int
	IsFirst  bool
	IsLast   bool
}

type WhitespaceSource struct{}

func (WhitespaceSource) ExtractContext(input string, cursorPos int) Context {
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
	contextStart int,
	contextEnd int,
) (newInput string, newCursorPos int) {
	runes := []rune(input)
	replacement := suggestion + " "
	newValue := string(runes[:contextStart]) + replacement + string(runes[contextEnd:])
	newCursorPos = contextStart + len([]rune(replacement))
	return newValue, newCursorPos
}

func (WhitespaceSource) ItemsToExclude(input string, cursorPos int) []string {
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

func ExtractWordAtCursor(input string, cursorPos int) WordInfo {
	if input == "" {
		return WordInfo{
			Word:     "",
			StartIdx: 0,
			EndIdx:   0,
			IsFirst:  true,
			IsLast:   true,
		}
	}

	runes := []rune(input)

	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(runes) {
		cursorPos = len(runes)
	}

	wordStart := 0
	for i := cursorPos - 1; i >= 0; i-- {
		if isWhitespace(runes[i]) {
			wordStart = i + 1
			break
		}
		wordStart = i
	}

	wordEnd := len(runes)
	for i := cursorPos; i < len(runes); i++ {
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
		StartIdx: wordStart,
		EndIdx:   wordEnd,
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
