package autocomplete

import (
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

type Model struct {
	ctx         *context.ProgramContext
	suggestions []string
	filtered    []string
	selected    int
	visible     bool
	maxVisible  int
	posX        int
	posY        int
	width       int
}

func NewModel(ctx *context.ProgramContext) Model {
	return Model{
		ctx:        ctx,
		visible:    false,
		selected:   0,
		maxVisible: 5,
		width:      30,
	}
}

var NextKey = key.NewBinding(
	key.WithKeys("down", "ctrl+n"),
	key.WithHelp("↓/ctrl+n", "next"),
)

func (m *Model) prevKey() key.Binding {
	return key.NewBinding(
		key.WithKeys("up", "ctrl+p"),
		key.WithHelp("↑/ctrl+p", "previous suggestion"),
	)
}

func (m *Model) SetSuggestions(suggestions []string) {
	m.suggestions = suggestions
}

func (m *Model) Filter(input string, excludeLabels []string) {
	currentLabel := extractCurrentLabel(input)

	excludeMap := make(map[string]bool)
	for _, label := range excludeLabels {
		excludeMap[strings.ToLower(strings.TrimSpace(label))] = true
	}

	m.filtered = filterAndRankSuggestions(
		m.suggestions,
		currentLabel,
		excludeMap,
		m.maxVisible,
	)

	m.selected = 0

	m.visible = len(m.filtered) > 0 && strings.TrimSpace(currentLabel) != ""
}

func (m *Model) Selected() string {
	if m.selected >= 0 && m.selected < len(m.filtered) {
		return m.filtered[m.selected]
	}
	return ""
}

func (m *Model) Next() {
	if len(m.filtered) > 0 {
		m.selected = (m.selected + 1) % len(m.filtered)
	}
}

func (m *Model) Prev() {
	if len(m.filtered) > 0 {
		m.selected--
		if m.selected < 0 {
			m.selected = len(m.filtered) - 1
		}
	}
}

func (m *Model) Show() {
	m.visible = true
}

func (m *Model) Hide() {
	m.visible = false
}

func (m *Model) IsVisible() bool {
	return m.visible
}

func (m *Model) SetPosition(x, y int) {
	m.posX = x
	m.posY = y
}

func (m *Model) SetWidth(width int) {
	m.width = width
}

func (m *Model) View() string {
	if !m.visible || len(m.filtered) == 0 {
		return ""
	}

	numVisible := min(len(m.filtered), m.maxVisible)

	var b strings.Builder

	popupStyle := m.ctx.Styles.Autocomplete.PopupStyle.Width(m.width)
	maxLabelWidth := m.width - popupStyle.GetHorizontalPadding()
	ellipsisWidth := lipgloss.Width(constants.Ellipsis)

	for i := 0; i < numVisible && i < len(m.filtered); i++ {
		label := m.filtered[i]
		if len(label) > maxLabelWidth {
			label = ansi.Truncate(label, maxLabelWidth-popupStyle.GetHorizontalPadding()-ellipsisWidth, constants.Ellipsis)
		}

		// Style based on selection
		if i == m.selected {
			// Selected row - use inverted colors
			b.WriteString(m.ctx.Styles.Autocomplete.SelectedStyle.Render(constants.SelectionIcon + label))
		} else {
			// Non-selected row
			b.WriteString(" " + label)
		}

		if i < numVisible-1 {
			b.WriteString("\n")
		}
	}

	return popupStyle.Render(b.String())
}

func extractCurrentLabel(input string) string {
	lastComma := strings.LastIndex(input, ",")
	if lastComma == -1 {
		return input
	}
	return strings.TrimSpace(input[lastComma+1:])
}

type suggestionMatch struct {
	label string
	score float64
}

func filterAndRankSuggestions(suggestions []string, currentLabel string, excludeMap map[string]bool, maxResults int) []string {
	currentLabelLower := strings.ToLower(currentLabel)

	var matches []suggestionMatch

	for _, suggestion := range suggestions {
		if excludeMap[strings.ToLower(suggestion)] {
			continue
		}

		suggestionLower := strings.ToLower(suggestion)

		if strings.HasPrefix(suggestionLower, currentLabelLower) {
			// Exact prefix match - highest score
			matches = append(matches, suggestionMatch{
				label: suggestion,
				score: 1.0,
			})
		} else if strings.Contains(suggestionLower, currentLabelLower) {
			// Contains match - lower score
			matches = append(matches, suggestionMatch{
				label: suggestion,
				score: 0.5,
			})
		} else if currentLabelLower != "" && fuzzyMatch(suggestionLower, currentLabelLower) {
			// Fuzzy match - lowest score
			matches = append(matches, suggestionMatch{
				label: suggestion,
				score: 0.3,
			})
		}
	}

	// Sort by score descending, then alphabetically
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		return matches[i].label < matches[j].label
	})

	// Return top results
	result := make([]string, 0, maxResults)
	for i := 0; i < len(matches) && i < maxResults; i++ {
		result = append(result, matches[i].label)
	}

	return result
}

func fuzzyMatch(text, pattern string) bool {
	patternLen := utf8.RuneCountInString(pattern)
	if patternLen == 0 {
		return true
	}

	pi := 0
	for _, r := range text {
		if pi < patternLen && r == rune(pattern[pi]) {
			pi++
			if pi == patternLen {
				break
			}
		}
	}

	return pi == patternLen
}
