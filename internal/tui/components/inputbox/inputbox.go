package inputbox

import (
	"fmt"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/cmp"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

type Model struct {
	ctx       *context.ProgramContext
	textArea  textarea.Model
	inputHelp help.Model
	prompt    string
	cmp       *cmp.Model
	cmpSource cmp.Source
}

var inputKeys = []key.Binding{
	key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("Ctrl+d", "submit")),
	key.NewBinding(key.WithKeys("ctrl+c", "esc"), key.WithHelp("Ctrl+c/esc", "cancel")),
	cmp.ToggleSuggestions,
}

func DefaultTextArea(ctx *context.ProgramContext) textarea.Model {
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
	return ta
}

func NewModel(ctx *context.ProgramContext, ta textarea.Model) Model {
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

func (m *Model) SetAutocomplete(cmp *cmp.Model) {
	m.cmp = cmp
}

func (m *Model) SetAutocompleteSource(src cmp.Source) {
	m.cmpSource = src
}

func (m Model) CurrentAutocompleteContext() cmp.Context {
	if m.cmpSource == nil {
		return cmp.Context{}
	}

	return m.cmpSource.ExtractContext(m.textArea.Value(), m.GetAbsoluteCursorPosition())
}

func (m Model) AutocompleteItemsToExclude() []string {
	if m.cmpSource == nil {
		return nil
	}

	return m.cmpSource.ItemsToExclude(m.textArea.Value(), m.GetAbsoluteCursorPosition())
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Allow toggling suggestions at any time
		if m.cmp != nil && key.Matches(msg, cmp.ToggleSuggestions) {
			if m.cmp.IsVisible() {
				m.cmp.Suppress()
				return m, nil
			}

			m.cmp.Unsuppress()
			currentContext := m.CurrentAutocompleteContext()
			m.cmp.Show(currentContext.Content, m.AutocompleteItemsToExclude())
			return m, nil
		}

		// Allow navigation/selection even if the popup is hidden (as long as there are filtered results)
		if m.cmp != nil &&
			(m.cmp.IsVisible() || m.cmp.HasSuggestions()) {
			switch {
			case key.Matches(msg, cmp.PrevKey):
				m.cmp.Prev()
				return m, nil
			case key.Matches(msg, cmp.NextKey):
				m.cmp.Next()
				return m, nil
			case m.cmp.Selected() != "" && key.Matches(msg, cmp.SelectKey):
				selected := m.cmp.Selected()
				if selected != "" && m.cmpSource != nil {
					currentValue := m.textArea.Value()
					currentContext := m.CurrentAutocompleteContext()
					newValue, newCursorPos := m.cmpSource.InsertSuggestion(
						currentValue,
						selected,
						currentContext.Start,
						currentContext.End,
					)
					m.textArea.SetValue(newValue)
					m.textArea.SetCursorColumn(newCursorPos.X)
					// Refresh cmp to exclude the newly-added item
					if m.AutocompleteItemsToExclude() != nil {
						newContext := m.CurrentAutocompleteContext()
						m.cmp.Show(newContext.Content, m.AutocompleteItemsToExclude())
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
	cmpView := ""
	if m.cmp != nil && m.cmp.IsVisible() {
		cmpView = m.cmp.View()
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
				cmpView,
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
