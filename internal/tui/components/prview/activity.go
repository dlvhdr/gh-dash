package prview

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/markdown"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

// diffHunkPreviewLines is the number of trailing context lines we keep when
// rendering a thread's diff hunk — matches the collapsed preview GitHub web
// shows above the first comment of a thread.
const diffHunkPreviewLines = 4

// activityItem is a tagged union of the top-level PR-activity kinds. Exactly
// one of review/issueComment/thread is set; at is the chronological key
// (Review.SubmittedAt, IssueComment.CreatedAt, or — for an implicit-Review
// promoted thread — its first comment's CreatedAt) used for top-level
// ordering. The thread variant exists only because GitHub creates a synthetic
// "implicit" Review wrapper (state=COMMENTED, body="") to host inline activity
// that wasn't part of a formal review submission; we strip those wrappers and
// surface their threads directly. See docs/adr/0001 for the three-case rule.
type activityItem struct {
	at           time.Time
	review       *data.Review
	issueComment *data.Comment
	thread       *reviewThread
}

// isImplicitReview is true for the GitHub-generated wrapper Review that hosts
// inline activity outside a formal "Start a review" flow. The signal is
// state=COMMENTED with an empty body — a real "just commented" Review always
// carries a body.
func isImplicitReview(r data.Review) bool {
	return r.State == "COMMENTED" && r.Body == ""
}

// reviewThread is the local-to-this-package view of one
// data.ReviewThreadsWithComments node. The upstream data type is an anonymous
// struct, so we copy what we need into this typed shape to keep the render
// code free of inline struct gymnastics.
type reviewThread struct {
	Id         string
	IsOutdated bool
	IsResolved bool
	Path       string
	Line       int
	StartLine  int
	Comments   []data.ReviewComment
}

func (m *Model) renderActivity() string {
	if !m.pr.Data.IsEnriched {
		return lipgloss.NewStyle().Render("Loading...")
	}

	width := m.getIndentedContentWidth()
	r := markdown.GetMarkdownRenderer(width)

	threads := m.collectThreads()
	items := m.collectActivityItems(threads)

	var rendered []string
	for _, it := range items {
		switch {
		case it.review != nil:
			rendered = append(rendered, m.renderReviewItem(*it.review, threads, r))
		case it.issueComment != nil:
			rendered = append(rendered, m.renderIssueComment(*it.issueComment, r))
		case it.thread != nil:
			rendered = append(rendered, m.renderThread(*it.thread, r))
		}
	}

	if len(rendered) == 0 {
		return renderEmptyState()
	}

	title := m.ctx.Styles.Common.MainTextStyle.MarginBottom(1).Underline(true).Render(
		fmt.Sprintf("%s  %d comments", constants.CommentsIcon, m.leafCommentCount(items)))
	body := lipgloss.JoinVertical(lipgloss.Left, rendered...)
	return lipgloss.JoinVertical(lipgloss.Left, title, body)
}

// collectActivityItems flattens the top-level streams (Reviews, IssueComments)
// into a single chronologically-sorted slice and applies the implicit-Review
// unwrap rule from docs/adr/0001:
//
//   - 0 threads opened — implicit Review only contributed replies into other
//     people's threads. Hide entirely; the reply already renders inside its
//     parent thread.
//   - 1 thread opened — promote the thread to a top-level activityItem.
//   - 2+ threads opened — promote each thread as a sibling top-level item.
//
// Non-implicit Reviews render their wrapper with their opened threads nested
// inside. ReviewThreads opened by non-implicit Reviews are intentionally not
// top-level.
func (m *Model) collectActivityItems(threads []reviewThread) []activityItem {
	enriched := m.pr.Data.Enriched
	items := make([]activityItem, 0, len(enriched.Reviews.Nodes)+len(enriched.Comments.Nodes))
	for i := range enriched.Reviews.Nodes {
		rev := &enriched.Reviews.Nodes[i]
		if isImplicitReview(*rev) {
			opened := threadsOpenedBy(rev.Id, threads)
			for j := range opened {
				t := &opened[j]
				key := rev.SubmittedAt
				if len(t.Comments) > 0 {
					key = t.Comments[0].CreatedAt
				}
				items = append(items, activityItem{at: key, thread: t})
			}
			continue
		}
		items = append(items, activityItem{at: rev.SubmittedAt, review: rev})
	}
	for i := range enriched.Comments.Nodes {
		ic := &enriched.Comments.Nodes[i]
		items = append(items, activityItem{at: ic.CreatedAt, issueComment: ic})
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].at.Before(items[j].at) })
	return items
}

