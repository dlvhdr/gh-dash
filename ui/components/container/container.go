package container

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/ui/context"
)

type Model struct {
	Ctx         *context.ProgramContext
	Title       string
	Subtitle    string
	Description string
	content     string
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	styles := m.Ctx.Styles.Card
	title := styles.Title.
		Render(m.Title)
	subtitle := styles.Subtitle.
		Render(m.Subtitle)
	titleContainer := styles.TitleContainer.
		Width(m.Ctx.MainContentWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Center, title, subtitle))

	description := m.Ctx.Styles.Card.Content.Render(m.Description)

	content := lipgloss.JoinVertical(lipgloss.Left, titleContainer, description)
	return m.Ctx.Styles.Card.Root.Copy().
		Width(m.Ctx.ScreenWidth).
		Render(content)
}

func (m *Model) SetWidth(width int) {
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.Ctx = ctx
}
