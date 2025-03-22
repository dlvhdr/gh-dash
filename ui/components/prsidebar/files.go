package prsidebar

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/data"
)

func (m *Model) renderChangedFiles() string {
	files := make([]string, 0)
	for _, file := range m.pr.Data.Files.Nodes {
		files = append(files, m.renderFile(file))
	}

	return lipgloss.JoinVertical(lipgloss.Left, files...)
}

func (m *Model) renderFile(file data.ChangedFile) string {
	icon := m.renderChangeTypeIcon(file.ChangeType)
	additions := lipgloss.NewStyle().Foreground(m.ctx.Theme.SuccessText).Width(5).Render(fmt.Sprintf("+%d", file.Additions))
	deletions := lipgloss.NewStyle().Foreground(m.ctx.Theme.ErrorText).Width(5).Render(fmt.Sprintf("-%d", file.Deletions))
	return lipgloss.NewStyle().MaxWidth(m.width).Render(lipgloss.JoinHorizontal(
		lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Center, additions, deletions),
		" ",
		icon,
		" ",
		file.Path,
	))
}

func (m *Model) renderChangeTypeIcon(changeType string) string {
	switch changeType {
	case "ADDED":
		return lipgloss.NewStyle().Foreground(m.ctx.Theme.SuccessText).Render("")
	case "DELETED":
		return lipgloss.NewStyle().Foreground(m.ctx.Theme.ErrorText).Render("")
	case "RENAMED":
		return lipgloss.NewStyle().Foreground(m.ctx.Theme.WarningText).Render("")
	case "COPIED":
		return lipgloss.NewStyle().Foreground(m.ctx.Theme.WarningText).Render("")
	case "MODIFIED":
		return lipgloss.NewStyle().Foreground(m.ctx.Theme.WarningText).Render("")
	case "CHANGED":
		return lipgloss.NewStyle().Foreground(m.ctx.Theme.WarningText).Render("")
	default:
		return ""
	}
}
