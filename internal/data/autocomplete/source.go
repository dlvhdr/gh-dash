package autocomplete

type Context struct {
	Start   int
	End     int
	Content string
}

type Source interface {
	ExtractContext(input string, cursorPos int) Context
	InsertSuggestion(input string, suggestion string, contextStart int, contextEnd int) (newInput string, newCursorPos int)
	ItemsToExclude(input string, cursorPos int) []string
}
