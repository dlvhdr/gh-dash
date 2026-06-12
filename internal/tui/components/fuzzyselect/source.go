package fuzzyselect

import tea "charm.land/bubbletea/v2"

type Context struct {
	Start   tea.Position
	End     tea.Position
	Content string
}

type LoaderContext struct {
	RepoOwner string
	RepoName  string
}

// Sources can load suggestions, return them based on the cursor position and insert them.
type Source interface {
	// ExtractContext returns a context with the current word under the cursor, where
	// it starts and ends.
	// This helps with knowing how to autocomplete the current word.
	ExtractContext(input string, cursorPos tea.Position) Context
	// TODO: use Inserter and remove it from Source
	InsertSuggestion(
		input string,
		suggestion string,
		contextStart tea.Position,
		contextEnd tea.Position,
	) (newInput string, newCursorPos tea.Position)
	ItemsToExclude(input string, cursorPos tea.Position) []string
	Suggestions(input string, cursorPos tea.Position) []Suggestion
	LoadSuggestions(ctx LoaderContext) error
}
