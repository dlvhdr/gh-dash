package prview

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
)

func unresolvedThreads(resolved ...bool) data.ReviewThreads {
	threads := data.ReviewThreads{TotalCount: len(resolved)}
	for _, r := range resolved {
		threads.Nodes = append(threads.Nodes, struct{ IsResolved bool }{IsResolved: r})
	}
	return threads
}

func TestStatusRowPillsAreBlankWhenThereIsNothingToShow(t *testing.T) {
	m := newTestModel(t, &data.PullRequestData{State: "OPEN", Mergeable: "MERGEABLE"}, nil, nil)

	require.Empty(t, m.renderLinkedIssuesPill())
	require.Empty(t, m.renderConflictsPill())
	require.Empty(t, m.renderUnresolvedThreadsPill())
}

func TestRenderLinkedIssuesPill(t *testing.T) {
	tests := []struct {
		count    int
		expected string
	}{
		{count: 1, expected: constants.IssueIcon + " 1 issue"},
		{count: 7, expected: constants.IssueIcon + " 7 issues"},
	}

	for _, tt := range tests {
		m := newTestModel(t, &data.PullRequestData{
			State:        "OPEN",
			LinkedIssues: data.LinkedIssues{TotalCount: tt.count},
		}, nil, nil)
		require.Contains(t, ansi.Strip(m.renderLinkedIssuesPill()), tt.expected)
	}
}

func TestRenderConflictsPill(t *testing.T) {
	m := newTestModel(t, &data.PullRequestData{State: "OPEN", Mergeable: "CONFLICTING"}, nil, nil)
	require.Contains(t, ansi.Strip(m.renderConflictsPill()), constants.ConflictIcon+" Conflicts")

	merged := newTestModel(
		t,
		&data.PullRequestData{State: "MERGED", Mergeable: "CONFLICTING"},
		nil,
		nil,
	)
	require.Empty(t, merged.renderConflictsPill())
}

func TestRenderUnresolvedThreadsPill(t *testing.T) {
	m := newTestModel(t, &data.PullRequestData{
		State:         "OPEN",
		ReviewThreads: unresolvedThreads(true, false, false),
	}, nil, nil)
	require.Contains(
		t,
		ansi.Strip(m.renderUnresolvedThreadsPill()),
		constants.UnresolvedIcon+" 2 unresolved",
	)

	allResolved := newTestModel(t, &data.PullRequestData{
		State:         "OPEN",
		ReviewThreads: unresolvedThreads(true, true),
	}, nil, nil)
	require.Empty(t, allResolved.renderUnresolvedThreadsPill())
}
