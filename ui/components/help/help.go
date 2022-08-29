package help

import (
	bbHelp "github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/ui/styles"
	"github.com/dlvhdr/gh-dash/utils"
)

type Model struct {
	help    bbHelp.Model
	ShowAll bool
}

func NewModel() Model {
	help := bbHelp.NewModel()
	help.Styles = bbHelp.Styles{
		ShortDesc:      helpTextStyle.Copy().Foreground(styles.DefaultTheme.FaintText),
		FullDesc:       helpTextStyle.Copy(),
		ShortSeparator: helpTextStyle.Copy().Foreground(styles.DefaultTheme.SecondaryBorder),
		FullSeparator:  helpTextStyle.Copy(),
		FullKey:        helpTextStyle.Copy().Foreground(styles.DefaultTheme.PrimaryText),
		ShortKey:       helpTextStyle.Copy(),
		Ellipsis:       helpTextStyle.Copy(),
	}

	return Model{
		help: help,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, utils.Keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			m.ShowAll = m.help.ShowAll
		}
	}

	return m, nil
}

func (m Model) View(ctx context.ProgramContext) string {
	if m.help.ShowAll {
		return styles.FooterStyle.Copy().
			Height(styles.ExpandedHelpHeight - 1).
			Width(ctx.ScreenWidth).
			Render(m.help.View(utils.Keys))
	}

	return styles.FooterStyle.Copy().
		Width(ctx.ScreenWidth).
		Render(m.help.View(utils.Keys))
}

func (m *Model) SetWidth(width int) {
	m.help.Width = width
}
