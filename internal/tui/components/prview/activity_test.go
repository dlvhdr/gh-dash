package prview

import (
	"regexp"
	"strings"
	"testing"
	"time"

	graphql "github.com/cli/shurcooL-graphql"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/markdown"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

func init() {
	markdown.InitializeMarkdownStyle(true)
}

// ansiRE matches CSI/SGR escape sequences emitted by lipgloss/glamour. Tests
// assert on rendered substrings; stripping ANSI is required when adjacent
// styled fragments are split across different styles (e.g. "↳ " then author).
var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func stripAnsi(s string) string { return ansiRE.ReplaceAllString(s, "") }

// reviewSpec is a compact test input for one PullRequestReview submission.
type reviewSpec struct {
	id          string // synthetic ID; threads opened by this Review point back to it
	state       string // APPROVED, CHANGES_REQUESTED, COMMENTED, PENDING
	body        string
	author      string
	submittedAt time.Time
}

// threadSpec is a compact test input for one ReviewThread.
type threadSpec struct {
	openedByReviewId string // links the thread to its parent reviewSpec.id
	path             string
	line             int
	startLine        int // 0 for single-line
	isOutdated       bool
	isResolved       bool
	diffHunk         string
	comments         []reviewCommentSpec
}

type reviewCommentSpec struct {
	author    string
	body      string
	createdAt time.Time
	replyToId string // empty for openers
}

// issueCommentSpec is a compact test input for one top-level PR conversation
// IssueComment (the github "Comment" type, not a ReviewComment).
type issueCommentSpec struct {
	author    string
	body      string
	createdAt time.Time
}

type activityTestOptions struct {
	reviews       []reviewSpec
	threads       []threadSpec
	issueComments []issueCommentSpec
}

func newTestModelForActivity(t *testing.T, opts activityTestOptions) Model {
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
		Config:           &cfg,
		Theme:            thm,
		Styles:           context.InitStyles(thm),
		MainContentWidth: 100,
	}

	enriched := data.EnrichedPullRequestData{}

	for _, r := range opts.reviews {
		enriched.Reviews.Nodes = append(enriched.Reviews.Nodes, data.Review{
			Id:          r.id,
			State:       r.state,
			Body:        r.body,
			SubmittedAt: r.submittedAt,
			Author: struct {
				Login string
			}{Login: r.author},
		})
	}
	enriched.Reviews.TotalCount = len(enriched.Reviews.Nodes)

	for _, ic := range opts.issueComments {
		enriched.Comments.Nodes = append(enriched.Comments.Nodes, data.Comment{
			Body:      ic.body,
			CreatedAt: ic.createdAt,
			UpdatedAt: ic.createdAt,
			Author: struct {
				Login string
			}{Login: ic.author},
		})
	}
	enriched.Comments.TotalCount = graphql.Int(len(enriched.Comments.Nodes))

	for _, th := range opts.threads {
		threadNode := struct {
			Id           string
			IsOutdated   bool
			IsResolved   bool
			OriginalLine int
			StartLine    int
			Line         int
			Path         string
			Comments     data.ReviewComments `graphql:"comments(first: 20)"`
		}{
			Id:         th.path + "#" + th.openedByReviewId,
			IsOutdated: th.isOutdated,
			IsResolved: th.isResolved,
			StartLine:  th.startLine,
			Line:       th.line,
			Path:       th.path,
		}
		for _, c := range th.comments {
			threadNode.Comments.Nodes = append(threadNode.Comments.Nodes, data.ReviewComment{
				Body:      c.body,
				CreatedAt: c.createdAt,
				DiffHunk:  th.diffHunk,
				StartLine: th.startLine,
				Line:      th.line,
				Author: struct {
					Login string
				}{Login: c.author},
				PullRequestReview: struct{ Id string }{Id: th.openedByReviewId},
				ReplyTo:           struct{ Id string }{Id: c.replyToId},
			})
		}
		threadNode.Comments.TotalCount = len(threadNode.Comments.Nodes)
		enriched.ReviewThreads.Nodes = append(enriched.ReviewThreads.Nodes, threadNode)
	}

	m := NewModel(ctx)
	m.ctx = ctx
	m.width = 100
	m.pr = &prrow.PullRequest{
		Ctx: ctx,
		Data: &prrow.Data{
			Primary: &data.PullRequestData{
				Repository: data.Repository{},
			},
			IsEnriched: true,
			Enriched:   enriched,
		},
	}
	return m
}

