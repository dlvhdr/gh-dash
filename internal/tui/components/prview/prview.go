package prview

import (
	"fmt"
	"image/color"
	"regexp"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/common"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/cmp"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/carousel"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/cmpcontroller"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prssection"
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
	foldBodyHeight   = 8
)

type Model struct {
	ctx             *context.ProgramContext
	sectionId       int
	pr              *prrow.PullRequest
	width           int
	carousel        carousel.Model
	editor          cmpcontroller.Controller
	summaryViewMore bool
}

var tabs = []string{" Overview", " Activity", " Commits", " Checks", " Files Changed"}

func NewModel(ctx *context.ProgramContext) Model {
	c := carousel.New(
		carousel.WithItems(tabs),
		carousel.WithWidth(ctx.MainContentWidth),
	)

	return Model{
		pr:       nil,
		carousel: c,
		editor:   cmpcontroller.New(ctx),
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	editor, cmd, submit, handled := m.editor.Update(msg)
	m.editor = editor

	if submit != nil {
		if m.pr == nil {
			return m, nil
		}

		sid := tasks.SectionIdentifier{Id: m.sectionId, Type: prssection.SectionType}

		switch submit.Mode {
		case cmpcontroller.ModeComment:
			if len(strings.TrimSpace(submit.Value)) != 0 {
				return m, tasks.CommentOnPR(m.ctx, sid, m.pr.Data.Primary, submit.Value)
			}
			return m, nil

		case cmpcontroller.ModeApprove:
			comment := ""
			if len(strings.TrimSpace(submit.Value)) != 0 {
				comment = submit.Value
			}
			return m, tasks.ApprovePR(m.ctx, sid, m.pr.Data.Primary, comment)

		case cmpcontroller.ModeAssign:
			usernames := cmp.AllWords(submit.Value)
			if len(usernames) > 0 {
				return m, tasks.AssignPR(m.ctx, sid, m.pr.Data.Primary, usernames)
			}
			return m, nil

		case cmpcontroller.ModeUnassign:
			usernames := cmp.AllWords(submit.Value)
			if len(usernames) > 0 {
				return m, tasks.UnassignPR(m.ctx, sid, m.pr.Data.Primary, usernames)
			}
			return m, nil

		case cmpcontroller.ModeLabel:
			labels := cmp.CurrentLabels(submit.Value)
			if len(labels) > 0 || len(m.pr.Data.Primary.Labels.Nodes) > 0 {
				return m, m.label(labels)
			}
			return m, nil
		}
	}
	if handled {
		return m, cmd
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, keys.PRKeys.PrevSidebarTab):
			m.carousel.MoveLeft()
		case key.Matches(keyMsg, keys.PRKeys.NextSidebarTab):
			m.carousel.MoveRight()
		}
	}

	return m, cmd
}

func (m Model) View() string {
	header := strings.Builder{}

	header.WriteString(m.renderFullNameAndNumber())
	header.WriteString("\n")

	header.WriteString(m.renderTitle())
	header.WriteString("\n\n")
	header.WriteString(m.renderBranches())
	header.WriteString("\n\n")
	header.WriteString(m.renderAuthor())
	header.WriteString("\n\n")
	header.WriteString(lipgloss.NewStyle().Width(m.width).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(m.ctx.Theme.FaintBorder).
		Render(m.carousel.View()),
	)

	header.WriteString("\n")

	body := strings.Builder{}

	switch m.carousel.SelectedItem() {
	case tabs[0]:
		reviewers := m.renderRequestedReviewers()
		if reviewers != "" {
			body.WriteString(reviewers)
			body.WriteString("\n\n")
		}

		labels := m.renderLabels()
		if labels != "" {
			body.WriteString(labels)
			body.WriteString("\n\n")
		}

		body.WriteString(m.renderSummary())
		body.WriteString("\n\n")
		body.WriteString(
			m.ctx.Styles.Common.MainTextStyle.MarginBottom(1).Underline(true).Render(" Changes"),
		)
		body.WriteString("\n")
		body.WriteString(m.renderChangesOverview())
		body.WriteString("\n\n")
		body.WriteString(
			m.ctx.Styles.Common.MainTextStyle.MarginBottom(1).Underline(true).Render(" Checks"),
		)
		body.WriteString("\n")
		body.WriteString(m.renderChecksOverview())

		if editorView := m.editor.View(); editorView != "" {
			body.WriteString(editorView)
		}

	case tabs[1]:
		body.WriteString(m.renderActivity())
	case tabs[2]:
		body.WriteString(m.renderCommits())
	case tabs[3]:
		body.WriteString(m.renderChecksOverview())
		body.WriteString("\n\n")
		body.WriteString(m.renderChecks())
	case tabs[4]:
		body.WriteString(m.renderChangedFiles())
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header.String(),
		lipgloss.NewStyle().Padding(0, m.ctx.Styles.Sidebar.ContentPadding).Render(body.String()),
	)
}

