package search

import (
	"fmt"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/cmpcontroller"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/inputbox"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

type Model struct {
	ctx          *context.ProgramContext
	initialValue string
	cmpctl       *cmpcontroller.Controller
}

type SearchOptions struct {
	Prefix       string
	InitialValue string
	Placeholder  string
}

func NewModel(ctx *context.ProgramContext, opts SearchOptions) Model {
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
			Text:        lipgloss.NewStyle().Foreground(ctx.Theme.PrimaryText),
		},
		Cursor: textinput.CursorStyle{
			Color: ctx.Theme.FaintText,
			Shape: tea.CursorBar,
			Blink: true,
		},
	})
	ti.Prompt = fmt.Sprintf(" %s ", opts.Prefix)

	ti.Blur()
	ti.SetValue(opts.InitialValue)
	ti.CursorStart()

	ctl := cmpcontroller.New(ctx, inputbox.ModelOpts{TextInput: &ti})

	m := Model{
		ctx:          ctx,
		initialValue: opts.InitialValue,
		cmpctl:       &ctl,
	}

	m.cmpctl.Exit()

	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	cmd, _ := m.cmpctl.Update(msg)
	return m, cmd
}

func (m Model) View(ctx *context.ProgramContext) string {
	s := m.ctx.Styles.Search.Root
	if m.cmpctl.Focused() {
		s = s.BorderForeground(m.ctx.Styles.Colors.OpenIssue)
	}
	return s.Render(m.cmpctl.View())
}

func (m *Model) CursorEnd() {
	m.cmpctl.CursorEnd()
}

func (m *Model) Focus() tea.Cmd {
	return m.cmpctl.Enter(cmpcontroller.EnterOptions{
		Mode:   cmpcontroller.ModeSearch,
		Prompt: "",
		Repo: cmpcontroller.RepoRef{
			NameWithOwner: "",
			Owner:         "",
			Name:          "",
		},
		SuggestionKind:                   cmpcontroller.SuggestionNone,
		EnterFetch:                       cmpcontroller.FetchNone,
		ConfirmDiscardOnCancel:           false,
		HideAutocompleteWhenContextEmpty: true,
		InitialValue:                     m.cmpctl.Value(),
	})
}

func (m *Model) Blur() {
	m.cmpctl.Exit()
}

func (m *Model) SetValue(val string) {
	m.cmpctl.SetValue(val)
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	oldWidth := m.cmpctl.Width()
	newWidth := m.getInputWidth(ctx)
	m.cmpctl.SetWidth(newWidth)
	if newWidth != oldWidth {
		m.cmpctl.CursorEnd()
	}
}

func (m *Model) getInputWidth(ctx *context.ProgramContext) int {
	return max(
		2,
		ctx.MainContentWidth-4,
	)
}

func (m Model) Value() string {
	return m.cmpctl.Value()
}
