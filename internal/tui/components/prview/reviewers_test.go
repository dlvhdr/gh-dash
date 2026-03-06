package prview

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

func newTestModel(t *testing.T, prData *data.PullRequestData) Model {
	return newTestModelWithWidth(t, prData, 0)
}

func newTestModelWithWidth(t *testing.T, prData *data.PullRequestData, width int) Model {
	t.Helper()
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../../../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	thm := theme.ParseTheme(&cfg)
	ctx := &context.ProgramContext{
		Config: &cfg,
		Theme:  thm,
		Styles: context.InitStyles(thm),
	}

	m := NewModel(ctx)
	m.ctx = ctx
	m.pr = &prrow.PullRequest{
		Ctx: ctx,
		Data: &prrow.Data{
			Primary:    prData,
			IsEnriched: true,
			Enriched: data.EnrichedPullRequestData{
				ReviewRequests: prData.ReviewRequests,
				Reviews:        prData.Reviews,
			},
		},
	}
	if width > 0 {
		m.width = width
	}
	return m
}

func TestRenderRequestedReviewers(t *testing.T) {
	testCases := map[string]struct {
		reviewRequests []data.ReviewRequestNode
		reviews        []data.Review
		wantContains   []string
		wantNotContain []string
	}{
		"empty review requests": {
			reviewRequests: []data.ReviewRequestNode{},
			reviews:        []data.Review{},
			wantContains:   []string{},
		},
		"single user awaiting review": {
			reviewRequests: []data.ReviewRequestNode{
				{
					AsCodeOwner: false,
					RequestedReviewer: struct {
						User      data.RequestedReviewerUser      `graphql:"... on User"`
						Team      data.RequestedReviewerTeam      `graphql:"... on Team"`
						Bot       data.RequestedReviewerBot       `graphql:"... on Bot"`
						Mannequin data.RequestedReviewerMannequin `graphql:"... on Mannequin"`
					}{
						User: data.RequestedReviewerUser{Login: "alice"},
					},
				},
			},
			reviews:      []data.Review{},
			wantContains: []string{"Reviewers", "@alice", constants.DotIcon},
		},
		"user who commented": {
			reviewRequests: []data.ReviewRequestNode{
				{
					AsCodeOwner: false,
					RequestedReviewer: struct {
						User      data.RequestedReviewerUser      `graphql:"... on User"`
						Team      data.RequestedReviewerTeam      `graphql:"... on Team"`
						Bot       data.RequestedReviewerBot       `graphql:"... on Bot"`
						Mannequin data.RequestedReviewerMannequin `graphql:"... on Mannequin"`
					}{
						User: data.RequestedReviewerUser{Login: "bob"},
					},
				},
			},
			reviews: []data.Review{
				{Author: struct{ Login string }{Login: "bob"}, State: "COMMENTED"},
			},
			wantContains:   []string{"Reviewers", "@bob", constants.CommentIcon},
			wantNotContain: []string{constants.WaitingIcon},
		},
		"code owner": {
			reviewRequests: []data.ReviewRequestNode{
				{
					AsCodeOwner: true,
					RequestedReviewer: struct {
						User      data.RequestedReviewerUser      `graphql:"... on User"`
						Team      data.RequestedReviewerTeam      `graphql:"... on Team"`
						Bot       data.RequestedReviewerBot       `graphql:"... on Bot"`
						Mannequin data.RequestedReviewerMannequin `graphql:"... on Mannequin"`
					}{
						User: data.RequestedReviewerUser{Login: "charlie"},
					},
				},
			},
			reviews:      []data.Review{},
			wantContains: []string{"Reviewers", "@charlie", constants.OwnerIcon},
		},
		"team reviewer": {
			reviewRequests: []data.ReviewRequestNode{
				{
					AsCodeOwner: false,
					RequestedReviewer: struct {
						User      data.RequestedReviewerUser      `graphql:"... on User"`
						Team      data.RequestedReviewerTeam      `graphql:"... on Team"`
						Bot       data.RequestedReviewerBot       `graphql:"... on Bot"`
						Mannequin data.RequestedReviewerMannequin `graphql:"... on Mannequin"`
					}{
						Team: data.RequestedReviewerTeam{Slug: "core-team", Name: "Core Team"},
					},
				},
			},
			reviews:        []data.Review{},
			wantContains:   []string{"Reviewers", "core-team"},
			wantNotContain: []string{"@core-team", constants.TeamIcon},
		},
		"user with bot in name but is code owner": {
			reviewRequests: []data.ReviewRequestNode{
				{
					AsCodeOwner: true,
					RequestedReviewer: struct {
						User      data.RequestedReviewerUser      `graphql:"... on User"`
						Team      data.RequestedReviewerTeam      `graphql:"... on Team"`
						Bot       data.RequestedReviewerBot       `graphql:"... on Bot"`
						Mannequin data.RequestedReviewerMannequin `graphql:"... on Mannequin"`
					}{
						User: data.RequestedReviewerUser{Login: "mdn-bot"},
					},
				},
			},
			reviews:      []data.Review{},
			wantContains: []string{"Reviewers", "@mdn-bot", constants.OwnerIcon},
		},
		"multiple reviewers": {
			reviewRequests: []data.ReviewRequestNode{
				{
					AsCodeOwner: false,
					RequestedReviewer: struct {
						User      data.RequestedReviewerUser      `graphql:"... on User"`
						Team      data.RequestedReviewerTeam      `graphql:"... on Team"`
						Bot       data.RequestedReviewerBot       `graphql:"... on Bot"`
						Mannequin data.RequestedReviewerMannequin `graphql:"... on Mannequin"`
					}{
						User: data.RequestedReviewerUser{Login: "alice"},
					},
				},
				{
					AsCodeOwner: true,
					RequestedReviewer: struct {
						User      data.RequestedReviewerUser      `graphql:"... on User"`
						Team      data.RequestedReviewerTeam      `graphql:"... on Team"`
						Bot       data.RequestedReviewerBot       `graphql:"... on Bot"`
						Mannequin data.RequestedReviewerMannequin `graphql:"... on Mannequin"`
					}{
						User: data.RequestedReviewerUser{Login: "bob"},
					},
				},
			},
			reviews:      []data.Review{},
			wantContains: []string{"Reviewers", "@alice", "@bob", ","},
		},
		"reviewer who approved": {
			reviewRequests: []data.ReviewRequestNode{},
			reviews: []data.Review{
				{Author: struct{ Login string }{Login: "alice"}, State: "APPROVED"},
			},
			wantContains: []string{"Reviewers", "@alice", constants.ApprovedIcon},
		},
		"reviewer who requested changes": {
			reviewRequests: []data.ReviewRequestNode{},
			reviews: []data.Review{
				{Author: struct{ Login string }{Login: "bob"}, State: "CHANGES_REQUESTED"},
			},
			wantContains: []string{"Reviewers", "@bob", constants.ChangesRequestedIcon},
		},
		"reviewer who only commented": {
			reviewRequests: []data.ReviewRequestNode{},
			reviews: []data.Review{
				{Author: struct{ Login string }{Login: "charlie"}, State: "COMMENTED"},
			},
			wantContains: []string{"Reviewers", "@charlie", constants.CommentIcon},
		},
		"reviewer who approved then commented": {
			reviewRequests: []data.ReviewRequestNode{},
			reviews: []data.Review{
				{Author: struct{ Login string }{Login: "alice"}, State: "APPROVED"},
				{Author: struct{ Login string }{Login: "alice"}, State: "COMMENTED"},
			},
			wantContains:   []string{"Reviewers", "@alice", constants.ApprovedIcon},
			wantNotContain: []string{constants.CommentIcon},
		},
		"reviewer who requested changes then commented": {
			reviewRequests: []data.ReviewRequestNode{},
			reviews: []data.Review{
				{Author: struct{ Login string }{Login: "bob"}, State: "CHANGES_REQUESTED"},
				{Author: struct{ Login string }{Login: "bob"}, State: "COMMENTED"},
			},
			wantContains:   []string{"Reviewers", "@bob", constants.ChangesRequestedIcon},
			wantNotContain: []string{constants.CommentIcon},
		},
		"mix of pending and completed reviews": {
			reviewRequests: []data.ReviewRequestNode{
				{
					AsCodeOwner: false,
					RequestedReviewer: struct {
						User      data.RequestedReviewerUser      `graphql:"... on User"`
						Team      data.RequestedReviewerTeam      `graphql:"... on Team"`
						Bot       data.RequestedReviewerBot       `graphql:"... on Bot"`
						Mannequin data.RequestedReviewerMannequin `graphql:"... on Mannequin"`
					}{
						User: data.RequestedReviewerUser{Login: "alice"},
					},
				},
			},
			reviews: []data.Review{
				{Author: struct{ Login string }{Login: "bob"}, State: "APPROVED"},
				{Author: struct{ Login string }{Login: "charlie"}, State: "CHANGES_REQUESTED"},
			},
			wantContains: []string{
				"Reviewers", "@alice", "@bob", "@charlie", constants.DotIcon,
				constants.ApprovedIcon, constants.ChangesRequestedIcon,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			prData := &data.PullRequestData{
				ReviewRequests: data.ReviewRequests{
					TotalCount: len(tc.reviewRequests),
					Nodes:      tc.reviewRequests,
				},
				Reviews: data.Reviews{
					TotalCount: len(tc.reviews),
					Nodes:      tc.reviews,
				},
			}

			m := newTestModel(t, prData)
			got := m.renderRequestedReviewers()

			if len(tc.wantContains) == 0 {
				require.Empty(t, got)
				return
			}

			for _, want := range tc.wantContains {
				require.True(t, strings.Contains(got, want),
					"expected output to contain %q, got: %q", want, got)
			}

			for _, notWant := range tc.wantNotContain {
				require.False(t, strings.Contains(got, notWant),
					"expected output to NOT contain %q, got: %q", notWant, got)
			}
		})
	}
}

