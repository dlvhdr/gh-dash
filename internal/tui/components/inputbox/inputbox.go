package inputbox

import (
	"fmt"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	dataautocomplete "github.com/dlvhdr/gh-dash/v4/internal/data/autocomplete"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/autocomplete"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

type Model struct {
	ctx                *context.ProgramContext
	textArea           textarea.Model
	inputHelp          help.Model
	prompt             string
	autocomplete       *autocomplete.Model
	autocompleteSource dataautocomplete.Source
}

var inputKeys = []key.Binding{
	key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("Ctrl+d", "submit")),
	key.NewBinding(key.WithKeys("ctrl+c", "esc"), key.WithHelp("Ctrl+c/esc", "cancel")),
	autocomplete.ToggleSuggestions,
}

func NewModel(ctx *context.ProgramContext) Model {
	ta := textarea.New()
	ta.ShowLineNumbers = true
	ta.Prompt = ""
	ta.CharLimit = 65536
	base := lipgloss.NewStyle()
	ta.SetStyles(textarea.Styles{
		Focused: textarea.StyleState{
			Base:       base,
			Text:       base.Foreground(ctx.Theme.PrimaryText),
			LineNumber: base.Foreground(ctx.Theme.FaintText),
			CursorLine: base.Background(ctx.Theme.FaintBorder).
				Foreground(ctx.Theme.PrimaryText),
			CursorLineNumber: base.Foreground(ctx.Theme.SecondaryText),
			Placeholder:      base.Foreground(ctx.Theme.FaintText),
			EndOfBuffer:      base.Foreground(ctx.Theme.FaintText),
		},
	})
	ta.Focus()

	h := help.New()
	h.Styles = ctx.Styles.Help.BubbleStyles
	return Model{
		ctx:       ctx,
		textArea:  ta,
		inputHelp: h,
		prompt:    "",
	}
}

func (m *Model) SetAutocomplete(ac *autocomplete.Model) {
	m.autocomplete = ac
}

func (m *Model) SetAutocompleteSource(src dataautocomplete.Source) {
	m.autocompleteSource = src
}

func (m Model) CurrentAutocompleteContext() dataautocomplete.Context {
	if m.autocompleteSource == nil {
		return dataautocomplete.Context{}
	}

	return m.autocompleteSource.ExtractContext(m.textArea.Value(), m.GetAbsoluteCursorPosition())
}

func (m Model) AutocompleteItemsToExclude() []string {
	if m.autocompleteSource == nil {
		return nil
	}

	return m.autocompleteSource.ItemsToExclude(m.textArea.Value(), m.GetAbsoluteCursorPosition())
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Allow toggling suggestions at any time
		if m.autocomplete != nil && key.Matches(msg, autocomplete.ToggleSuggestions) {
			if m.autocomplete.IsVisible() {
				m.autocomplete.Suppress()
				return m, nil
			}

			m.autocomplete.Unsuppress()
			currentContext := m.CurrentAutocompleteContext()
			m.autocomplete.Show(currentContext.Content, m.AutocompleteItemsToExclude())
			return m, nil
		}

		// Allow navigation/selection even if the popup is hidden (as long as there are filtered results)
		if m.autocomplete != nil &&
			(m.autocomplete.IsVisible() || m.autocomplete.HasSuggestions()) {
			switch {
			case key.Matches(msg, autocomplete.PrevKey):
				m.autocomplete.Prev()
				return m, nil
			case key.Matches(msg, autocomplete.NextKey):
				m.autocomplete.Next()
				return m, nil
			case m.autocomplete.Selected() != "" && key.Matches(msg, autocomplete.SelectKey):
				selected := m.autocomplete.Selected()
				if selected != "" && m.autocompleteSource != nil {
					currentValue := m.textArea.Value()
					currentContext := m.CurrentAutocompleteContext()
					newValue, newCursorPos := m.autocompleteSource.InsertSuggestion(
						currentValue,
						selected,
						currentContext.Start,
						currentContext.End,
					)
					m.textArea.SetValue(newValue)
					m.textArea.SetCursorColumn(newCursorPos.X)
					// Refresh autocomplete to exclude the newly-added item
					if m.AutocompleteItemsToExclude() != nil {
						newContext := m.CurrentAutocompleteContext()
						m.autocomplete.Show(newContext.Content, m.AutocompleteItemsToExclude())
					}
				}
				return m, nil
			}
		}
	}

	m.textArea, cmd = m.textArea.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return lipgloss.NewStyle().
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(m.ctx.Theme.SecondaryBorder).
		MarginTop(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				fmt.Sprintf("%s\n", m.prompt),
				m.textArea.View(),
				lipgloss.NewStyle().
					MarginTop(1).
					Render(m.inputHelp.ShortHelpView(inputKeys)),
			),
		)
}

func (m Model) ViewWithAutocomplete() string {
	autocompleteView := ""
	if m.autocomplete != nil && m.autocomplete.IsVisible() {
		autocompleteView = m.autocomplete.View()
	}

	return lipgloss.NewStyle().
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(m.ctx.Theme.SecondaryBorder).
		MarginTop(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				fmt.Sprintf("%s\n", m.prompt),
				m.textArea.View(),
				autocompleteView,
				lipgloss.NewStyle().
					MarginTop(1).
					Render(m.inputHelp.ShortHelpView(inputKeys)),
			),
		)
}

func (m *Model) Value() string {
	return m.textArea.Value()
}

func (m *Model) SetValue(s string) {
	m.textArea.SetValue(s)
}

func (m *Model) Blur() {
	m.textArea.Blur()
}

func (m *Model) Focus() tea.Cmd {
	return m.textArea.Focus()
}

func (m *Model) SetWidth(width int) {
	m.textArea.SetWidth(width)
}

func (m *Model) SetHeight(height int) {
	m.textArea.SetHeight(height)
}

func (m *Model) SetPrompt(prompt string) {
	m.prompt = prompt
}

func (m *Model) Reset() {
	m.textArea.Reset()
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
	m.inputHelp.Styles = ctx.Styles.Help.BubbleStyles
}

func (m *Model) GetAbsoluteCursorPosition() tea.Position {
	line := m.textArea.Line()
	col := m.textArea.Column()
	return tea.Position{X: col, Y: line}
}

func (m *Model) CursorEnd() {
	m.textArea.CursorEnd()
}
