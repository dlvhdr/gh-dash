package autocomplete

type UserMentionSource struct{}

func (UserMentionSource) ExtractContext(input string, cursorPos int) Context {
	if input == "" {
		return Context{}
	}

	runes := []rune(input)

	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(runes) {
		cursorPos = len(runes)
	}

	userStart := 0
	for i := cursorPos - 1; i >= 0; i-- {
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
	for i := cursorPos; i < len(runes); i++ {
		if isWordBoundary(runes[i]) {
			userEnd = i
			break
		}
	}

	return Context{
		Start:   userStart,
		End:     userEnd,
		Content: string(runes[userStart+1 : userEnd]),
	}
}

func (UserMentionSource) InsertSuggestion(
	input string,
	suggestion string,
	contextStart int,
	contextEnd int,
) (newInput string, newCursorPos int) {
	runes := []rune(input)
	replacement := "@" + suggestion + " "
	newValue := string(runes[:contextStart]) + replacement + string(runes[contextEnd:])
	newCursorPos = contextStart + len([]rune(replacement))
	return newValue, newCursorPos
}

func (UserMentionSource) ItemsToExclude(input string, cursorPos int) []string {
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
