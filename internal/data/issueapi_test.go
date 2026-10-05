package data

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssueData_HasParent(t *testing.T) {
	t.Run("returns false when Parent is nil", func(t *testing.T) {
		issue := IssueData{
			Number: 1,
			Title:  "Top-level issue",
			Parent: nil,
		}
		require.False(t, issue.HasParent())
	})

	t.Run("returns true when Parent is present", func(t *testing.T) {
		issue := IssueData{
			Number: 2,
			Title:  "Child issue",
			Parent: &ParentIssue{
				Number: 1,
				Title:  "Parent issue",
				State:  "OPEN",
			},
		}
		require.True(t, issue.HasParent())
	})
}

func TestIssueData_HasSubIssues(t *testing.T) {
	t.Run("returns false when no sub-issues exist", func(t *testing.T) {
		issue := IssueData{Number: 1}
		require.False(t, issue.HasSubIssues())
	})

	t.Run("returns true when SubIssues TotalCount is greater than zero", func(t *testing.T) {
		issue := IssueData{
			Number: 1,
			SubIssues: SubIssues{
				TotalCount: 3,
			},
		}
		require.True(t, issue.HasSubIssues())
	})

	t.Run("returns true when SubIssues Nodes is non-empty", func(t *testing.T) {
		issue := IssueData{
			Number: 1,
			SubIssues: SubIssues{
				Nodes: []SubIssue{
					{Number: 2, Title: "Sub task", State: "OPEN"},
				},
			},
		}
		require.True(t, issue.HasSubIssues())
	})

	t.Run("returns true when SubIssuesSummary has positive total", func(t *testing.T) {
		issue := IssueData{
			Number: 1,
			SubIssuesSummary: SubIssuesSummary{
				Total: 2,
			},
		}
		require.True(t, issue.HasSubIssues())
	})
}

func TestIssueData_GetSubIssuesProgress(t *testing.T) {
	t.Run("uses SubIssuesSummary when total is greater than zero", func(t *testing.T) {
		issue := IssueData{
			Number: 1,
			SubIssuesSummary: SubIssuesSummary{
				Total:            10,
				Completed:        4,
				PercentCompleted: 40,
			},
		}
		completed, total, percent := issue.GetSubIssuesProgress()
		require.Equal(t, 4, completed)
		require.Equal(t, 10, total)
		require.Equal(t, 40, percent)
	})

	t.Run("calculates from SubIssues nodes when summary is empty", func(t *testing.T) {
		issue := IssueData{
			Number: 1,
			SubIssues: SubIssues{
				TotalCount: 3,
				Nodes: []SubIssue{
					{Number: 10, State: "CLOSED"},
					{Number: 11, State: "OPEN"},
					{Number: 12, State: "CLOSED"},
				},
			},
		}
		completed, total, percent := issue.GetSubIssuesProgress()
		require.Equal(t, 2, completed)
		require.Equal(t, 3, total)
		require.Equal(t, 66, percent)
	})

	t.Run("handles zero sub-issues gracefully", func(t *testing.T) {
		issue := IssueData{Number: 1}
		completed, total, percent := issue.GetSubIssuesProgress()
		require.Equal(t, 0, completed)
		require.Equal(t, 0, total)
		require.Equal(t, 0, percent)
	})
}

func TestIssueData_UnmarshalJSON_SubIssuesAndParent(t *testing.T) {
	rawJSON := `{
		"number": 42,
		"title": "Main task",
		"state": "OPEN",
		"parent": {
			"number": 10,
			"title": "Epic initiative",
			"state": "OPEN",
			"url": "https://github.com/org/repo/issues/10",
			"repository": {
				"name": "repo",
				"nameWithOwner": "org/repo"
			}
		},
		"subIssues": {
			"totalCount": 2,
			"nodes": [
				{
					"number": 43,
					"title": "Child A",
					"state": "CLOSED",
					"url": "https://github.com/org/repo/issues/43",
					"repository": {
						"name": "repo",
						"nameWithOwner": "org/repo"
					}
				},
				{
					"number": 44,
					"title": "Child B",
					"state": "OPEN",
					"url": "https://github.com/org/repo/issues/44",
					"repository": {
						"name": "repo",
						"nameWithOwner": "org/repo"
					}
				}
			]
		},
		"subIssuesSummary": {
			"total": 2,
			"completed": 1,
			"percentCompleted": 50
		}
	}`

	var issue IssueData
	err := json.Unmarshal([]byte(rawJSON), &issue)
	require.NoError(t, err)

	require.Equal(t, 42, issue.Number)
	require.Equal(t, "Main task", issue.Title)

	require.NotNil(t, issue.Parent)
	require.Equal(t, 10, issue.Parent.Number)
	require.Equal(t, "Epic initiative", issue.Parent.Title)
	require.Equal(t, "org/repo", issue.Parent.Repository.NameWithOwner)

	require.Equal(t, 2, issue.SubIssues.TotalCount)
	require.Len(t, issue.SubIssues.Nodes, 2)
	require.Equal(t, 43, issue.SubIssues.Nodes[0].Number)
	require.Equal(t, "CLOSED", issue.SubIssues.Nodes[0].State)
	require.Equal(t, 44, issue.SubIssues.Nodes[1].Number)
	require.Equal(t, "OPEN", issue.SubIssues.Nodes[1].State)

	require.Equal(t, 2, issue.SubIssuesSummary.Total)
	require.Equal(t, 1, issue.SubIssuesSummary.Completed)
	require.Equal(t, 50, issue.SubIssuesSummary.PercentCompleted)
}
