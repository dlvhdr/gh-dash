package autocomplete

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractLabelAtCursor(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		cursorPos   int
		wantLabel   string
		wantStart   int
		wantEnd     int
		wantIsFirst bool
		wantIsLast  bool
	}{
		{
			name:        "empty input",
			input:       "",
			cursorPos:   0,
			wantLabel:   "",
			wantStart:   0,
			wantEnd:     0,
			wantIsFirst: true,
			wantIsLast:  true,
		},
		{
			name:        "single label, cursor at start",
			input:       "bug",
			cursorPos:   0,
			wantLabel:   "bug",
			wantStart:   0,
			wantEnd:     3,
			wantIsFirst: true,
			wantIsLast:  true,
		},
		{
			name:        "single label, cursor at end",
			input:       "bug",
			cursorPos:   3,
			wantLabel:   "bug",
			wantStart:   0,
			wantEnd:     3,
			wantIsFirst: true,
			wantIsLast:  true,
		},
		{
			name:        "multiple labels, cursor on first",
			input:       "bug, feature",
			cursorPos:   2,
			wantLabel:   "bug",
			wantStart:   0,
			wantEnd:     3,
			wantIsFirst: true,
			wantIsLast:  false,
		},
		{
			name:        "multiple labels, cursor on second",
			input:       "bug, feature",
			cursorPos:   8,
			wantLabel:   "feature",
			wantStart:   4, // Position right after the comma (includes space before)
			wantEnd:     12,
			wantIsFirst: false,
			wantIsLast:  true,
		},
		{
			name:        "labels with whitespace",
			input:       "bug, feature",
			cursorPos:   3,
			wantLabel:   "bug",
			wantStart:   0,
			wantEnd:     3,
			wantIsFirst: true,
			wantIsLast:  false,
		},
		{
			name:        "three labels, cursor on middle",
			input:       "bug, feature, docs",
			cursorPos:   10,
			wantLabel:   "feature",
			wantStart:   4, // Position right after the comma (includes space before)
			wantEnd:     12,
			wantIsFirst: false,
			wantIsLast:  false,
		},
		{
			name:        "unicode label",
			input:       "🔴-bug",
			cursorPos:   3,
			wantLabel:   "🔴-bug",
			wantStart:   0,
			wantEnd:     5, // 🔴 is 1 rune, then -, b, u, g = 5 runes total
			wantIsFirst: true,
			wantIsLast:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := ExtractLabelAtCursor(tc.input, tc.cursorPos)
			require.Equal(t, tc.wantLabel, got.Label, "label mismatch")
			require.Equal(t, tc.wantStart, got.StartIdx, "start index mismatch")
			require.Equal(t, tc.wantEnd, got.EndIdx, "end index mismatch")
			require.Equal(t, tc.wantIsFirst, got.IsFirst, "IsFirst mismatch")
			require.Equal(t, tc.wantIsLast, got.IsLast, "IsLast mismatch")
		})
	}
}

func TestLabelContextExtractor(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		cursorPos int
		wantCtx   string
		wantStart int
		wantEnd   int
	}{
		{
			name:      "first label",
			input:     "bug, feature",
			cursorPos: 2,
			wantCtx:   "bug",
			wantStart: 0,
			wantEnd:   3,
		},
		{
			name:      "second label",
			input:     "bug, feature",
			cursorPos: 8,
			wantCtx:   "feature",
			wantStart: 4,
			wantEnd:   12,
		},
		{
			name:      "empty input",
			input:     "",
			cursorPos: 0,
			wantCtx:   "",
			wantStart: 0,
			wantEnd:   0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, start, end := LabelContextExtractor(tc.input, tc.cursorPos)
			require.Equal(t, tc.wantCtx, ctx)
			require.Equal(t, tc.wantStart, start)
			require.Equal(t, tc.wantEnd, end)
		})
	}
}

func TestCurrentLabels(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
		{
			name:  "whitespace only",
			input: "   ",
			want:  nil,
		},
		{
			name:  "single label",
			input: "bug",
			want:  []string{"bug"},
		},
		{
			name:  "multiple labels",
			input: "bug, feature, docs",
			want:  []string{"bug", "feature", "docs"},
		},
		{
			name:  "labels with whitespace",
			input: "  bug  ,  feature  ,  docs  ",
			want:  []string{"bug", "feature", "docs"},
		},
		{
			name:  "empty label in middle",
			input: "bug,,feature",
			want:  []string{"bug", "feature"},
		},
		{
			name:  "trailing comma",
			input: "bug, feature,",
			want:  []string{"bug", "feature"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := CurrentLabels(tc.input)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestLabelLabelItemsToExclude(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		cursorPos int
		want      []string
	}{
		{
			name:      "empty input",
			input:     "",
			cursorPos: 0,
			want:      nil,
		},
		{
			name:      "single label - nothing to exclude",
			input:     "bug",
			cursorPos: 0,
			want:      []string{},
		},
		{
			name:      "two labels, first is current",
			input:     "bug, feature",
			cursorPos: 0,
			want:      []string{"feature"},
		},
		{
			name:      "two labels, second is current",
			input:     "bug, feature",
			cursorPos: 5,
			want:      []string{"bug"},
		},
		{
			name:      "three labels, middle is current",
			input:     "bug, feature, docs",
			cursorPos: 5,
			want:      []string{"bug", "docs"},
		},
		{
			name:      "with whitespace",
			input:     "  bug  ,  feature  ,  docs  ",
			cursorPos: 2,
			want:      []string{"feature", "docs"},
		},
		{
			name:      "two labels, second is current, cursor in middle",
			input:     "bug, feature",
			cursorPos: 8,
			want:      []string{"bug"},
		},
		{
			name:      "three labels, middle is current, cursor in middle",
			input:     "bug, feature, docs",
			cursorPos: 8,
			want:      []string{"bug", "docs"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := LabelItemsToExclude(tc.input, tc.cursorPos)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestLabelSuggestionInserter(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		suggestion    string
		contextStart  int
		contextEnd    int
		wantNewInput  string
		wantNewCursor int
	}{
		{
			name:          "replace first label",
			input:         "bu, feature, docs",
			suggestion:    "bug",
			contextStart:  0,
			contextEnd:    2,
			wantNewInput:  "bug, feature, docs",
			wantNewCursor: 5, // len("bug, ")
		},
		{
			name:          "replace middle label",
			input:         "bug, fea, docs",
			suggestion:    "feature",
			contextStart:  5,
			contextEnd:    8,
			wantNewInput:  "bug,  feature, docs", // Note: double space after comma
			wantNewCursor: 15,
		},
		{
			name:          "replace last label",
			input:         "bug, feature, do",
			suggestion:    "docs",
			contextStart:  14,
			contextEnd:    16,
			wantNewInput:  "bug, feature,  docs, ", // Note: double space and trailing comma
			wantNewCursor: 21,
		},
		{
			name:          "empty input",
			input:         "",
			suggestion:    "bug",
			contextStart:  0,
			contextEnd:    0,
			wantNewInput:  "bug, ",
			wantNewCursor: 5,
		},
		{
			name:          "single label partial",
			input:         "fe",
			suggestion:    "feature",
			contextStart:  0,
			contextEnd:    2,
			wantNewInput:  "feature, ",
			wantNewCursor: 9,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotInput, gotCursor := LabelSuggestionInserter(tc.input, tc.suggestion, tc.contextStart, tc.contextEnd)
			require.Equal(t, tc.wantNewInput, gotInput, "new input mismatch")
			require.Equal(t, tc.wantNewCursor, gotCursor, "cursor position mismatch")
		})
	}
}