func (m *Model) renderFullNameAndNumber() string {
	return common.RenderPreviewHeader(
		m.ctx.Theme,
		m.width,
		fmt.Sprintf(
			"%s · #%d",
			m.pr.Data.Primary.GetRepoNameWithOwner(),
			m.pr.Data.Primary.GetNumber(),
		),
	)
}

func (m *Model) renderTitle() string {
	return common.RenderPreviewTitle(
		m.ctx.Theme,
		m.ctx.Styles.Common,
		m.width,
		m.pr.Data.Primary.Title,
	)
}

func (m *Model) renderBranches() string {
	return lipgloss.JoinHorizontal(lipgloss.Left,
		" ",
		m.renderStatusPill(),
		" ",
		lipgloss.NewStyle().
			Foreground(m.ctx.Theme.SecondaryText).
			Render(m.pr.Data.Primary.BaseRefName+"  "+m.pr.Data.Primary.HeadRefName))
}

func (m *Model) renderStatusPill() string {
	var bgColor color.Color
	switch m.pr.Data.Primary.State {
	case "OPEN":
		if m.pr.Data.Primary.IsDraft {
			bgColor = m.ctx.Theme.FaintText.Dark
		} else {
			bgColor = m.ctx.Styles.Colors.OpenPR.Dark
		}
	case "CLOSED":
		bgColor = m.ctx.Styles.Colors.ClosedPR.Dark
	case "MERGED":
		bgColor = m.ctx.Styles.Colors.MergedPR.Dark
	}

	return m.ctx.Styles.PrView.PillStyle.
		BorderForeground(bgColor).
		Background(bgColor).
		Render(m.pr.RenderState())
}

func (m *Model) renderLabels() string {
	width := m.getIndentedContentWidth()
	labels := m.pr.Data.Primary.Labels.Nodes
	style := m.ctx.Styles.PrView.PillStyle
	if len(labels) == 0 {
		return ""
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.ctx.Styles.Common.MainTextStyle.Underline(true).Bold(true).Render(
			fmt.Sprintf("%s Labels", constants.LabelsIcon)),
		"",
		common.RenderLabels(labels, common.LabelOpts{
			Width:     width,
			PillStyle: style,
		}),
	)
}

type reviewerItem struct {
	text string
}

