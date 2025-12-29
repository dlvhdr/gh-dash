package sidebar

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/keys"
)

const (
	ResizeZoneID     = "sidebar-resize"
	MinPreviewWidth  = 30
	MaxPreviewWidth  = 150
	ResizeHandleChar = "│"
)

// ResizeMsg is sent when the sidebar is resized via mouse drag
type ResizeMsg struct {
	NewWidth int
}

// ResizeStartMsg indicates the start of a resize drag operation
type ResizeStartMsg struct{}

// ResizeEndMsg indicates the end of a resize drag operation
type ResizeEndMsg struct{}

type Model struct {
	IsOpen     bool
	data       string
	viewport   viewport.Model
	ctx        *context.ProgramContext
	emptyState string
	isResizing bool
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
			m.viewport.HalfPageDown()

		case key.Matches(msg, keys.Keys.PageUp):
			m.viewport.HalfPageUp()
		}

	case tea.MouseMsg:
		return m.handleMouseMsg(msg)
	}

	return m, nil
}

func (m Model) handleMouseMsg(msg tea.MouseMsg) (Model, tea.Cmd) {
	if !m.IsOpen || m.ctx == nil {
		return m, nil
	}

	// Handle scroll wheel in the sidebar area
	sidebarStartX := m.ctx.ScreenWidth - m.ctx.PreviewWidth
	if m.ctx.PreviewWidth <= 0 {
		sidebarStartX = m.ctx.ScreenWidth - m.ctx.Config.Defaults.Preview.Width
	}
	if msg.X >= sidebarStartX {
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			m.viewport.LineUp(3)
			return m, nil
		case tea.MouseButtonWheelDown:
			m.viewport.LineDown(3)
			return m, nil
		}
	}

	// Handle resize zone interactions
	if zone.Get(ResizeZoneID).InBounds(msg) {
		switch msg.Action {
		case tea.MouseActionPress:
			if msg.Button == tea.MouseButtonLeft {
				m.isResizing = true
				return m, func() tea.Msg { return ResizeStartMsg{} }
			}
		}
	}

	// Handle drag while resizing
	if m.isResizing {
		switch msg.Action {
		case tea.MouseActionMotion:
			// Calculate new width based on mouse position
			// Mouse X is relative to the terminal, sidebar is on the right
			// New width = ScreenWidth - MouseX
			newWidth := m.ctx.ScreenWidth - msg.X
			if newWidth < MinPreviewWidth {
				newWidth = MinPreviewWidth
			}
			if newWidth > MaxPreviewWidth {
				newWidth = MaxPreviewWidth
			}
			// Don't let the sidebar take more than 70% of the screen
			maxWidth := int(float64(m.ctx.ScreenWidth) * 0.7)
			if newWidth > maxWidth {
				newWidth = maxWidth
			}
			return m, func() tea.Msg { return ResizeMsg{NewWidth: newWidth} }

		case tea.MouseActionRelease:
			m.isResizing = false
			return m, func() tea.Msg { return ResizeEndMsg{} }
		}
	}

	return m, nil
}

// IsResizing returns whether a resize operation is in progress
func (m Model) IsResizing() bool {
	return m.isResizing
}

// SetResizing sets the resizing state
func (m *Model) SetResizing(resizing bool) {
	m.isResizing = resizing
}

func (m Model) View() string {
	if !m.IsOpen {
		return ""
	}

	height := m.ctx.MainContentHeight
	width := m.ctx.PreviewWidth
	if width <= 0 {
		width = m.ctx.Config.Defaults.Preview.Width
	}

	// Create the resize handle (left border) as a zone for mouse interaction
	resizeHandle := m.renderResizeHandle(height)

	// Content style without the left border (we'll add it separately)
	contentStyle := lipgloss.NewStyle().
		Height(height).
		Width(width - 1). // Subtract 1 for the resize handle
		MaxWidth(width - 1)

	var content string
	if m.data == "" {
		content = contentStyle.Align(lipgloss.Center).Render(
			lipgloss.PlaceVertical(height, lipgloss.Center, m.emptyState),
		)
	} else {
		content = contentStyle.Render(lipgloss.JoinVertical(
			lipgloss.Top,
			m.viewport.View(),
			m.ctx.Styles.Sidebar.PagerStyle.
				Render(fmt.Sprintf("%d%%", int(m.viewport.ScrollPercent()*100))),
		))
	}

	// Join the resize handle and content horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, resizeHandle, content)
}

func (m Model) renderResizeHandle(height int) string {
	// Create a vertical line as the resize handle
	handleStyle := lipgloss.NewStyle().
		Foreground(m.ctx.Theme.PrimaryBorder).
		Width(1).
		Height(height)

	// Build the handle string (vertical line)
	handle := ""
	for i := 0; i < height; i++ {
		handle += ResizeHandleChar
		if i < height-1 {
			handle += "\n"
		}
	}

	return zone.Mark(ResizeZoneID, handleStyle.Render(handle))
}

func (m *Model) SetContent(data string) {
	m.data = data
	m.viewport.SetContent(data)
}

func (m *Model) GetSidebarContentWidth() int {
	if m.ctx == nil || m.ctx.Config == nil {
		return 0
	}
	width := m.ctx.PreviewWidth
	if width <= 0 {
		width = m.ctx.Config.Defaults.Preview.Width
	}
	return width - m.ctx.Styles.Sidebar.BorderWidth
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
	m.viewport.Width = m.GetSidebarContentWidth() - 1 // Account for resize handle
}
