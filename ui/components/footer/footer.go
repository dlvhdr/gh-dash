package footer

import (
	"strings"

	bbHelp "github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/config"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/ui/keys"
	"github.com/dlvhdr/gh-dash/utils"
)

type Model struct {
	ctx          *context.ProgramContext
	leftSection  *string
	rightSection *string
	help         bbHelp.Model
	ShowAll      bool
}

func NewModel(ctx context.ProgramContext) Model {
	help := bbHelp.NewModel()
	help.ShowAll = true
	help.Styles = ctx.Styles.Help.BubbleStyles
	l := ""
	r := ""
	return Model{
		ctx:          &ctx,
		help:         help,
		leftSection:  &l,
		rightSection: &r,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Keys.Help):
			m.ShowAll = !m.ShowAll
		}
	}

	return m, nil
}

func (m Model) View() string {
	keymap := keys.GetKeyMap(m.ctx.View)

	helpIndicator := lipgloss.NewStyle().
		Background(m.ctx.Theme.FaintText).
		Foreground(m.ctx.Theme.SelectedBackground).
		Padding(0, 1).
		Render("? help")
	viewSwitcher := m.renderViewSwitcher(*m.ctx)
	leftSection := ""
	if m.leftSection != nil {
		leftSection = *m.leftSection
	}
	rightSection := ""
	if m.rightSection != nil {
		rightSection = *m.rightSection
	}
	spacing := lipgloss.NewStyle().
		Background(m.ctx.Theme.SelectedBackground).
		Render(
			strings.Repeat(
				" ",
				utils.Max(0,
					m.ctx.ScreenWidth-lipgloss.Width(
						viewSwitcher,
					)-lipgloss.Width(leftSection)-
						lipgloss.Width(rightSection)-
						lipgloss.Width(
							helpIndicator,
						),
				)))

	footer := m.ctx.Styles.Common.FooterStyle.Copy().
		Render(lipgloss.JoinHorizontal(lipgloss.Top, viewSwitcher, leftSection, spacing, rightSection, helpIndicator))

	if m.ShowAll {
		fullHelp := m.help.View(keymap)
		return lipgloss.JoinVertical(lipgloss.Top, footer, fullHelp)
	}

	return footer
}

func (m *Model) SetWidth(width int) {
	m.help.Width = width
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
	m.help.Styles = ctx.Styles.Help.BubbleStyles
}

func (m *Model) renderViewSwitcher(ctx context.ProgramContext) string {
	view := ""
	if ctx.View == config.PRsView {
		view = " PRs"
	} else {
		view = " Issues"
	}

	return ctx.Styles.Tabs.ViewSwitcher.Copy().
		Render(view)
}

func (m *Model) SetLeftSection(leftSection string) {
	*m.leftSection = leftSection
}

func (m *Model) SetRightSection(rightSection string) {
	*m.rightSection = rightSection
}