func (m *Model) renderRequestedReviewers() string {
	if !m.pr.Data.IsEnriched {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			m.ctx.Styles.Common.MainTextStyle.Underline(true).Bold(true).Render(
				fmt.Sprintf("%s Reviewers", constants.CodeReviewIcon)),
			"",
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				m.ctx.Styles.Common.WaitingGlyph,
				" ",
				m.ctx.Styles.Common.FaintTextStyle.Render("Loading..."),
			),
		)
	}

	reviewRequests := m.pr.Data.Enriched.ReviewRequests.Nodes
	reviews := m.pr.Data.Enriched.Reviews.Nodes
	suggestedReviewers := m.pr.Data.Enriched.SuggestedReviewers

	if len(reviewRequests) == 0 && len(reviews) == 0 && len(suggestedReviewers) == 0 {
		return ""
	}

	reviewStates := make(map[string]string)
	for _, review := range reviews {
		login := review.Author.Login
		existingState := reviewStates[login]
		// Don't override APPROVED or CHANGES_REQUESTED with COMMENTED
		if review.State == "COMMENTED" &&
			(existingState == "APPROVED" || existingState == "CHANGES_REQUESTED") {
			continue
		}
		reviewStates[login] = review.State
	}

	reviewerItems := make([]reviewerItem, 0)
	faintStyle := m.ctx.Styles.Common.FaintTextStyle
	reviewerStyle := lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText)
	successStyle := lipgloss.NewStyle().Foreground(m.ctx.Theme.SuccessText)
	errorStyle := lipgloss.NewStyle().Foreground(m.ctx.Theme.ErrorText)

	shownReviewers := make(map[string]bool)

	for _, req := range reviewRequests {
		displayName := req.GetReviewerDisplayName()
		if displayName == "" {
			continue
		}
		shownReviewers[displayName] = true

		var reviewerStr string
		stateIcon := ""
		if state, hasReview := reviewStates[displayName]; hasReview && state == "COMMENTED" {
			stateIcon = m.ctx.Styles.Common.CommentGlyph
		} else {
			stateIcon = m.ctx.Styles.Common.WaitingDotGlyph
		}

		if req.IsTeam() {
			reviewerStr += reviewerStyle.Render(displayName)
		} else {
			reviewerStr += reviewerStyle.Render("@" + displayName)
		}

		if req.AsCodeOwner {
			reviewerStr = lipgloss.JoinHorizontal(lipgloss.Top,
				faintStyle.Render(constants.OwnerIcon), " ", reviewerStr)
		}
		reviewerStr = lipgloss.JoinHorizontal(lipgloss.Top, stateIcon, " ", reviewerStr)

		reviewerItems = append(reviewerItems, reviewerItem{text: reviewerStr})
	}

	for login, state := range reviewStates {
		if shownReviewers[login] {
			continue
		}
		if state != "APPROVED" && state != "CHANGES_REQUESTED" && state != "COMMENTED" {
			continue
		}
		shownReviewers[login] = true

		var stateIcon string
		switch state {
		case "APPROVED":
			stateIcon = successStyle.Render(constants.ApprovedIcon)
		case "CHANGES_REQUESTED":
			stateIcon = errorStyle.Render(constants.ChangesRequestedIcon)
		case "COMMENTED":
			stateIcon = m.ctx.Styles.Common.CommentGlyph
		}
		reviewerStr := stateIcon + " " + reviewerStyle.Render("@"+login)

		reviewerItems = append(reviewerItems, reviewerItem{text: reviewerStr})
	}

	// Show suggested reviewers (= code owners) who haven't been requested or reviewed yet
	for _, suggested := range suggestedReviewers {
		login := suggested.Reviewer.Login
		if shownReviewers[login] {
			continue
		}
		if suggested.IsAuthor {
			continue
		}
		shownReviewers[login] = true

		reviewerStr := lipgloss.JoinHorizontal(lipgloss.Top,
			faintStyle.Render(constants.OwnerIcon), " ",
			faintStyle.Render("@"+login),
		)

		reviewerItems = append(reviewerItems, reviewerItem{text: reviewerStr})
	}

	if len(reviewerItems) == 0 {
		return ""
	}

	width := m.getIndentedContentWidth()
	var rows []string
	var currentRow strings.Builder
	currentRowWidth := 0

	for i, item := range reviewerItems {
		itemWidth := lipgloss.Width(item.text)
		separator := ", "
		separatorWidth := lipgloss.Width(separator)

		// Check if adding this item would exceed the width
		needsSeparator := i < len(reviewerItems)-1
		totalItemWidth := itemWidth
		if needsSeparator {
			totalItemWidth += separatorWidth
		}

		if currentRowWidth > 0 && currentRowWidth+totalItemWidth > width {
			// Start a new row
			rows = append(rows, currentRow.String())
			currentRow.Reset()
			currentRowWidth = 0
		}

		currentRow.WriteString(item.text)
		currentRowWidth += itemWidth

		if needsSeparator {
			currentRow.WriteString(separator)
			currentRowWidth += separatorWidth
		}
	}

	// Add the last row
	if currentRow.Len() > 0 {
		rows = append(rows, currentRow.String())
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.ctx.Styles.Common.MainTextStyle.Underline(true).Bold(true).Render(
			fmt.Sprintf("%s Reviewers", constants.CodeReviewIcon)),
		"",
		strings.Join(rows, "\n"),
	)
}

