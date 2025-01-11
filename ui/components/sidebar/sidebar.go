package sidebar

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/ui/context"
	"github.com/dlvhdr/gh-dash/v4/ui/keys"
)

type Model struct {
	IsOpen     bool
	data       string
	viewport   viewport.Model
	ctx        *context.ProgramContext
	emptyState string
}

func NewModel() Model {
	return Model{
		IsOpen: false,
		data:   "",
		viewport: viewport.Model{
			Width:  0,
			Height: 0,
		},
		ctx:        nil,
		emptyState: "Nothing selected...",
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Keys.PageDown):
			m.viewport.HalfViewDown()

		case key.Matches(msg, keys.Keys.PageUp):
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
	style := m.ctx.Styles.Sidebar.Root.
		Height(height).
		MaxHeight(height).
		Width(m.ctx.Config.Defaults.Preview.Width).
		MaxWidth(m.ctx.Config.Defaults.Preview.Width)

	if m.data == "" {
		return style.Align(lipgloss.Center).Render(
			lipgloss.PlaceVertical(height, lipgloss.Center, m.emptyState),
		)
	}

	return style.Render(lipgloss.JoinVertical(
		lipgloss.Top,
		m.viewport.View(),
		m.ctx.Styles.Sidebar.PagerStyle.
			Render(fmt.Sprintf("%d%%", int(m.viewport.ScrollPercent()*100))),
	))
}

func (m *Model) SetContent(data string) {
	m.data = data
	m.viewport.SetContent(data)
}

func (m *Model) GetSidebarContentWidth() int {
	if m.ctx.Config == nil {
		return 0
	}
	return m.ctx.Config.Defaults.Preview.Width - 2*m.ctx.Styles.Sidebar.ContentPadding - m.ctx.Styles.Sidebar.BorderWidth
}

func (m *Model) ScrollToTop() {
	m.viewport.GotoTop()
}

func (m *Model) ScrollToBottom() {
	m.viewport.GotoBottom()
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	if ctx == nil {
		return
	}
	m.ctx = ctx
	m.viewport.Height = m.ctx.MainContentHeight - m.ctx.Styles.Sidebar.PagerHeight
	m.viewport.Width = m.GetSidebarContentWidth()
}
