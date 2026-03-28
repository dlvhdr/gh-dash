package issueview

import (
	"fmt"
	"image/color"
	"regexp"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	dataautocomplete "github.com/dlvhdr/gh-dash/v4/internal/data/autocomplete"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/common"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/detailedit"
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

type Model struct {
	ctx       *context.ProgramContext
	issue     *issuerow.Issue
	sectionId int
	width     int
	editor    detailedit.Controller
}

func NewModel(ctx *context.ProgramContext) Model {
	return Model{
		issue:  nil,
		editor: detailedit.New(ctx),
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd, *IssueAction) {
	editor, cmd, submit, handled := m.editor.Update(msg)
	m.editor = editor

	if submit != nil {
		if m.issue == nil {
			return m, nil, nil
		}

		sid := tasks.SectionIdentifier{Id: m.sectionId, Type: issuessection.SectionType}

		switch submit.Mode {
		case detailedit.ModeComment:
			if len(strings.TrimSpace(submit.Value)) != 0 {
				return m, tasks.CommentOnIssue(m.ctx, sid, m.issue.Data, submit.Value), nil
			}
			return m, nil, nil

		case detailedit.ModeAssign:
			usernames := dataautocomplete.AllWords(submit.Value)
			if len(usernames) > 0 {
				return m, tasks.AssignIssue(m.ctx, sid, m.issue.Data, usernames), nil
			}
			return m, nil, nil

		case detailedit.ModeUnassign:
			usernames := dataautocomplete.AllWords(submit.Value)
			if len(usernames) > 0 {
				return m, tasks.UnassignIssue(m.ctx, sid, m.issue.Data, usernames), nil
			}
			return m, nil, nil

		case detailedit.ModeLabel:
			labels := dataautocomplete.CurrentLabels(submit.Value)
			if len(labels) > 0 || len(m.issue.Data.Labels.Nodes) > 0 {
				return m, tasks.LabelIssue(
					m.ctx,
					sid,
					m.issue.Data,
					labels,
					m.issue.Data.Labels.Nodes,
				), nil
			}
			return m, nil, nil
		}
	}
	if handled {
		return m, cmd, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, keys.IssueKeys.Label):
			return m, nil, &IssueAction{Type: IssueActionLabel}
		case key.Matches(keyMsg, keys.IssueKeys.Assign):
			return m, nil, &IssueAction{Type: IssueActionAssign}
		case key.Matches(keyMsg, keys.IssueKeys.Unassign):
			return m, nil, &IssueAction{Type: IssueActionUnassign}
		case key.Matches(keyMsg, keys.IssueKeys.Comment):
			return m, nil, &IssueAction{Type: IssueActionComment}
		case key.Matches(keyMsg, keys.IssueKeys.Checkout):
			return m, nil, &IssueAction{Type: IssueActionCheckout}
		case key.Matches(keyMsg, keys.IssueKeys.Close):
			return m, nil, &IssueAction{Type: IssueActionClose}
		case key.Matches(keyMsg, keys.IssueKeys.Reopen):
			return m, nil, &IssueAction{Type: IssueActionReopen}
		}
	}

	return m, cmd, nil
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

	if editorView := m.editor.View(); editorView != "" {
		s.WriteString(editorView)
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
	var bgColor color.Color
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
		BorderForeground(bgColor).
		Background(bgColor).
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
				m.ctx.Theme),
				" ", lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render(strings.ToLower(authorAssociation))),
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
		return lipgloss.NewStyle().
			Italic(true).
			Foreground(m.ctx.Theme.FaintText).
			Render("No description provided.")
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

	return common.RenderLabels(labels, common.LabelOpts{
		Width:     width,
		PillStyle: style,
	})
}

func (m *Model) getIndentedContentWidth() int {
	return m.width - 6
}

func (m *Model) SetWidth(width int) {
	m.width = width
	m.editor.SetWidth(width)
}

func (m *Model) SetSectionId(id int) {
	m.sectionId = id
}

func (m *Model) SetRow(issueData *data.IssueData) {
	if issueData == nil {
		m.issue = nil
	} else {
		m.issue = &issuerow.Issue{Ctx: m.ctx, Data: *issueData}
	}
}

// EnrichIssueComments fetches comments for the current issue (for GitLab)
func (m *Model) EnrichIssueComments() tea.Cmd {
	if m.issue == nil {
		return nil
	}
	// Only fetch if we don't have comments yet
	if len(m.issue.Data.Comments.Nodes) > 0 {
		return nil
	}
	issueUrl := m.issue.Data.Url
	return func() tea.Msg {
		comments, err := data.FetchIssueComments(issueUrl)
		return IssueCommentsMsg{
			IssueUrl: issueUrl,
			Comments: comments,
			Err:      err,
		}
	}
}

