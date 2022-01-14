package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/ui/markdown"
)

type Sidebar struct {
	model Model
	pr    PullRequest
}

func (m *Model) renderSidebar() string {
	if !m.isSidebarOpen {
		return ""
	}

	height := m.sidebarViewport.Height + pagerHeight
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

	return style.Copy().Render(lipgloss.JoinVertical(
		lipgloss.Left,
		m.sidebarViewport.View(),
		pagerStyle.Copy().Render(fmt.Sprintf("%d%%", int(m.sidebarViewport.ScrollPercent()*100))),
	))
}

func (m *Model) setSidebarViewportContent() {
	pr := m.getCurrPr()
	if pr == nil {
		return
	}

	sidebar := Sidebar{
		model: *m,
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
	s.WriteString(sidebar.renderActivity())

	m.sidebarViewport.SetContent(s.String())
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

	body = strings.TrimSpace(body)
	if body == "" {
		return lipgloss.NewStyle().Italic(true).Render("No description provided.")
	}

	markdownRenderer := markdown.GetMarkdownRenderer(width)
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

func (m *Model) getSidebarWidth() int {
	return m.config.Defaults.Preview.Width
}