const sampleDiffHunk = `@@ -38,7 +38,11 @@ func (m *Model) renderActivity() string {
 	var activities []RenderedActivity
 	var comments []comment
-	if !m.pr.Data.IsEnriched {
+	if !m.pr.Data.IsEnriched {
+		return bodyStyle.Render("Loading...")
+	}
+
+	if len(m.pr.Data.Enriched.ReviewThreads.Nodes) == 0 {
 		return bodyStyle.Render("Loading...")
 	}`

// Cycle 1 (tracer): A Review opening one ReviewThread (one ReviewComment)
// renders the Review pill, the thread's path:line header, the diff-hunk
// content, and the comment body.
func TestRenderActivity_ReviewWithOneThread(t *testing.T) {
	submittedAt := time.Date(2026, 5, 13, 16, 33, 24, 0, time.UTC)
	createdAt := submittedAt.Add(-2 * time.Second)
	opts := activityTestOptions{
		reviews: []reviewSpec{
			{
				id:          "REVIEW_1",
				state:       "APPROVED",
				body:        "LGTM with one nit.",
				author:      "alice",
				submittedAt: submittedAt,
			},
		},
		threads: []threadSpec{
			{
				openedByReviewId: "REVIEW_1",
				path:             "internal/data/prapi.go",
				line:             42,
				diffHunk:         sampleDiffHunk,
				comments: []reviewCommentSpec{
					{
						author:    "alice",
						body:      "Could we extract this into a helper?",
						createdAt: createdAt,
					},
				},
			},
		},
	}

	m := newTestModelForActivity(t, opts)
	got := m.renderActivity()

	require.Contains(t, got, "alice", "review author should appear")
	require.Contains(t, got, "LGTM with one nit.", "review body should appear")
	require.Contains(t, got, "internal/data/prapi.go:42", "thread path:line anchor should appear")
	require.Contains(t, got, "ReviewThreads.Nodes",
		"diff hunk content should appear (truncated to last 4 lines, which contain this fragment)")
	require.Contains(t, got, "Could we extract this", "comment body should appear")
	require.False(t, strings.Contains(got, "no comments"), "should not show empty state")
}

// Cycle 2: A multi-line ReviewThread (StartLine != Line) renders the anchor
// header as "path:startLine-line" rather than "path:line".
func TestRenderActivity_MultiLineAnchor(t *testing.T) {
	submittedAt := time.Date(2026, 5, 13, 16, 33, 24, 0, time.UTC)
	opts := activityTestOptions{
		reviews: []reviewSpec{
			{id: "R1", state: "COMMENTED", body: "x", author: "bob", submittedAt: submittedAt},
		},
		threads: []threadSpec{
			{
				openedByReviewId: "R1",
				path:             "internal/data/prapi.go",
				startLine:        38,
				line:             42,
				diffHunk:         sampleDiffHunk,
				comments: []reviewCommentSpec{
					{author: "bob", body: "Multi-line concern", createdAt: submittedAt},
				},
			},
		},
	}

	m := newTestModelForActivity(t, opts)
	got := m.renderActivity()

	require.Contains(t, got, "internal/data/prapi.go:38-42",
		"multi-line anchor should render as path:start-end")
	require.NotContains(t, got, "internal/data/prapi.go:42 ",
		"should not render the single-line path:line form for a multi-line thread")
}

// Cycle 3: A ReviewThread with IsOutdated=true shows an [Outdated] badge in
// its header. The thread content stays visible (we deliberately do not
// collapse — see Q4a in the spec).
func TestRenderActivity_OutdatedBadge(t *testing.T) {
	submittedAt := time.Date(2026, 5, 13, 16, 33, 24, 0, time.UTC)
	opts := activityTestOptions{
		reviews: []reviewSpec{
			{id: "R1", state: "COMMENTED", body: "x", author: "bob", submittedAt: submittedAt},
		},
		threads: []threadSpec{
			{
				openedByReviewId: "R1",
				path:             "old/file.go",
				line:             10,
				isOutdated:       true,
				diffHunk:         sampleDiffHunk,
				comments: []reviewCommentSpec{
					{author: "bob", body: "Stale concern", createdAt: submittedAt},
				},
			},
		},
	}

	m := newTestModelForActivity(t, opts)
	got := m.renderActivity()

	require.Contains(t, got, "Outdated", "outdated badge should appear")
	require.Contains(t, got, "Stale concern", "outdated thread body should still render")
}

