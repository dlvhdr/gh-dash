package search

import (
	"fmt"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/cmpcontroller"
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

	ctl := cmpcontroller.New(ctx, ta)

	m := Model{
		ctx:          ctx,
		initialValue: opts.InitialValue,
		cmpctl:       &ctl,
	}

	w := m.getInputWidth(m.ctx)
	m.cmpctl.SetWidth(w)
	m.cmpctl.Exit()

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	cmd, _ := m.cmpctl.Update(msg)
	return m, cmd
}

func (m Model) View(ctx *context.ProgramContext) string {
	return m.ctx.Styles.Search.Root.Render(m.cmpctl.View())
}

func (m *Model) Focus() tea.Cmd {
	return m.cmpctl.Enter(cmpcontroller.EnterOptions{
		Mode:   cmpcontroller.ModeSearch,
		Prompt: "",
		Repo: cmpcontroller.RepoRef{
			NameWithOwner: "dlvhdr/gh-dash",
			Owner:         "dlvhdr",
			Name:          "gh-dash",
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
	if m.ctx.SidebarOpen {
		return max(
			2,
			ctx.MainContentWidth,
		)
	}

	return max(
		2,
		ctx.MainContentWidth-4,
	)
}

func (m Model) Value() string {
	return m.cmpctl.Value()
}
