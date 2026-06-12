package fuzzyselect

import (
	tea "charm.land/bubbletea/v2"
)

// ListSource implements the Source interface.
type ListSource struct {
	Options []Suggestion
}

func (src *ListSource) ExtractContext(input string, cursorPos tea.Position) Context {
	info := ExtractLabelAtCursor(input, cursorPos)
	return Context{
		Start:   info.StartIdx,
		End:     info.EndIdx,
		Content: info.Label,
	}
}

func (src *ListSource) Suggestions(input string, cursorPos tea.Position) []Suggestion {
	return src.Options
}

func (src *ListSource) InsertSuggestion(
	input string,
	suggestion string,
	contextStart tea.Position,
	contextEnd tea.Position,
) (newInput string, newCursorPos tea.Position) {
	return input + suggestion, tea.Position{X: contextEnd.X + len(suggestion), Y: contextEnd.Y}
}

func (*ListSource) ItemsToExclude(input string, cursorPos tea.Position) []string {
	return nil
}

func (src *ListSource) LoadSuggestions(ctx LoaderContext) error {
	return nil
}
