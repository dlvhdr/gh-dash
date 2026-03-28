package autocomplete

import (
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
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

func (s suggestionList) Len() int {
	return len(s.items)
}

var (
	NextKey = key.NewBinding(
		key.WithKeys("down", "ctrl+n"),
		key.WithHelp("↓/Ctrl+n", "next"),
	)
	PrevKey = key.NewBinding(
		key.WithKeys("up", "ctrl+p"),
		key.WithHelp("↑/Ctrl+p", "previous"),
	)
	SelectKey = key.NewBinding(
		key.WithKeys("tab", "enter", "ctrl+y"),
		key.WithHelp("tab/enter/Ctrl+y", "select"),
	)
	RefreshSuggestionsKey = key.NewBinding(
		key.WithKeys("ctrl+f"),
		key.WithHelp("Ctrl+f", "refresh suggestions"),
	)
	ToggleSuggestions = key.NewBinding(
		key.WithKeys("ctrl+h"),
		key.WithHelp("Ctrl+h", "toggle suggestions"),
	)
)

var suggestionKeys = []key.Binding{
	NextKey,
	PrevKey,
	SelectKey,
	RefreshSuggestionsKey,
}

const (
	FetchStateIdle FetchState = iota
	FetchStateLoading
	FetchStateSuccess
	FetchStateError
)

// ClearFetchStatusMsg is sent to clear the fetch status after a delay
type ClearFetchStatusMsg struct{}

// FetchSuggestionsRequestedMsg requests that the current view fetch suggestions from upstream.
//
// When Force is true the fetch should bypass any local cache and request fresh
// data from the gh CLI.
type FetchSuggestionsRequestedMsg struct {
	Force bool
}

// NewFetchSuggestionsRequestedCmd returns a tea.Cmd that emits a
// FetchSuggestionsRequestedMsg with the given force flag.
func NewFetchSuggestionsRequestedCmd(force bool) tea.Cmd {
	return func() tea.Msg { return FetchSuggestionsRequestedMsg{Force: force} }
}

type Model struct {
	ctx            *context.ProgramContext
	suggestionHelp help.Model
	suggestions    []Suggestion
	filtered       []Suggestion
	selected       int
	visible        bool
	maxVisible     int
	width          int
	fetchState     FetchState
	fetchError     error
	spinner        spinner.Model
	// whether the user explicitly hid the suggestions; when true
	// Show() will not re-open the popup automatically until Unsuppress()
	hiddenByUser bool
}

func NewModel(ctx *context.ProgramContext) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(ctx.Theme.SecondaryText)

	h := help.New()
	h.Styles = ctx.Styles.Help.BubbleStyles
	return Model{
		ctx:            ctx,
		suggestionHelp: h,
		visible:        false,
		selected:       0,
		maxVisible:     4,
		width:          30,
		fetchState:     FetchStateIdle,
		spinner:        sp,
	}
}

func (m *Model) SetSuggestions(suggestions []Suggestion) {
	m.suggestions = suggestions
}

