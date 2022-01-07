package ui

import (
	"regexp"
	"sort"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/utils"
)

type Sidebar struct {
	model Model
	pr    PullRequest
}

func (m Model) renderSidebar() string {
	if !m.isSidebarOpen {
		return ""
	}

	height := m.viewport.Height + mainContentPadding*2
	style := sideBarStyle.Copy().
		Height(height).
		MaxHeight(height).
		Width(m.getSidebarWidth()).
		MaxWidth(m.getSidebarWidth())

	pr := m.getCurrPr()
	if pr == nil {
		return style.Copy().Align(lipgloss.Center).Render(
			lipgloss.PlaceVertical(height, lipgloss.Center, "Select a Pull Request..."),
		)
	}

	sidebar := Sidebar{
		model: m,
		pr:    *pr,
	}

	s := strings.Builder{}
	s.WriteString(sidebar.renderTitle())
	s.WriteString("\n")
	s.WriteString(sidebar.renderBranches())
	s.WriteString("\n\n")
	s.WriteString(sidebar.renderPills())
	s.WriteString("\n\n")
	s.WriteString(sidebar.renderDescription())
	s.WriteString("\n\n")
	s.WriteString(sidebar.renderChecks())
	s.WriteString("\n\n")
	s.WriteString(sidebar.renderComments())

	return style.Copy().Render(s.String())
}

func (sidebar *Sidebar) renderTitle() string {
	return mainTextStyle.Copy().Width(sidebar.model.getSidebarWidth() - 6).
		Render(sidebar.pr.Data.Title)
}

func (sidebar *Sidebar) renderBranches() string {
	return lipgloss.NewStyle().
		Foreground(secondaryText).
		Render(sidebar.pr.Data.BaseRefName + "  " + sidebar.pr.Data.HeadRefName)
}

func (sidebar *Sidebar) renderStatusPill() string {
	bgColor := ""
	switch sidebar.pr.Data.State {
	case "OPEN":
		bgColor = openPR.Dark
	case "CLOSED":
		bgColor = closedPR.Dark
	case "MERGED":
		bgColor = mergedPR.Dark
	}

	return pillStyle.
		Background(lipgloss.Color(bgColor)).
		Render(sidebar.pr.renderState())
}

func (sidebar *Sidebar) renderMergeablePill() string {
	status := sidebar.pr.Data.Mergeable
	if status == "CONFLICTING" {
		return pillStyle.Copy().
			Background(warningText).
			Render(" Merge Conflicts")
	} else if status == "MERGEABLE" {
		return pillStyle.Copy().
			Background(successText).
			Render(" Mergeable")
	}

	return ""
}

func (sidebar *Sidebar) renderChecksPill() string {
	status := sidebar.pr.getStatusChecksRollup()
	if status == "FAILURE" {
		return pillStyle.Copy().
			Background(warningText).
			Render(" Checks")
	} else if status == "PENDING" {
		return pillStyle.Copy().
			Background(faintText).
			Render(waitingGlyph + " Checks")
	}

	return pillStyle.Copy().
		Background(successText).
		Foreground(subtleIndigo).
		Render(" Checks")
}

func (sidebar *Sidebar) renderPills() string {
	statusPill := sidebar.renderStatusPill()
	mergeablePill := sidebar.renderMergeablePill()
	checksPill := sidebar.renderChecksPill()
	return lipgloss.JoinHorizontal(lipgloss.Left, statusPill, " ", mergeablePill, " ", checksPill)
}

func (sidebar Sidebar) renderDescription() string {
	width := sidebar.model.getSidebarWidth() - 6
	regex := regexp.MustCompile("(?U)<!--(.|[[:space:]])*-->")
	body := regex.ReplaceAllString(sidebar.pr.Data.Body, "")

	regex = regexp.MustCompile("\n\n")
	body = regex.ReplaceAllString(body, "\n")
	body = strings.TrimSpace(body)

	if body == "" {
		return lipgloss.NewStyle().Italic(true).Render("No description provided.")
	}

	// TODO: create style JSON file and load it somewhere once
	style := glamour.DefaultStyles["dark"]
	indentToken := ""
	indent := uint(0)
	style.Document.Indent = &indent
	style.Document.IndentToken = &indentToken
	style.Document.Margin = &indent
	markdownRenderer, _ := glamour.NewTermRenderer(
		glamour.WithStyles(*style),
		glamour.WithWordWrap(width),
	)
	rendered, err := markdownRenderer.Render(body)
	if err != nil {
		return ""
	}

	return lipgloss.NewStyle().
		MaxHeight(10).
		Width(width).
		MaxWidth(width).
		Align(lipgloss.Left).
		Render(rendered)
}

