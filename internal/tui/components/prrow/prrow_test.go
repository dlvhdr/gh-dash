package prrow

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
	graphql "github.com/cli/shurcooL-graphql"
	checks "github.com/dlvhdr/x/gh-checks"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/table"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
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

func TestRenderLabels(t *testing.T) {
	tests := []struct {
		name         string
		pr           *PullRequest
		isSelected   bool
		wantContains []string
		wantNewlines int
	}{
		{
			name: "nil Data returns empty string",
			pr: &PullRequest{
				Data: nil,
				Ctx:  &context.ProgramContext{},
			},
		},
		{
			name: "nil Primary returns empty string",
			pr: &PullRequest{
				Data: &Data{Primary: nil},
				Ctx:  &context.ProgramContext{},
			},
		},
		{
			name: "empty labels returns empty string",
			pr: &PullRequest{
				Data: &Data{
					Primary: &data.PullRequestData{
						Labels: data.PRLabels{
							Nodes: []data.Label{},
						},
					},
				},
				Ctx: &context.ProgramContext{},
			},
		},
		{
			name: "single label returns non-empty string",
			pr: &PullRequest{
				Data: &Data{
					Primary: &data.PullRequestData{
						Labels: data.PRLabels{
							Nodes: []data.Label{
								{Name: "bug", Color: "FF0000"},
							},
						},
					},
				},
				Ctx: &context.ProgramContext{
					Config: &config.Config{
						Theme: &config.ThemeConfig{},
					},
				},
				Columns: []table.Column{
					{Title: constants.LabelsIcon, ComputedWidth: 20},
				},
			},
			wantContains: []string{"bug"},
		},
		{
			name: "compact labels keep overflow summary on one line",
			pr: &PullRequest{
				Data: &Data{
					Primary: &data.PullRequestData{
						Labels: data.PRLabels{
							Nodes: []data.Label{
								{Name: "bug", Color: "FF0000"},
								{Name: "fix", Color: "00FF00"},
								{Name: "chore", Color: "0000FF"},
							},
						},
					},
				},
				Ctx: &context.ProgramContext{
					Config: &config.Config{
						Theme: &config.ThemeConfig{
							Ui: config.UIThemeConfig{
								Table: config.TableUIThemeConfig{Compact: true},
							},
						},
					},
				},
				Columns: []table.Column{
					{Title: constants.LabelsIcon, ComputedWidth: 12},
				},
			},
			wantContains: []string{"bug", "fix", "+1"},
			wantNewlines: 0,
		},
		{
			name: "selected labels keep content on selected rows",
			pr: &PullRequest{
				Data: &Data{
					Primary: &data.PullRequestData{
						Labels: data.PRLabels{
							Nodes: []data.Label{
								{Name: "bug", Color: "FF0000"},
								{Name: "fix", Color: "00FF00"},
							},
						},
					},
				},
				Ctx: &context.ProgramContext{
					Config: &config.Config{
						Theme: &config.ThemeConfig{},
					},
				},
				Columns: []table.Column{
					{Title: constants.LabelsIcon, ComputedWidth: 20},
				},
			},
			isSelected:   true,
			wantContains: []string{"bug", "fix"},
			wantNewlines: 0,
		},
		{
			name: "full labels keep overflow summary across two lines",
			pr: &PullRequest{
				Data: &Data{
					Primary: &data.PullRequestData{
						Labels: data.PRLabels{
							Nodes: []data.Label{
								{Name: "bug", Color: "FF0000"},
								{Name: "fix", Color: "00FF00"},
								{Name: "chore", Color: "0000FF"},
								{Name: "feature", Color: "AAAAAA"},
							},
						},
					},
				},
				Ctx: &context.ProgramContext{
					Config: &config.Config{
						Theme: &config.ThemeConfig{},
					},
				},
				Columns: []table.Column{
					{Title: constants.LabelsIcon, ComputedWidth: 14},
				},
			},
			wantContains: []string{"bug", "fix", "chore", "+1"},
			wantNewlines: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.pr.Ctx.Theme.SelectedBackground = compat.AdaptiveColor{
				Light: lipgloss.Color("7"),
				Dark:  lipgloss.Color("7"),
			}
			result := tt.pr.renderLabels(tt.isSelected)

			// For nil/empty cases, expect empty string
			if tt.pr.Data == nil ||
				tt.pr.Data.Primary == nil ||
				len(tt.pr.Data.Primary.Labels.Nodes) == 0 {
				if result != "" {
					t.Errorf("renderLabels() = %q, want empty string", result)
				}
				return
			}

			if result == "" {
				t.Errorf(
					"renderLabels() returned empty string for %d labels",
					len(tt.pr.Data.Primary.Labels.Nodes),
				)
			}

			if strings.Count(result, "\n") != tt.wantNewlines {
				t.Errorf(
					"renderLabels() newline count = %d, want %d",
					strings.Count(result, "\n"),
					tt.wantNewlines,
				)
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("renderLabels() = %q, want substring %q", result, want)
				}
			}
		})
	}
}
