package prview

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
	checks "github.com/dlvhdr/x/gh-checks"
)

func (m *Model) renderCommits() string {
	main := m.ctx.Styles.Common.MainTextStyle
	faint := m.ctx.Styles.Common.FaintTextStyle
	fainter := lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintBorder)

	if !m.pr.Data.IsEnriched {
		return lipgloss.JoinHorizontal(lipgloss.Top, m.ctx.Styles.Common.WaitingGlyph, " ", faint.Render("Loading..."))
	}

	commits := m.pr.Data.Enriched.AllCommits.Nodes
	heading := m.ctx.Styles.Common.MainTextStyle.MarginBottom(1).Underline(true).Render(
		fmt.Sprintf("%s  %d commits", constants.CommitIcon, len(commits)))

	rendered := make([]string, len(commits))
	for i, commit := range commits {
		commit := commit.Commit
		name := commit.Author.User.Login
		if name == "" {
			name = commit.Author.Name
		}
		left := fmt.Sprintf(
			"%s %s",
			faint.Render(constants.VerticalCommitIcon),
			main.Render(commit.MessageHeadline),
		)
		right := faint.Render(commit.AbbreviatedOid)
		wright := lipgloss.Width(right)
		left = ansi.Truncate(left, max(0, m.getIndentedContentWidth()-wright-1), constants.Ellipsis)
		pad := fainter.Render(" " + strings.Repeat(constants.HorizontalLineIcon,
			max(1, m.getIndentedContentWidth()-lipgloss.Width(left)-wright)-1) + " ")

		title := lipgloss.JoinHorizontal(lipgloss.Top, left, pad, right)

		statsStr := ""
		if commit.StatusCheckRollup.Contexts.TotalCount > 0 {
			stats := m.getStatusCheckRollupStats(commit.StatusCheckRollup)
			statsStr = lipgloss.JoinHorizontal(lipgloss.Top,
				" ",
				faint.Render(constants.SmallDotIcon),
				" ",
				m.commitStateSign(commit.StatusCheckRollup.State),
				" ",
				faint.Render(fmt.Sprintf("%d/%d", stats.succeeded,
					commit.StatusCheckRollup.Contexts.TotalCount)),
			)
		}

		desc := lipgloss.JoinHorizontal(lipgloss.Top,
			fainter.Render("│ "),
			faint.Render(fmt.Sprintf("@%s", name)),
			faint.Render(" committed "),
			faint.Render(utils.TimeElapsed(commit.CommittedDate)),
			faint.Render(" ago"),
			statsStr,
		)
		rendered[i] = lipgloss.JoinVertical(lipgloss.Left, title, desc)
	}

	res := heading
	for i, r := range rendered {
		res = lipgloss.JoinVertical(lipgloss.Left, res, r)
		if i < len(rendered)-1 {
			res = lipgloss.JoinVertical(lipgloss.Left, res, fainter.Render("│"))
		}
	}

	return res
}

func (m *Model) commitStateSign(state checks.CommitState) string {
	switch state {
	case checks.CommitStateError, checks.CommitStateFailure:
		return m.ctx.Styles.Common.FailureGlyph
	case checks.CommitStatePending, checks.CommitStateExpected, checks.CommitStateUnknown:
		return m.ctx.Styles.Common.WaitingGlyph
	case checks.CommitStateSuccess:
		return m.ctx.Styles.Common.SuccessGlyph
	}

	return ""
}
