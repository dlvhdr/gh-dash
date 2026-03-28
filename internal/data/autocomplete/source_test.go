package autocomplete

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func TestExtractLabelAtCursor(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		cursorPos   tea.Position
		wantLabel   string
		wantStart   tea.Position
		wantEnd     tea.Position
		wantIsFirst bool
		wantIsLast  bool
	}{
		{
			name:        "empty input",
			input:       "",
			cursorPos:   tea.Position{},
			wantLabel:   "",
			wantStart:   tea.Position{X: 0},
			wantEnd:     tea.Position{X: 0},
			wantIsFirst: true,
			wantIsLast:  true,
		},
		{
			name:        "single label, cursor at start",
			input:       "bug",
			cursorPos:   tea.Position{},
			wantLabel:   "bug",
			wantStart:   tea.Position{X: 0},
			wantEnd:     tea.Position{X: 3},
			wantIsFirst: true,
			wantIsLast:  true,
		},
		{
			name:        "single label, cursor at end",
			input:       "bug",
			cursorPos:   tea.Position{X: 3},
			wantLabel:   "bug",
			wantStart:   tea.Position{X: 0},
			wantEnd:     tea.Position{X: 3},
			wantIsFirst: true,
			wantIsLast:  true,
		},
		{
			name:        "multiple labels, cursor on first",
			input:       "bug, feature",
			cursorPos:   tea.Position{X: 2},
			wantLabel:   "bug",
			wantStart:   tea.Position{X: 0},
			wantEnd:     tea.Position{X: 3},
			wantIsFirst: true,
			wantIsLast:  false,
		},
		{
			name:        "multiple labels, cursor on second",
			input:       "bug, feature",
			cursorPos:   tea.Position{X: 8},
			wantLabel:   "feature",
			wantStart:   tea.Position{X: 4},
			wantEnd:     tea.Position{X: 12},
			wantIsFirst: false,
			wantIsLast:  true,
		},
		{
			name:        "three labels, cursor on middle",
			input:       "bug, feature, docs",
			cursorPos:   tea.Position{X: 10},
			wantLabel:   "feature",
			wantStart:   tea.Position{X: 4},
			wantEnd:     tea.Position{X: 12},
			wantIsFirst: false,
			wantIsLast:  false,
		},
		{
			name:        "unicode label",
			input:       "🔴-bug",
			cursorPos:   tea.Position{X: 3},
			wantLabel:   "🔴-bug",
			wantStart:   tea.Position{X: 0},
			wantEnd:     tea.Position{X: 5},
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
		cursorPos tea.Position
		want      []string
	}{
		{name: "empty input", input: "", cursorPos: tea.Position{X: 0}, want: nil},
		{
			name:      "single label - nothing to exclude",
			input:     "bug",
			cursorPos: tea.Position{X: 0},
			want:      []string{},
		},
		{
			name:      "two labels, first is current",
			input:     "bug, feature",
			cursorPos: tea.Position{X: 0},
			want:      []string{"feature"},
		},
		{
			name:      "two labels, second is current",
			input:     "bug, feature",
			cursorPos: tea.Position{X: 5},
			want:      []string{"bug"},
		},
		{
			name:      "three labels, middle is current",
			input:     "bug, feature, docs",
			cursorPos: tea.Position{X: 8},
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
		contextStart  tea.Position
		contextEnd    tea.Position
		wantNewInput  string
		wantNewCursor tea.Position
	}{
		{
			name:          "replace first label",
			input:         "bu, feature, docs",
			suggestion:    "bug",
			contextStart:  tea.Position{X: 0},
			contextEnd:    tea.Position{X: 2},
			wantNewInput:  "bug, feature, docs",
			wantNewCursor: tea.Position{X: 5},
		},
		{
			name:          "replace middle label",
			input:         "bug, fea, docs",
			suggestion:    "feature",
			contextStart:  tea.Position{X: 5},
			contextEnd:    tea.Position{X: 8},
			wantNewInput:  "bug, feature, docs",
			wantNewCursor: tea.Position{X: 14},
		},
		{
			name:          "replace last label",
			input:         "bug, feature, do",
			suggestion:    "docs",
			contextStart:  tea.Position{X: 14},
			contextEnd:    tea.Position{X: 16},
			wantNewInput:  "bug, feature, docs, ",
			wantNewCursor: tea.Position{X: 20},
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
	ctx := source.ExtractContext("foo bar baz", tea.Position{X: 5})
	require.Equal(
		t,
		Context{Start: tea.Position{X: 4}, End: tea.Position{X: 7}, Content: "bar"},
		ctx,
	)
	require.Equal(
		t,
		[]string{"foo", "baz"},
		source.ItemsToExclude("foo bar baz", tea.Position{X: 5}),
	)

	newInput, newCursor := source.InsertSuggestion(
		"foo ba baz",
		"bar",
		tea.Position{X: 4},
		tea.Position{X: 6},
	)
	require.Equal(t, "foo bar  baz", newInput)
	require.Equal(t, tea.Position{X: 8}, newCursor)
}

func TestUserMentionSource(t *testing.T) {
	source := UserMentionSource{}
	require.Equal(t, Context{}, source.ExtractContext("hello world", tea.Position{X: 5}))
	require.Equal(
		t,
		Context{Start: tea.Position{X: 6}, End: tea.Position{X: 7}, Content: ""},
		source.ExtractContext("hello @", tea.Position{X: 7}),
	)
	require.Equal(
		t,
		Context{Start: tea.Position{X: 0}, End: tea.Position{X: 1}, Content: ""},
		source.ExtractContext("@", tea.Position{X: 1}),
	)
	require.Equal(
		t,
		Context{Start: tea.Position{X: 6}, End: tea.Position{X: 11}, Content: "octo"},
		source.ExtractContext("hello @octo", tea.Position{X: 9}),
	)
	require.Equal(
		t,
		Context{},
		source.ExtractContext("hello @octo ", tea.Position{X: 12}),
	)

	require.Equal(
		t,
		Context{},
		source.ExtractContext("hello @octo"+string('\n'), tea.Position{Y: 1, X: 0}),
	)

	newInput, newCursor := source.InsertSuggestion(
		"hello @oc",
		"octo",
		tea.Position{X: 6},
		tea.Position{X: 9},
	)
	require.Equal(t, "hello @octo ", newInput)
	require.Equal(t, tea.Position{X: 12}, newCursor)
	require.Nil(t, source.ItemsToExclude("hello @oc", tea.Position{X: 8}))
}

func TestUserMentionSourceWithNewLines(t *testing.T) {
	source := UserMentionSource{}
	newInput, newCursor := source.InsertSuggestion(
		`hello @octo

yes
@oc`,
		"octo",
		tea.Position{Y: 3, X: 0},
		tea.Position{Y: 3, X: 3},
	)
	require.Equal(t, `hello @octo

yes
@octo `, newInput)
	require.Equal(t, tea.Position{Y: 0, X: 6}, newCursor)
}
