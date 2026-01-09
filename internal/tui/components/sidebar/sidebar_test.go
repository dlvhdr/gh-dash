package sidebar

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

func init() {
	zone.NewGlobal()
}

func TestSidebar(t *testing.T) {
	t.Run("Should return empty string when not open", func(t *testing.T) {
		m := NewModel()
		ctx := createTestContext()
		m.UpdateProgramContext(ctx)
		m.IsOpen = false

		view := m.View()
		if view != "" {
			t.Errorf("Expected empty string when sidebar is closed, got %q", view)
		}
	})

	t.Run("Should set and get resizing state", func(t *testing.T) {
		m := NewModel()
		ctx := createTestContext()
		m.UpdateProgramContext(ctx)
		m.IsOpen = true

		// Test SetResizing and IsResizing
		if m.IsResizing() {
			t.Error("Expected IsResizing to be false initially")
		}

		m.SetResizing(true)
		if !m.IsResizing() {
			t.Error("Expected IsResizing to be true after SetResizing(true)")
		}

		m.SetResizing(false)
		if m.IsResizing() {
			t.Error("Expected IsResizing to be false after SetResizing(false)")
		}
	})

	t.Run("Should calculate new width on mouse motion during resize", func(t *testing.T) {
		m := NewModel()
		ctx := createTestContext()
		m.UpdateProgramContext(ctx)
		m.IsOpen = true
		m.SetResizing(true)

		// Simulate mouse motion
		newX := 60 // This should result in width = ScreenWidth - 60 = 40
		msg := tea.MouseMsg{
			X:      newX,
			Y:      10,
			Button: tea.MouseButtonNone,
			Action: tea.MouseActionMotion,
		}

		_, cmd := m.Update(msg)

		if cmd == nil {
			t.Error("Expected a ResizeMsg command during drag")
		}

		// Execute the command to get the message
		result := cmd()
		resizeMsg, ok := result.(ResizeMsg)
		if !ok {
			t.Errorf("Expected ResizeMsg, got %T", result)
		}

		expectedWidth := ctx.ScreenWidth - newX
		if resizeMsg.NewWidth != expectedWidth {
			t.Errorf("Expected new width %d, got %d", expectedWidth, resizeMsg.NewWidth)
		}
	})

	t.Run("Should clamp resize width to minimum", func(t *testing.T) {
		m := NewModel()
		ctx := createTestContext()
		m.UpdateProgramContext(ctx)
		m.IsOpen = true
		m.SetResizing(true)

		// Simulate mouse motion far to the right (would result in very small width)
		msg := tea.MouseMsg{
			X:      ctx.ScreenWidth - 10, // Would result in width = 10
			Y:      10,
			Button: tea.MouseButtonNone,
			Action: tea.MouseActionMotion,
		}

		_, cmd := m.Update(msg)

		result := cmd()
		resizeMsg, ok := result.(ResizeMsg)
		if !ok {
			t.Errorf("Expected ResizeMsg, got %T", result)
		}

		if resizeMsg.NewWidth < MinPreviewWidth {
			t.Errorf("Width should be at least %d, got %d", MinPreviewWidth, resizeMsg.NewWidth)
		}
	})

	t.Run("Should clamp resize width to maximum", func(t *testing.T) {
		m := NewModel()
		ctx := createTestContext()
		m.UpdateProgramContext(ctx)
		m.IsOpen = true
		m.SetResizing(true)

		// Simulate mouse motion far to the left (would result in very large width)
		msg := tea.MouseMsg{
			X:      0, // Would result in width = ScreenWidth
			Y:      10,
			Button: tea.MouseButtonNone,
			Action: tea.MouseActionMotion,
		}

		_, cmd := m.Update(msg)

		result := cmd()
		resizeMsg, ok := result.(ResizeMsg)
		if !ok {
			t.Errorf("Expected ResizeMsg, got %T", result)
		}

		maxWidth := int(float64(ctx.ScreenWidth) * 0.7)
		if resizeMsg.NewWidth > maxWidth && resizeMsg.NewWidth > MaxPreviewWidth {
			t.Errorf("Width should be at most %d or %d, got %d", maxWidth, MaxPreviewWidth, resizeMsg.NewWidth)
		}
	})

	t.Run("Should stop resizing on mouse release", func(t *testing.T) {
		m := NewModel()
		ctx := createTestContext()
		m.UpdateProgramContext(ctx)
		m.IsOpen = true
		m.SetResizing(true)

		msg := tea.MouseMsg{
			X:      50,
			Y:      10,
			Button: tea.MouseButtonLeft,
			Action: tea.MouseActionRelease,
		}

		updatedModel, cmd := m.Update(msg)
		m = updatedModel

		if m.IsResizing() {
			t.Error("Expected IsResizing to be false after mouse release")
		}

		if cmd == nil {
			t.Error("Expected a ResizeEndMsg command on release")
		}
	})

	t.Run("GetSidebarContentWidth should use PreviewWidth", func(t *testing.T) {
		m := NewModel()
		ctx := createTestContext()
		ctx.PreviewWidth = 80
		m.UpdateProgramContext(ctx)

		width := m.GetSidebarContentWidth()
		expectedWidth := ctx.PreviewWidth - ctx.Styles.Sidebar.BorderWidth
		if width != expectedWidth {
			t.Errorf("Expected content width %d, got %d", expectedWidth, width)
		}
	})
}

func createTestContext() *context.ProgramContext {
	cfg := config.Config{
		Defaults: config.Defaults{
			Preview: config.PreviewConfig{
				Open:  true,
				Width: 50,
			},
		},
		Theme: &config.ThemeConfig{
			Ui: config.UIThemeConfig{
				SectionsShowCount: true,
				Table: config.TableUIThemeConfig{
					ShowSeparator: true,
					Compact:       false,
				},
			},
		},
	}

	ctx := &context.ProgramContext{
		Config:       &cfg,
		ScreenWidth:  100,
		ScreenHeight: 30,
		PreviewWidth: 50,
	}

	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)

	return ctx
}