// Cycle 4: A ReviewThread with IsResolved=true shows a [Resolved] badge in
// its header. Like Outdated, content stays visible.
func TestRenderActivity_ResolvedBadge(t *testing.T) {
	submittedAt := time.Date(2026, 5, 13, 16, 33, 24, 0, time.UTC)
	opts := activityTestOptions{
		reviews: []reviewSpec{
			{id: "R1", state: "COMMENTED", body: "x", author: "bob", submittedAt: submittedAt},
		},
		threads: []threadSpec{
			{
				openedByReviewId: "R1",
				path:             "internal/data/prapi.go",
				line:             10,
				isResolved:       true,
				diffHunk:         sampleDiffHunk,
				comments: []reviewCommentSpec{
					{author: "bob", body: "Closed convo", createdAt: submittedAt},
				},
			},
		},
	}

	m := newTestModelForActivity(t, opts)
	got := m.renderActivity()

	require.Contains(t, got, "Resolved", "resolved badge should appear")
	require.Contains(t, got, "Closed convo", "resolved thread body should still render")
}

// Cycle 6: Reviews and IssueComments interleave in chronological order at the
// top level. We key Reviews by SubmittedAt and IssueComments by CreatedAt and
// expect the rendered output to contain them in that order regardless of the
// source-stream order they arrived in.
func TestRenderActivity_ChronologicalOrdering(t *testing.T) {
	t0 := time.Date(2026, 5, 13, 9, 0, 0, 0, time.UTC)
	opts := activityTestOptions{
		reviews: []reviewSpec{
			{id: "R_LATE", state: "APPROVED", body: "review-late-body", author: "carol", submittedAt: t0.Add(10 * time.Hour)},
			{id: "R_EARLY", state: "CHANGES_REQUESTED", body: "review-early-body", author: "alice", submittedAt: t0},
		},
		issueComments: []issueCommentSpec{
			{author: "bob", body: "issue-comment-body", createdAt: t0.Add(5 * time.Hour)},
		},
	}

	m := newTestModelForActivity(t, opts)
	got := stripAnsi(m.renderActivity())

	earlyIdx := strings.Index(got, "review-early-body")
	icIdx := strings.Index(got, "issue-comment-body")
	lateIdx := strings.Index(got, "review-late-body")

	require.NotEqual(t, -1, earlyIdx, "early review body should render")
	require.NotEqual(t, -1, icIdx, "issue comment body should render")
	require.NotEqual(t, -1, lateIdx, "late review body should render")
	require.Less(t, earlyIdx, icIdx, "earlier review should render before later issue comment")
	require.Less(t, icIdx, lateIdx, "earlier issue comment should render before later review")
}

// Cycle 7: An implicit Review (state=COMMENTED, body="") that opened exactly
// one ReviewThread is unwrapped: the thread renders as a top-level item, and
// no "X reviewed" wrapper appears for the synthetic Review.
func TestRenderActivity_ImplicitReviewOneThreadPromoted(t *testing.T) {
	t0 := time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC)
	opts := activityTestOptions{
		reviews: []reviewSpec{
			{id: "IMPLICIT", state: "COMMENTED", body: "", author: "alice", submittedAt: t0},
		},
		threads: []threadSpec{
			{
				openedByReviewId: "IMPLICIT",
				path:             "internal/data/prapi.go",
				line:             42,
				diffHunk:         sampleDiffHunk,
				comments: []reviewCommentSpec{
					{author: "alice", body: "bare-thread-body", createdAt: t0},
				},
			},
		},
	}

	m := newTestModelForActivity(t, opts)
	got := stripAnsi(m.renderActivity())

	require.Contains(t, got, "bare-thread-body", "promoted thread body should render")
	require.Contains(t, got, "internal/data/prapi.go:42", "thread anchor should render")
	require.NotContains(t, got, "reviewed", "implicit Review wrapper should be stripped (no 'reviewed' header)")
}

