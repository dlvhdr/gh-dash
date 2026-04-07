package search

import (
	"fmt"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/cmp"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

type Model struct {
	ctx          *context.ProgramContext
	initialValue string
	textInput    textinput.Model
	cmp *cmp.Model
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
	base := lipgloss.NewStyle()
	ti.SetStyles(textinput.Styles{
		Focused: textinput.StyleState{
			Placeholder: lipgloss.NewStyle().Foreground(ctx.Theme.FaintText),
			Prompt:      base.Foreground(ctx.Theme.SecondaryText),
			Text:        base.Foreground(ctx.Theme.PrimaryText),
		},
		Blurred: textinput.StyleState{
			Placeholder: lipgloss.NewStyle().Foreground(ctx.Theme.FaintText),
			Prompt:      base.Foreground(ctx.Theme.SecondaryText),
			Text:        lipgloss.NewStyle().Foreground(ctx.Theme.FaintText),
		},
		Cursor: textinput.CursorStyle{
			Color: ctx.Theme.FaintText,
			Shape: tea.CursorBar,
			Blink: true,
		},
	})
	ti.Prompt = prompt
	ti.Blur()
	ti.SetValue(opts.InitialValue)
	ti.CursorStart()
	m := Model{
		ctx:          ctx,
		textInput:    ti,
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

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
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
	// leave space for at least 2 characters - one character of the input and 1 for the cursor
	// - deduce 4 - 2 for the padding, 2 for the borders
	// - deduce 1 for the cursor
	// - deduce 1 for the spacing between the prompt and text
	return max(
		2,
		ctx.MainContentWidth-lipgloss.Width(m.textInput.Prompt)-4-1-1,
	) // borders + cursor
}

func (m Model) Value() string {
	return m.textInput.Value()
}
