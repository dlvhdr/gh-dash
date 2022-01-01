package ui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/data"
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
		MaxHeight(height)
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
	s.WriteString("\n\n")
	s.WriteString(sidebar.renderPills())
	s.WriteString("\n")
	s.WriteString(sidebar.renderDescription())
	s.WriteString("\n\n")
	s.WriteString(sidebar.renderChecks())

	return style.Copy().Render(s.String())
}

func (sidebar *Sidebar) renderTitle() string {
	return mainTextStyle.Copy().Width(sideBarWidth - 6).
		Render(sidebar.pr.Data.Title)
}

func (sidebar *Sidebar) renderStatusPill() string {
	color := ""
	switch sidebar.pr.Data.State {
	case "OPEN":
		color = subtleIndigo.Dark
	case "CLOSED":
		color = faintText.Dark
	case "MERGED":
		color = "#123123"
	}

	return pillStyle.
		Background(lipgloss.Color(color)).
		Render(sidebar.pr.renderState())
}

func (sidebar *Sidebar) renderMergeablePill() string {
	if sidebar.pr.Data.Mergeable == "MERGEABLE" {
		return ""
	}

	return pillStyle.
		Background(subtleIndigo).
		Render("Mergeable")
}

func (sidebar *Sidebar) renderPills() string {
	statusPill := sidebar.renderStatusPill()
	mergeablePill := sidebar.renderMergeablePill()
	return lipgloss.JoinHorizontal(lipgloss.Left, statusPill, " ", mergeablePill)
}

func (sidebar Sidebar) renderDescription() string {
	// if sidebar.pr.Data.Body == "" {
	// 	return lipgloss.NewStyle().Italic(true).MarginTop(1).Render("No description provided.")
	// }

	width := sideBarWidth - 6
	regex := regexp.MustCompile("(?U)<!--(.|[[:space:]])*-->")
	body := regex.ReplaceAllString(sidebar.pr.Data.Body, "")

	regex = regexp.MustCompile("\n\n")
	body = regex.ReplaceAllString(body, "\n")
	body = strings.TrimSpace(body)

	if body == "" {
		return lipgloss.NewStyle().Italic(true).MarginTop(1).Render("No description provided.")
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
			status := string(checkRun.Status)
			renderedStatus := renderCheckRunConclusion(&status, string(checkRun.Conclusion))
			name := renderCheckRunName(checkRun)
			checks = append(checks, lipgloss.JoinHorizontal(lipgloss.Left, renderedStatus, " ", name))
		} else if node.Typename == "StatusContext" {
			statusContext := node.StatusContext
			status := renderCheckRunConclusion(nil, string(statusContext.State))
			checks = append(checks, lipgloss.JoinHorizontal(lipgloss.Left, status, " ", string(statusContext.Context)))
		}
	}

	if len(checks) == 0 {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			lipgloss.NewStyle().
				Italic(true).
				Render("No checks to display..."),
		)
	}

	renderedChecks := lipgloss.JoinVertical(lipgloss.Left, checks...)
	return lipgloss.JoinVertical(lipgloss.Left, title, renderedChecks)
}

func renderCheckRunName(checkRun data.CheckRun) string {
	workflow := string(checkRun.CheckSuite.WorkflowRun.Workflow.Name)
	name := string(checkRun.Name)
	if workflow == "" {
		return name
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, workflow, "/", name)
}

func renderCheckRunConclusion(status *string, conclusion string) string {
	conclusionStr := string(conclusion)
	if status != nil && data.IsStatusWaiting(*status) {
		return lipgloss.NewStyle().Foreground(faintBorder).Render("")
	}

	if conclusionStr == "FAILURE" || conclusionStr == "TIMED_OUT" || conclusionStr == "STARTUP_FAILURE" {
		return lipgloss.NewStyle().Foreground(warningText).Render("")
	}

	return lipgloss.NewStyle().Foreground(successText).Render("")
}