// Cycle 8: An implicit Review (state=COMMENTED, body="") that opened zero
// ReviewThreads — i.e. it only contributed replies into other people's
// threads — is hidden entirely. The reply already renders inside its parent
// thread, so surfacing the empty wrapper would be redundant noise.
func TestRenderActivity_ImplicitReviewZeroThreadsHidden(t *testing.T) {
	t0 := time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC)
	opts := activityTestOptions{
		reviews: []reviewSpec{
			{id: "REAL", state: "APPROVED", body: "real-review-body", author: "alice", submittedAt: t0},
			// Implicit Review contributes a reply into REAL's thread; opens no threads itself.
			{id: "IMPLICIT_REPLIER", state: "COMMENTED", body: "", author: "bob", submittedAt: t0.Add(time.Hour)},
		},
		threads: []threadSpec{
			{
				openedByReviewId: "REAL",
				path:             "internal/data/prapi.go",
				line:             42,
				diffHunk:         sampleDiffHunk,
				comments: []reviewCommentSpec{
					{author: "alice", body: "opener-body", createdAt: t0},
					{author: "bob", body: "reply-body", createdAt: t0.Add(time.Hour)},
				},
			},
		},
	}

	m := newTestModelForActivity(t, opts)
	got := stripAnsi(m.renderActivity())

	require.Contains(t, got, "real-review-body", "real review body should render")
	require.Contains(t, got, "reply-body", "reply should still render inside the parent thread")
	bobReviewedIdx := strings.Index(got, "bob reviewed")
	require.Equal(t, -1, bobReviewedIdx, "implicit reply-only Review wrapper must not appear as 'bob reviewed'")
}

// Cycle 9: An implicit Review (state=COMMENTED, body="") that opened 2+
// ReviewThreads in a single submission is unwrapped into multiple bare
// top-level threads (one activityItem per thread). No "X reviewed" wrapper
// appears for the synthetic Review.
func TestRenderActivity_ImplicitReviewMultipleThreadsPromoted(t *testing.T) {
	t0 := time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC)
	opts := activityTestOptions{
		reviews: []reviewSpec{
			{id: "IMPLICIT", state: "COMMENTED", body: "", author: "alice", submittedAt: t0},
		},
		threads: []threadSpec{
			{
				openedByReviewId: "IMPLICIT",
				path:             "a/first.go",
				line:             10,
				diffHunk:         sampleDiffHunk,
				comments: []reviewCommentSpec{
					{author: "alice", body: "first-thread-body", createdAt: t0},
				},
			},
			{
				openedByReviewId: "IMPLICIT",
				path:             "b/second.go",
				line:             20,
				diffHunk:         sampleDiffHunk,
				comments: []reviewCommentSpec{
					{author: "alice", body: "second-thread-body", createdAt: t0.Add(time.Second)},
				},
			},
		},
	}

	m := newTestModelForActivity(t, opts)
	got := stripAnsi(m.renderActivity())

	require.Contains(t, got, "first-thread-body", "first promoted thread should render")
	require.Contains(t, got, "second-thread-body", "second promoted thread should render")
	require.Contains(t, got, "a/first.go:10", "first thread anchor should render")
	require.Contains(t, got, "b/second.go:20", "second thread anchor should render")
	require.NotContains(t, got, "reviewed", "implicit Review wrapper should be stripped")
}

// Cycle 10: The activity title's "N comments" count matches GitHub web — it
// counts leaf-level comments (non-empty review bodies + all thread comments +
// all issue comments), not top-level rendered items. Implicit reviews (empty
// body) never contribute, since their wrapper is stripped.
func TestRenderActivity_TitleCountIsLeafComments(t *testing.T) {
	t0 := time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC)
	opts := activityTestOptions{
		reviews: []reviewSpec{
			{id: "REAL", state: "APPROVED", body: "real-review-body", author: "alice", submittedAt: t0},
			// Implicit review — should not contribute to the count.
			{id: "IMPLICIT", state: "COMMENTED", body: "", author: "bob", submittedAt: t0.Add(time.Hour)},
		},
		threads: []threadSpec{
			{
				openedByReviewId: "REAL",
				path:             "internal/data/prapi.go",
				line:             42,
				comments: []reviewCommentSpec{
					{author: "alice", body: "opener", createdAt: t0},
					{author: "bob", body: "reply", createdAt: t0.Add(time.Hour)},
				},
			},
		},
		issueComments: []issueCommentSpec{
			{author: "carol", body: "issue-comment", createdAt: t0.Add(2 * time.Hour)},
		},
	}

	m := newTestModelForActivity(t, opts)
	got := stripAnsi(m.renderActivity())

	// Leaf count: 1 review body + 2 thread comments + 1 issue comment = 4.
	require.Contains(t, got, "4 comments",
		"title should show leaf-comment count (1 review body + 2 thread comments + 1 issue comment)")
}

