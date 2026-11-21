package issueview

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
)

func newPillTestModel(t *testing.T, issue data.IssueData) Model {
	t.Helper()
	m := NewModel(newTestContext(t))
	m.SetWidth(80)
	m.SetRow(&issue)
	return m
}

func TestStatusRowPillsAreBlankWhenThereIsNothingToShow(t *testing.T) {
	m := newPillTestModel(t, data.IssueData{State: "OPEN"})

	require.Empty(t, m.renderTypePill())
	require.Empty(t, m.renderLinkedPrsPill())
}

func TestRenderTypePill(t *testing.T) {
	m := newPillTestModel(t, data.IssueData{
		State:     "OPEN",
		IssueType: &data.IssueType{Name: "Bug", Color: "RED"},
	})
	require.Contains(t, ansi.Strip(m.renderTypePill()), "Bug")
}

func TestRenderLinkedPrsPill(t *testing.T) {
	tests := []struct {
		name     string
		prs      []data.LinkedPullRequest
		expected string
	}{
		{name: "one PR", prs: []data.LinkedPullRequest{{State: "OPEN"}}, expected: " 1 PR"},
		{
			name:     "several PRs",
			prs:      []data.LinkedPullRequest{{State: "MERGED"}, {State: "OPEN", IsDraft: true}},
			expected: " 2 PRs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newPillTestModel(t, data.IssueData{
				State:     "OPEN",
				LinkedPrs: data.LinkedPullRequests{TotalCount: len(tt.prs), Nodes: tt.prs},
			})
			require.Contains(t, ansi.Strip(m.renderLinkedPrsPill()), constants.OpenIcon+tt.expected)
		})
	}
}

func TestViewPutsPillsOnTheStatusRow(t *testing.T) {
	m := newPillTestModel(t, data.IssueData{
		Title:     "Example issue",
		State:     "OPEN",
		IssueType: &data.IssueType{Name: "Bug", Color: "RED"},
		LinkedPrs: data.LinkedPullRequests{
			TotalCount: 1,
			Nodes:      []data.LinkedPullRequest{{State: "OPEN"}},
		},
	})

	var statusRow string
	for _, line := range splitLines(ansi.Strip(m.View())) {
		if containsAll(line, "Open", "Bug", "1 PR") {
			statusRow = line
		}
	}
	require.NotEmpty(t, statusRow, "no line holds the state, type, and linked-PR pills together")
}

func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}
