package carousel

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
)

func init() {
	zone.NewGlobal()
}

func TestCarousel(t *testing.T) {
	t.Run("Should create carousel with items", func(t *testing.T) {
		items := []string{"Tab 1", "Tab 2", "Tab 3"}
		c := New(WithItems(items), WithWidth(100))

		if len(c.Items()) != len(items) {
			t.Errorf("Expected %d items, got %d", len(items), len(c.Items()))
		}

		if c.Cursor() != 0 {
			t.Errorf("Expected cursor at 0, got %d", c.Cursor())
		}
	})

	t.Run("Should set cursor position", func(t *testing.T) {
		items := []string{"Tab 1", "Tab 2", "Tab 3"}
		c := New(WithItems(items), WithWidth(100))

		c.SetCursor(2)
		if c.Cursor() != 2 {
			t.Errorf("Expected cursor at 2, got %d", c.Cursor())
		}

		// Should clamp to valid range
		c.SetCursor(10)
		if c.Cursor() != 2 {
			t.Errorf("Expected cursor clamped to 2, got %d", c.Cursor())
		}

		c.SetCursor(-1)
		if c.Cursor() != 0 {
			t.Errorf("Expected cursor clamped to 0, got %d", c.Cursor())
		}
	})

	t.Run("Should move left and right", func(t *testing.T) {
		items := []string{"Tab 1", "Tab 2", "Tab 3"}
		c := New(WithItems(items), WithWidth(100))

		c.SetCursor(1)
		c.MoveLeft()
		if c.Cursor() != 0 {
			t.Errorf("Expected cursor at 0 after MoveLeft, got %d", c.Cursor())
		}

		c.MoveRight()
		if c.Cursor() != 1 {
			t.Errorf("Expected cursor at 1 after MoveRight, got %d", c.Cursor())
		}

		// Should not go below 0
		c.SetCursor(0)
		c.MoveLeft()
		if c.Cursor() != 0 {
			t.Errorf("Expected cursor to stay at 0, got %d", c.Cursor())
		}

		// Should not go above max
		c.SetCursor(2)
		c.MoveRight()
		if c.Cursor() != 2 {
			t.Errorf("Expected cursor to stay at 2, got %d", c.Cursor())
		}
	})

	t.Run("Should render items with zone markers", func(t *testing.T) {
		items := []string{"Tab 1", "Tab 2", "Tab 3"}
		c := New(WithItems(items), WithWidth(100))
		c.UpdateSize()

		view := c.View()
		if view == "" {
			t.Error("Expected non-empty view")
		}

		// The view should contain the items (after zone.Scan)
		scanned := zone.Scan(view)
		for _, item := range items {
			if !strings.Contains(scanned, item) {
				t.Errorf("Expected view to contain %q", item)
			}
		}
	})

	t.Run("HandleClick should return index for clicked tab", func(t *testing.T) {
		items := []string{"Tab 1", "Tab 2", "Tab 3"}
		c := New(WithItems(items), WithWidth(200))
		c.UpdateSize()

		// Render the view to set up zones
		view := c.View()
		_ = zone.Scan(view)

		// Click inside the first tab's zone
		// Zone IDs now include the zonePrefix, so construct it the same way as tabZoneID()
		zoneID := fmt.Sprintf("%s%s%d", TabZonePrefix, c.zonePrefix, 0)
		z := zone.Get(zoneID)
		if z.IsZero() {
			t.Fatal("Expected zone to be registered")
		}
		msg := tea.MouseMsg{
			X:      z.StartX,
			Y:      z.StartY,
			Button: tea.MouseButtonLeft,
			Action: tea.MouseActionRelease,
		}

		result := c.HandleClick(msg)
		if result != 0 {
			t.Errorf("Expected tab index 0, got %d", result)
		}
	})

	t.Run("HandleClick should return -1 for empty items", func(t *testing.T) {
		c := New(WithItems([]string{}), WithWidth(100))
		c.UpdateSize()

		msg := tea.MouseMsg{
			X:      1,
			Y:      1,
			Button: tea.MouseButtonLeft,
			Action: tea.MouseActionRelease,
		}

		result := c.HandleClick(msg)
		if result != -1 {
			t.Errorf("Expected -1 for empty items, got %d", result)
		}
	})

	t.Run("HandleClick should return -1 for click outside tabs", func(t *testing.T) {
		items := []string{"Tab 1", "Tab 2", "Tab 3"}
		c := New(WithItems(items), WithWidth(100))
		c.UpdateSize()

		// Click way outside any tab area
		msg := tea.MouseMsg{
			X:      1000,
			Y:      1000,
			Button: tea.MouseButtonLeft,
			Action: tea.MouseActionRelease,
		}

		result := c.HandleClick(msg)
		if result != -1 {
			t.Errorf("Expected -1 for click outside tabs, got %d", result)
		}
	})
}