// Cycle 11: A DISMISSED review renders with a state glyph in its header (web
// signals "dismissed" with explicit affordance and so should we). Before this
// fix, renderReviewDecision returned "" for DISMISSED, leaving just whitespace
// before the author name.
func TestRenderActivity_DismissedReviewHasGlyph(t *testing.T) {
	t0 := time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC)
	opts := activityTestOptions{
		reviews: []reviewSpec{
			{id: "R_DISMISSED", state: "DISMISSED", body: "stale-approval", author: "jasper", submittedAt: t0},
		},
	}

	m := newTestModelForActivity(t, opts)
	got := stripAnsi(m.renderActivity())

	// The glyph used is ↺ (faint, "rotated/dismissed"). Test asserts presence
	// rather than a specific glyph so the visual can be tweaked without churn.
	require.Contains(t, got, "↺", "DISMISSED review header should include a state glyph")
	require.Contains(t, got, "stale-approval", "DISMISSED review body should still render")
	require.Contains(t, got, "jasper reviewed", "DISMISSED review still gets a reviewer header")
}

// Cycle 12: A diff hunk longer than 4 content lines is truncated to its last 4
// (closest to the anchor), matching GitHub web's collapsed preview. Earlier
// lines are dropped and an ellipsis marker indicates truncation. The @@ header
// is never shown — it's just position metadata.
func TestRenderActivity_DiffHunkTruncatedToLast4Lines(t *testing.T) {
	t0 := time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC)
	// 8 content lines after the @@ header; only the last 4 should survive.
	longHunk := "@@ -0,0 +1,8 @@\n" +
		"+line-one-FAR-FROM-ANCHOR\n" +
		"+line-two-FAR-FROM-ANCHOR\n" +
		"+line-three-FAR-FROM-ANCHOR\n" +
		"+line-four-FAR-FROM-ANCHOR\n" +
		"+line-five-KEEP\n" +
		"+line-six-KEEP\n" +
		"+line-seven-KEEP\n" +
		"+line-eight-KEEP-ANCHOR"
	opts := activityTestOptions{
		reviews: []reviewSpec{
			{id: "R1", state: "COMMENTED", body: "x", author: "alice", submittedAt: t0},
		},
		threads: []threadSpec{
			{
				openedByReviewId: "R1",
				path:             "new_file.py",
				line:             8,
				diffHunk:         longHunk,
				comments: []reviewCommentSpec{
					{author: "alice", body: "concern", createdAt: t0},
				},
			},
		},
	}

	m := newTestModelForActivity(t, opts)
	got := stripAnsi(m.renderActivity())

	require.Contains(t, got, "line-eight-KEEP-ANCHOR", "last hunk line should render")
	require.Contains(t, got, "line-five-KEEP", "4th-from-last line should render")
	require.NotContains(t, got, "line-four-FAR-FROM-ANCHOR", "5th-from-last line should be truncated")
	require.NotContains(t, got, "line-one-FAR-FROM-ANCHOR", "first line should be truncated")
	require.NotContains(t, got, "@@ -0,0", "the @@ header itself should never render")
}

// Cycle 13: When the anchor + badges together exceed the panel's content
// width, the badges fall to a second line under the anchor instead of being
// clipped. Badges carry "is this thread still actionable?" signal; truncating
// them silently loses information.
func TestRenderActivity_BadgesWrapWhenAnchorOverflows(t *testing.T) {
	t0 := time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC)
	longPath := strings.Repeat("very/deeply/nested/", 6) + "module.py"
	opts := activityTestOptions{
		reviews: []reviewSpec{
			{id: "R1", state: "COMMENTED", body: "x", author: "alice", submittedAt: t0},
		},
		threads: []threadSpec{
			{
				openedByReviewId: "R1",
				path:             longPath,
				line:             9999,
				isOutdated:       true,
				isResolved:       true,
				comments: []reviewCommentSpec{
					{author: "alice", body: "concern", createdAt: t0},
				},
			},
		},
	}

	m := newTestModelForActivity(t, opts)
	got := stripAnsi(m.renderActivity())

	require.Contains(t, got, longPath+":9999", "anchor should render in full")
	require.Contains(t, got, "Outdated", "Outdated badge should still appear even when anchor is wide")
	require.Contains(t, got, "Resolved", "Resolved badge should still appear even when anchor is wide")

	// Wrap is detected by the badges sitting on their own line — i.e. the
	// anchor and the first badge are separated by a newline, not just a space.
	anchorIdx := strings.Index(got, longPath+":9999")
	resolvedIdx := strings.Index(got, "[Resolved]")
	require.Greater(t, resolvedIdx, anchorIdx, "badges should follow the anchor")
	between := got[anchorIdx:resolvedIdx]
	require.Contains(t, between, "\n",
		"badges should wrap to a new line when they don't fit alongside the anchor")
}

