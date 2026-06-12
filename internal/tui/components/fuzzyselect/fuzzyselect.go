// Package fuzzyselect houses the generic fuzzy-find select component
// It receives a completion Source that loads the suggestions asyncly
package fuzzyselect

import (
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
	"github.com/sahilm/fuzzy"
)

// Suggestion represents an autocomplete entry
// Value is the text inserted into the input; Detail is optional display-only context
type Suggestion struct {
	Value  string
	Detail string
}

// suggestionList wraps a slice of suggestions to implement fuzzy.Source
type suggestionList struct {
	items []Suggestion
}

func (s suggestionList) String(i int) string {
	return s.items[i].Value
}

type FetchState int

const (
	FetchStateIdle FetchState = iota
	FetchStateLoading
	FetchStateSuccess
	FetchStateError
)

func (s suggestionList) Len() int {
	return len(s.items)
}

var (
	NextKey = key.NewBinding(
		key.WithKeys("down", "ctrl+n"),
		key.WithHelp("↓/ctrl+n", "next"),
	)
	PrevKey = key.NewBinding(
		key.WithKeys("up", "ctrl+p"),
		key.WithHelp("↑/ctrl+p", "previous"),
	)
	SelectKey = key.NewBinding(
		key.WithKeys("ctrl+y"),
		key.WithHelp("ctrl+y", "select"),
	)
	RefreshSuggestionsKey = key.NewBinding(
		key.WithKeys("ctrl+f"),
		key.WithHelp("ctrl+f", "refresh"),
	)
	ToggleSuggestions = key.NewBinding(
		key.WithKeys("ctrl+h"),
		key.WithHelp("ctrl+h", "toggle"),
	)
)

type keyMap struct{}

func (km keyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		NextKey, PrevKey, SelectKey, RefreshSuggestionsKey, ToggleSuggestions,
	}
}

func (km keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{NextKey, PrevKey, SelectKey},
		{RefreshSuggestionsKey, ToggleSuggestions},
	}
}

// ClearFetchStatusMsg is sent to clear the fetch status after a delay
type ClearFetchStatusMsg struct{}

type Model struct {
	ctx        *context.ProgramContext
	styles     context.SelectStyles
	help       help.Model
	filtered   []Suggestion
	selected   int
	visible    bool
	maxVisible int
	width      int
	fetchState FetchState
	fetchError error
	spinner    spinner.Model
	// whether the user explicitly hid the suggestions; when true
	// Show() will not re-open the popup automatically until Unsuppress()
	hiddenByUser bool
	Source       Source
}

func NewModel(ctx *context.ProgramContext, src Source) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(ctx.Theme.SecondaryText)

	h := help.New()
	h.ShowAll = false
	h.Styles = ctx.Styles.Help.BubbleStyles
	return Model{
		ctx:        ctx,
		help:       h,
		styles:     ctx.Styles.Select,
		visible:    false,
		selected:   0,
		maxVisible: 4,
		width:      30,
		fetchState: FetchStateIdle,
		spinner:    sp,
		Source:     src,
	}
}

func (m *Model) Filter(input string, cmpCtx Context, excludeItems []string) {
	if m.Source == nil {
		return
	}

	excludeMap := make(map[string]bool)
	for _, item := range excludeItems {
		excludeMap[strings.ToLower(strings.TrimSpace(item))] = true
	}

	suggestions := m.Source.Suggestions(input, cmpCtx.Start)
	log.Debug("fuzzyselect.Filter suggestions", "ctx", cmpCtx, "len(suggestions)", len(suggestions))

	// Filter excluded items first
	filteredSuggestions := make(
		[]Suggestion,
		0,
	)

	for _, suggestion := range suggestions {
		if excluded, ok := excludeMap[strings.ToLower(strings.TrimSpace(suggestion.Value))]; !ok ||
			!excluded {
			filteredSuggestions = append(filteredSuggestions, suggestion)
		}
	}
	if cmpCtx.Content == "" || len(filteredSuggestions) == 0 {
		m.filtered = filteredSuggestions
		if len(m.filtered) > m.maxVisible {
			m.filtered = m.filtered[:m.maxVisible]
		}
		return
	}

	// Use fuzzy.FindFrom with suggestionList as Source
	list := suggestionList{items: filteredSuggestions}
	matches := fuzzy.FindFrom(cmpCtx.Content, list)
	log.Debug("fuzzyselect.Filter matches", "ctx", cmpCtx, "len(matches)", len(matches))

	// Collect matched items up to maxResults
	m.filtered = make([]Suggestion, 0, m.maxVisible)
	for _, match := range matches {
		if len(m.filtered) >= m.maxVisible {
			break
		}
		m.filtered = append(m.filtered, filteredSuggestions[match.Index])
	}
}

func (m *Model) Show() {
	m.selected = 0
	m.visible = true
	if !m.hiddenByUser {
		m.visible = true
	}
}

func (m *Model) Selected() string {
	if !m.IsVisible() {
		return ""
	}

	if m.selected >= 0 && m.selected < len(m.filtered) {
		return m.filtered[m.selected].Value
	}
	return ""
}

func (m *Model) Next() {
	if len(m.filtered) > 0 {
		m.selected = (m.selected + 1) % len(m.filtered)
	}
}

func (m *Model) Prev() {
	if len(m.filtered) == 0 {
		return
	}
	m.selected--
	if m.selected < 0 {
		m.selected = len(m.filtered) - 1
	}
}

func (m *Model) Hide() {
	m.visible = false
}

// Reset clears all autocomplete state including filtered suggestions, selection,
// and visibility flags. Use this when switching between different input modes
// (e.g., from labeling to commenting) to prevent stale suggestions from leaking.
func (m *Model) Reset() {
	m.filtered = nil
	m.selected = 0
	m.visible = false
	m.hiddenByUser = false
}