func (m *Model) Show(currentItem string, excludeItems []string) {
	excludeMap := make(map[string]bool)
	for _, item := range excludeItems {
		excludeMap[strings.ToLower(strings.TrimSpace(item))] = true
	}

	// Filter excluded items first
	filteredSuggestions := make([]Suggestion, 0, len(m.suggestions))
	for _, suggestion := range m.suggestions {
		if !excludeMap[strings.ToLower(strings.TrimSpace(suggestion.Value))] {
			filteredSuggestions = append(filteredSuggestions, suggestion)
		}
	}

	if currentItem == "" || len(filteredSuggestions) == 0 {
		m.filtered = filteredSuggestions
		if len(m.filtered) > m.maxVisible {
			m.filtered = m.filtered[:m.maxVisible]
		}
		m.selected = 0
		// respect suppression: don't auto-show if suppressed
		m.UpdateVisible()
		return
	}

	// Use fuzzy.FindFrom with suggestionList as Source
	list := suggestionList{items: filteredSuggestions}
	matches := fuzzy.FindFrom(currentItem, list)

	// Collect matched items up to maxResults
	m.filtered = make([]Suggestion, 0, m.maxVisible)
	for _, match := range matches {
		if len(m.filtered) >= m.maxVisible {
			break
		}
		m.filtered = append(m.filtered, filteredSuggestions[match.Index])
	}

	m.selected = 0
	// respect suppression: don't auto-show if suppressed
	m.UpdateVisible()
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

func (m *Model) SetWidth(width int) {
	m.width = max(0, width)
}

type columnLayout struct {
	valueWidth  int
	detailWidth int
	gapWidth    int
}

func (m *Model) computeColumnLayout(numVisible, totalContentWidth int) columnLayout {
	layout := columnLayout{
		valueWidth: totalContentWidth,
	}

	maxValueWidth := 0
	hasAnyDetail := false
	for i := range numVisible {
		suggestion := m.filtered[i]
		maxValueWidth = max(maxValueWidth, lipgloss.Width(suggestion.Value))
		hasAnyDetail = hasAnyDetail || strings.TrimSpace(suggestion.Detail) != ""
	}
	if !hasAnyDetail || totalContentWidth <= 0 {
		return layout
	}

	layout.gapWidth = min(
		constants.AutocompleteColumnGap,
		max(0, totalContentWidth-constants.AutocompleteMinDetailWidth),
	)
	maxValueForDetails := max(
		0,
		totalContentWidth-layout.gapWidth-constants.AutocompleteMinDetailWidth,
	)
	if maxValueForDetails <= 0 {
		return layout
	}

	preferredValueWidth := min(
		maxValueWidth,
		max(
			constants.AutocompleteMinValueWidth,
			(totalContentWidth*constants.AutocompletePreferredValueRatioNum)/constants.AutocompletePreferredValueRatioDen,
		),
	)
	layout.valueWidth = max(1, min(preferredValueWidth, maxValueForDetails))
	layout.detailWidth = max(0, totalContentWidth-layout.valueWidth-layout.gapWidth)

	return layout
}

func (m *Model) View() string {
	if !m.visible || len(m.filtered) == 0 {
		return ""
	}

	numVisible := min(len(m.filtered), m.maxVisible)
	numRows := m.maxVisible
	if numRows <= 0 {
		numRows = numVisible
	}

	var b strings.Builder

	popupStyle := m.ctx.Styles.Autocomplete.PopupStyle.Width(m.width)
	maxLabelWidth := m.width - popupStyle.GetHorizontalPadding()

	selectedPrefix := constants.SelectionIcon + " "
	normalPrefix := "  "
	selectedPrefixWidth := lipgloss.Width(selectedPrefix)
	normalPrefixWidth := lipgloss.Width(normalPrefix)
	maxPrefixWidth := max(selectedPrefixWidth, normalPrefixWidth)
	totalContentWidth := max(0, maxLabelWidth-maxPrefixWidth)
	layout := m.computeColumnLayout(numVisible, totalContentWidth)
	valueColumnStyle := lipgloss.NewStyle().Width(layout.valueWidth)
	detailColumnStyle := lipgloss.NewStyle().
		Width(layout.detailWidth).
		Foreground(m.ctx.Theme.FaintText)
	ellipsisWidth := lipgloss.Width(constants.Ellipsis)

	for i := 0; i < numRows; i++ {
		suggestion := Suggestion{}
		if i < numVisible {
			suggestion = m.filtered[i]
		}
		value := suggestion.Value
		detail := suggestion.Detail

		if layout.valueWidth > 0 && lipgloss.Width(value) > layout.valueWidth {
			value = ansi.Truncate(
				value,
				max(0, layout.valueWidth-ellipsisWidth),
				constants.Ellipsis,
			)
		}
		if layout.detailWidth > 0 && lipgloss.Width(detail) > layout.detailWidth {
			detail = ansi.Truncate(
				detail,
				max(0, layout.detailWidth-ellipsisWidth),
				constants.Ellipsis,
			)
		}

		rowText := value
		if layout.detailWidth > 0 {
			rowText = lipgloss.JoinHorizontal(
				lipgloss.Left,
				valueColumnStyle.Render(value),
				strings.Repeat(" ", layout.gapWidth),
				detailColumnStyle.Render(detail),
			)
		}

		if i < numVisible && i == m.selected {
			b.WriteString(m.ctx.Styles.Autocomplete.SelectedStyle.Render(selectedPrefix + rowText))
		} else {
			b.WriteString(normalPrefix + rowText)
		}

		if i < numRows-1 {
			b.WriteString("\n")
		}
	}

	var statusView string
	switch m.fetchState {
	case FetchStateLoading:
		statusView = m.spinner.View() + m.ctx.Styles.Common.FaintTextStyle.Render(
			"Fetching suggestions"+constants.Ellipsis,
		)
	case FetchStateSuccess:
		statusView = m.ctx.Styles.Common.SuccessGlyph + m.ctx.Styles.Common.FaintTextStyle.Render(
			" Suggestions loaded",
		)
	case FetchStateError:
		errMsg := m.ctx.Styles.Common.FailureGlyph + m.ctx.Styles.Common.FaintTextStyle.Render(
			" Failed to fetch suggestions",
		)
		if m.fetchError != nil {
			errMsg = m.fetchError.Error()
		}
		statusView = m.ctx.Styles.Common.FailureGlyph + " " + errMsg
	}

	return popupStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			b.String(),
			statusView,
			lipgloss.NewStyle().
				Render(m.suggestionHelp.ShortHelpView(suggestionKeys)),
		),
	)
}

func (m *Model) SetFetchLoading() tea.Cmd {
	m.fetchState = FetchStateLoading
	m.fetchError = nil

	placeholders := make([]Suggestion, 0, m.maxVisible)
	for i := 0; i < m.maxVisible; i++ {
		placeholders = append(placeholders, Suggestion{})
	}
	m.filtered = placeholders
	m.selected = 0

	m.UpdateVisible()

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

func (m *Model) UpdateVisible() {
	if m.hiddenByUser {
		m.visible = false
	} else {
		m.visible = len(m.filtered) > 0
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
	m.suggestionHelp.Styles = ctx.Styles.Help.BubbleStyles
}
