package prview

import (
	"fmt"

	"charm.land/lipgloss/v2"
	checks "github.com/dlvhdr/x/gh-checks"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
)

func (m *Model) renderStack() string {
	faint := m.ctx.Styles.Common.FaintTextStyle

	if !m.pr.Data.IsEnriched {
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.ctx.Styles.Common.WaitingGlyph,
			" ",
			faint.Render("Loading..."),
		)
	}

	entry := m.pr.Data.Enriched.StackEntry
	if !entry.IsStacked() {
		return faint.Render("This pull request isn't part of a stack.")
	}

	entries := entry.SortedEntries()
	title := fmt.Sprintf("%s  Stack of %d", constants.StackIcon, entry.Stack.Size)
	if len(entries) < entry.Stack.Size {
		title = fmt.Sprintf("%s (showing %d)", title, len(entries))
	}
	heading := m.ctx.Styles.Common.MainTextStyle.MarginBottom(1).Underline(true).Render(title)

	fainter := lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintBorder)
	res := heading

	current := 0
	if m.pr.Data.Primary != nil {
		current = m.pr.Data.Primary.Number
	}

	for i := len(entries) - 1; i >= 0; i-- {
		res = lipgloss.JoinVertical(
			lipgloss.Left,
			res,
			m.renderStackEntry(
				entries[i],
				entries[i].PullRequest.Number == current,
			),
		)
		if i > 0 {
			res = lipgloss.JoinVertical(lipgloss.Left, res, fainter.Render("  │"))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, res,
		fainter.Render("  │"),
		lipgloss.JoinHorizontal(lipgloss.Top,
			fainter.Render("  └ "),
			faint.Render(entry.Stack.BaseRefName),
		),
	)
}

func (m *Model) renderStackEntry(node data.StackEntryNode, isCurrent bool) string {
	main := m.ctx.Styles.Common.MainTextStyle
	faint := m.ctx.Styles.Common.FaintTextStyle
	fainter := lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintBorder)

	pr := node.PullRequest
	titleStyle := main
	if isCurrent {
		titleStyle = main.Bold(true)
	}

	marker := "  "
	if isCurrent {
		marker = m.ctx.Styles.Common.MainTextStyle.Render(constants.SelectionIcon) + " "
	}

	left := lipgloss.JoinHorizontal(lipgloss.Top,
		marker,
		m.stackStateGlyph(pr),
		" ",
		faint.Render(fmt.Sprintf("#%d", pr.Number)),
		" ",
		titleStyle.Render(pr.Title),
	)

	title := m.renderDottedRow(left, faint.Render(pr.HeadRefName))

	desc := lipgloss.JoinHorizontal(lipgloss.Top,
		fainter.Render("  │ "),
		faint.Render(fmt.Sprintf("@%s", pr.Author.Login)),
		faint.Render(" · "),
		m.stackReviewStatus(pr),
	)
	if ci := m.stackCiStatus(pr); ci != "" {
		desc = lipgloss.JoinHorizontal(lipgloss.Top, desc, faint.Render(" · "), ci)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, desc)
}

func (m *Model) stackStateGlyph(pr data.StackPullRequest) string {
	return components.RenderPRStateGlyph(m.ctx, pr.State, pr.IsDraft, pr.IsInMergeQueue)
}

func (m *Model) stackReviewStatus(pr data.StackPullRequest) string {
	faint := m.ctx.Styles.Common.FaintTextStyle

	switch pr.ReviewDecision {
	case "APPROVED":
		return lipgloss.NewStyle().Foreground(m.ctx.Theme.SuccessText).Render("Approved")
	case "CHANGES_REQUESTED":
		return lipgloss.NewStyle().Foreground(m.ctx.Theme.ErrorText).Render("Changes requested")
	case "REVIEW_REQUIRED":
		return faint.Render("Review required")
	default:
		return faint.Render("No review")
	}
}

func (m *Model) stackCiStatus(pr data.StackPullRequest) string {
	nodes := pr.Commits.Nodes
	if len(nodes) == 0 {
		return ""
	}

	return m.commitStateSign(checks.CommitState(nodes[0].Commit.StatusCheckRollup.State))
}
