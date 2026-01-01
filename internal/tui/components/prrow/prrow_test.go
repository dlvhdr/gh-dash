package prrow

import (
	"testing"

	graphql "github.com/cli/shurcooL-graphql"
	checks "github.com/dlvhdr/x/gh-checks"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

func TestGetStatusChecksRollup(t *testing.T) {
	tests := []struct {
		name     string
		pr       *PullRequest
		expected checks.CommitState
	}{
		{
			name:     "nil Data returns Unknown",
			pr:       &PullRequest{Data: nil},
			expected: checks.CommitStateUnknown,
		},
		{
			name:     "nil Primary returns Unknown",
			pr:       &PullRequest{Data: &Data{Primary: nil}},
			expected: checks.CommitStateUnknown,
		},
		{
			name: "empty Commits returns Unknown",
			pr: &PullRequest{
				Data: &Data{
					Primary: &data.PullRequestData{
						Commits: data.Commits{
							Nodes: []struct {
								Commit struct {
									Deployments struct {
										Nodes []struct {
											Task        graphql.String
											Description graphql.String
										}
									} `graphql:"deployments(last: 10)"`
									CommitUrl         graphql.String
									StatusCheckRollup struct {
										State graphql.String
									}
								}
							}{},
						},
					},
				},
			},
			expected: checks.CommitStateUnknown,
		},
		{
			name: "SUCCESS state returns Success",
			pr: &PullRequest{
				Data: &Data{
					Primary: &data.PullRequestData{
						Commits: data.Commits{
							Nodes: []struct {
								Commit struct {
									Deployments struct {
										Nodes []struct {
											Task        graphql.String
											Description graphql.String
										}
									} `graphql:"deployments(last: 10)"`
									CommitUrl         graphql.String
									StatusCheckRollup struct {
										State graphql.String
									}
								}
							}{
								{
									Commit: struct {
										Deployments struct {
											Nodes []struct {
												Task        graphql.String
												Description graphql.String
											}
										} `graphql:"deployments(last: 10)"`
										CommitUrl         graphql.String
										StatusCheckRollup struct {
											State graphql.String
										}
									}{
										StatusCheckRollup: struct {
											State graphql.String
										}{
											State: "SUCCESS",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: checks.CommitStateSuccess,
		},
		{
			name: "FAILURE state returns Failure",
			pr: &PullRequest{
				Data: &Data{
					Primary: &data.PullRequestData{
						Commits: data.Commits{
							Nodes: []struct {
								Commit struct {
									Deployments struct {
										Nodes []struct {
											Task        graphql.String
											Description graphql.String
										}
									} `graphql:"deployments(last: 10)"`
									CommitUrl         graphql.String
									StatusCheckRollup struct {
										State graphql.String
									}
								}
							}{
								{
									Commit: struct {
										Deployments struct {
											Nodes []struct {
												Task        graphql.String
												Description graphql.String
											}
										} `graphql:"deployments(last: 10)"`
										CommitUrl         graphql.String
										StatusCheckRollup struct {
											State graphql.String
										}
									}{
										StatusCheckRollup: struct {
											State graphql.String
										}{
											State: "FAILURE",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: checks.CommitStateFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pr.GetStatusChecksRollup()
			if result != tt.expected {
				t.Errorf("GetStatusChecksRollup() = %v, want %v", result, tt.expected)
			}
		})
	}
}
