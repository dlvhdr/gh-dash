package inputbox

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/autocomplete"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

type Model struct {
	ctx          *context.ProgramContext
	textArea     textarea.Model
	inputHelp    help.Model
	prompt       string
	autocomplete *autocomplete.Model

	// OnSuggestionSelected is called when a user selects an autocomplete suggestion.
	// It receives the selected suggestion, current cursor position, and current value in inputbox.
	// It should return the new value for the inputbox and new cursor position after insertion.
	OnSuggestionSelected func(selected string, cursorPos int, currentValue string) (newValue string, newCursorPos int)

	// CurrentContext extracts the "current context" (e.g., partial label being typed)
	// at the given cursor position, used for filtering autocomplete suggestions.
	CurrentContext func(cursorPos int, currentValue string) string

	// SuggestionsToExclude parses the current value in the inputbox and returns all complete items,
	// used to exclude already-entered items from autocomplete suggestions.
	SuggestionsToExclude func(currentValue string) []string
}

var inputKeys = []key.Binding{
	key.NewBinding(key.WithKeys(tea.KeyCtrlD.String()), key.WithHelp("Ctrl+d", "submit")),
	key.NewBinding(key.WithKeys(tea.KeyCtrlC.String(), tea.KeyEsc.String()), key.WithHelp("Ctrl+c/esc", "cancel")),
	autocomplete.ToggleSuggestions,
}

func NewModel(ctx *context.ProgramContext) Model {
	ta := textarea.New()
	ta.ShowLineNumbers = true
	ta.Prompt = ""
	ta.CharLimit = 65536
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().
		Background(ctx.Theme.FaintBorder).
		Foreground(ctx.Theme.PrimaryText)
	ta.FocusedStyle.LineNumber = lipgloss.NewStyle().Foreground(ctx.Theme.FaintText)
	ta.FocusedStyle.CursorLineNumber = lipgloss.NewStyle().Foreground(ctx.Theme.SecondaryText)
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(ctx.Theme.FaintText)
	ta.FocusedStyle.Text = lipgloss.NewStyle().Foreground(ctx.Theme.PrimaryText)
	ta.FocusedStyle.EndOfBuffer = lipgloss.NewStyle().Foreground(ctx.Theme.FaintText)
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
			currentValue := m.textArea.Value()
			cursorPos := m.GetCursorPosition()
			var currentLabel string
			var existingLabels []string
			if m.CurrentContext != nil {
				currentLabel = m.CurrentContext(cursorPos, currentValue)
			}
			if m.SuggestionsToExclude != nil {
				existingLabels = m.SuggestionsToExclude(currentValue)
			}
			m.autocomplete.Show(currentLabel, existingLabels)
			return m, nil
		}

		// Allow navigation/selection even if the popup is hidden (as long as there are filtered results)
		if m.autocomplete != nil && (m.autocomplete.IsVisible() || m.autocomplete.HasSuggestions()) {
			switch {
			case key.Matches(msg, autocomplete.PrevKey):
				m.autocomplete.Prev()
				return m, nil
			case key.Matches(msg, autocomplete.NextKey):
				m.autocomplete.Next()
				return m, nil
			case key.Matches(msg, autocomplete.SelectKey):
				selected := m.autocomplete.Selected()
				if selected != "" && m.OnSuggestionSelected != nil {
					currentValue := m.textArea.Value()
					cursorPos := m.GetCursorPosition()
					newValue, newCursorPos := m.OnSuggestionSelected(selected, cursorPos, currentValue)
					m.textArea.SetValue(newValue)
					m.textArea.SetCursor(newCursorPos)
				}
				m.autocomplete.Hide()
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
	if m.autocomplete != nil {
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

// GetCursorPosition returns the cursor position within the current logical line
// in runes. This correctly handles multi-byte Unicode characters since the
// textarea internally uses rune-based positioning via [][]rune.
//
// Use this for single-line input contexts like comma-separated labels.
// For multi-line contexts (e.g., @mentions in comments), use GetAbsoluteCursorPosition.
func (m *Model) GetCursorPosition() int {
	lineInfo := m.textArea.LineInfo()
	return lineInfo.StartColumn + lineInfo.ColumnOffset
}
