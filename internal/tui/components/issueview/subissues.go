package issueview

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
	"github.com/charmbracelet/x/ansi"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

func (m *Model) renderParent() string {
	if m.issue == nil || !m.issue.Data.HasParent() {
		return ""
	}

	parent := m.issue.Data.Parent
	stateIcon := ""
	var stateColor compat.AdaptiveColor
	if parent.State == "OPEN" {
		stateColor = m.ctx.Styles.Colors.OpenIssue
	} else {
		stateColor = m.ctx.Styles.Colors.ClosedIssue
		stateIcon = ""
	}

	prefix := lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render("Sub-issue of ")
	icon := lipgloss.NewStyle().Foreground(stateColor).Render(stateIcon)

	parentRef := fmt.Sprintf("#%d", parent.Number)
	if parent.Repository.NameWithOwner != "" &&
		parent.Repository.NameWithOwner != m.issue.Data.GetRepoNameWithOwner() {
		parentRef = fmt.Sprintf("%s#%d", parent.Repository.NameWithOwner, parent.Number)
	}
	refStyle := lipgloss.NewStyle().
		Foreground(m.ctx.Theme.SecondaryText).
		Bold(true).
		Render(parentRef)

	titleStyle := lipgloss.NewStyle().Foreground(m.ctx.Theme.PrimaryText)
	renderedTitle := ""
	if parent.Title != "" {
		renderedTitle = " · " + titleStyle.Render(parent.Title)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, prefix, icon, " ", refStyle, renderedTitle)
}

func (m *Model) renderSubIssues() string {
	if m.issue == nil || !m.issue.Data.HasSubIssues() {
		return ""
	}

	title := m.renderSubIssuesTitle()
	body := m.renderSubIssuesList()

	bodyStyle := lipgloss.NewStyle().PaddingLeft(2)
	return lipgloss.JoinVertical(lipgloss.Left, title, bodyStyle.Render(body))
}

func (m *Model) renderSubIssuesTitle() string {
	completed, total, _ := m.issue.Data.GetSubIssuesProgress()
	titleText := fmt.Sprintf(" Sub-issues (%d/%d)", completed, total)
	return m.ctx.Styles.Common.MainTextStyle.
		MarginBottom(1).
		Underline(true).
		Render(titleText)
}

func (m *Model) renderSubIssuesList() string {
	nodes := m.issue.Data.SubIssues.Nodes
	if len(nodes) == 0 {
		completed, total, _ := m.issue.Data.GetSubIssuesProgress()
		return lipgloss.NewStyle().
			Italic(true).
			Foreground(m.ctx.Theme.FaintText).
			Render(fmt.Sprintf("%d sub-issues (%d completed)", total, completed))
	}

	width := m.getIndentedContentWidth() - 4
	var items []string
	for _, sub := range nodes {
		items = append(items, m.renderSubIssueItem(sub, width))
	}

	return lipgloss.JoinVertical(lipgloss.Left, items...)
}

func (m *Model) renderSubIssueItem(sub data.SubIssue, width int) string {
	stateIcon := ""
	var stateColor compat.AdaptiveColor
	if sub.State == "OPEN" {
		stateColor = m.ctx.Styles.Colors.OpenIssue
	} else {
		stateColor = m.ctx.Styles.Colors.ClosedIssue
		stateIcon = ""
	}

	icon := lipgloss.NewStyle().Foreground(stateColor).Render(stateIcon)

	refText := fmt.Sprintf("#%d", sub.Number)
	if sub.Repository.NameWithOwner != "" &&
		sub.Repository.NameWithOwner != m.issue.Data.GetRepoNameWithOwner() {
		refText = fmt.Sprintf("%s#%d", sub.Repository.NameWithOwner, sub.Number)
	}
	ref := lipgloss.NewStyle().Foreground(m.ctx.Theme.SecondaryText).Render(refText)

	title := lipgloss.NewStyle().Foreground(m.ctx.Theme.PrimaryText).Render(sub.Title)

	line := lipgloss.JoinHorizontal(lipgloss.Top, icon, " ", ref, " ", title)
	if width > 0 && lipgloss.Width(line) > width {
		line = ansi.Truncate(line, width, "…")
	}
	return line
}