func (sidebar Sidebar) renderChecks() string {
	title := mainTextStyle.Copy().MarginBottom(1).Render("ﱔ Checks")

	commits := sidebar.pr.Data.Commits.Nodes
	if len(commits) == 0 {
		return ""
	}

	lastCommit := commits[0]
	var checks []string
	for _, node := range lastCommit.Commit.StatusCheckRollup.Contexts.Nodes {
		if node.Typename == "CheckRun" {
			checkRun := node.CheckRun
			renderedStatus := renderCheckRunConclusion(checkRun)
			name := renderCheckRunName(checkRun)
			checks = append(checks, lipgloss.JoinHorizontal(lipgloss.Left, renderedStatus, " ", name))
		} else if node.Typename == "StatusContext" {
			statusContext := node.StatusContext
			status := renderStatusContextConclusion(statusContext)
			checks = append(checks, lipgloss.JoinHorizontal(lipgloss.Left, status, " ", renderStatusContextName(statusContext)))
		}
	}

	if len(checks) == 0 {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			lipgloss.NewStyle().
				Italic(true).
				PaddingLeft(2).
				Width(sidebar.model.getSidebarWidth()-6).
				Render("No checks to display..."),
		)
	}

	renderedChecks := lipgloss.JoinVertical(lipgloss.Left, checks...)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		lipgloss.NewStyle().PaddingLeft(2).Width(sidebar.model.getSidebarWidth()-6).Render(renderedChecks),
	)
}

func (sidebar Sidebar) renderComments() string {
	width := sidebar.model.getSidebarWidth() - 8
	markdownStyle := glamour.DefaultStyles["dark"]
	indentToken := ""
	indent := uint(0)
	markdownStyle.Document.BlockPrefix = indentToken
	markdownStyle.Document.Prefix = indentToken
	markdownStyle.Document.BlockSuffix = indentToken
	markdownStyle.Document.Indent = &indent
	markdownStyle.Document.IndentToken = &indentToken
	markdownStyle.Document.Margin = &indent
	markdownStyle.Paragraph.Indent = &indent
	markdownStyle.Paragraph.IndentToken = &indentToken
	markdownStyle.Paragraph.Margin = &indent
	markdownStyle.Paragraph.Prefix = indentToken
	markdownStyle.Paragraph.BlockPrefix = indentToken
	markdownStyle.Paragraph.BlockSuffix = indentToken
	markdownRenderer, _ := glamour.NewTermRenderer(
		glamour.WithStyles(*markdownStyle),
		glamour.WithWordWrap(width),
	)
	commentNodes := sidebar.pr.Data.Comments.Nodes
	sort.Slice(commentNodes, func(i, j int) bool {
		return commentNodes[i].UpdatedAt.After(commentNodes[j].UpdatedAt)
	})
	var renderedComments []string
	for _, comment := range commentNodes {
		header := lipgloss.JoinHorizontal(lipgloss.Left,
			mainTextStyle.Copy().Render(comment.Author.Login),
			" ",
			lipgloss.NewStyle().Foreground(faintText).Render(utils.TimeElapsed(comment.UpdatedAt)),
		)
		body, err := markdownRenderer.Render(comment.Body)
		if err != nil {
			continue
		}
		renderedComments = append(renderedComments, lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			body,
		))
	}
	if len(renderedComments) == 0 {
		renderedComments = append(
			renderedComments,
			lipgloss.NewStyle().Italic(true).Render("No comments..."),
		)
	}

	title := mainTextStyle.Copy().MarginBottom(1).Render(" Comments")
	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		lipgloss.NewStyle().PaddingLeft(2).Render(
			lipgloss.JoinVertical(lipgloss.Left, renderedComments...),
		),
	)
}

func (m Model) getSidebarWidth() int {
	return m.config.Defaults.Preview.Width
}
