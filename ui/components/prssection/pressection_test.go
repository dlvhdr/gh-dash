package prssection

import (
	"testing"

	"github.com/dlvhdr/gh-dash/data"
)

func Test_excludeArchivedPullRequests(t *testing.T) {
	m := &Model{}

	tests := []struct {
		name     string
		prs      []data.PullRequestData
		expected int
	}{
		{
			name: "Not all pull requests are archived",
			prs: []data.PullRequestData{
				{
					Repository: struct {
						NameWithOwner string
						IsArchived    bool
					}{
						IsArchived: false,
					},
				},
				{
					Repository: struct {
						NameWithOwner string
						IsArchived    bool
					}{
						IsArchived: false,
					},
				},
			},
			expected: 2,
		},
		{
			name: "All pull requests are archived",
			prs: []data.PullRequestData{
				{
					Repository: struct {
						NameWithOwner string
						IsArchived    bool
					}{
						IsArchived: true,
					},
				},
				{
					Repository: struct {
						NameWithOwner string
						IsArchived    bool
					}{
						IsArchived: true,
					},
				},
			},
			expected: 0,
		},
		{
			name: "There is only one archived pull request",
			prs: []data.PullRequestData{
				{
					Repository: struct {
						NameWithOwner string
						IsArchived    bool
					}{
						IsArchived: false,
					},
				},
				{
					Repository: struct {
						NameWithOwner string
						IsArchived    bool
					}{
						IsArchived: true,
					},
				},
			},
			expected: 1,
		},
		{
			name:     "Empty pull requests",
			prs:      []data.PullRequestData{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prs := m.excludeArchivedPullRequests(tt.prs)
			actual := len(prs)
			if actual != tt.expected {
				t.Errorf("atcual : %d, expected: %d", actual, tt.expected)
			}
		})
	}
}