func (m *Model) renderAuthor() string {
	authorAssociation := m.pr.Data.Primary.AuthorAssociation
	if authorAssociation == "" {
		authorAssociation = "unknown role"
	}
	time := lipgloss.NewStyle().Render(utils.TimeElapsed(m.pr.Data.Primary.CreatedAt))
	return lipgloss.JoinHorizontal(lipgloss.Top,
		" by ",
		lipgloss.NewStyle().Foreground(m.ctx.Theme.PrimaryText).Render(
			lipgloss.NewStyle().Bold(true).Render("@"+m.pr.Data.Primary.Author.Login)),
		lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render(
			lipgloss.JoinHorizontal(lipgloss.Top, " ⋅ ", time, " ago", " ⋅ ")),
		lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render(
			lipgloss.JoinHorizontal(lipgloss.Top, data.GetAuthorRoleIcon(m.pr.Data.Primary.AuthorAssociation,
				m.ctx.Theme),
				" ", lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render(strings.ToLower(authorAssociation))),
		),
	)
}

func (m *Model) renderSummary() string {
	width := m.getIndentedContentWidth()
	// Strip HTML comments from body and cleanup body.
	body := htmlCommentRegex.ReplaceAllString(m.pr.Data.Primary.Body, "")
	body = lineCleanupRegex.ReplaceAllString(body, "")

	desc := m.ctx.Styles.Common.MainTextStyle.Bold(true).Underline(true).Render(" Summary")
	title := lipgloss.JoinVertical(
		lipgloss.Left,
		desc,
		"",
	)
	sbody := lipgloss.NewStyle().Width(m.getIndentedContentWidth())
	body = strings.TrimSpace(body)
	if body == "" {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			sbody.Italic(true).Foreground(m.ctx.Theme.FaintText).Render("No description provided."),
		)
	}

	markdownRenderer := markdown.GetMarkdownRenderer(width)
	rendered, err := markdownRenderer.Render(body)
	if err != nil {
		return ""
	}

	bodyHeight := lipgloss.Height(rendered)
	if !m.summaryViewMore && bodyHeight > foldBodyHeight {
		rendered = lipgloss.NewStyle().MaxHeight(foldBodyHeight).Render(rendered)
		rendered = lipgloss.JoinVertical(lipgloss.Left,
			rendered,
			"",
			lipgloss.PlaceHorizontal(m.getIndentedContentWidth(), lipgloss.Center,
				lipgloss.JoinHorizontal(
					lipgloss.Top,
					lipgloss.NewStyle().Bold(true).Italic(true).Render("Press "),
					lipgloss.NewStyle().
						Background(m.ctx.Theme.SelectedBackground).
						Foreground(m.ctx.Theme.PrimaryText).
						Render("e"),
					lipgloss.NewStyle().Bold(true).Italic(true).Render(" to read more..."),
				),
			),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title,
		lipgloss.NewStyle().
			Width(width).
			MaxWidth(width).
			Align(lipgloss.Left).
			Render(rendered),
	)
}

func (m *Model) SetSectionId(id int) {
	m.sectionId = id
}

func (m *Model) SetRow(d *prrow.Data) {
	if d == nil {
		m.pr = nil
	} else {
		m.pr = &prrow.PullRequest{Ctx: m.ctx, Data: d}
	}
}

