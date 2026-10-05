package issueview

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

func TestRenderParent(t *testing.T) {
	ctx := newTestContext(t)
	m := NewModel(ctx)
	m.SetWidth(80)

	t.Run("returns empty string when issue has no parent", func(t *testing.T) {
		m.SetRow(&data.IssueData{
			Number: 1,
			Title:  "Top-level issue",
			Parent: nil,
		})
		require.Empty(t, m.renderParent())
	})

	t.Run("renders parent indicator with number and title when open", func(t *testing.T) {
		m.SetRow(&data.IssueData{
			Number: 2,
			Title:  "Child task",
			Parent: &data.ParentIssue{
				Number: 1,
				Title:  "Parent epic",
				State:  "OPEN",
			},
		})
		rendered := ansi.Strip(m.renderParent())
		require.NotEmpty(t, rendered)
		require.Contains(t, rendered, "Sub-issue of")
		require.Contains(t, rendered, "#1")
		require.Contains(t, rendered, "Parent epic")
	})

	t.Run("renders parent indicator when parent is closed", func(t *testing.T) {
		m.SetRow(&data.IssueData{
			Number: 3,
			Title:  "Another child",
			Parent: &data.ParentIssue{
				Number: 5,
				Title:  "Completed epic",
				State:  "CLOSED",
			},
		})
		rendered := ansi.Strip(m.renderParent())
		require.NotEmpty(t, rendered)
		require.Contains(t, rendered, "#5")
		require.Contains(t, rendered, "Completed epic")
	})

	t.Run("renders cross-repository parent with repository prefix", func(t *testing.T) {
		issue := &data.IssueData{
			Number: 4,
			Title:  "Cross repo child",
			Parent: &data.ParentIssue{
				Number: 99,
				Title:  "Cross repo parent",
				State:  "OPEN",
			},
		}
		issue.Repository.NameWithOwner = "owner/current-repo"
		issue.Parent.Repository.NameWithOwner = "owner/other-repo"

		m.SetRow(issue)
		rendered := ansi.Strip(m.renderParent())
		require.Contains(t, rendered, "owner/other-repo#99")
	})
}

func TestRenderSubIssues(t *testing.T) {
	ctx := newTestContext(t)
	m := NewModel(ctx)
	m.SetWidth(80)

	t.Run("returns empty string when issue has no sub-issues", func(t *testing.T) {
		m.SetRow(&data.IssueData{
			Number: 1,
			Title:  "Leaf issue",
		})
		require.Empty(t, m.renderSubIssues())
	})

	t.Run("renders sub-issues list with status and progress", func(t *testing.T) {
		m.SetRow(&data.IssueData{
			Number: 10,
			Title:  "Parent issue",
			SubIssues: data.SubIssues{
				TotalCount: 2,
				Nodes: []data.SubIssue{
					{
						Number: 11,
						Title:  "First sub-issue",
						State:  "CLOSED",
					},
					{
						Number: 12,
						Title:  "Second sub-issue",
						State:  "OPEN",
					},
				},
			},
			SubIssuesSummary: data.SubIssuesSummary{
				Total:            2,
				Completed:        1,
				PercentCompleted: 50,
			},
		})

		rendered := ansi.Strip(m.renderSubIssues())
		require.NotEmpty(t, rendered)
		require.Contains(t, rendered, "Sub-issues (1/2)")
		require.Contains(t, rendered, "#11")
		require.Contains(t, rendered, "First sub-issue")
		require.Contains(t, rendered, "#12")
		require.Contains(t, rendered, "Second sub-issue")
	})

	t.Run("renders summary fallback when nodes slice is empty", func(t *testing.T) {
		m.SetRow(&data.IssueData{
			Number: 20,
			Title:  "Parent without fetched nodes",
			SubIssuesSummary: data.SubIssuesSummary{
				Total:            5,
				Completed:        2,
				PercentCompleted: 40,
			},
		})

		rendered := ansi.Strip(m.renderSubIssues())
		require.NotEmpty(t, rendered)
		require.Contains(t, rendered, "Sub-issues (2/5)")
		require.Contains(t, rendered, "5 sub-issues (2 completed)")
	})

	t.Run("renders cross-repository sub-issue with repository prefix", func(t *testing.T) {
		issue := &data.IssueData{
			Number: 30,
			Title:  "Multi-repo parent",
			SubIssues: data.SubIssues{
				TotalCount: 1,
				Nodes: []data.SubIssue{
					{
						Number: 50,
						Title:  "External task",
						State:  "OPEN",
					},
				},
			},
		}
		issue.Repository.NameWithOwner = "owner/main-repo"
		issue.SubIssues.Nodes[0].Repository.NameWithOwner = "owner/worker-repo"

		m.SetRow(issue)
		rendered := ansi.Strip(m.renderSubIssues())
		require.Contains(t, rendered, "owner/worker-repo#50")
	})

	t.Run("View includes parent and sub-issues when present", func(t *testing.T) {
		m.SetRow(&data.IssueData{
			Number: 100,
			Title:  "Intermediate issue",
			Body:   "Issue description text",
			State:  "OPEN",
			Parent: &data.ParentIssue{
				Number: 90,
				Title:  "Root epic",
				State:  "OPEN",
			},
			SubIssues: data.SubIssues{
				TotalCount: 1,
				Nodes: []data.SubIssue{
					{
						Number: 101,
						Title:  "Leaf task",
						State:  "OPEN",
					},
				},
			},
		})

		viewOutput := ansi.Strip(m.View())
		require.True(t, strings.Contains(viewOutput, "Sub-issue of"))
		require.True(t, strings.Contains(viewOutput, "#90"))
		require.True(t, strings.Contains(viewOutput, "Root epic"))
		require.True(t, strings.Contains(viewOutput, "Sub-issues"))
		require.True(t, strings.Contains(viewOutput, "#101"))
		require.True(t, strings.Contains(viewOutput, "Leaf task"))
	})
}
