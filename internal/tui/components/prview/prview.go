package prview

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/common"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/carousel"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/inputbox"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prssection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/keys"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/markdown"
	tuitheme "github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

var (
	htmlCommentRegex = regexp.MustCompile("(?U)<!--(.|[[:space:]])*-->")
	lineCleanupRegex = regexp.MustCompile(`((\n)+|^)([^\r\n]*\|[^\r\n]*(\n)?)+`)
	commentPrompt    = "Leave a comment..."
	approvalPrompt   = "Approve with comment..."
	foldBodyHeight   = 8
)

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
	summaryViewMore   bool

	inputBox inputbox.Model
}

var tabs = []string{" Overview", " Checks", " Activity", " Files Changed"}

func NewModel(ctx *context.ProgramContext) Model {
	inputBox := inputbox.NewModel(ctx)
	inputBox.SetHeight(common.InputBoxHeight)

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
		carousel:      c,

		inputBox: inputBox,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmds  []tea.Cmd
		cmd   tea.Cmd
		taCmd tea.Cmd
	)

	switch msg := msg.(type) {
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
					m.inputBox.SetPrompt(commentPrompt)
					m.ShowConfirmCancel = false
					return m, nil
				}
				m.inputBox.SetPrompt(commentPrompt)
				m.ShowConfirmCancel = false
			}

			m.inputBox, taCmd = m.inputBox.Update(msg)
			cmds = append(cmds, cmd, taCmd)
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
				return m, cmd

			case tea.KeyEsc, tea.KeyCtrlC:
				if m.shouldCancelComment() {
					return m, nil
				}
			default:
				m.inputBox.SetPrompt(approvalPrompt)
				m.ShowConfirmCancel = false
			}

			m.inputBox, taCmd = m.inputBox.Update(msg)
			cmds = append(cmds, cmd, taCmd)
		} else if m.isAssigning {
			switch msg.Type {
			case tea.KeyCtrlD:
				usernames := strings.Fields(m.inputBox.Value())
				if len(usernames) > 0 {
					cmd = m.assign(usernames)
				}
				m.inputBox.Blur()
				m.isAssigning = false
				return m, cmd

			case tea.KeyEsc, tea.KeyCtrlC:
				m.inputBox.Blur()
				m.isAssigning = false
				return m, nil
			}

			m.inputBox, taCmd = m.inputBox.Update(msg)
			cmds = append(cmds, cmd, taCmd)
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
		} else {
			switch {
			case key.Matches(msg, keys.PRKeys.PrevSidebarTab):
				m.carousel.MoveLeft()
				return m, nil
			case key.Matches(msg, keys.PRKeys.NextSidebarTab):
				m.carousel.MoveRight()
				return m, nil
			}
			return m, nil
		}
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
	header.WriteString(lipgloss.NewStyle().Width(m.width).Background(m.ctx.Theme.MainBackground).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(m.ctx.Theme.FaintBorder).
		Render(m.carousel.View()),
	)

	header.WriteString("\n")

	body := strings.Builder{}

	switch m.carousel.SelectedItem() {
	case tabs[0]:
		labels := m.renderLabels()
		if labels != "" {
			body.WriteString(labels)
			body.WriteString("\n\n")
		}

		body.WriteString(m.renderSummary())
		body.WriteString("\n\n")
		body.WriteString(m.ctx.Styles.Common.MainTextStyle.MarginBottom(1).Underline(true).Width(m.getIndentedContentWidth()).Render(" Changes"))
		body.WriteString("\n")
		body.WriteString(m.renderChangesOverview())
		body.WriteString("\n\n")
		body.WriteString(m.ctx.Styles.Common.MainTextStyle.MarginBottom(1).Underline(true).Width(m.getIndentedContentWidth()).Render(" Checks"))
		body.WriteString("\n")
		body.WriteString(m.renderChecksOverview())

		if m.isCommenting || m.isApproving || m.isAssigning || m.isUnassigning {
			body.WriteString(m.inputBox.View())
		}

	case tabs[1]:
		body.WriteString(m.renderChecksOverview())
		body.WriteString("\n\n")
		body.WriteString(m.renderChecks())

	case tabs[2]:
		body.WriteString(m.renderActivity())
	case tabs[3]:
		body.WriteString(m.renderChangedFiles())
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		header.String(),
		lipgloss.NewStyle().Padding(0, m.ctx.Styles.Sidebar.ContentPadding).Background(m.ctx.Theme.MainBackground).Render(body.String()),
	)
	return lipgloss.NewStyle().Background(m.ctx.Theme.MainBackground).Render(content)
}

