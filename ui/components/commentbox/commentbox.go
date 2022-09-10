package commentbox

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/ui/styles"
)

type Model struct {
	textArea    textarea.Model
	commentHelp help.Model
}

var commentKeys = []key.Binding{
	key.NewBinding(key.WithKeys(tea.KeyCtrlD.String()), key.WithHelp("Ctrl+d", "submit")),
	key.NewBinding(key.WithKeys(tea.KeyCtrlC.String(), tea.KeyEsc.String()), key.WithHelp("Ctrl+c/esc", "cancel")),
}

func NewModel() Model {
	ta := textarea.New()
	ta.ShowLineNumbers = true
	ta.Prompt = ""
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().
		Background(styles.DefaultTheme.FaintBorder).
		Foreground(styles.DefaultTheme.PrimaryText)
	ta.FocusedStyle.LineNumber = lipgloss.NewStyle().Foreground(styles.DefaultTheme.FaintText)
	ta.FocusedStyle.CursorLineNumber = lipgloss.NewStyle().Foreground(styles.DefaultTheme.SecondaryText)
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(styles.DefaultTheme.FaintText)
	ta.FocusedStyle.Text = lipgloss.NewStyle().Foreground(styles.DefaultTheme.PrimaryText)
	ta.FocusedStyle.EndOfBuffer = lipgloss.NewStyle().Foreground(styles.DefaultTheme.FaintText)
	ta.Focus()

	helpTextStyle := lipgloss.NewStyle().Foreground(styles.DefaultTheme.SecondaryText)
	h := help.NewModel()
	h.Styles = help.Styles{
		ShortDesc:      helpTextStyle.Copy().Foreground(styles.DefaultTheme.FaintText),
		FullDesc:       helpTextStyle.Copy(),
		ShortSeparator: helpTextStyle.Copy().Foreground(styles.DefaultTheme.SecondaryBorder),
		FullSeparator:  helpTextStyle.Copy(),
		FullKey:        helpTextStyle.Copy().Foreground(styles.DefaultTheme.PrimaryText),
		ShortKey:       helpTextStyle.Copy(),
		Ellipsis:       helpTextStyle.Copy(),
	}

	return Model{
		textArea:    ta,
		commentHelp: h,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.textArea, cmd = m.textArea.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return lipgloss.NewStyle().
		Width(m.textArea.Width()).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(styles.DefaultTheme.SecondaryBorder).
		MarginTop(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				fmt.Sprint("Leave a comment...\n"),
				m.textArea.View(),
				lipgloss.NewStyle().
					MarginTop(1).
					Render(m.commentHelp.ShortHelpView(commentKeys)),
			),
		)

}

func (m *Model) Value() string {
	return m.textArea.Value()
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

func (m *Model) Reset() {
	m.textArea.Reset()
}
