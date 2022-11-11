package help

import (
	bbHelp "github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/ui/common"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/ui/keys"
)

type Model struct {
	ctx     *context.ProgramContext
	help    bbHelp.Model
	ShowAll bool
}

func NewModel(ctx context.ProgramContext) Model {
	help := bbHelp.NewModel()
	help.Styles = ctx.Styles.Help.BubbleStyles
	return Model{
		ctx:  &ctx,
		help: help,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			m.ShowAll = m.help.ShowAll
		}
	}

	return m, nil
}

func (m Model) View() string {
	keymap := keys.GetKeyMap(m.ctx.View)
	if m.help.ShowAll {
		return m.ctx.Styles.Common.FooterStyle.Copy().
			Height(common.ExpandedHelpHeight - 1).
			Width(m.ctx.ScreenWidth).
			Render(m.help.View(keymap))
	}

	return m.ctx.Styles.Common.FooterStyle.Copy().
		Width(m.ctx.ScreenWidth).
		Render(m.help.View(keymap))
}

func (m *Model) SetWidth(width int) {
	m.help.Width = width
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
	m.help.Styles = ctx.Styles.Help.BubbleStyles
}