func (m *Model) renderFullNameAndNumber() string {
	return lipgloss.NewStyle().
		PaddingLeft(1).
		Width(m.width).
		Background(m.ctx.Theme.SelectedBackground).
		Foreground(m.ctx.Theme.SecondaryText).
		Render(fmt.Sprintf("%s · #%d", m.pr.Data.Primary.GetRepoNameWithOwner(), m.pr.Data.Primary.GetNumber()))
}

func (m *Model) renderTitle() string {
	return lipgloss.NewStyle().Height(3).Width(m.width).Background(
		m.ctx.Theme.SelectedBackground).PaddingLeft(1).Render(
		lipgloss.PlaceVertical(3, lipgloss.Center, m.ctx.Styles.Common.MainTextStyle.
			Background(m.ctx.Theme.SelectedBackground).
			Render(m.pr.Data.Primary.Title),
		),
	)
}

func (m *Model) renderBranches() string {
	bgStyle := lipgloss.NewStyle().Background(m.ctx.Theme.MainBackground)
	content := lipgloss.JoinHorizontal(lipgloss.Left,
		bgStyle.Render(" "),
		m.renderStatusPill(),
		bgStyle.Render(" "),
		lipgloss.NewStyle().
			Foreground(m.ctx.Theme.SecondaryText).
			Background(m.ctx.Theme.MainBackground).
			Render(m.pr.Data.Primary.BaseRefName+"  "+m.pr.Data.Primary.HeadRefName))
	return bgStyle.Width(m.getIndentedContentWidth()).Render(content)
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
		m.ctx.Styles.Common.MainTextStyle.Underline(true).Bold(true).Render("Labels"),
		"",
		common.RenderLabels(width, labels, style),
	)
}

func (m *Model) renderAuthor() string {
	authorAssociation := m.pr.Data.Primary.AuthorAssociation
	if authorAssociation == "" {
		authorAssociation = "unknown role"
	}
	bgStyle := lipgloss.NewStyle().Background(m.ctx.Theme.MainBackground)
	faintStyle := lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Background(m.ctx.Theme.MainBackground)
	time := bgStyle.Render(utils.TimeElapsed(m.pr.Data.Primary.CreatedAt))
	content := lipgloss.JoinHorizontal(lipgloss.Top,
		bgStyle.Render(" by "),
		lipgloss.NewStyle().Foreground(m.ctx.Theme.PrimaryText).Background(m.ctx.Theme.MainBackground).Bold(true).Render("@"+m.pr.Data.Primary.Author.Login),
		faintStyle.Render(" ⋅ "),
		time,
		faintStyle.Render(" ago ⋅ "),
		data.GetAuthorRoleIcon(m.pr.Data.Primary.AuthorAssociation, m.ctx.Theme),
		bgStyle.Render(" "),
		faintStyle.Render(strings.ToLower(authorAssociation)),
	)
	return lipgloss.NewStyle().Width(m.getIndentedContentWidth()).Background(m.ctx.Theme.MainBackground).Render(content)
}

