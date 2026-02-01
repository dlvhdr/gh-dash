package autocomplete

import (
	"strings"
)

// LabelInfo contains information about a label at a specific cursor position
// in a comma-separated list of labels.
// StartIdx and EndIdx are rune indices (not byte indices).
type LabelInfo struct {
	Label    string
	StartIdx int
	EndIdx   int
	IsFirst  bool
	IsLast   bool
}

type WordInfo struct {
	Word     string
	StartIdx int
	EndIdx   int
	IsFirst  bool
	IsLast   bool
}

// LabelContextExtractor extracts label context from a comma-separated list
// at the given cursor position. Returns the current label text and its start/end positions.
func LabelContextExtractor(input string, cursorPos int) (context string, start int, end int) {
	info := ExtractLabelAtCursor(input, cursorPos)
	return info.Label, info.StartIdx, info.EndIdx
}

// ExtractLabelAtCursor extracts information about the label at the given cursor position
// in a comma-separated list. It considers the entire word containing the cursor as the
// current label. The cursor position and returned indices are rune-based (not byte-based)
// to correctly handle multi-byte Unicode characters.
func ExtractLabelAtCursor(input string, cursorPos int) LabelInfo {
	if input == "" {
		return LabelInfo{
			Label:    "",
			StartIdx: 0,
			EndIdx:   0,
			IsFirst:  true,
			IsLast:   true,
		}
	}

	runes := []rune(input)

	// Clamp cursor position to valid range
	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(runes) {
		cursorPos = len(runes)
	}

	// Find the comma before the cursor (or start of string)
	startIdx := 0
	for i := cursorPos - 1; i >= 0; i-- {
		if runes[i] == ',' {
			startIdx = i + 1
			break
		}
	}

	// Find the comma after the cursor (or end of string)
	endIdx := len(runes)
	for i := cursorPos; i < len(runes); i++ {
		if runes[i] == ',' {
			endIdx = i
			break
		}
	}

	// Extract and trim the label
	label := strings.TrimSpace(string(runes[startIdx:endIdx]))

	// Determine if this is the first or last label
	isFirst := startIdx == 0
	isLast := endIdx == len(runes)

	return LabelInfo{
		Label:    label,
		StartIdx: startIdx,
		EndIdx:   endIdx,
		IsFirst:  isFirst,
		IsLast:   isLast,
	}
}

// CurrentLabels splits the input by commas and returns all trimmed, non-empty labels.
func CurrentLabels(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	labels := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			labels = append(labels, trimmed)
		}
	}
	return labels
}

// LabelItemsToExclude returns all non-empty labels from the input.
// This is used to exclude already-entered labels from autocomplete suggestions.
func LabelItemsToExclude(input string, cursorPos int) []string {
	if strings.TrimSpace(input) == "" {
		return nil
	}

	// Get all labels
	currentLabels := CurrentLabels(input)
	if currentLabels == nil {
		return nil
	}

	// Filter out all non-empty labels (already entered labels)
	excluded := make([]string, 0, len(currentLabels))
	for _, label := range currentLabels {
		if label != "" {
			excluded = append(excluded, label)
		}
	}

	return excluded
}

