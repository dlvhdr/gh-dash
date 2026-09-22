package prview

import (
	"strings"
	"testing"
	"time"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/markdown"
)

func TestRenderActivityGroupsReviewThreadWithDiffHunk(t *testing.T) {
	markdown.InitializeMarkdownStyle(true)

	m := newTestModelForAction(t)
	m.width = 100

	firstCommentAt := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	replyAt := firstCommentAt.Add(time.Minute)
	generalCommentAt := firstCommentAt.Add(2 * time.Minute)

	m.pr.Data.Enriched.ReviewThreads.Nodes = []data.ReviewThread{
		{
			Path:       "internal/server.go",
			Line:       42,
			IsOutdated: true,
			Comments: data.ReviewComments{
				Nodes: []data.ReviewComment{
					{
						Author:    struct{ Login string }{Login: "reviewer"},
						Body:      "Please guard this path.",
						DiffHunk:  "@@ -40,3 +40,4 @@\n-oldCall()\n+newCall()",
						UpdatedAt: firstCommentAt,
					},
					{
						Author:    struct{ Login string }{Login: "author"},
						Body:      "Good catch, updated.",
						UpdatedAt: replyAt,
					},
				},
			},
		},
	}
	m.pr.Data.Enriched.Comments.Nodes = []data.Comment{
		{
			Author:    struct{ Login string }{Login: "maintainer"},
			Body:      "General note.",
			UpdatedAt: generalCommentAt,
		},
	}

	rendered := m.renderActivity()

	for _, want := range []string{
		"internal/server.go#l42",
		"outdated",
		"-oldCall()",
		"+newCall()",
		"Please guard this path.",
		"Good catch, updated.",
		"General note.",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("renderActivity() missing %q:\n%s", want, rendered)
		}
	}

	firstCommentIndex := strings.Index(rendered, "Please guard this path.")
	replyIndex := strings.Index(rendered, "Good catch, updated.")
	generalCommentIndex := strings.Index(rendered, "General note.")

	if !(firstCommentIndex < replyIndex && replyIndex < generalCommentIndex) {
		t.Fatalf("comments rendered out of order:\n%s", rendered)
	}
}
