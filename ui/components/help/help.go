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
	help    bbHelp.Model
	ShowAll bool
}

func NewModel(ctx context.ProgramContext) Model {
	help := bbHelp.NewModel()
	help.Styles = bbHelp.Styles{
		ShortDesc:      ctx.Styles.Help.Text.Copy().Foreground(ctx.Theme.FaintText),
		FullDesc:       ctx.Styles.Help.Text.Copy(),
		ShortSeparator: ctx.Styles.Help.Text.Copy().Foreground(ctx.Theme.SecondaryBorder),
		FullSeparator:  ctx.Styles.Help.Text.Copy(),
		FullKey:        ctx.Styles.Help.Text.Copy().Foreground(ctx.Theme.PrimaryText),
		ShortKey:       ctx.Styles.Help.Text.Copy(),
		Ellipsis:       ctx.Styles.Help.Text.Copy(),
	}

	return Model{
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

func (m Model) View(ctx context.ProgramContext) string {
	keymap := keys.GetKeyMap(ctx.View)
	if m.help.ShowAll {
		return ctx.Styles.Common.FooterStyle.Copy().
			Height(common.ExpandedHelpHeight - 1).
			Width(ctx.ScreenWidth).
			Render(m.help.View(keymap))
	}

	return ctx.Styles.Common.FooterStyle.Copy().
		Width(ctx.ScreenWidth).
		Render(m.help.View(keymap))
}

func (m *Model) SetWidth(width int) {
	m.help.Width = width
}
