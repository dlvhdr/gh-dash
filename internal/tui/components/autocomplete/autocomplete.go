package autocomplete

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/sahilm/fuzzy"
)

// suggestionList wraps a slice of strings to implement fuzzy.Source
type suggestionList struct {
	items []string
}

func (s suggestionList) String(i int) string {
	return s.items[i]
}

type FetchState int

func (s suggestionList) Len() int {
	return len(s.items)
}

var (
	NextKey               = key.NewBinding(key.WithKeys(tea.KeyDown.String(), tea.KeyCtrlN.String()), key.WithHelp("↓/Ctrl+n", "next"))
	PrevKey               = key.NewBinding(key.WithKeys(tea.KeyUp.String(), tea.KeyCtrlP.String()), key.WithHelp("↑/Ctrl+p", "previous"))
	SelectKey             = key.NewBinding(key.WithKeys(tea.KeyTab.String(), tea.KeyEnter.String(), tea.KeyCtrlY.String()), key.WithHelp("tab/enter/Ctrl+y", "select"))
	RefreshSuggestionsKey = key.NewBinding(key.WithKeys(tea.KeyCtrlF.String()), key.WithHelp("Ctrl+f", "refresh suggestions"))
	ToggleSuggestions     = key.NewBinding(key.WithKeys(tea.KeyCtrlH.String()), key.WithHelp("Ctrl+h", "toggle suggestions"))
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
	suggestions    []string
	filtered       []string
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

func (m *Model) SetSuggestions(suggestions []string) {
	m.suggestions = suggestions
}

func (m *Model) Show(currentItem string, excludeItems []string) {
	excludeMap := make(map[string]bool)
	for _, item := range excludeItems {
		excludeMap[strings.ToLower(strings.TrimSpace(item))] = true
	}

	// Filter excluded labels first
	var filteredSuggestions []string
	for _, suggestion := range m.suggestions {
		if !excludeMap[strings.ToLower(strings.TrimSpace(suggestion))] {
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
	m.filtered = make([]string, 0, m.maxVisible)
	for _, match := range matches {
		if len(m.filtered) >= m.maxVisible {
			break
		}
		m.filtered = append(m.filtered, match.Str)
	}

	m.selected = 0
	// respect suppression: don't auto-show if suppressed
	m.UpdateVisible()
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
			label = ansi.Truncate(label, maxLabelWidth-ellipsisWidth, constants.Ellipsis)
		}

		// Style based on selection
		if i == m.selected {
			// Selected row - use inverted colors
			b.WriteString(m.ctx.Styles.Autocomplete.SelectedStyle.Render(constants.SelectionIcon + " " + label))
		} else {
			// Non-selected row
			b.WriteString("  " + label)
		}

		if i < numVisible-1 {
			b.WriteString("\n")
		}
	}

	var statusView string
	switch m.fetchState {
	case FetchStateLoading:
		statusView = m.spinner.View() + m.ctx.Styles.Common.FaintTextStyle.Render("Fetching suggestions"+constants.Ellipsis)
	case FetchStateSuccess:
		statusView = m.ctx.Styles.Common.SuccessGlyph + m.ctx.Styles.Common.FaintTextStyle.Render(" Suggestions loaded")
	case FetchStateError:
		errMsg := m.ctx.Styles.Common.FailureGlyph + m.ctx.Styles.Common.FaintTextStyle.Render(" Failed to fetch suggestions")
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

	placeholders := make([]string, 0, m.maxVisible)
	for i := 0; i < m.maxVisible; i++ {
		placeholders = append(placeholders, "")
	}
	m.filtered = placeholders
	m.selected = 0

	m.UpdateVisible()

	return m.spinner.Tick
}

func (m *Model) SetFetchSuccess() tea.Cmd {
	m.fetchState = FetchStateSuccess
	m.fetchError = nil
	return m.clearFetchStatusAfterDelay()
}

func (m *Model) SetFetchError(err error) tea.Cmd {
	m.fetchState = FetchStateError
	m.fetchError = err
	return m.clearFetchStatusAfterDelay()
}

// clearFetchStatusAfterDelay returns a command that will send a ClearFetchStatusMsg after 2 seconds
func (m *Model) clearFetchStatusAfterDelay() tea.Cmd {
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
