package prview

import (
	"fmt"
	"os/exec"
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
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/carousel"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/inputbox"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prssection"
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

// RepoLabelsFetchedMsg is sent when repository labels are successfully fetched
type RepoLabelsFetchedMsg struct {
	Labels []data.Label
}

// RepoLabelsFetchFailedMsg is sent when repository label fetching fails
type RepoLabelsFetchFailedMsg struct {
	Err error
}

// RepoUsersFetchedMsg is sent when repository users are successfully fetched
type RepoUsersFetchedMsg struct {
	Users []data.User
}

// RepoUsersFetchFailedMsg is sent when repository user fetching fails
type RepoUsersFetchFailedMsg struct {
	Err error
}

// PrActionErrorMsg is sent when a PR action fails
type PrActionErrorMsg struct {
	Err error
}

// PrLabeledMsg is sent when a PR is successfully labeled
type PrLabeledMsg struct{}

type Model struct {
	ctx       *context.ProgramContext
	sectionId int
	pr        *prrow.PullRequest
	width     int
	carousel  carousel.Model

	ShowConfirmCancel bool
	isCommenting      bool
	isApproving       bool
	isAssigning       bool
	isUnassigning     bool
	isLabeling        bool
	summaryViewMore   bool

	inputBox   inputbox.Model
	ac         *autocomplete.Model
	repoLabels []data.Label
	repoUsers  []data.User
}

var tabs = []string{" Overview", " Activity", " Commits", " Checks", " Files Changed"}

func NewModel(ctx *context.ProgramContext) Model {
	inputBox := inputbox.NewModel(ctx)

	// Set up autocomplete for labeling
	ac := autocomplete.NewModel(ctx)
	inputBox.ContextExtractor = autocomplete.LabelContextExtractor
	inputBox.SuggestionInserter = autocomplete.LabelSuggestionInserter
	inputBox.ItemsToExclude = autocomplete.LabelItemsToExclude
	inputBox.SetAutocomplete(&ac)

	c := carousel.New(
		carousel.WithItems(tabs),
		carousel.WithWidth(ctx.MainContentWidth),
	)

	return Model{
		pr: nil,

		isCommenting:  false,
		isApproving:   false,
		isAssigning:   false,
		isUnassigning: false,
		isLabeling:    false,
		carousel:      c,

		inputBox:   inputBox,
		ac:         &ac,
		repoLabels: nil,
		repoUsers:  nil,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmds  []tea.Cmd
		cmd   tea.Cmd
		taCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case RepoLabelsFetchedMsg:
		clearCmd := m.ac.SetFetchSuccess()
		m.repoLabels = msg.Labels
		labelNames := data.LabelNames(msg.Labels)
		m.ac.SetSuggestions(labelNames)
		if m.isLabeling {
			cursorPos := m.inputBox.CursorPosition()
			currentLabel, _, _ := autocomplete.LabelContextExtractor(m.inputBox.Value(), cursorPos)
			existingLabels := autocomplete.LabelItemsToExclude(m.inputBox.Value(), cursorPos)
			m.ac.Show(currentLabel, existingLabels)
		}
		return m, clearCmd

	case RepoLabelsFetchFailedMsg:
		clearCmd := m.ac.SetFetchError(msg.Err)
		return m, clearCmd

	case RepoUsersFetchedMsg:
		clearCmd := m.ac.SetFetchSuccess()
		m.repoUsers = msg.Users
		m.ac.SetSuggestions(data.UserLogins(msg.Users))
		if m.isCommenting {
			cursorPos := m.inputBox.CursorPosition()
			mention, _, _ := autocomplete.UserMentionContextExtractor(m.inputBox.Value(), cursorPos)
			if mention != "" {
				m.ac.Show(mention, nil)
			}
		} else if m.isApproving || m.isAssigning {
			cursorPos := m.inputBox.CursorPosition()
			word, _, _ := autocomplete.WhitespaceContextExtractor(m.inputBox.Value(), cursorPos)
			existingWords := autocomplete.WhitespaceItemsToExclude(m.inputBox.Value(), cursorPos)
			m.ac.Show(word, existingWords)
		}
		return m, clearCmd

	case RepoUsersFetchFailedMsg:
		clearCmd := m.ac.SetFetchError(msg.Err)
		return m, clearCmd

	case autocomplete.FetchSuggestionsRequestedMsg:
		if m.isLabeling {
			// If this is a forced refresh (e.g., via Ctrl+f), clear the cached labels
			// for this repo so FetchRepoLabels will actually call the gh CLI.
			if msg.Force {
				if m.pr != nil {
					repoName := m.pr.Data.Primary.GetRepoNameWithOwner()
					data.ClearRepoLabelCache(repoName)
				}
			}
			cmd := m.fetchLabels()
			return m, cmd
		} else if m.isCommenting || m.isApproving {
			// If this is a forced refresh (e.g., via Ctrl+f), clear the cached users
			// for this repo so FetchRepoCollaborators will actually call the gh CLI.
			if msg.Force {
				if m.pr != nil {
					repoName := m.pr.Data.Primary.GetRepoNameWithOwner()
					data.ClearRepoUserCache(repoName)
				}
			}
			cmd := m.fetchUsers()
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		if m.isCommenting {
			switch msg.Type {
			case tea.KeyCtrlD:
				if len(strings.Trim(m.inputBox.Value(), " ")) != 0 {
					cmd = m.comment(m.inputBox.Value())
				}
				m.inputBox.Blur()
				m.isCommenting = false
				return m, cmd

			case tea.KeyEsc, tea.KeyCtrlC:
				if !m.ShowConfirmCancel {
					m.shouldCancelComment()
				}

			default:
				if msg.String() == "Y" || msg.String() == "y" {
					if m.shouldCancelComment() {
						return m, nil
					}
				}
				if m.ShowConfirmCancel && (msg.String() == "N" || msg.String() == "n") {
					m.inputBox.SetPrompt(constants.CommentPrompt)
					m.ShowConfirmCancel = false
					return m, nil
				}
				m.inputBox.SetPrompt(constants.CommentPrompt)
				m.ShowConfirmCancel = false
			}

			// Track @-mention context before and after the keystroke
			previousCursorPos := m.inputBox.CursorPosition()
			previousValue := m.inputBox.Value()
			previousMention, _, _ := autocomplete.UserMentionContextExtractor(previousValue, previousCursorPos)

			m.inputBox, taCmd = m.inputBox.Update(msg)
			cmds = append(cmds, cmd, taCmd)

			// Check for @-mention context change after the keystroke
			currentCursorPos := m.inputBox.CursorPosition()
			currentValue := m.inputBox.Value()
			currentMention, _, _ := autocomplete.UserMentionContextExtractor(currentValue, currentCursorPos)

			if currentMention != previousMention {
				if currentMention != "" {
					// User is typing an @-mention, show autocomplete
					m.ac.Show(currentMention, nil)
				} else {
					// No longer in an @-mention context
					m.ac.Hide()
				}
			}
		} else if m.isApproving {
			switch msg.Type {
			case tea.KeyCtrlD:
				comment := ""
				if len(strings.Trim(m.inputBox.Value(), " ")) != 0 {
					comment = m.inputBox.Value()
				}
				cmd = m.approve(comment)
				m.inputBox.Blur()
				m.isApproving = false
				m.ac.Hide()
				return m, cmd

			case tea.KeyEsc, tea.KeyCtrlC:
				if m.shouldCancelComment() {
					return m, nil
				}
			default:
				m.inputBox.SetPrompt(constants.ApprovalPrompt)
				m.ShowConfirmCancel = false
			}

			// Track current user context before and after the keystroke
			previousCursorPos := m.inputBox.CursorPosition()
			previousValue := m.inputBox.Value()
			previousUser, _, _ := autocomplete.WhitespaceContextExtractor(previousValue, previousCursorPos)

			m.inputBox, taCmd = m.inputBox.Update(msg)
			cmds = append(cmds, cmd, taCmd)

			// Check for user context change after the keystroke
			currentCursorPos := m.inputBox.CursorPosition()
			currentValue := m.inputBox.Value()
			currentUser, _, _ := autocomplete.WhitespaceContextExtractor(currentValue, currentCursorPos)

			if currentUser != previousUser {
				// Always show autocomplete for approve mode (even with empty word)
				existingUsers := autocomplete.WhitespaceItemsToExclude(currentValue, currentCursorPos)
				m.ac.Show(currentUser, existingUsers)
			}
		} else if m.isAssigning {
			switch msg.Type {
			case tea.KeyCtrlD:
				usernames := strings.Fields(m.inputBox.Value())
				if len(usernames) > 0 {
					cmd = m.assign(usernames)
				}
				m.inputBox.Blur()
				m.isAssigning = false
				m.ac.Hide()
				return m, cmd

			case tea.KeyEsc, tea.KeyCtrlC:
				m.inputBox.Blur()
				m.isAssigning = false
				m.ac.Hide()
				return m, nil
			}

			// Track current word context before and after the keystroke
			previousCursorPos := m.inputBox.CursorPosition()
			previousValue := m.inputBox.Value()
			previousWord, _, _ := autocomplete.WhitespaceContextExtractor(previousValue, previousCursorPos)

			m.inputBox, taCmd = m.inputBox.Update(msg)
			cmds = append(cmds, cmd, taCmd)

			// Check for word context change after the keystroke
			currentCursorPos := m.inputBox.CursorPosition()
			currentValue := m.inputBox.Value()
			currentWord, _, _ := autocomplete.WhitespaceContextExtractor(currentValue, currentCursorPos)

			if currentWord != previousWord {
				// Always show autocomplete for assign mode (even with empty word)
				existingWords := autocomplete.WhitespaceItemsToExclude(currentValue, currentCursorPos)
				m.ac.Show(currentWord, existingWords)
			}
		} else if m.isUnassigning {
			switch msg.Type {
			case tea.KeyCtrlD:
				usernames := strings.Fields(m.inputBox.Value())
				if len(usernames) > 0 {
					cmd = m.unassign(usernames)
				}
				m.inputBox.Blur()
				m.isUnassigning = false
				return m, cmd

			case tea.KeyEsc, tea.KeyCtrlC:
				m.inputBox.Blur()
				m.isUnassigning = false
				return m, nil
			}

			m.inputBox, taCmd = m.inputBox.Update(msg)
			cmds = append(cmds, cmd, taCmd)
		} else if m.isLabeling {
			switch msg.Type {
			case tea.KeyCtrlD:
				labels := autocomplete.CurrentLabels(m.inputBox.Value())
				if len(labels) > 0 {
					cmd = m.label(labels)
				}
				m.inputBox.Blur()
				m.isLabeling = false
				m.ac.Hide()
				return m, cmd

			case tea.KeyEsc, tea.KeyCtrlC:
				m.inputBox.Blur()
				m.isLabeling = false
				m.ac.Hide()
				return m, nil
			}

			// Track label context before and after the keystroke
			previousCursorPos := m.inputBox.CursorPosition()
			previousValue := m.inputBox.Value()
			previousLabel, _, _ := autocomplete.LabelContextExtractor(previousValue, previousCursorPos)

			m.inputBox, taCmd = m.inputBox.Update(msg)
			cmds = append(cmds, cmd, taCmd)

			// Check for label context change after the keystroke
			currentCursorPos := m.inputBox.CursorPosition()
			currentValue := m.inputBox.Value()
			currentLabel, _, _ := autocomplete.LabelContextExtractor(currentValue, currentCursorPos)

			if currentLabel != previousLabel {
				existingLabels := autocomplete.LabelItemsToExclude(currentValue, currentCursorPos)
				m.ac.Show(currentLabel, existingLabels)
			}
		} else {
			switch {
			case key.Matches(msg, keys.PRKeys.PrevSidebarTab):
				m.carousel.MoveLeft()
			case key.Matches(msg, keys.PRKeys.NextSidebarTab):
				m.carousel.MoveRight()
			}
		}
	}

	switch msg.(type) {
	case spinner.TickMsg, autocomplete.ClearFetchStatusMsg:
		var acCmd tea.Cmd
		*m.ac, acCmd = m.ac.Update(msg)
		cmds = append(cmds, acCmd)
	}

	return m, tea.Batch(cmds...)
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
		body.WriteString(m.ctx.Styles.Common.MainTextStyle.MarginBottom(1).Underline(true).Render(" Changes"))
		body.WriteString("\n")
		body.WriteString(m.renderChangesOverview())
		body.WriteString("\n\n")
		body.WriteString(m.ctx.Styles.Common.MainTextStyle.MarginBottom(1).Underline(true).Render(" Checks"))
		body.WriteString("\n")
		body.WriteString(m.renderChecksOverview())

		if m.isCommenting || m.isApproving || m.isLabeling || m.isAssigning {
			body.WriteString(m.inputBox.ViewWithAutocomplete())
		} else if m.isUnassigning {
			body.WriteString(m.inputBox.View())
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
	return common.RenderPreviewHeader(m.ctx.Theme, m.width,
		fmt.Sprintf("%s · #%d", m.pr.Data.Primary.GetRepoNameWithOwner(), m.pr.Data.Primary.GetNumber()))
}

func (m *Model) renderTitle() string {
	return common.RenderPreviewTitle(m.ctx.Theme, m.ctx.Styles.Common, m.width, m.pr.Data.Primary.Title)
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
	bgColor := ""
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
		BorderForeground(lipgloss.Color(bgColor)).
		Background(lipgloss.Color(bgColor)).
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
		common.RenderLabels(width, labels, style),
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
			lipgloss.JoinHorizontal(lipgloss.Top, m.ctx.Styles.Common.WaitingGlyph, " ", m.ctx.Styles.Common.FaintTextStyle.Render("Loading...")),
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
		if review.State == "COMMENTED" && (existingState == "APPROVED" || existingState == "CHANGES_REQUESTED") {
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
				m.ctx.Theme), " ", lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render(strings.ToLower(authorAssociation))),
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
				lipgloss.JoinHorizontal(lipgloss.Top,
					lipgloss.NewStyle().Bold(true).Italic(true).Render("Press "),
					lipgloss.NewStyle().Background(m.ctx.Theme.SelectedBackground).Foreground(m.ctx.Theme.PrimaryText).Render("e"),
					lipgloss.NewStyle().Bold(true).Italic(true).Render(" to read more...")),
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
	m.inputBox.SetWidth(width)
	m.ac.SetWidth(width - 4)
}

func (m *Model) IsTextInputBoxFocused() bool {
	return m.isCommenting || m.isAssigning || m.isApproving || m.isUnassigning || m.isLabeling
}

func (m *Model) GetIsCommenting() bool {
	return m.isCommenting
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
	m.inputBox.UpdateProgramContext(ctx)
	m.ac.UpdateProgramContext(ctx)
	m.carousel.SetStyles(
		carousel.Styles{
			Item:     lipgloss.NewStyle().Padding(0, 1).Foreground(m.ctx.Theme.FaintText),
			Selected: lipgloss.NewStyle().Padding(0, 1).Bold(true),
		},
	)
}

func (m *Model) shouldCancelComment() bool {
	if !m.ShowConfirmCancel {
		m.inputBox.SetPrompt(lipgloss.NewStyle().Foreground(m.ctx.Theme.ErrorText).Render("Discard comment? (y/N)"))
		m.ShowConfirmCancel = true
		return false
	}
	m.inputBox.Blur()
	m.isCommenting = false
	m.isApproving = false
	m.ac.Hide()
	m.ShowConfirmCancel = false
	return true
}

func (m *Model) SetIsCommenting(isCommenting bool) tea.Cmd {
	if m.pr == nil {
		return nil
	}

	if !m.isCommenting && isCommenting {
		m.inputBox.Reset()
		m.ac.Reset() // Clear any stale autocomplete state (e.g., from labeling)

		// Set up user mention autocomplete for commenting
		m.inputBox.ContextExtractor = autocomplete.UserMentionContextExtractor
		m.inputBox.SuggestionInserter = autocomplete.UserMentionSuggestionInserter
		m.inputBox.ItemsToExclude = autocomplete.UserMentionItemsToExclude
	}
	m.isCommenting = isCommenting
	m.inputBox.SetPrompt(constants.CommentPrompt)

	if isCommenting {
		// Fetch users for autocomplete if not already cached
		repoName := m.pr.Data.Primary.GetRepoNameWithOwner()
		if users, ok := data.CachedRepoUsers(repoName); ok {
			m.repoUsers = users
			m.ac.SetSuggestions(data.UserLogins(users))
			cursorPos := m.inputBox.CursorPosition()
			mention, _, _ := autocomplete.UserMentionContextExtractor(m.inputBox.Value(), cursorPos)
			if mention != "" {
				m.ac.Show(mention, nil)
			}
			return tea.Sequence(textarea.Blink, m.inputBox.Focus())
		}
		return tea.Sequence(m.fetchUsersSilent(), textarea.Blink, m.inputBox.Focus())
	}
	return nil
}

func (m *Model) getIndentedContentWidth() int {
	return m.width - 3*m.ctx.Styles.Sidebar.ContentPadding
}

func (m *Model) GetIsApproving() bool {
	return m.isApproving
}

func (m *Model) SetIsApproving(isApproving bool) tea.Cmd {
	if m.pr == nil {
		return nil
	}

	if !m.isApproving && isApproving {
		m.ac.Reset()
		m.inputBox.Reset()

		// Set up whitespace-based autocomplete for approving (users are whitespace-separated)
		m.inputBox.ContextExtractor = autocomplete.WhitespaceContextExtractor
		m.inputBox.SuggestionInserter = autocomplete.WhitespaceSuggestionInserter
		m.inputBox.ItemsToExclude = autocomplete.WhitespaceItemsToExclude
	}
	m.isApproving = isApproving
	m.inputBox.SetPrompt(constants.ApprovalPrompt)
	m.inputBox.SetValue(m.ctx.Config.Defaults.PrApproveComment)

	if isApproving {
		repoName := m.pr.Data.Primary.GetRepoNameWithOwner()
		if users, ok := data.CachedRepoUsers(repoName); ok {
			m.repoUsers = users
			m.ac.SetSuggestions(data.UserLogins(users))
			// Show autocomplete immediately for current word at cursor
			cursorPos := m.inputBox.CursorPosition()
			currentWord, _, _ := autocomplete.WhitespaceContextExtractor(m.inputBox.Value(), cursorPos)
			existingWords := autocomplete.WhitespaceItemsToExclude(m.inputBox.Value(), cursorPos)
			m.ac.Show(currentWord, existingWords)
			return tea.Sequence(textarea.Blink, m.inputBox.Focus())
		}
		// Fetch users asynchronously in background (no loading UI)
		return tea.Sequence(m.fetchUsers(), textarea.Blink, m.inputBox.Focus())
	}
	return nil
}

func (m *Model) GetIsAssigning() bool {
	return m.isAssigning
}

func (m *Model) SetIsAssigning(isAssigning bool) tea.Cmd {
	if m.pr == nil {
		return nil
	}

	if !m.isAssigning && isAssigning {
		m.inputBox.Reset()
		m.ac.Reset() // Clear any stale autocomplete state (e.g., from labeling)

		// Set up whitespace-based autocomplete for assigning (users are whitespace-separated)
		m.inputBox.ContextExtractor = autocomplete.WhitespaceContextExtractor
		m.inputBox.SuggestionInserter = autocomplete.WhitespaceSuggestionInserter
		m.inputBox.ItemsToExclude = autocomplete.WhitespaceItemsToExclude
	}
	m.isAssigning = isAssigning
	m.inputBox.SetPrompt(constants.AssignPrompt)
	if !m.userAssignedToPr(m.ctx.User) {
		m.inputBox.SetValue(m.ctx.User)
	}

	// Reset autocomplete
	m.ac.Hide()
	m.ac.SetSuggestions(nil)

	if isAssigning {
		repoName := m.pr.Data.Primary.GetRepoNameWithOwner()
		if users, ok := data.CachedRepoUsers(repoName); ok {
			m.repoUsers = users
			m.ac.SetSuggestions(data.UserLogins(users))
			// Show autocomplete immediately for current word at cursor
			cursorPos := m.inputBox.CursorPosition()
			currentWord, _, _ := autocomplete.WhitespaceContextExtractor(m.inputBox.Value(), cursorPos)
			existingWords := autocomplete.WhitespaceItemsToExclude(m.inputBox.Value(), cursorPos)
			m.ac.Show(currentWord, existingWords)
			return tea.Sequence(textarea.Blink, m.inputBox.Focus())
		}
		// Fetch users asynchronously in background (no loading UI)
		return tea.Sequence(m.fetchUsers(), textarea.Blink, m.inputBox.Focus())
	}
	return nil
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
	return m.isUnassigning
}

func (m *Model) SetIsUnassigning(isUnassigning bool) tea.Cmd {
	if m.pr == nil {
		return nil
	}

	if !m.isUnassigning && isUnassigning {
		m.inputBox.Reset()
	}
	m.isUnassigning = isUnassigning
	m.inputBox.SetPrompt(constants.UnassignPrompt)
	m.inputBox.SetValue(strings.Join(m.prAssignees(), "\n"))

	if isUnassigning {
		return tea.Sequence(textarea.Blink, m.inputBox.Focus())
	}
	return nil
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
	return m.isLabeling
}

// SetIsLabeling enters or exits labeling mode
func (m *Model) SetIsLabeling(isLabeling bool) tea.Cmd {
	if m.pr == nil {
		return nil
	}

	if !m.isLabeling && isLabeling {
		m.inputBox.Reset()

		// Set up label autocomplete for labeling
		m.inputBox.ContextExtractor = autocomplete.LabelContextExtractor
		m.inputBox.SuggestionInserter = autocomplete.LabelSuggestionInserter
		m.inputBox.ItemsToExclude = autocomplete.LabelItemsToExclude
	}
	m.isLabeling = isLabeling
	m.inputBox.SetPrompt(constants.LabelPrompt)

	// Pre-populate with current labels
	labels := make([]string, 0)
	for _, label := range m.pr.Data.Primary.Labels.Nodes {
		labels = append(labels, label.Name)
	}
	labels = append(labels, "")
	m.inputBox.SetValue(strings.Join(labels, ", "))

	// Reset autocomplete
	m.ac.Hide()
	m.ac.SetSuggestions(nil)

	// Trigger label fetching for autocomplete
	if isLabeling {
		repoName := m.pr.Data.Primary.GetRepoNameWithOwner()
		if labels, ok := data.CachedRepoLabels(repoName); ok {
			// Use cached labels
			m.repoLabels = labels
			m.ac.SetSuggestions(data.LabelNames(labels))
			cursorPos := m.inputBox.CursorPosition()
			currentLabel, _, _ := autocomplete.LabelContextExtractor(m.inputBox.Value(), cursorPos)
			existingLabels := autocomplete.LabelItemsToExclude(m.inputBox.Value(), cursorPos)
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
		repoName := m.pr.Data.Primary.GetRepoNameWithOwner()
		labels, err := data.FetchRepoLabels(repoName)
		if err != nil {
			return RepoLabelsFetchFailedMsg{Err: err}
		}
		return RepoLabelsFetchedMsg{Labels: labels}
	}

	return tea.Batch(spinnerTickCmd, fetchCmd)
}

// fetchUsers returns a command to fetch repository users for @-mention autocomplete
// This shows a loading UI - use when user explicitly requests a refresh (e.g., Ctrl+f)
func (m *Model) fetchUsers() tea.Cmd {
	spinnerTickCmd := m.ac.SetFetchLoading()

	fetchCmd := func() tea.Msg {
		repoName := m.pr.Data.Primary.GetRepoNameWithOwner()
		users, err := data.FetchRepoUsers(repoName)
		if err != nil {
			return RepoUsersFetchFailedMsg{Err: err}
		}
		return RepoUsersFetchedMsg{Users: users}
	}

	return tea.Batch(spinnerTickCmd, fetchCmd)
}

// fetchUsersSilent returns a command to fetch repository users without showing loading UI
// Use this for background fetching when entering commenting/approving modes
func (m *Model) fetchUsersSilent() tea.Cmd {
	return func() tea.Msg {
		repoName := m.pr.Data.Primary.GetRepoNameWithOwner()
		users, err := data.FetchRepoUsers(repoName)
		if err != nil {
			return RepoUsersFetchFailedMsg{Err: err}
		}
		return RepoUsersFetchedMsg{Users: users}
	}
}

// prLabels returns the current labels of the PR as a slice of strings
func (m *Model) prLabels() []string {
	var labels []string
	for _, n := range m.pr.Data.Primary.Labels.Nodes {
		labels = append(labels, n.Name)
	}
	return labels
}

// label executes the label command via gh CLI
func (m *Model) label(labels []string) tea.Cmd {
	prNumber := m.pr.Data.Primary.GetNumber()
	repoName := m.pr.Data.Primary.GetRepoNameWithOwner()

	cmd := func() tea.Msg {
		cmd := exec.Command("gh", "pr", "edit", fmt.Sprintf("%d", prNumber),
			"-R", repoName, "--add-label", strings.Join(labels, ","))
		output, err := cmd.CombinedOutput()
		if err != nil {
			return PrActionErrorMsg{Err: fmt.Errorf("failed to label PR: %s", string(output))}
		}
		return PrLabeledMsg{}
	}

	return cmd
}