func TestRenderRequestedReviewersWrapping(t *testing.T) {
	// Create multiple reviewers that would exceed a narrow width
	reviewRequests := []data.ReviewRequestNode{}
	for _, name := range []string{"alice", "bob", "charlie", "david", "eve"} {
		reviewRequests = append(reviewRequests, data.ReviewRequestNode{
			AsCodeOwner: false,
			RequestedReviewer: struct {
				User      data.RequestedReviewerUser      `graphql:"... on User"`
				Team      data.RequestedReviewerTeam      `graphql:"... on Team"`
				Bot       data.RequestedReviewerBot       `graphql:"... on Bot"`
				Mannequin data.RequestedReviewerMannequin `graphql:"... on Mannequin"`
			}{
				User: data.RequestedReviewerUser{Login: name},
			},
		})
	}

	prData := &data.PullRequestData{
		ReviewRequests: data.ReviewRequests{
			TotalCount: len(reviewRequests),
			Nodes:      reviewRequests,
		},
		Reviews: data.Reviews{
			TotalCount: 0,
			Nodes:      []data.Review{},
		},
	}

	// Use a narrow width to force wrapping
	m := newTestModelWithWidth(t, prData, 40)
	got := m.renderRequestedReviewers()

	// Should contain all reviewers
	for _, name := range []string{"@alice", "@bob", "@charlie", "@david", "@eve"} {
		require.True(t, strings.Contains(got, name),
			"expected output to contain %q, got: %q", name, got)
	}

	// Count newlines in the reviewer list (after "Reviewers" title)
	lines := strings.Split(got, "\n")
	// Should have: "Reviewers", empty line, then multiple lines of reviewers
	require.Greater(t, len(lines), 3,
		"expected output to wrap to multiple lines, got %d lines: %q", len(lines), got)
}

