package search

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/ui/context"
)

type Model struct {
	ctx          *context.ProgramContext
	sectionType  string
	initialValue string
	textInput    textinput.Model
}

func NewModel(sectionType string, ctx *context.ProgramContext, initialValue string) Model {
	prompt := fmt.Sprintf(" is:%s ", sectionType)
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.Width = getInputWidth(ctx, prompt)
	ti.PromptStyle = ti.PromptStyle.Foreground(ctx.Theme.SecondaryText)
	ti.Prompt = prompt
	ti.TextStyle = ti.TextStyle.Faint(true)
	ti.Blur()
	ti.SetValue(initialValue)
	ti.CursorStart()

	return Model{
		ctx:          ctx,
		textInput:    ti,
		initialValue: initialValue,
		sectionType:  sectionType,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) View(ctx context.ProgramContext) string {
	return lipgloss.NewStyle().
		Width(ctx.MainContentWidth - 4).
		MaxHeight(3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.ctx.Theme.PrimaryBorder).
		Render(m.textInput.View())
}

func (m *Model) Focus() {
	m.textInput.TextStyle = m.textInput.TextStyle.Faint(false)
	m.textInput.CursorEnd()
	m.textInput.Focus()
}

func (m *Model) Blur() {
	m.textInput.TextStyle = m.textInput.TextStyle.Faint(true)
	m.textInput.CursorStart()
	m.textInput.Blur()
}

func (m *Model) SetValue(val string) {
	m.textInput.SetValue(val)
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.textInput.Width = getInputWidth(ctx, m.textInput.Prompt)
	m.textInput.SetValue(m.textInput.Value())
	m.textInput.Blur()
}

func getInputWidth(ctx *context.ProgramContext, prompt string) int {
	return ctx.MainContentWidth - lipgloss.Width(prompt) - 6
}

func (m Model) Value() string {
	return m.textInput.Value()
}
