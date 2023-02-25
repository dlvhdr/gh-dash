package sectioncard

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/ui/components/prssection"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/utils"
)

type Model struct {
	Ctx      *context.ProgramContext
	Title    string
	Subtitle string
	Section  *prssection.Model
	loaded   bool
}

func NewModel(ctx *context.ProgramContext) Model {
	return Model{
		Ctx: ctx,
	}
}

func (m *Model) SetDimensions(width, height int) {
	m.Section.SetDimensions(width, utils.Max(0, height))
}

func (m Model) View() string {
	styles := m.Ctx.Styles.Card
	title := styles.Title.
		Render(m.Title)

	subtitle := m.Subtitle
	if m.loaded {
		subtitle = fmt.Sprintf("(%d)", m.Section.TotalCount)
	}
	subtitle = styles.Subtitle.
		Render(subtitle)

	titleContainer := styles.TitleContainer.
		Width(m.Ctx.MainContentWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Center, title, subtitle))

	content := m.Ctx.Styles.Card.Content.Render(m.Section.View())

	root := lipgloss.JoinVertical(lipgloss.Left, titleContainer, content)
	return m.Ctx.Styles.Card.Root.Copy().
		Width(m.Ctx.MainContentWidth).
		Render(root)
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.Ctx = ctx
	m.Section.UpdateProgramContext(ctx)
	m.SetDimensions(ctx.MainContentWidth, len(m.Section.Table.Rows))
}

func (m *Model) SetPrs(msg prssection.SectionPullRequestsFetchedMsg) {
	m.Section.SetPrs(msg)

	limit := 0
	if m.Section.Config.Limit != nil {
		limit = *m.Section.Config.Limit
	}
	m.SetDimensions(
		m.Ctx.MainContentWidth,
		utils.Max(limit, len(m.Section.Table.Rows))+1,
	)
	m.loaded = true
}
