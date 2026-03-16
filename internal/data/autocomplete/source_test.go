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
			wantStart:   4,
			wantEnd:     12,
			wantIsFirst: false,
			wantIsLast:  true,
		},
		{
			name:        "three labels, cursor on middle",
			input:       "bug, feature, docs",
			cursorPos:   10,
			wantLabel:   "feature",
			wantStart:   4,
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
			wantEnd:     5,
			wantIsFirst: true,
			wantIsLast:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := ExtractLabelAtCursor(tc.input, tc.cursorPos)
			require.Equal(t, tc.wantLabel, got.Label)
			require.Equal(t, tc.wantStart, got.StartIdx)
			require.Equal(t, tc.wantEnd, got.EndIdx)
			require.Equal(t, tc.wantIsFirst, got.IsFirst)
			require.Equal(t, tc.wantIsLast, got.IsLast)
		})
	}
}

func TestLabelSourceItemsToExclude(t *testing.T) {
	source := LabelSource{}
	testCases := []struct {
		name      string
		input     string
		cursorPos int
		want      []string
	}{
		{name: "empty input", input: "", cursorPos: 0, want: nil},
		{name: "single label - nothing to exclude", input: "bug", cursorPos: 0, want: []string{}},
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
			cursorPos: 8,
			want:      []string{"bug", "docs"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, source.ItemsToExclude(tc.input, tc.cursorPos))
		})
	}
}

func TestLabelSourceInsertSuggestion(t *testing.T) {
	source := LabelSource{}
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
			wantNewCursor: 5,
		},
		{
			name:          "replace middle label",
			input:         "bug, fea, docs",
			suggestion:    "feature",
			contextStart:  5,
			contextEnd:    8,
			wantNewInput:  "bug, feature, docs",
			wantNewCursor: 14,
		},
		{
			name:          "replace last label",
			input:         "bug, feature, do",
			suggestion:    "docs",
			contextStart:  14,
			contextEnd:    16,
			wantNewInput:  "bug, feature, docs, ",
			wantNewCursor: 20,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotInput, gotCursor := source.InsertSuggestion(
				tc.input,
				tc.suggestion,
				tc.contextStart,
				tc.contextEnd,
			)
			require.Equal(t, tc.wantNewInput, gotInput)
			require.Equal(t, tc.wantNewCursor, gotCursor)
		})
	}
}

func TestWhitespaceSource(t *testing.T) {
	source := WhitespaceSource{}
	ctx := source.ExtractContext("foo bar baz", 5)
	require.Equal(t, Context{Start: 4, End: 7, Content: "bar"}, ctx)
	require.Equal(t, []string{"foo", "baz"}, source.ItemsToExclude("foo bar baz", 5))

	newInput, newCursor := source.InsertSuggestion("foo ba baz", "bar", 4, 6)
	require.Equal(t, "foo bar  baz", newInput)
	require.Equal(t, 8, newCursor)
}

func TestUserMentionSource(t *testing.T) {
	source := UserMentionSource{}
	require.Equal(t, Context{}, source.ExtractContext("hello world", 5))
	require.Equal(t, Context{Start: 6, End: 7, Content: ""}, source.ExtractContext("hello @", 7))
	require.Equal(t, Context{Start: 0, End: 1, Content: ""}, source.ExtractContext("@", 1))
	require.Equal(
		t,
		Context{Start: 6, End: 11, Content: "octo"},
		source.ExtractContext("hello @octo", 9),
	)

	newInput, newCursor := source.InsertSuggestion("hello @oc", "octo", 6, 9)
	require.Equal(t, "hello @octo ", newInput)
	require.Equal(t, 12, newCursor)
	require.Nil(t, source.ItemsToExclude("hello @oc", 8))
}
