package themeselector

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

// ThemeSelectedMsg is sent when a theme is selected
type ThemeSelectedMsg struct {
	ThemeID string
}

// ThemeSelectorClosedMsg is sent when the selector is closed without selection
type ThemeSelectorClosedMsg struct{}

type Model struct {
	ctx          *context.ProgramContext
	themes       []config.AvailableTheme
	cursor       int
	currentTheme string
	width        int
	height       int
	visible      bool
}

func NewModel(ctx *context.ProgramContext) Model {
	themes, _ := config.LoadAvailableThemes()
	currentTheme := config.GetCurrentTheme()

	// Find current theme index
	cursor := 0
	for i, t := range themes {
		if t.ID == currentTheme {
			cursor = i
			break
		}
	}

	return Model{
		ctx:          ctx,
		themes:       themes,
		cursor:       cursor,
		currentTheme: currentTheme,
		visible:      false,
	}
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) Show() {
	m.visible = true
	// Reload themes in case user added new ones
	themes, _ := config.LoadAvailableThemes()
	m.themes = themes
}

func (m *Model) Hide() {
	m.visible = false
}

func (m Model) IsVisible() bool {
	return m.visible
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if m.cursor < len(m.themes)-1 {
				m.cursor++
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if m.cursor < len(m.themes) {
				selectedTheme := m.themes[m.cursor]
				m.currentTheme = selectedTheme.ID
				m.visible = false
				return m, func() tea.Msg {
					return ThemeSelectedMsg{ThemeID: selectedTheme.ID}
				}
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q", "t"))):
			m.visible = false
			return m, func() tea.Msg {
				return ThemeSelectorClosedMsg{}
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	if !m.visible {
		return ""
	}

	// Modal styling - use theme colors
	modalWidth := 40
	modalHeight := min(len(m.themes)+4, 20)

	// Use theme colors
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.ctx.Theme.InvertedText).
		Background(m.ctx.Theme.PrimaryBorder).
		Padding(0, 1).
		Width(modalWidth - 2).
		Align(lipgloss.Center)

	itemStyle := lipgloss.NewStyle().
		Foreground(m.ctx.Theme.PrimaryText).
		Padding(0, 2).
		Width(modalWidth - 2)

	selectedStyle := itemStyle.
		Background(m.ctx.Theme.SelectedBackground).
		Foreground(m.ctx.Theme.PrimaryText)

	currentMarker := lipgloss.NewStyle().
		Foreground(m.ctx.Theme.SuccessText).
		Render(" ●")

	// Build theme list
	var items []string
	items = append(items, titleStyle.Render("Select Theme"))
	items = append(items, "")

	visibleStart := 0
	visibleEnd := len(m.themes)
	maxVisible := modalHeight - 4

	if len(m.themes) > maxVisible {
		if m.cursor >= maxVisible/2 {
			visibleStart = m.cursor - maxVisible/2
		}
		visibleEnd = visibleStart + maxVisible
		if visibleEnd > len(m.themes) {
			visibleEnd = len(m.themes)
			visibleStart = visibleEnd - maxVisible
		}
	}

	for i := visibleStart; i < visibleEnd; i++ {
		theme := m.themes[i]
		name := theme.Name
		if theme.ID == m.currentTheme {
			name = name + currentMarker
		}

		if i == m.cursor {
			items = append(items, selectedStyle.Render(fmt.Sprintf("> %s", name)))
		} else {
			items = append(items, itemStyle.Render(fmt.Sprintf("  %s", name)))
		}
	}

	items = append(items, "")
	helpStyle := lipgloss.NewStyle().
		Foreground(m.ctx.Theme.FaintText).
		Padding(0, 2).
		Width(modalWidth - 2)
	items = append(items, helpStyle.Render("↑/↓: navigate • enter: select • esc: close"))

	content := lipgloss.JoinVertical(lipgloss.Left, items...)

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.ctx.Theme.PrimaryBorder).
		Width(modalWidth).
		Background(m.ctx.Theme.MainBackground).
		Padding(0)

	modal := modalStyle.Render(content)

	// Center the modal on screen
	screen := lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		modal,
		lipgloss.WithWhitespaceBackground(m.ctx.Theme.MainBackground),
	)
	return lipgloss.NewStyle().
		Background(m.ctx.Theme.MainBackground).
		Width(m.width).
		Height(m.height).
		Render(screen)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