type EnrichedPrMsg struct {
	Id   int
	Type string
	Data data.EnrichedPullRequestData
	Err  error
}

func (m *Model) EnrichCurrRow() tea.Cmd {
	if m == nil || m.pr == nil || m.pr.Data.IsEnriched {
		return nil
	}
	url := m.pr.Data.Primary.Url
	return func() tea.Msg {
		d, err := data.FetchPullRequest(url)
		return EnrichedPrMsg{
			Id:   m.sectionId,
			Type: prssection.SectionType,
			Data: d,
			Err:  err,
		}
	}
}

func (m *Model) SetWidth(width int) {
	m.width = width
	m.carousel.SetWidth(width)
	m.editor.SetWidth(width)
}

func (m *Model) IsTextInputBoxFocused() bool {
	return m.editor.Active()
}

func (m *Model) GetIsCommenting() bool {
	return m.editor.Mode() == cmpcontroller.ModeComment
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
	m.editor.UpdateProgramContext(ctx)
	m.carousel.SetStyles(
		carousel.Styles{
			Item:     lipgloss.NewStyle().Padding(0, 1).Foreground(m.ctx.Theme.FaintText),
			Selected: lipgloss.NewStyle().Padding(0, 1).Bold(true),
		},
	)
}

func (m *Model) SetIsCommenting(isCommenting bool) tea.Cmd {
	if m.pr == nil {
		return nil
	}

	if !isCommenting {
		if m.editor.Mode() == cmpcontroller.ModeComment {
			m.editor = m.editor.Exit()
		}
		return nil
	}

	editor, cmd := m.editor.Enter(cmpcontroller.EnterOptions{
		Mode:                             cmpcontroller.ModeComment,
		Prompt:                           constants.CommentPrompt,
		Source:                           cmp.UserMentionSource{},
		Repo:                             m.repoRef(),
		SuggestionKind:                   cmpcontroller.SuggestionUsers,
		EnterFetch:                       cmpcontroller.FetchSilent,
		ConfirmDiscardOnCancel:           true,
		HideAutocompleteWhenContextEmpty: true,
	})
	m.editor = editor
	return cmd
}

func (m *Model) getIndentedContentWidth() int {
	return m.width - 3*m.ctx.Styles.Sidebar.ContentPadding
}

func (m *Model) GetIsApproving() bool {
	return m.editor.Mode() == cmpcontroller.ModeApprove
}

func (m *Model) SetIsApproving(isApproving bool) tea.Cmd {
	if m.pr == nil {
		return nil
	}

	if !isApproving {
		if m.editor.Mode() == cmpcontroller.ModeApprove {
			m.editor = m.editor.Exit()
		}
		return nil
	}

	editor, cmd := m.editor.Enter(cmpcontroller.EnterOptions{
		Mode:                             cmpcontroller.ModeApprove,
		Prompt:                           constants.ApprovalPrompt,
		InitialValue:                     m.ctx.Config.Defaults.PrApproveComment,
		Source:                           cmp.WhitespaceSource{},
		Repo:                             m.repoRef(),
		SuggestionKind:                   cmpcontroller.SuggestionUsers,
		EnterFetch:                       cmpcontroller.FetchSilent,
		ConfirmDiscardOnCancel:           true,
		HideAutocompleteWhenContextEmpty: false,
	})
	m.editor = editor
	return cmd
}

func (m *Model) GetIsAssigning() bool {
	return m.editor.Mode() == cmpcontroller.ModeAssign
}

func (m *Model) SetIsAssigning(isAssigning bool) tea.Cmd {
	if m.pr == nil {
		return nil
	}

	if !isAssigning {
		if m.editor.Mode() == cmpcontroller.ModeAssign {
			m.editor = m.editor.Exit()
		}
		return nil
	}

	initialValue := ""
	if !m.userAssignedToPr(m.ctx.User) {
		initialValue = m.ctx.User
	}

	editor, cmd := m.editor.Enter(cmpcontroller.EnterOptions{
		Mode:                             cmpcontroller.ModeAssign,
		Prompt:                           constants.AssignPrompt,
		InitialValue:                     initialValue,
		Source:                           cmp.WhitespaceSource{},
		Repo:                             m.repoRef(),
		SuggestionKind:                   cmpcontroller.SuggestionUsers,
		EnterFetch:                       cmpcontroller.FetchSilent,
		HideAutocompleteWhenContextEmpty: false,
	})
	m.editor = editor
	return cmd
}