// Suppress hides the popup immediately and prevents it from being shown again
// automatically until `Unsuppress()` is called. The underlying filtered results
// are still updated while suppressed so navigation and selection keys will
// operate on up-to-date suggestions even though the popup is not visible.
func (m *Model) Suppress() {
	m.hiddenByUser = true
	m.visible = false
}

// Unsuppress clears the user hide flag and allows auto-showing again.
func (m *Model) Unsuppress() {
	m.hiddenByUser = false
}

func (m *Model) IsVisible() bool {
	return m.visible
}

// HasSuggestions returns true if there are filtered suggestions available.
func (m *Model) HasSuggestions() bool {
	return len(m.filtered) > 0
}

func (m *Model) Width() int {
	return m.width
}

func (m *Model) SetWidth(width int) {
	m.width = max(0, width)
	m.help.SetWidth(m.width)
}

func (m *Model) View() string {
	if !m.visible {
		return ""
	}

	numVisible := min(len(m.filtered), m.maxVisible)
	numRows := m.maxVisible
	if numRows <= 0 {
		numRows = numVisible
	}

	var b strings.Builder

	helpStyle := m.styles.HelpStyle.Width(m.width)
	selectedPrefix := m.styles.Pointer.Render(constants.SelectionIcon + " ")
	normalPrefix := "  "
	valueColStyle := lipgloss.NewStyle().Bold(true).Foreground(m.ctx.Theme.PrimaryText)
	detailColStyle := lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText)

	var rows []string
	selectedBgStyle := lipgloss.NewStyle().Background(m.ctx.Theme.SelectedBackground)
	maxRowWidth := m.width
	for i := 0; i < numRows; i++ {
		if i >= numVisible {
			continue
		}

		suggestion := m.filtered[i]
		value := suggestion.Value
		detail := suggestion.Detail

		bg := lipgloss.NewStyle()
		selected := i < numVisible && i == m.selected
		if selected {
			bg = selectedBgStyle
		}

		rowText := lipgloss.JoinHorizontal(
			lipgloss.Left,
			bg.Render(valueColStyle.Render(value)),
			bg.Render(" "),
			bg.Render(detailColStyle.Render(detail)),
		)

		row := ""
		if i < numVisible && i == m.selected {
			char := selectedPrefix
			row = utils.RemoveLastReset(
				char + m.styles.SelectedStyle.Render(rowText),
			)
		} else {
			row = m.styles.ItemStyle.Render(normalPrefix + rowText)
		}

		rows = append(rows, row)
		maxRowWidth = max(
			maxRowWidth,
			lipgloss.Width(row),
		)
	}

	if m.selected < len(rows) {
		rows[m.selected] = selectedBgStyle.Width(maxRowWidth).
			Render(rows[m.selected])
	}
	if len(rows) > 0 {
		b.WriteString(lipgloss.JoinVertical(lipgloss.Left, rows...))
	}

	var statusView string
	switch m.fetchState {
	case FetchStateLoading:
		statusView = m.spinner.View() + m.ctx.Styles.Common.FaintTextStyle.Render(
			"fetching suggestions"+constants.Ellipsis,
		)
	case FetchStateSuccess:
	case FetchStateError:
		errMsg := m.ctx.Styles.Common.FailureGlyph + m.ctx.Styles.Common.FaintTextStyle.Render(
			" failed to fetch suggestions",
		)
		if m.fetchError != nil {
			errMsg = m.fetchError.Error()
		}
		statusView = m.ctx.Styles.Common.FailureGlyph + " " + errMsg
	}

	parts := make([]string, 0)
	if filteredRows := b.String(); filteredRows != "" {
		parts = append(parts, filteredRows)
	} else if m.fetchState == FetchStateSuccess || m.fetchState == FetchStateIdle {
		parts = append(parts, m.styles.NoResults.Render("no results"+constants.Ellipsis))
	}
	if statusView != "" {
		parts = append(parts, m.styles.Status.Render(statusView))
	}
	parts = append(parts, helpStyle.Render(m.help.View(keyMap{})))

	return m.styles.PopupStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			parts...,
		),
	)
}

func (m *Model) Styles() context.SelectStyles {
	return m.styles
}

func (m *Model) SetStyles(styles context.SelectStyles) {
	m.styles = styles
}

func (m *Model) SetFetchLoading() tea.Cmd {
	m.fetchState = FetchStateLoading
	m.fetchError = nil

	m.selected = 0

	m.ShowIfHasContent()

	return m.spinner.Tick
}

func (m *Model) SetFetchSuccess() tea.Cmd {
	m.fetchState = FetchStateSuccess
	m.fetchError = nil
	return m.clearFetchStatus()
}

func (m *Model) SetFetchError(err error) tea.Cmd {
	m.fetchState = FetchStateError
	m.fetchError = err
	return m.clearFetchStatus()
}

// clearFetchStatus returns a command that will send a ClearFetchStatusMsg after 2 seconds
func (m *Model) clearFetchStatus() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return ClearFetchStatusMsg{}
	})
}

func (m *Model) ShowIfHasContent() {
	if m.hiddenByUser {
		m.visible = false
	} else if len(m.filtered) > 0 {
		m.visible = true
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.fetchState == FetchStateLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case ClearFetchStatusMsg:
		// Only clear if we're in a success or error state (not loading or already idle)
		if m.fetchState == FetchStateSuccess || m.fetchState == FetchStateError {
			m.fetchState = FetchStateIdle
			m.fetchError = nil
		}
	}
	return m, nil
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
	m.help.Styles = ctx.Styles.Help.BubbleStyles
}