// leafCommentCount returns the leaf-level "N comments" figure GitHub web's
// Conversation tab badge displays: inline thread comments + top-level
// IssueComments + review-summary cards that carry a body — except a
// state=COMMENTED review whose body has zero inline comments, which the badge
// silently excludes (it treats those as redundant with a plain issue-comment,
// even though the timeline still renders the card).
func (m *Model) leafCommentCount(items []activityItem) int {
	n := 0
	threads := m.collectThreads()
	for _, it := range items {
		switch {
		case it.review != nil:
			inlineCount := 0
			for _, t := range threadsOpenedBy(it.review.Id, threads) {
				inlineCount += len(t.Comments)
			}
			if it.review.Body != "" && !(it.review.State == "COMMENTED" && inlineCount == 0) {
				n++
			}
			n += inlineCount
		case it.thread != nil:
			n += len(it.thread.Comments)
		case it.issueComment != nil:
			n++
		}
	}
	return n
}

func (m *Model) renderIssueComment(c data.Comment, r glamour.TermRenderer) string {
	pill := m.renderAuthorTimePill(c.Author.Login, c.CreatedAt)
	body, _ := r.Render(c.Body)
	return lipgloss.JoinVertical(lipgloss.Left, pill, body)
}

// renderAuthorTimePill draws the standard rounded-border "{author} {time}"
// header used by IssueComments and ReviewComment openers. Replies use a
// compact ↳-prefix instead and bypass this helper.
func (m *Model) renderAuthorTimePill(login string, at time.Time) string {
	return lipgloss.NewStyle().
		Width(m.getIndentedContentWidth()).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(m.ctx.Theme.FaintBorder).
		Render(lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.ctx.Styles.Common.MainTextStyle.Render(login),
			" ",
			lipgloss.NewStyle().
				Foreground(m.ctx.Theme.FaintText).
				Render(utils.TimeElapsed(at)),
		))
}

// collectThreads copies the enriched-PR review threads into the local
// reviewThread shape, keyed in encounter order (the order GitHub returned
// them, which is reverse-chronological by default).
func (m *Model) collectThreads() []reviewThread {
	out := make([]reviewThread, 0, len(m.pr.Data.Enriched.ReviewThreads.Nodes))
	for _, t := range m.pr.Data.Enriched.ReviewThreads.Nodes {
		out = append(out, reviewThread{
			Id:         t.Id,
			IsOutdated: t.IsOutdated,
			IsResolved: t.IsResolved,
			Path:       t.Path,
			Line:       t.Line,
			StartLine:  t.StartLine,
			Comments:   t.Comments.Nodes,
		})
	}
	return out
}

// threadsOpenedBy returns every thread whose first ReviewComment points back
// to the given Review via its PullRequestReview.Id link. That is GitHub's
// canonical "this Review opened this thread" relationship.
func threadsOpenedBy(reviewId string, threads []reviewThread) []reviewThread {
	var out []reviewThread
	for _, t := range threads {
		if len(t.Comments) == 0 {
			continue
		}
		if t.Comments[0].PullRequestReview.Id == reviewId {
			out = append(out, t)
		}
	}
	return out
}