func (m *Model) renderSummary() string {
	width := m.getIndentedContentWidth()
	// Strip HTML comments from body and cleanup body.
	body := htmlCommentRegex.ReplaceAllString(m.pr.Data.Primary.Body, "")
	body = lineCleanupRegex.ReplaceAllString(body, "")

	desc := m.ctx.Styles.Common.MainTextStyle.Bold(true).Underline(true).Width(width).Render(" Summary")
	title := lipgloss.JoinVertical(
		lipgloss.Left,
		desc,
		"",
	)
	sbody := lipgloss.NewStyle().Width(m.getIndentedContentWidth()).Background(m.ctx.Theme.MainBackground)
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
		rendered = lipgloss.NewStyle().MaxHeight(foldBodyHeight).Background(m.ctx.Theme.MainBackground).Render(rendered)
		moreLineRaw := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Bold(true).Italic(true).Background(m.ctx.Theme.MainBackground).Render("Press "),
			lipgloss.NewStyle().Background(m.ctx.Theme.SelectedBackground).Foreground(m.ctx.Theme.PrimaryText).Render("e"),
			lipgloss.NewStyle().Bold(true).Italic(true).Background(m.ctx.Theme.MainBackground).Render(" to read more..."),
		)
		bgStyle := lipgloss.NewStyle().Background(m.ctx.Theme.MainBackground)
		indentWidth := m.getIndentedContentWidth()
		padding := indentWidth - lipgloss.Width(moreLineRaw)
		if padding < 0 {
			padding = 0
		}
		leftPad := padding / 2
		rightPad := padding - leftPad
		moreLine := bgStyle.Render(strings.Repeat(" ", leftPad)) +
			moreLineRaw +
			bgStyle.Render(strings.Repeat(" ", rightPad))
		rendered = lipgloss.JoinVertical(lipgloss.Left,
			rendered,
			"",
			moreLine,
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title,
		lipgloss.NewStyle().
			Width(width).
			MaxWidth(width).
			Align(lipgloss.Left).
			Background(m.ctx.Theme.MainBackground).
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
}

func (m *Model) IsTextInputBoxFocused() bool {
	return m.isCommenting || m.isAssigning || m.isApproving || m.isUnassigning
}

func (m *Model) GetIsCommenting() bool {
	return m.isCommenting
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
	m.inputBox.UpdateProgramContext(ctx)
	isDarkBackground := tuitheme.HasDarkBackground(m.ctx.Theme)
	itemStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Background(m.ctx.Theme.MainBackground)
	if isDarkBackground {
		itemStyle = itemStyle.Foreground(m.ctx.Theme.FaintText)
	} else {
		itemStyle = itemStyle.Foreground(m.ctx.Theme.SecondaryText)
	}
	m.carousel.SetStyles(
		carousel.Styles{
			Item: itemStyle,
			Selected: lipgloss.NewStyle().
				Padding(0, 1).
				Bold(true).
				Foreground(m.ctx.Theme.PrimaryText).
				Background(m.ctx.Theme.SelectedBackground),
			OverflowIndicator: lipgloss.NewStyle().
				Foreground(m.ctx.Theme.FaintText).
				Background(m.ctx.Theme.MainBackground),
			Separator: lipgloss.NewStyle().
				Foreground(m.ctx.Theme.SecondaryBorder).
				Background(m.ctx.Theme.MainBackground),
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
	m.ShowConfirmCancel = false
	return true
}

func (m *Model) SetIsCommenting(isCommenting bool) tea.Cmd {
	if m.pr == nil {
		return nil
	}

	if !m.isCommenting && isCommenting {
		m.inputBox.Reset()
	}
	m.isCommenting = isCommenting
	m.inputBox.SetPrompt(commentPrompt)

	if isCommenting {
		return tea.Sequence(textarea.Blink, m.inputBox.Focus())
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
		m.inputBox.Reset()
	}
	m.isApproving = isApproving
	m.inputBox.SetPrompt(approvalPrompt)
	m.inputBox.SetValue(m.ctx.Config.Defaults.PrApproveComment)

	if isApproving {
		return tea.Sequence(textarea.Blink, m.inputBox.Focus())
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
	}
	m.isAssigning = isAssigning
	m.inputBox.SetPrompt("Assign users (whitespace-separated)...")
	if !m.userAssignedToPr(m.ctx.User) {
		m.inputBox.SetValue(m.ctx.User)
	}

	if isAssigning {
		return tea.Sequence(textarea.Blink, m.inputBox.Focus())
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
	m.inputBox.SetPrompt("Unassign users (whitespace-separated)...")
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