func TestRenderRequestedReviewersLoading(t *testing.T) {
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../../../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	require.NoError(t, err)

	thm := theme.ParseTheme(&cfg)
	ctx := &context.ProgramContext{
		Config: &cfg,
		Theme:  thm,
		Styles: context.InitStyles(thm),
	}

	m := NewModel(ctx)
	m.ctx = ctx
	m.pr = &prrow.PullRequest{
		Ctx: ctx,
		Data: &prrow.Data{
			Primary:    &data.PullRequestData{},
			IsEnriched: false, // Not yet enriched - should show loading
		},
	}

	got := m.renderRequestedReviewers()

	require.True(t, strings.Contains(got, "Reviewers"),
		"expected output to contain 'Reviewers' title, got: %q", got)
	require.True(t, strings.Contains(got, "Loading..."),
		"expected output to contain 'Loading...', got: %q", got)
}

func TestRenderSuggestedReviewers(t *testing.T) {
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../../../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	require.NoError(t, err)

	thm := theme.ParseTheme(&cfg)
	ctx := &context.ProgramContext{
		Config: &cfg,
		Theme:  thm,
		Styles: context.InitStyles(thm),
	}

	m := NewModel(ctx)
	m.ctx = ctx
	m.pr = &prrow.PullRequest{
		Ctx: ctx,
		Data: &prrow.Data{
			Primary:    &data.PullRequestData{},
			IsEnriched: true,
			Enriched: data.EnrichedPullRequestData{
				ReviewRequests: data.ReviewRequests{},
				Reviews:        data.Reviews{},
				SuggestedReviewers: []data.SuggestedReviewer{
					{
						IsAuthor:    false,
						IsCommenter: false,
						Reviewer:    struct{ Login string }{Login: "codeowner1"},
					},
					{
						IsAuthor:    true, // Should be skipped
						IsCommenter: false,
						Reviewer:    struct{ Login string }{Login: "author"},
					},
				},
			},
		},
	}

	got := m.renderRequestedReviewers()

	require.True(t, strings.Contains(got, "Reviewers"),
		"expected output to contain 'Reviewers' title, got: %q", got)
	require.True(t, strings.Contains(got, "@codeowner1"),
		"expected output to contain suggested reviewer '@codeowner1', got: %q", got)
	require.True(t, strings.Contains(got, constants.OwnerIcon),
		"expected output to contain owner icon for suggested reviewer, got: %q", got)
	require.False(t, strings.Contains(got, "@author"),
		"expected output to NOT contain '@author' (PR author), got: %q", got)
}