func (m *Model) userAssignedToPr(login string) bool {
	for _, a := range m.pr.Data.Primary.Assignees.Nodes {
		if login == a.Login {
			return true
		}
	}
	return false
}

func (m *Model) GetIsUnassigning() bool {
	return m.editor.Mode() == cmpcontroller.ModeUnassign
}

func (m *Model) SetIsUnassigning(isUnassigning bool) tea.Cmd {
	if m.pr == nil {
		return nil
	}

	if !isUnassigning {
		if m.editor.Mode() == cmpcontroller.ModeUnassign {
			m.editor = m.editor.Exit()
		}
		return nil
	}

	editor, cmd := m.editor.Enter(cmpcontroller.EnterOptions{
		Mode:         cmpcontroller.ModeUnassign,
		Prompt:       constants.UnassignPrompt,
		InitialValue: strings.Join(m.prAssignees(), "\n"),
		Repo:         m.repoRef(),
	})
	m.editor = editor
	return cmd
}

func (m *Model) prAssignees() []string {
	var assignees []string
	for _, n := range m.pr.Data.Primary.Assignees.Nodes {
		assignees = append(assignees, n.Login)
	}
	return assignees
}

func (m *Model) GoToFirstTab() {
	m.carousel.SetCursor(0)
}

func (m *Model) GoToActivityTab() {
	m.carousel.SetCursor(1) // Activity is the second tab (index 1)
}

func (m Model) SelectedTab() string {
	return m.carousel.SelectedItem()
}

func (m *Model) SetSummaryViewMore() {
	m.summaryViewMore = true
}

func (m *Model) SetSummaryViewLess() {
	m.summaryViewMore = false
}

func (m *Model) SetEnrichedPR(data data.EnrichedPullRequestData) {
	if m.pr.Data.Primary.Url == data.Url {
		m.pr.Data.Enriched = data
		m.pr.Data.IsEnriched = true
	}
}

func (m *Model) GetIsLabeling() bool {
	return m.editor.Mode() == cmpcontroller.ModeLabel
}

// SetIsLabeling enters or exits labeling mode
func (m *Model) SetIsLabeling(isLabeling bool) tea.Cmd {
	if m.pr == nil {
		return nil
	}

	if !isLabeling {
		if m.editor.Mode() == cmpcontroller.ModeLabel {
			m.editor = m.editor.Exit()
		}
		return nil
	}

	labels := make([]string, 0, len(m.pr.Data.Primary.Labels.Nodes)+1)
	for _, label := range m.pr.Data.Primary.Labels.Nodes {
		labels = append(labels, label.Name)
	}
	labels = append(labels, "")

	editor, cmd := m.editor.Enter(cmpcontroller.EnterOptions{
		Mode:                             cmpcontroller.ModeLabel,
		Prompt:                           constants.LabelPrompt,
		InitialValue:                     strings.Join(labels, ", "),
		Source:                           cmp.LabelSource{},
		Repo:                             m.repoRef(),
		SuggestionKind:                   cmpcontroller.SuggestionLabels,
		EnterFetch:                       cmpcontroller.FetchSilent,
		HideAutocompleteWhenContextEmpty: false,
	})
	m.editor = editor
	return cmd
}

func (m *Model) repoRef() cmpcontroller.RepoRef {
	owner, repo := m.pr.Data.Primary.GetRepoNameAndOwner()
	return cmpcontroller.RepoRef{
		NameWithOwner: m.pr.Data.Primary.GetRepoNameWithOwner(),
		Owner:         owner,
		Name:          repo,
	}
}
