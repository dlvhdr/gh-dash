package issuerow

import (
	"testing"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

func TestRenderType(t *testing.T) {
	tests := []struct {
		name      string
		issueType *data.IssueType
		expected  string
	}{
		{name: "no type renders the placeholder", issueType: nil, expected: constants.EmptyCell},
		{name: "Bug", issueType: &data.IssueType{Name: "Bug", Color: "RED"}, expected: "Bug"},
		{
			name:      "Feature",
			issueType: &data.IssueType{Name: "Feature", Color: "BLUE"},
			expected:  "Feature",
		},
		{
			name:      "unknown color",
			issueType: &data.IssueType{Name: "Epic", Color: "TEAL"},
			expected:  "Epic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := &Issue{
				Ctx:  &context.ProgramContext{},
				Data: data.IssueData{IssueType: tt.issueType},
			}
			require.Equal(t, tt.expected, ansi.Strip(issue.renderType()))
		})
	}
}

func solid(hex string) compat.AdaptiveColor {
	return compat.AdaptiveColor{Light: lipgloss.Color(hex), Dark: lipgloss.Color(hex)}
}

func TestRenderLinkedPrs(t *testing.T) {
	const (
		openSeq   = "38;2;0;255;0m"
		mergedSeq = "38;2;255;0;255m"
		draftSeq  = "38;2;128;128;128m"
	)
	open := data.LinkedPullRequest{State: "OPEN"}
	draft := data.LinkedPullRequest{State: "OPEN", IsDraft: true}
	merged := data.LinkedPullRequest{State: "MERGED"}

	tests := []struct {
		name     string
		prs      []data.LinkedPullRequest
		text     string
		colorSeq string
	}{
		{name: "no PRs renders the placeholder", prs: nil, text: constants.EmptyCell},
		{
			name:     "one open PR",
			prs:      []data.LinkedPullRequest{open},
			text:     "1",
			colorSeq: openSeq,
		},
		{
			name:     "open wins over merged and draft",
			prs:      []data.LinkedPullRequest{merged, draft, open},
			text:     "3",
			colorSeq: openSeq,
		},
		{
			name:     "draft wins over merged",
			prs:      []data.LinkedPullRequest{merged, draft},
			text:     "2",
			colorSeq: draftSeq,
		},
		{
			name:     "all merged",
			prs:      []data.LinkedPullRequest{merged, merged},
			text:     "2",
			colorSeq: mergedSeq,
		},
	}

	ctx := &context.ProgramContext{Theme: theme.Theme{FaintText: solid("#808080")}}
	ctx.Styles.Colors.OpenPR = solid("#00FF00")
	ctx.Styles.Colors.MergedPR = solid("#FF00FF")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := &Issue{
				Ctx: ctx,
				Data: data.IssueData{
					LinkedPrs: data.LinkedPullRequests{TotalCount: len(tt.prs), Nodes: tt.prs},
				},
			}
			out := issue.renderLinkedPrs()
			require.Equal(t, tt.text, ansi.Strip(out))
			if tt.colorSeq != "" {
				require.Contains(t, out, tt.colorSeq)
			}
		})
	}
}
