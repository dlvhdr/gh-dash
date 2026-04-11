package search

import (
	"fmt"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

type Model struct {
	ctx          *context.ProgramContext
	initialValue string
	textInput    textarea.Model
}

type SearchOptions struct {
	Prefix       string
	InitialValue string
	Placeholder  string
}

func NewModel(ctx *context.ProgramContext, opts SearchOptions) Model {
	ta := textarea.New()
	ta.Placeholder = opts.Placeholder
	base := lipgloss.NewStyle()
	ta.SetStyles(textarea.Styles{
		Focused: textarea.StyleState{
			Placeholder: lipgloss.NewStyle().Foreground(ctx.Theme.FaintText),
			Prompt:      base.Foreground(ctx.Theme.SecondaryText),
			Text:        base.Foreground(ctx.Theme.PrimaryText),
		},
		Blurred: textarea.StyleState{
			Placeholder: lipgloss.NewStyle().Foreground(ctx.Theme.FaintText),
			Prompt:      base.Foreground(ctx.Theme.SecondaryText),
			Text:        lipgloss.NewStyle().Foreground(ctx.Theme.FaintText),
		},
		Cursor: textarea.CursorStyle{
			Color: ctx.Theme.FaintText,
			Shape: tea.CursorBar,
			Blink: true,
		},
	})
	ta.Prompt = fmt.Sprintf(" %s ", opts.Prefix)

	// act as an input to allow reuse of autocomplete
	ta.MaxHeight = 1
	ta.SetHeight(1)

	ta.ShowLineNumbers = false
	ta.Blur()
	ta.SetValue(opts.InitialValue)
	ta.CursorStart()

	m := Model{
		ctx:          ctx,
		textInput:    ta,
		initialValue: opts.InitialValue,
	}

	w := m.getInputWidth(m.ctx)
	m.textInput.SetWidth(w)

	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	cmds := make([]tea.Cmd, 0)

	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View(ctx *context.ProgramContext) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.ctx.Theme.PrimaryBorder).
		Render(m.textInput.View())
}

func (m *Model) Focus() tea.Cmd {
	m.textInput.CursorEnd()
	return m.textInput.Focus()
}

func (m *Model) Blur() {
	m.textInput.CursorStart()
	m.textInput.Blur()
}

func (m *Model) SetValue(val string) {
	m.textInput.SetValue(val)
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	oldWidth := m.textInput.Width()
	newWidth := m.getInputWidth(ctx)
	m.textInput.SetWidth(newWidth)
	if m.textInput.Width() != oldWidth {
		m.textInput.CursorEnd()
	}
}

func (m *Model) getInputWidth(ctx *context.ProgramContext) int {
	return max(
		2,
		ctx.MainContentWidth,
	)
}

func (m Model) Value() string {
	return m.textInput.Value()
}
