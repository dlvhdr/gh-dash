package search

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

type Model struct {
	ctx          *context.ProgramContext
	initialValue string
	textInput    textinput.Model
}

type SearchOptions struct {
	Prefix       string
	InitialValue string
	Placeholder  string
}

func NewModel(ctx *context.ProgramContext, opts SearchOptions) Model {
	prompt := fmt.Sprintf(" %s ", opts.Prefix)
	ti := textinput.New()
	ti.Placeholder = opts.Placeholder
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(ctx.Theme.FaintText)
	ti.Width = ctx.MainContentWidth - lipgloss.Width(prompt) - 6
	ti.PromptStyle = ti.PromptStyle.Foreground(ctx.Theme.SecondaryText)
	ti.Prompt = prompt
	ti.TextStyle = ti.TextStyle.Faint(true)
	ti.Cursor.Style = ti.Cursor.Style.Faint(true)
	ti.Cursor.TextStyle = ti.Cursor.TextStyle.Faint(true)
	ti.Blur()
	ti.SetValue(opts.InitialValue)
	ti.CursorStart()

	return Model{
		ctx:          ctx,
		textInput:    ti,
		initialValue: opts.InitialValue,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	m.textInput.Width = m.getInputWidth(m.ctx)
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) View(ctx *context.ProgramContext) string {
	return lipgloss.NewStyle().
		Width(ctx.MainContentWidth - 4).
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
	oldWidth := m.textInput.Width
	m.textInput.Width = m.getInputWidth(ctx)
	if m.textInput.Width != oldWidth {
		log.Debug("search width changed, blurring", "mainContentWidth",
			m.ctx.MainContentWidth, "oldWidth", oldWidth, "newWidth", m.textInput.Width)
		m.textInput.CursorEnd()
	}
}

func (m *Model) getInputWidth(ctx *context.ProgramContext) int {
	// leave space for at least 2 characters - one character of the input and 1 for the cursor
	// - deduce 4 - 2 for the padding, 2 for the borders
	// - deduce 1 for the cursor
	// - deduce 1 for the spacing between the prompt and text
	return max(2, ctx.MainContentWidth-lipgloss.Width(m.textInput.Prompt)-4-1-1) // borders + cursor
}

func (m Model) Value() string {
	return m.textInput.Value()
}
