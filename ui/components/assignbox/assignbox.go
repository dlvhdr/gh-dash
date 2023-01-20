package assignbox

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/ui/context"
)

type Model struct {
	ctx        *context.ProgramContext
	textArea   textarea.Model
	assignHelp help.Model
}

var assignKeys = []key.Binding{
	key.NewBinding(key.WithKeys(tea.KeyCtrlD.String()), key.WithHelp("Ctrl+d", "submit")),
	key.NewBinding(key.WithKeys(tea.KeyCtrlC.String(), tea.KeyEsc.String()), key.WithHelp("Ctrl+c/esc", "cancel")),
}

func NewModel(ctx *context.ProgramContext) Model {
	ta := textarea.New()
	ta.ShowLineNumbers = true
	ta.Prompt = ""
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

	h := help.NewModel()
	h.Styles = ctx.Styles.Help.BubbleStyles
	return Model{
		ctx:        ctx,
		textArea:   ta,
		assignHelp: h,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.textArea, cmd = m.textArea.Update(msg)
	return m, cmd
}

func (m Model) View(assign bool) string {
	var instructionText string
	if assign {
		instructionText = "Assign users (whitespace-separated)... "
	} else {
		instructionText = "Unassign users (whitespace-separated)... "
	}

	return lipgloss.NewStyle().
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(m.ctx.Theme.SecondaryBorder).
		MarginTop(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				fmt.Sprintf("%s\n", instructionText),
				m.textArea.View(),
				lipgloss.NewStyle().
					MarginTop(1).
					Render(m.assignHelp.ShortHelpView(assignKeys)),
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

func (m *Model) SetHeight(height int) {
	m.textArea.SetHeight(height)
}

func (m *Model) Reset() {
	m.textArea.Reset()
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
	m.assignHelp.Styles = ctx.Styles.Help.BubbleStyles
}