// IssueCommentsMsg is sent when issue comments are fetched
type IssueCommentsMsg struct {
	IssueUrl string
	Comments []data.IssueComment
	Err      error
}

// SetIssueComments updates the issue with fetched comments
func (m *Model) SetIssueComments(issueUrl string, comments []data.IssueComment) {
	if m.issue != nil && m.issue.Data.Url == issueUrl {
		m.issue.Data.Comments.Nodes = comments
		m.issue.Data.Comments.TotalCount = len(comments)
	}
}

func (m *Model) IsTextInputBoxFocused() bool {
	return m.editor.Active()
}

func (m *Model) GetIsCommenting() bool {
	return m.editor.Mode() == detailedit.ModeComment
}

func (m *Model) SetIsCommenting(isCommenting bool) tea.Cmd {
	if m.issue == nil {
		return nil
	}

	if !isCommenting {
		if m.editor.Mode() == detailedit.ModeComment {
			m.editor = m.editor.Exit()
		}
		return nil
	}

	editor, cmd := m.editor.Enter(detailedit.EnterOptions{
		Mode:                             detailedit.ModeComment,
		Prompt:                           constants.CommentPrompt,
		Source:                           dataautocomplete.UserMentionSource{},
		Repo:                             m.repoRef(),
		SuggestionKind:                   detailedit.SuggestionUsers,
		EnterFetch:                       detailedit.FetchSilent,
		ConfirmDiscardOnCancel:           true,
		HideAutocompleteWhenContextEmpty: true,
	})
	m.editor = editor
	return cmd
}

func (m *Model) GetIsAssigning() bool {
	return m.editor.Mode() == detailedit.ModeAssign
}

func (m *Model) SetIsAssigning(isAssigning bool) tea.Cmd {
	if m.issue == nil {
		return nil
	}

	if !isAssigning {
		if m.editor.Mode() == detailedit.ModeAssign {
			m.editor = m.editor.Exit()
		}
		return nil
	}

	initialValue := ""
	if !m.userAssignedToIssue(m.ctx.User) {
		initialValue = m.ctx.User
	}

	editor, cmd := m.editor.Enter(detailedit.EnterOptions{
		Mode:                             detailedit.ModeAssign,
		Prompt:                           constants.AssignPrompt,
		InitialValue:                     initialValue,
		Source:                           dataautocomplete.WhitespaceSource{},
		Repo:                             m.repoRef(),
		SuggestionKind:                   detailedit.SuggestionUsers,
		EnterFetch:                       detailedit.FetchSilent,
		HideAutocompleteWhenContextEmpty: false,
	})
	m.editor = editor
	return cmd
}

func (m *Model) SetIsLabeling(isLabeling bool) tea.Cmd {
	if m.issue == nil {
		return nil
	}

	if !isLabeling {
		if m.editor.Mode() == detailedit.ModeLabel {
			m.editor = m.editor.Exit()
		}
		return nil
	}

	labels := make([]string, 0, len(m.issue.Data.Labels.Nodes)+1)
	for _, label := range m.issue.Data.Labels.Nodes {
		labels = append(labels, label.Name)
	}
	labels = append(labels, "")

	editor, cmd := m.editor.Enter(detailedit.EnterOptions{
		Mode:                             detailedit.ModeLabel,
		Prompt:                           constants.LabelPrompt,
		InitialValue:                     strings.Join(labels, ", "),
		Source:                           dataautocomplete.LabelSource{},
		Repo:                             m.repoRef(),
		SuggestionKind:                   detailedit.SuggestionLabels,
		EnterFetch:                       detailedit.FetchSilent,
		HideAutocompleteWhenContextEmpty: false,
	})
	m.editor = editor
	return cmd
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
	return m.editor.Mode() == detailedit.ModeUnassign
}

func (m *Model) SetIsUnassigning(isUnassigning bool) tea.Cmd {
	if m.issue == nil {
		return nil
	}

	if !isUnassigning {
		if m.editor.Mode() == detailedit.ModeUnassign {
			m.editor = m.editor.Exit()
		}
		return nil
	}

	editor, cmd := m.editor.Enter(detailedit.EnterOptions{
		Mode:         detailedit.ModeUnassign,
		Prompt:       constants.UnassignPrompt,
		InitialValue: strings.Join(m.issueAssignees(), "\n"),
		Repo:         m.repoRef(),
	})
	m.editor = editor
	return cmd
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
	m.editor.UpdateProgramContext(ctx)
}

func (m *Model) repoRef() detailedit.RepoRef {
	owner, repo := m.issue.Data.GetRepoNameAndOwner()
	return detailedit.RepoRef{
		NameWithOwner: m.issue.Data.GetRepoNameWithOwner(),
		Owner:         owner,
		Name:          repo,
	}
}
