package prompt

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
)

type Model struct {
	ctx    *context.ProgramContext
	prompt textinput.Model
}

func NewModel(ctx *context.ProgramContext) Model {
	ti := textinput.New()
	ti.Focus()
	ti.Blur()
	ti.CursorStart()

	return Model{
		ctx:    ctx,
		prompt: ti,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.prompt, cmd = m.prompt.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return m.prompt.View()
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) Blur() {
	m.prompt.Blur()
}

func (m *Model) Focus() tea.Cmd {
	return m.prompt.Focus()
}

func (m *Model) SetValue(value string) {
	m.prompt.SetValue(value)
}

func (m *Model) Value() string {
	return m.prompt.Value()
}

func (m *Model) SetPrompt(prompt string) {
	m.prompt.Prompt = prompt
}

func (m *Model) Reset() {
	m.prompt.Reset()
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
}
