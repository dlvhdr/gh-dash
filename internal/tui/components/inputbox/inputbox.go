package inputbox

import (
	"fmt"
	"strings"

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
}

var inputKeys = []key.Binding{
	key.NewBinding(key.WithKeys(tea.KeyCtrlD.String()), key.WithHelp("Ctrl+d", "submit")),
	key.NewBinding(key.WithKeys(tea.KeyCtrlC.String(), tea.KeyEsc.String()), key.WithHelp("Ctrl+c/esc", "cancel")),
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
		if m.autocomplete != nil && m.autocomplete.IsVisible() {
			switch msg.Type {
			case tea.KeyUp, tea.KeyCtrlP:
				m.autocomplete.Prev()
				return m, nil
			case tea.KeyDown, tea.KeyCtrlN:
				m.autocomplete.Next()
				return m, nil
			case tea.KeyEnter, tea.KeyTab:
				selected := m.autocomplete.GetSelected()
				if selected != "" {
					currentInput := m.textArea.Value()
					currentLabel := extractCurrentLabel(currentInput)
					newLabel := selected

					if currentLabel != "" {
						newInput := strings.ReplaceAll(
							currentInput,
							currentLabel,
							newLabel,
						)
						m.textArea.SetValue(newInput)
					}

					newValue := m.textArea.Value()
					if !strings.HasSuffix(newValue, ", ") {
						m.textArea.SetValue(newValue + ", ")
					}
					m.textArea.SetCursor(len(m.textArea.Value()))
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
	baseView := m.View()
	autocompleteView := ""
	if m.autocomplete != nil {
		autocompleteView = m.autocomplete.View()
	}
	return baseView + "\n" + autocompleteView
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

func extractCurrentLabel(input string) string {
	lastComma := strings.LastIndex(input, ",")
	if lastComma == -1 {
		return input
	}
	return strings.TrimSpace(input[lastComma+1:])
}

func (m *Model) GetCurrentLabel() string {
	return extractCurrentLabel(m.textArea.Value())
}

func (m *Model) GetAllLabels() []string {
	value := m.textArea.Value()
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
