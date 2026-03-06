package issueview

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/common"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/autocomplete"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/inputbox"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/issuerow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/issuessection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tasks"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/keys"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/markdown"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

var (
	htmlCommentRegex = regexp.MustCompile("(?U)<!--(.|[[:space:]])*-->")
	lineCleanupRegex = regexp.MustCompile(`((\n)+|^)([^\r\n]*\|[^\r\n]*(\n)?)+`)
)

type RepoLabelsFetchedMsg struct {
	Labels []data.Label
}

type RepoLabelsFetchFailedMsg struct {
	Err error
}

type Model struct {
	ctx       *context.ProgramContext
	issue     *issuerow.Issue
	sectionId int
	width     int

	ShowConfirmCancel bool
	isCommenting      bool
	isLabeling        bool
	isAssigning       bool
	isUnassigning     bool

	inputBox   inputbox.Model
	ac         *autocomplete.Model
	repoLabels []data.Label
}

func NewModel(ctx *context.ProgramContext) Model {
	inputBox := inputbox.NewModel(ctx)
	linesToAdjust := 5
	inputBox.SetHeight(common.InputBoxHeight - linesToAdjust)

	inputBox.OnSuggestionSelected = handleLabelSelection
	inputBox.CurrentContext = labelAtCursor
	inputBox.SuggestionsToExclude = allLabels

	ac := autocomplete.NewModel(ctx)
	inputBox.SetAutocomplete(&ac)

	return Model{
		issue: nil,

		isCommenting:  false,
		isLabeling:    false,
		isAssigning:   false,
		isUnassigning: false,

		inputBox:   inputBox,
		ac:         &ac,
		repoLabels: nil,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd, *IssueAction) {
	var (
		cmds  []tea.Cmd
		cmd   tea.Cmd
		taCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case RepoLabelsFetchedMsg:
		clearCmd := m.ac.SetFetchSuccess()
		m.repoLabels = msg.Labels
		labelNames := data.GetLabelNames(msg.Labels)
		m.ac.SetSuggestions(labelNames)
		if m.isLabeling {
			cursorPos := m.inputBox.GetCursorPosition()
			currentLabel := labelAtCursor(cursorPos, m.inputBox.Value())
			existingLabels := allLabels(m.inputBox.Value())
			m.ac.Show(currentLabel, existingLabels)
		}
		return m, clearCmd, nil

	case RepoLabelsFetchFailedMsg:
		clearCmd := m.ac.SetFetchError(msg.Err)
		return m, clearCmd, nil

	case autocomplete.FetchSuggestionsRequestedMsg:
		// Only fetch when we're in labeling mode (where labels are relevant)
		if m.isLabeling {
			// If this is a forced refresh (e.g., via Ctrl+F), clear the cached labels
			// for this repo so FetchRepoLabels will actually call the gh CLI.
			if msg.Force {
				if m.issue != nil {
					repoName := m.issue.Data.GetRepoNameWithOwner()
					data.ClearRepoLabelCache(repoName)
				}
			}
			cmd := m.fetchLabels()
			return m, cmd, nil
		}
		return m, nil, nil

	case tea.KeyMsg:
		if m.isCommenting {
			switch msg.Type {
			case tea.KeyCtrlD:
				if len(strings.Trim(m.inputBox.Value(), " ")) != 0 {
					sid := tasks.SectionIdentifier{Id: m.sectionId, Type: issuessection.SectionType}
					cmd = tasks.CommentOnIssue(m.ctx, sid, m.issue.Data, m.inputBox.Value())
				}
				m.inputBox.Blur()
				m.isCommenting = false
				return m, cmd, nil

			case tea.KeyEsc, tea.KeyCtrlC:
				if !m.ShowConfirmCancel {
					m.shouldCancelComment()
				}
			default:
				if msg.String() == "Y" || msg.String() == "y" {
					if m.shouldCancelComment() {
						return m, nil, nil
					}
				}
				if m.ShowConfirmCancel && (msg.String() == "N" || msg.String() == "n") {
					m.inputBox.SetPrompt(constants.CommentPrompt)
					m.ShowConfirmCancel = false
					return m, nil, nil
				}
				m.inputBox.SetPrompt(constants.CommentPrompt)
				m.ShowConfirmCancel = false
			}

			m.inputBox, taCmd = m.inputBox.Update(msg)
			cmds = append(cmds, cmd, taCmd)
		} else if m.isLabeling {
			switch msg.Type {
			case tea.KeyCtrlD:
				labels := allLabels(m.inputBox.Value())
				if len(labels) > 0 {
					sid := tasks.SectionIdentifier{Id: m.sectionId, Type: issuessection.SectionType}
					cmd = tasks.LabelIssue(m.ctx, sid, m.issue.Data, labels, m.issue.Data.Labels.Nodes)
				}
				m.inputBox.Blur()
				m.isLabeling = false
				m.ac.Hide()
				return m, cmd, nil

			case tea.KeyEsc, tea.KeyCtrlC:
				m.inputBox.Blur()
				m.isLabeling = false
				m.ac.Hide()
				return m, nil, nil
			}

			if key.Matches(msg, autocomplete.RefreshSuggestionsKey) {
				if m.issue != nil {
					repoName := m.issue.Data.GetRepoNameWithOwner()
					data.ClearRepoLabelCache(repoName)
				}
				cmds = append(cmds, m.fetchLabels())
			}

			previousCursorPos := m.inputBox.GetCursorPosition()
			previousValue := m.inputBox.Value()
			previousLabel := labelAtCursor(previousCursorPos, previousValue)

			m.inputBox, taCmd = m.inputBox.Update(msg)
			cmds = append(cmds, cmd, taCmd)

			currentCursorPos := m.inputBox.GetCursorPosition()
			currentValue := m.inputBox.Value()
			currentLabel := labelAtCursor(currentCursorPos, currentValue)

			if currentLabel != previousLabel {
				existingLabels := allLabels(currentValue)
				m.ac.Show(currentLabel, existingLabels)
			}
		} else if m.isAssigning {
			switch msg.Type {
			case tea.KeyCtrlD:
				usernames := strings.Fields(m.inputBox.Value())
				if len(usernames) > 0 {
					sid := tasks.SectionIdentifier{Id: m.sectionId, Type: issuessection.SectionType}
					cmd = tasks.AssignIssue(m.ctx, sid, m.issue.Data, usernames)
				}
				m.inputBox.Blur()
				m.isAssigning = false
				return m, cmd, nil

			case tea.KeyEsc, tea.KeyCtrlC:
				m.inputBox.Blur()
				m.isAssigning = false
				return m, nil, nil
			}

			m.inputBox, taCmd = m.inputBox.Update(msg)
			cmds = append(cmds, cmd, taCmd)
		} else if m.isUnassigning {
			switch msg.Type {
			case tea.KeyCtrlD:
				usernames := strings.Fields(m.inputBox.Value())
				if len(usernames) > 0 {
					sid := tasks.SectionIdentifier{Id: m.sectionId, Type: issuessection.SectionType}
					cmd = tasks.UnassignIssue(m.ctx, sid, m.issue.Data, usernames)
				}
				m.inputBox.Blur()
				m.isUnassigning = false
				return m, cmd, nil

			case tea.KeyEsc, tea.KeyCtrlC:
				m.inputBox.Blur()
				m.isUnassigning = false
				return m, nil, nil
			}

			m.inputBox, taCmd = m.inputBox.Update(msg)
			cmds = append(cmds, cmd, taCmd)
		} else {
			switch {
			case key.Matches(msg, keys.IssueKeys.Label):
				return m, nil, &IssueAction{Type: IssueActionLabel}
			case key.Matches(msg, keys.IssueKeys.Assign):
				return m, nil, &IssueAction{Type: IssueActionAssign}
			case key.Matches(msg, keys.IssueKeys.Unassign):
				return m, nil, &IssueAction{Type: IssueActionUnassign}
			case key.Matches(msg, keys.IssueKeys.Comment):
				return m, nil, &IssueAction{Type: IssueActionComment}
			case key.Matches(msg, keys.IssueKeys.Close):
				return m, nil, &IssueAction{Type: IssueActionClose}
			case key.Matches(msg, keys.IssueKeys.Reopen):
				return m, nil, &IssueAction{Type: IssueActionReopen}
			}
			return m, nil, nil
		}
	}

	switch msg.(type) {
	case spinner.TickMsg, autocomplete.ClearFetchStatusMsg:
		var acCmd tea.Cmd
		*m.ac, acCmd = m.ac.Update(msg)
		cmds = append(cmds, acCmd)
	}

	return m, tea.Batch(cmds...), nil
}

func (m Model) View() string {
	s := strings.Builder{}

	s.WriteString(m.renderFullNameAndNumber())
	s.WriteString("\n")

	s.WriteString(m.renderTitle())
	s.WriteString("\n\n")
	s.WriteString(m.renderStatusPill())
	s.WriteString("\n\n")
	s.WriteString(m.renderAuthor())
	s.WriteString("\n\n")

	labels := m.renderLabels()
	if labels != "" {
		s.WriteString(labels)
		s.WriteString("\n\n")
	}

	s.WriteString(m.renderBody())
	s.WriteString("\n\n")
	s.WriteString(m.renderActivity())

	if m.isCommenting || m.isAssigning || m.isUnassigning {
		s.WriteString(m.inputBox.View())
	} else if m.isLabeling {
		s.WriteString(m.inputBox.ViewWithAutocomplete())
	}

	return lipgloss.NewStyle().Padding(0, m.ctx.Styles.Sidebar.ContentPadding).Render(s.String())
}

func (m *Model) renderFullNameAndNumber() string {
	return common.RenderPreviewHeader(m.ctx.Theme, m.width,
		fmt.Sprintf("#%d · %s", m.issue.Data.GetNumber(), m.issue.Data.GetRepoNameWithOwner()))
}

func (m *Model) renderTitle() string {
	return common.RenderPreviewTitle(m.ctx.Theme, m.ctx.Styles.Common, m.width, m.issue.Data.Title)
}

func (m *Model) renderStatusPill() string {
	bgColor := ""
	content := ""
	switch m.issue.Data.State {
	case "OPEN":
		bgColor = m.ctx.Styles.Colors.OpenIssue.Dark
		content = " Open"
	case "CLOSED":
		bgColor = m.ctx.Styles.Colors.ClosedIssue.Dark
		content = " Closed"
	}

	return m.ctx.Styles.PrView.PillStyle.
		BorderForeground(lipgloss.Color(bgColor)).
		Background(lipgloss.Color(bgColor)).
		Render(content)
}

func (m *Model) renderAuthor() string {
	authorAssociation := m.issue.Data.AuthorAssociation
	if authorAssociation == "" {
		authorAssociation = "unknown role"
	}
	time := lipgloss.NewStyle().Render(utils.TimeElapsed(m.issue.Data.CreatedAt))
	return lipgloss.JoinHorizontal(lipgloss.Top,
		" by ",
		lipgloss.NewStyle().Foreground(m.ctx.Theme.PrimaryText).Render(
			lipgloss.NewStyle().Bold(true).Render("@"+m.issue.Data.Author.Login)),
		lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render(
			lipgloss.JoinHorizontal(lipgloss.Top, " ⋅ ", time, " ago", " ⋅ ")),
		lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render(
			lipgloss.JoinHorizontal(lipgloss.Top, data.GetAuthorRoleIcon(m.issue.Data.AuthorAssociation,
				m.ctx.Theme), " ", lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render(strings.ToLower(authorAssociation))),
		),
	)
}

func (m *Model) renderBody() string {
	width := m.getIndentedContentWidth()
	// Strip HTML comments from body and cleanup body.
	body := htmlCommentRegex.ReplaceAllString(m.issue.Data.Body, "")
	body = lineCleanupRegex.ReplaceAllString(body, "")

	body = strings.TrimSpace(body)
	if body == "" {
		return lipgloss.NewStyle().Italic(true).Foreground(m.ctx.Theme.FaintText).Render("No description provided.")
	}

	markdownRenderer := markdown.GetMarkdownRenderer(width)
	rendered, err := markdownRenderer.Render(body)
	if err != nil {
		return ""
	}

	return lipgloss.NewStyle().
		Width(width).
		MaxWidth(width).
		Align(lipgloss.Left).
		Render(rendered)
}

func (m *Model) renderLabels() string {
	width := m.getIndentedContentWidth()
	labels := m.issue.Data.Labels.Nodes
	style := m.ctx.Styles.PrView.PillStyle

	return common.RenderLabels(width, labels, style)
}

func (m *Model) getIndentedContentWidth() int {
	return m.width - 6
}

func (m *Model) SetWidth(width int) {
	m.width = width
	m.inputBox.SetWidth(width)
	m.ac.SetWidth(width - 4)
}

func (m *Model) SetSectionId(id int) {
	m.sectionId = id
}

func (m *Model) SetRow(data *data.IssueData) {
	if data == nil {
		m.issue = nil
	} else {
		m.issue = &issuerow.Issue{Ctx: m.ctx, Data: *data}
	}
}

func (m *Model) IsTextInputBoxFocused() bool {
	return m.isCommenting || m.isAssigning || m.isUnassigning || m.isLabeling
}

func (m *Model) GetIsCommenting() bool {
	return m.isCommenting
}

func (m *Model) shouldCancelComment() bool {
	if !m.ShowConfirmCancel {
		m.inputBox.SetPrompt(lipgloss.NewStyle().Foreground(m.ctx.Theme.ErrorText).Render("Discard comment? (y/N)"))
		m.ShowConfirmCancel = true
		return false
	}
	m.inputBox.Blur()
	m.isCommenting = false
	m.ShowConfirmCancel = false
	return true
}

func (m *Model) SetIsCommenting(isCommenting bool) tea.Cmd {
	if m.issue == nil {
		return nil
	}

	if !m.isCommenting && isCommenting {
		m.inputBox.Reset()
		m.ac.Reset() // Clear any stale autocomplete state (e.g., from labeling)
	}
	m.isCommenting = isCommenting
	m.inputBox.SetPrompt(constants.CommentPrompt)

	if isCommenting {
		return tea.Sequence(textarea.Blink, m.inputBox.Focus())
	}
	return nil
}

func (m *Model) GetIsAssigning() bool {
	return m.isAssigning
}

func (m *Model) SetIsAssigning(isAssigning bool) tea.Cmd {
	if m.issue == nil {
		return nil
	}

	if !m.isAssigning && isAssigning {
		m.inputBox.Reset()
		m.ac.Reset() // Clear any stale autocomplete state (e.g., from labeling)
	}
	m.isAssigning = isAssigning
	m.inputBox.SetPrompt(constants.AssignPrompt)
	if !m.userAssignedToIssue(m.ctx.User) {
		m.inputBox.SetValue(m.ctx.User)
	}

	if isAssigning {
		return tea.Sequence(textarea.Blink, m.inputBox.Focus())
	}
	return nil
}

func (m *Model) SetIsLabeling(isLabeling bool) tea.Cmd {
	if m.issue == nil {
		return nil
	}

	if !m.isLabeling && isLabeling {
		m.inputBox.Reset()
	}
	m.isLabeling = isLabeling
	m.inputBox.SetPrompt(constants.LabelPrompt)

	labels := make([]string, 0)
	for _, label := range m.issue.Data.Labels.Nodes {
		labels = append(labels, label.Name)
	}
	labels = append(labels, "")
	m.inputBox.SetValue(strings.Join(labels, ", "))

	// Reset autocomplete
	m.ac.Hide()
	m.ac.SetSuggestions(nil)

	// Trigger label fetching for autocomplete
	if isLabeling {
		repoName := m.issue.Data.GetRepoNameWithOwner()
		if labels, ok := data.GetCachedRepoLabels(repoName); ok {
			// Use cached labels
			m.repoLabels = labels
			m.ac.SetSuggestions(data.GetLabelNames(labels))
			cursorPos := m.inputBox.GetCursorPosition()
			currentLabel := labelAtCursor(cursorPos, m.inputBox.Value())
			existingLabels := allLabels(m.inputBox.Value())
			m.ac.Show(currentLabel, existingLabels)
			return tea.Sequence(textarea.Blink, m.inputBox.Focus())
		} else {
			// Fetch labels asynchronously
			return tea.Sequence(m.fetchLabels(), textarea.Blink, m.inputBox.Focus())
		}
	}
	return nil
}

// fetchLabels returns a command to fetch repository labels
func (m *Model) fetchLabels() tea.Cmd {
	spinnerTickCmd := m.ac.SetFetchLoading()

	fetchCmd := func() tea.Msg {
		repoName := m.issue.Data.GetRepoNameWithOwner()
		labels, err := data.FetchRepoLabels(repoName)
		if err != nil {
			return RepoLabelsFetchFailedMsg{Err: err}
		}
		return RepoLabelsFetchedMsg{Labels: labels}
	}

	return tea.Batch(spinnerTickCmd, fetchCmd)
}

func (m *Model) userAssignedToIssue(login string) bool {
	for _, a := range m.issue.Data.Assignees.Nodes {
		if login == a.Login {
			return true
		}
	}
	return false
}

func (m *Model) GetIsUnassigning() bool {
	return m.isUnassigning
}

func (m *Model) SetIsUnassigning(isUnassigning bool) tea.Cmd {
	if m.issue == nil {
		return nil
	}

	if !m.isUnassigning && isUnassigning {
		m.inputBox.Reset()
		m.ac.Reset() // Clear any stale autocomplete state (e.g., from labeling)
	}
	m.isUnassigning = isUnassigning
	m.inputBox.SetPrompt(constants.UnassignPrompt)
	m.inputBox.SetValue(strings.Join(m.issueAssignees(), "\n"))

	if isUnassigning {
		return tea.Sequence(textarea.Blink, m.inputBox.Focus())
	}
	return nil
}

func (m *Model) issueAssignees() []string {
	var assignees []string
	for _, n := range m.issue.Data.Assignees.Nodes {
		assignees = append(assignees, n.Login)
	}
	return assignees
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
	m.inputBox.UpdateProgramContext(ctx)
	if m.ac != nil {
		m.ac.UpdateProgramContext(ctx)
	}
}