// Cycle 5: Within a thread, the second-and-later ReviewComments render with a
// compact "↳ {author}" reply prefix instead of the full pill that the opener
// uses. This signals "this is a reply, not a new thread starter."
func TestRenderActivity_ReplyPrefix(t *testing.T) {
	t0 := time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC)
	opts := activityTestOptions{
		reviews: []reviewSpec{
			{id: "R1", state: "COMMENTED", body: "x", author: "alice", submittedAt: t0},
		},
		threads: []threadSpec{
			{
				openedByReviewId: "R1",
				path:             "internal/data/prapi.go",
				line:             42,
				diffHunk:         sampleDiffHunk,
				comments: []reviewCommentSpec{
					{author: "alice", body: "opener-body", createdAt: t0},
					{author: "bob", body: "reply-body", createdAt: t0.Add(time.Hour)},
				},
			},
		},
	}

	m := newTestModelForActivity(t, opts)
	got := stripAnsi(m.renderActivity())

	require.Contains(t, got, "opener-body", "opener comment body should render")
	require.Contains(t, got, "reply-body", "reply comment body should render")
	require.Contains(t, got, "↳ bob", "reply should be prefixed with ↳ {author}")
}

// Cycle 14: GitHub web's Conversation N badge silently excludes review-summary
// cards where state=COMMENTED, body is non-empty, and inline-comment count is
// zero — it treats them as redundant with a plain IssueComment. Reviews with
// any other non-COMMENTED state (DISMISSED/CHANGES_REQUESTED/APPROVED) still
// count, and COMMENTED reviews that DO host inline comments also count.
func TestRenderActivity_CountExcludesCommentedReviewWithNoInlines(t *testing.T) {
	t0 := time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC)
	opts := activityTestOptions{
		reviews: []reviewSpec{
			// COMMENTED + body + 0 inline → excluded from count (bot fallback
			// pattern, e.g. gemini "I am having trouble creating individual
			// review comments").
			{id: "R_COMMENTED_BODY_NO_INLINE", state: "COMMENTED", body: "trouble-body", author: "gemini-bot", submittedAt: t0},
			// COMMENTED + body + inline > 0 → counted.
			{id: "R_COMMENTED_BODY_WITH_INLINE", state: "COMMENTED", body: "review-with-inlines", author: "alice", submittedAt: t0.Add(time.Hour)},
			// DISMISSED + body + 0 inline → counted (state != COMMENTED).
			{id: "R_DISMISSED_NO_INLINE", state: "DISMISSED", body: "dismissed-body", author: "jasper", submittedAt: t0.Add(2 * time.Hour)},
		},
		threads: []threadSpec{
			{
				openedByReviewId: "R_COMMENTED_BODY_WITH_INLINE",
				path:             "internal/data/prapi.go",
				line:             42,
				comments: []reviewCommentSpec{
					{author: "alice", body: "inline-1", createdAt: t0.Add(time.Hour)},
				},
			},
		},
		issueComments: []issueCommentSpec{
			{author: "carol", body: "ic-1", createdAt: t0.Add(3 * time.Hour)},
		},
	}

	// Counted: R_COMMENTED_BODY_WITH_INLINE (1) + R_DISMISSED_NO_INLINE (1)
	//          + 1 inline + 1 IC = 4.
	// R_COMMENTED_BODY_NO_INLINE is excluded — that's the GitHub-web rule.
	m := newTestModelForActivity(t, opts)
	got := stripAnsi(m.renderActivity())
	require.Contains(t, got, "4 comments",
		"COMMENTED review with body but no inline comments should NOT contribute to the count")
}