// LabelSuggestionInserter inserts a selected label into a comma-separated list.
// It handles proper spacing and cursor positioning.
// Returns the new input string and the new cursor position.
func LabelSuggestionInserter(input string, suggestion string, contextStart int, contextEnd int) (newInput string, newCursorPos int) {
	// Use the working logic: extract label info using cursor position
	labelInfo := ExtractLabelAtCursor(input, contextStart)
	runes := []rune(input)

	// Build replacement with consistent spacing and trailing comma
	var replacement string
	if labelInfo.IsFirst {
		replacement = suggestion + ", "
	} else {
		replacement = " " + suggestion + ", "
	}

	// Determine what comes after the current label
	// Skip existing comma and spaces if present to avoid duplication
	remainingInput := string(runes[labelInfo.EndIdx:])
	// Remove the comma
	remainingInput = strings.TrimPrefix(remainingInput, ",")
	// Skip any spaces after the comma
	remainingInput = strings.TrimLeft(remainingInput, " \t")

	// Build new input by replacing the label at cursor position
	newValue := string(runes[:labelInfo.StartIdx]) + replacement + remainingInput

	// Position cursor after the ", " we added
	newCursorPos = labelInfo.StartIdx + len([]rune(replacement))

	return newValue, newCursorPos
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// isWordBoundary returns true if the rune is a word boundary (whitespace or punctuation)
func isWordBoundary(r rune) bool {
	return isWhitespace(r) || r == ',' || r == '.' || r == '!' || r == '?' || r == ';' || r == ':' || r == '(' || r == ')' || r == '[' || r == ']' || r == '{' || r == '}' || r == '<' || r == '>' || r == '"' || r == '\'' || r == '`'
}

// UserMentionContextExtractor extracts @-mention context from input at the given cursor position.
// It detects if the cursor is in a word that starts with '@' and returns the username being typed
// (without the '@' prefix), along with the start and end positions of the mention.
// If the cursor is not in an @-mention context, it returns empty context and -1 for positions.
func UserMentionContextExtractor(input string, cursorPos int) (context string, start int, end int) {
	if input == "" {
		return "", -1, -1
	}

	runes := []rune(input)

	// Clamp cursor position to valid range
	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(runes) {
		cursorPos = len(runes)
	}

	// Find the start of the current word by looking backwards for word boundaries
	wordStart := 0
	for i := cursorPos - 1; i >= 0; i-- {
		if isWordBoundary(runes[i]) {
			wordStart = i + 1
			break
		}
		wordStart = i
	}

	// Check if this word starts with '@'
	if wordStart >= len(runes) {
		return "", -1, -1
	}
	if runes[wordStart] != '@' {
		return "", -1, -1
	}

	// Find the end of the current word by looking forwards for word boundaries
	wordEnd := len(runes)
	for i := cursorPos; i < len(runes); i++ {
		if isWordBoundary(runes[i]) {
			wordEnd = i
			break
		}
	}

	// Extract the mention text (without the '@' prefix)
	mentionStart := wordStart + 1 // Skip the '@'
	mentionText := string(runes[mentionStart:wordEnd])

	return mentionText, wordStart, wordEnd
}

// UserMentionSuggestionInserter inserts a selected username into an @-mention.
// It replaces the @-mention at the given context position with "@username".
// Returns the new input string and the new cursor position (after the inserted username).
func UserMentionSuggestionInserter(input string, suggestion string, contextStart int, contextEnd int) (newInput string, newCursorPos int) {
	runes := []rune(input)

	// Build replacement with @ prefix and a trailing space
	replacement := "@" + suggestion + " "

	// Build new input by replacing the mention at context position
	newValue := string(runes[:contextStart]) + replacement + string(runes[contextEnd:])

	// Position cursor after the inserted mention and the trailing space
	newCursorPos = contextStart + len([]rune(replacement))

	return newValue, newCursorPos
}

// UserMentionItemsToExclude returns nil for @-mentions since we don't exclude any users.
// Unlike labels where we exclude already-entered labels, for user mentions we want to
// allow mentioning the same user multiple times.
func UserMentionItemsToExclude(input string, cursorPos int) []string {
	return nil
}

// WhitespaceContextExtractor extracts the current whitespace-separated word at the cursor position.
// Unlike @-mentions which require a '@' prefix, this treats any word bounded by whitespace
// as a valid context for autocomplete.
func WhitespaceContextExtractor(input string, cursorPos int) (context string, start int, end int) {
	info := ExtractWordAtCursor(input, cursorPos)
	return info.Word, info.StartIdx, info.EndIdx
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

	// Clamp cursor position to valid range
	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(runes) {
		cursorPos = len(runes)
	}

	// Find the start of the current word by looking backwards for whitespace
	wordStart := 0
	for i := cursorPos - 1; i >= 0; i-- {
		if isWhitespace(runes[i]) {
			wordStart = i + 1
			break
		}
		wordStart = i
	}

	// Find the end of the current word by looking forwards for whitespace
	wordEnd := len(runes)
	for i := cursorPos; i < len(runes); i++ {
		if isWhitespace(runes[i]) {
			wordEnd = i
			break
		}
	}

	// Extract the word text
	wordText := strings.TrimSpace(string((runes[wordStart:wordEnd])))

	// Determine if this is the first or last label
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

// WhitespaceSuggestionInserter inserts a selected suggestion into a whitespace-separated word.
// It replaces the word at the given context position with the suggestion and adds a trailing space.
func WhitespaceSuggestionInserter(input string, suggestion string, contextStart int, contextEnd int) (newInput string, newCursorPos int) {
	runes := []rune(input)

	// Build replacement with trailing space for easy continuation
	replacement := suggestion + " "

	// Build new input by replacing the word at context position
	newValue := string(runes[:contextStart]) + replacement + string(runes[contextEnd:])

	// Position cursor after the inserted suggestion and the trailing space
	newCursorPos = contextStart + len([]rune(replacement))

	return newValue, newCursorPos
}

// WhitespaceItemsToExclude returns all whitespace-separated words from the input
// except the one at the given context range.
func WhitespaceItemsToExclude(input string, cursorPos int) []string {
	if strings.TrimSpace(input) == "" {
		return nil
	}

	// Get all words split by whitespace
	wordInfo := ExtractWordAtCursor(input, cursorPos)

	allWords := AllWords(input)
	if allWords == nil {
		return nil
	}

	// Filter out the word at cursor position by comparing text
	excluded := make([]string, 0, len(allWords))
	for _, word := range allWords {
		if word != wordInfo.Word {
			excluded = append(excluded, word)
		}
	}

	return excluded
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
