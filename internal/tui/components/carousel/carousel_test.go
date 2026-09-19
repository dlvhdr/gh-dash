package carousel

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/exp/golden"
)

func TestOverflow(t *testing.T) {
	w := 69
	m := NewModel(
		WithHeight(1),
		WithWidth(w),
		WithOverflowIndicators("←", "→"),
		WithSeparators(),
		WithItems(
			[]string{
				" Mine (0) ",
				" Review (1) ",
				" All (560) ",
				" Scratch (0) ",
				" Scratch (0) ",
			},
		),
	)
	v := lipgloss.JoinVertical(
		lipgloss.Left,
		strings.Repeat("-", w),
		lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			Render(m.View()),
	)
	golden.RequireEqual(t, v)
}

func TestOverflow_LastItemSelected(t *testing.T) {
	w := 69
	m := NewModel(
		WithHeight(1),
		WithWidth(w),
		WithOverflowIndicators("←", "→"),
		WithSeparators(),
		WithItems(
			[]string{
				" Mine (0) ",
				" Review (1) ",
				" All (560) ",
				" Scratch (0) ",
				" Scratch (0) ",
			},
		),
	)
	m.SetCursor(4)
	v := lipgloss.JoinVertical(
		lipgloss.Left,
		strings.Repeat("-", w),
		lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			Render(m.View()),
	)
	golden.RequireEqual(t, v)
}

func TestOverflow_WithRightIndicator(t *testing.T) {
	w := 69
	m := NewModel(
		WithHeight(1),
		WithWidth(w),
		WithOverflowIndicators("←", "→"),
		WithSeparators(),
		WithItems(
			[]string{
				" Mine (0) ",
				" Review (1) ",
				" All (560) ",
				" Scratch (1) ",
				" Scratch (2) ",
				" Scratch (3) ",
			},
		),
	)
	v := lipgloss.JoinVertical(
		lipgloss.Left,
		strings.Repeat("-", w),
		lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			Render(m.View()),
	)
	golden.RequireEqual(t, v)
}

func TestOverflow_WithLeftIndicator(t *testing.T) {
	w := 69
	m := NewModel(
		WithHeight(1),
		WithWidth(w),
		WithOverflowIndicators("←", "→"),
		WithSeparators(),
		WithItems(
			[]string{
				" Mine (0) ",
				" Review (1) ",
				" All (560) ",
				" Scratch (1) ",
				" Scratch (2) ",
				" Scratch (3) ",
			},
		),
	)
	m.SetCursor(4)
	v := lipgloss.JoinVertical(
		lipgloss.Left,
		strings.Repeat("-", w),
		lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			Render(m.View()),
	)
	golden.RequireEqual(t, v)
}

func TestNoItems(t *testing.T) {
	w := 69
	m := NewModel(
		WithHeight(1),
		WithWidth(w),
		WithOverflowIndicators("←", "→"),
		WithSeparators(),
		WithItems(
			[]string{},
		),
	)
	v := lipgloss.JoinVertical(
		lipgloss.Left,
		strings.Repeat("-", w),
		lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			Render(m.View()),
	)
	golden.RequireEqual(t, v)
}
