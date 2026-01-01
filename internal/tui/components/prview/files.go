package prview

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

func (m *Model) renderChangesOverview() string {
	changes := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(m.ctx.Theme.FaintBorder).
		Width(m.getIndentedContentWidth()).
		Padding(1).
		Background(m.ctx.Theme.MainBackground)

	commits := lipgloss.NewStyle().
		Width(m.getIndentedContentWidth()).
		Padding(1).
		Background(m.ctx.Theme.MainBackground)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(m.ctx.Theme.FaintBorder).
		Width(m.getIndentedContentWidth()).
		Background(m.ctx.Theme.MainBackground)

	textStyle := lipgloss.NewStyle().Background(m.ctx.Theme.MainBackground)
	faintStyle := lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Background(m.ctx.Theme.MainBackground)
	time := utils.TimeElapsed(m.pr.Data.Primary.UpdatedAt)
	changesLine := lipgloss.JoinHorizontal(lipgloss.Top,
		faintStyle.Render(" "),
		textStyle.Render(fmt.Sprintf("%d files changed", m.pr.Data.Primary.Files.TotalCount)),
		textStyle.Render(" "),
		m.pr.RenderLinesWithBackground(m.ctx.Theme.MainBackground),
	)
	changesLine = stripANSIReset(changesLine)
	commitsLine := lipgloss.JoinHorizontal(lipgloss.Top,
		faintStyle.Render(" "),
		textStyle.Render(fmt.Sprintf("%d commits", m.pr.Data.Primary.Commits.TotalCount)),
		textStyle.Render(" "),
		faintStyle.Render(fmt.Sprintf("%s ago", time)),
	)
	commitsLine = stripANSIReset(commitsLine)
	return box.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			changes.Render(
				changesLine,
			),
			commits.Render(
				commitsLine,
			),
		),
	)
}

func (m *Model) renderChangedFiles() string {
	files := make([]string, 0)
	for _, file := range m.pr.Data.Primary.Files.Nodes {
		files = append(files, m.renderFile(file))
	}

	return lipgloss.JoinVertical(lipgloss.Left, files...)
}

func (m *Model) renderFile(file data.ChangedFile) string {
	bgStyle := lipgloss.NewStyle().Background(m.ctx.Theme.MainBackground)
	icon := m.renderChangeTypeIcon(file.ChangeType)
	additions := lipgloss.NewStyle().Foreground(m.ctx.Theme.SuccessText).Background(m.ctx.Theme.MainBackground).Width(5).Render(fmt.Sprintf("+%d", file.Additions))
	deletions := lipgloss.NewStyle().Foreground(m.ctx.Theme.ErrorText).Background(m.ctx.Theme.MainBackground).Width(5).Render(fmt.Sprintf("-%d", file.Deletions))
	prefix := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.JoinHorizontal(lipgloss.Top, additions, deletions),
		bgStyle.Render(" "),
		icon,
		bgStyle.Render(" "))

	path := file.Path
	remaining := m.getIndentedContentWidth() - lipgloss.Width(prefix)
	if len(path) > remaining {
		path = lipgloss.JoinVertical(lipgloss.Left, path[0:remaining], bgStyle.Render(" ")+path[remaining:])
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		prefix,
		bgStyle.Render(path),
	)
}

func (m *Model) renderChangeTypeIcon(changeType string) string {
	switch changeType {
	case "ADDED":
		return lipgloss.NewStyle().Foreground(m.ctx.Theme.SuccessText).Background(m.ctx.Theme.MainBackground).Render("")
	case "DELETED":
		return lipgloss.NewStyle().Foreground(m.ctx.Theme.ErrorText).Background(m.ctx.Theme.MainBackground).Render("")
	case "RENAMED":
		return lipgloss.NewStyle().Foreground(m.ctx.Theme.WarningText).Background(m.ctx.Theme.MainBackground).Render("")
	case "COPIED":
		return lipgloss.NewStyle().Foreground(m.ctx.Theme.WarningText).Background(m.ctx.Theme.MainBackground).Render("")
	case "MODIFIED":
		return lipgloss.NewStyle().Foreground(m.ctx.Theme.WarningText).Background(m.ctx.Theme.MainBackground).Render("")
	case "CHANGED":
		return lipgloss.NewStyle().Foreground(m.ctx.Theme.WarningText).Background(m.ctx.Theme.MainBackground).Render("")
	default:
		return ""
	}
}