func (m *Model) renderReviewItem(
	review data.Review,
	threads []reviewThread,
	r glamour.TermRenderer,
) string {
	header := m.renderReviewHeader(review)
	body, _ := r.Render(review.Body)

	parts := []string{header, body}
	for _, t := range threadsOpenedBy(review.Id, threads) {
		parts = append(parts, m.renderThread(t, r))
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// joinedWithSpaces returns the input slice with a single-space separator
// interleaved between each element, suitable for JoinHorizontal. Mirrors
// strings.Join's shape but for lipgloss horizontal layout.
func joinedWithSpaces(parts []string) []string {
	if len(parts) == 0 {
		return nil
	}
	out := make([]string, 0, len(parts)*2-1)
	for i, p := range parts {
		if i > 0 {
			out = append(out, " ")
		}
		out = append(out, p)
	}
	return out
}

// previewDiffHunk strips the leading "@@ … @@" position header and keeps only
// the last diffHunkPreviewLines content lines (the ones closest to the comment
// anchor). For brand-new files the raw hunk is the entire file body — without
// this preview, a single inline comment can scroll for pages.
func previewDiffHunk(hunk string) string {
	lines := strings.Split(hunk, "\n")
	if len(lines) > 0 && strings.HasPrefix(lines[0], "@@") {
		lines = lines[1:]
	}
	if len(lines) <= diffHunkPreviewLines {
		return strings.Join(lines, "\n")
	}
	tail := lines[len(lines)-diffHunkPreviewLines:]
	return strings.Join(append([]string{"…"}, tail...), "\n")
}

func (m *Model) renderThread(t reviewThread, r glamour.TermRenderer) string {
	// Anchor format mirrors CLI conventions (grep, stack traces, IDE links):
	// "path:line" for single-line comments, "path:start-end" for multi-line.
	// GitHub returns StartLine == 0 for single-line comments.
	anchor := fmt.Sprintf("%s:%d", t.Path, t.Line)
	if t.StartLine > 0 && t.StartLine != t.Line {
		anchor = fmt.Sprintf("%s:%d-%d", t.Path, t.StartLine, t.Line)
	}
	anchorRendered := lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render(anchor)
	var badges []string
	if t.IsOutdated {
		badges = append(badges,
			lipgloss.NewStyle().Foreground(m.ctx.Theme.WarningText).Render("[Outdated]"))
	}
	if t.IsResolved {
		badges = append(badges,
			lipgloss.NewStyle().Foreground(m.ctx.Theme.SuccessText).Render("[Resolved]"))
	}
	header := anchorRendered
	if len(badges) > 0 {
		badgeRow := lipgloss.JoinHorizontal(lipgloss.Top, joinedWithSpaces(badges)...)
		// Try to fit anchor + " " + badges on one line; wrap badges to a
		// second line if they'd overflow the content width.
		oneLine := lipgloss.JoinHorizontal(lipgloss.Top, anchorRendered, " ", badgeRow)
		if lipgloss.Width(oneLine) <= m.getIndentedContentWidth() {
			header = oneLine
		} else {
			header = lipgloss.JoinVertical(lipgloss.Left, anchorRendered, badgeRow)
		}
	}

	hunk := ""
	if len(t.Comments) > 0 && t.Comments[0].DiffHunk != "" {
		rendered, _ := r.Render("```diff\n" + previewDiffHunk(t.Comments[0].DiffHunk) + "\n```")
		hunk = rendered
	}

	commentParts := make([]string, 0, len(t.Comments))
	for i, c := range t.Comments {
		if i == 0 {
			commentParts = append(commentParts, m.renderReviewComment(c, r))
		} else {
			commentParts = append(commentParts, m.renderReviewCommentReply(c, r))
		}
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		append([]string{header, hunk}, commentParts...)...,
	)
}

func (m *Model) renderReviewComment(c data.ReviewComment, r glamour.TermRenderer) string {
	pill := m.renderAuthorTimePill(c.Author.Login, c.CreatedAt)
	body, _ := r.Render(c.Body)
	return lipgloss.JoinVertical(lipgloss.Left, pill, body)
}

// renderReviewCommentReply renders a non-opener comment in a thread with a
// compact "↳ {author} {time}" prefix and no pill, distinguishing it from the
// thread opener and saving vertical space when threads are deep.
func (m *Model) renderReviewCommentReply(c data.ReviewComment, r glamour.TermRenderer) string {
	prefix := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render("↳ "),
		m.ctx.Styles.Common.MainTextStyle.Render(c.Author.Login),
		" ",
		lipgloss.NewStyle().
			Foreground(m.ctx.Theme.FaintText).
			Render(utils.TimeElapsed(c.CreatedAt)),
	)
	body, _ := r.Render(c.Body)
	return lipgloss.JoinVertical(lipgloss.Left, prefix, body)
}

func renderEmptyState() string {
	return lipgloss.NewStyle().Italic(true).Render("No comments...")
}

func (m *Model) renderReviewHeader(review data.Review) string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		m.renderReviewDecision(review.State),
		" ",
		m.ctx.Styles.Common.MainTextStyle.Render(review.Author.Login),
		" ",
		lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render(
			"reviewed "+utils.TimeElapsed(review.SubmittedAt)),
	)
}

func (m *Model) renderReviewDecision(decision string) string {
	switch decision {
	case "PENDING":
		return m.ctx.Styles.Common.WaitingGlyph
	case "COMMENTED":
		return lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render("󰈈")
	case "APPROVED":
		return m.ctx.Styles.Common.SuccessGlyph
	case "CHANGES_REQUESTED":
		return m.ctx.Styles.Common.FailureGlyph
	case "DISMISSED":
		return lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render("↺")
	}
	return ""
}
