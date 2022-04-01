package sidebar

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/ui/constants"
	"github.com/dlvhdr/gh-dash/ui/context"
)

type Model struct {
	IsOpen     bool
	viewport   viewport.Model
	ctx        *context.ProgramContext
	emptyState string
}

func NewModel() Model {
	return Model{
		IsOpen: false,
		viewport: viewport.Model{
			Width:  0,
			Height: 0,
		},
		ctx:        nil,
		emptyState: "",
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, constants.Keys.PageDown):
			m.viewport.HalfViewDown()

		case key.Matches(msg, constants.Keys.PageUp):
			m.viewport.HalfViewUp()
		}
	}

	return m, nil
}

func (m Model) View() string {
	if !m.IsOpen {
		return ""
	}

	height := m.ctx.MainContentHeight
	style := sideBarStyle.Copy().
		Height(height).
		MaxHeight(height).
		Width(m.ctx.Config.Defaults.Preview.Width).
		MaxWidth(m.ctx.Config.Defaults.Preview.Width)

	if m.viewport.View() == "" {
		return style.Copy().Align(lipgloss.Center).Render(
			lipgloss.PlaceVertical(height, lipgloss.Center, m.emptyState),
		)
	}

	return style.Copy().Render(lipgloss.JoinVertical(
		lipgloss.Top,
		m.viewport.View(),
		pagerStyle.Copy().Render(fmt.Sprintf("%d%%", int(m.viewport.ScrollPercent()*100))),
	))
}

func (m *Model) SetContent(data string) {
	m.viewport.SetContent(data)
}

func (m *Model) GetSidebarContentWidth() int {
	if m.ctx.Config == nil {
		return 0
	}
	return m.ctx.Config.Defaults.Preview.Width - 2*contentPadding - borderWidth
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	if ctx == nil {
		return
	}
	m.ctx = ctx
	m.viewport.Height = m.ctx.MainContentHeight - pagerHeight
	m.viewport.Width = m.GetSidebarContentWidth()
}
