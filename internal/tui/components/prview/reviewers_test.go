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
	t.Helper()
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag: "../../../config/testdata/test-config.yml",
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
			Primary: prData,
		},
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
			wantContains: []string{"reviewers:", "@alice", constants.WaitingIcon},
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
			wantContains:   []string{"reviewers:", "@bob", constants.CommentIcon},
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
			wantContains: []string{"reviewers:", "@charlie", constants.OwnerIcon},
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
			reviews:      []data.Review{},
			wantContains: []string{"reviewers:", "core-team", constants.TeamIcon},
		},
		"bot reviewer no annotation": {
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
			reviews:        []data.Review{},
			wantContains:   []string{"reviewers:", "@mdn-bot"},
			wantNotContain: []string{constants.OwnerIcon},
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
			wantContains: []string{"reviewers:", "@alice", "@bob", ","},
		},
		"reviewer who approved": {
			reviewRequests: []data.ReviewRequestNode{},
			reviews: []data.Review{
				{Author: struct{ Login string }{Login: "alice"}, State: "APPROVED"},
			},
			wantContains: []string{"reviewers:", "@alice", constants.SuccessIcon},
		},
		"reviewer who requested changes": {
			reviewRequests: []data.ReviewRequestNode{},
			reviews: []data.Review{
				{Author: struct{ Login string }{Login: "bob"}, State: "CHANGES_REQUESTED"},
			},
			wantContains: []string{"reviewers:", "@bob", constants.FailureIcon},
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
			wantContains: []string{"reviewers:", "@alice", "@bob", "@charlie", constants.WaitingIcon, constants.SuccessIcon, constants.FailureIcon},
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
