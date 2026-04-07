package cmp

import tea "charm.land/bubbletea/v2"

type Context struct {
	Start   tea.Position
	End     tea.Position
	Content string
}

type Source interface {
	// ExtraContext returns a context with the current word under the cursor, where
	// it starts and ends.
	// This helps with knowing how to autocomplete the current word.
	ExtractContext(input string, cursorPos tea.Position) Context
	InsertSuggestion(
		input string,
		suggestion string,
		contextStart tea.Position,
		contextEnd tea.Position,
	) (newInput string, newCursorPos tea.Position)
	ItemsToExclude(input string, cursorPos tea.Position) []string
}
