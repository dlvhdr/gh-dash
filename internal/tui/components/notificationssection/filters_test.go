package notificationssection

import (
	"testing"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

func TestParseNotificationFilters(t *testing.T) {
	tests := []struct {
		name                  string
		search                string
		wantReadState         data.NotificationReadState
		wantIsDone            bool
		wantExplicitUnread    bool
		wantIncludeBookmarked bool
		wantRepoCount         int
	}{
		{
			name:                  "empty search defaults to unread with bookmarks",
			search:                "",
			wantReadState:         data.NotificationStateUnread,
			wantIsDone:            false,
			wantExplicitUnread:    false,
			wantIncludeBookmarked: true,
			wantRepoCount:         0,
		},
		{
			name:                  "explicit is:unread excludes bookmarks",
			search:                "is:unread",
			wantReadState:         data.NotificationStateUnread,
			wantIsDone:            false,
			wantExplicitUnread:    true,
			wantIncludeBookmarked: false,
			wantRepoCount:         0,
		},
		{
			name:                  "is:read excludes bookmarks",
			search:                "is:read",
			wantReadState:         data.NotificationStateRead,
			wantIsDone:            false,
			wantExplicitUnread:    false,
			wantIncludeBookmarked: false,
			wantRepoCount:         0,
		},
		{
			name:                  "is:all excludes bookmarks",
			search:                "is:all",
			wantReadState:         data.NotificationStateAll,
			wantIsDone:            false,
			wantExplicitUnread:    false,
			wantIncludeBookmarked: false,
			wantRepoCount:         0,
		},
		{
			name:                  "is:done sets IsDone flag",
			search:                "is:done",
			wantReadState:         data.NotificationStateUnread,
			wantIsDone:            true,
			wantExplicitUnread:    false,
			wantIncludeBookmarked: true,
			wantRepoCount:         0,
		},
		{
			name:                  "is:unread and is:read together becomes is:all",
			search:                "is:unread is:read",
			wantReadState:         data.NotificationStateAll,
			wantIsDone:            false,
			wantExplicitUnread:    false,
			wantIncludeBookmarked: false,
			wantRepoCount:         0,
		},
		{
			name:                  "repo filter is extracted",
			search:                "repo:owner/repo",
			wantReadState:         data.NotificationStateUnread,
			wantIsDone:            false,
			wantExplicitUnread:    false,
			wantIncludeBookmarked: true,
			wantRepoCount:         1,
		},
		{
			name:                  "multiple repo filters",
			search:                "repo:owner/repo1 repo:owner/repo2",
			wantReadState:         data.NotificationStateUnread,
			wantIsDone:            false,
			wantExplicitUnread:    false,
			wantIncludeBookmarked: true,
			wantRepoCount:         2,
		},
		{
			name:                  "combined filters",
			search:                "repo:owner/repo is:unread",
			wantReadState:         data.NotificationStateUnread,
			wantIsDone:            false,
			wantExplicitUnread:    true,
			wantIncludeBookmarked: false,
			wantRepoCount:         1,
		},
		{
			name:                  "random text preserves defaults",
			search:                "some random search text",
			wantReadState:         data.NotificationStateUnread,
			wantIsDone:            false,
			wantExplicitUnread:    false,
			wantIncludeBookmarked: true,
			wantRepoCount:         0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := parseNotificationFilters(tt.search, false)

			if filters.ReadState != tt.wantReadState {
				t.Errorf("ReadState = %v, want %v", filters.ReadState, tt.wantReadState)
			}
			if filters.IsDone != tt.wantIsDone {
				t.Errorf("IsDone = %v, want %v", filters.IsDone, tt.wantIsDone)
			}
			if filters.ExplicitUnread != tt.wantExplicitUnread {
				t.Errorf("ExplicitUnread = %v, want %v", filters.ExplicitUnread, tt.wantExplicitUnread)
			}
			if filters.IncludeBookmarked != tt.wantIncludeBookmarked {
				t.Errorf("IncludeBookmarked = %v, want %v", filters.IncludeBookmarked, tt.wantIncludeBookmarked)
			}
			if len(filters.RepoFilters) != tt.wantRepoCount {
				t.Errorf("RepoFilters count = %d, want %d", len(filters.RepoFilters), tt.wantRepoCount)
			}
		})
	}
}

func TestParseReasonFilters(t *testing.T) {
	tests := []struct {
		name     string
		search   string
		expected []string
	}{
		{
			name:     "no reason filter",
			search:   "is:unread",
			expected: []string{},
		},
		{
			name:     "single reason filter",
			search:   "reason:author",
			expected: []string{data.ReasonAuthor},
		},
		{
			name:     "multiple reason filters",
			search:   "reason:author reason:mention",
			expected: []string{data.ReasonAuthor, data.ReasonMention},
		},
		{
			name:   "reason:participating expands to multiple reasons",
			search: "reason:participating",
			expected: []string{
				data.ReasonAuthor,
				data.ReasonComment,
				data.ReasonMention,
				data.ReasonReviewRequested,
				data.ReasonAssign,
				data.ReasonStateChange,
			},
		},
		{
			name:     "reason:review-requested normalizes to review_requested",
			search:   "reason:review-requested",
			expected: []string{data.ReasonReviewRequested},
		},
		{
			name:     "reason:team-mention normalizes to team_mention",
			search:   "reason:team-mention",
			expected: []string{data.ReasonTeamMention},
		},
		{
			name:     "reason:ci-activity normalizes to ci_activity",
			search:   "reason:ci-activity",
			expected: []string{data.ReasonCIActivity},
		},
		{
			name:     "reason:security-alert normalizes to security_alert",
			search:   "reason:security-alert",
			expected: []string{data.ReasonSecurityAlert},
		},
		{
			name:     "reason:state-change normalizes to state_change",
			search:   "reason:state-change",
			expected: []string{data.ReasonStateChange},
		},
		{
			name:     "reason filter mixed with other text",
			search:   "is:unread reason:author some text",
			expected: []string{data.ReasonAuthor},
		},
		{
			name:     "direct reason values pass through",
			search:   "reason:subscribed reason:comment",
			expected: []string{data.ReasonSubscribed, data.ReasonComment},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseReasonFilters(tt.search)

			if len(result) != len(tt.expected) {
				t.Errorf("parseReasonFilters(%q) returned %d items, want %d", tt.search, len(result), len(tt.expected))
				t.Errorf("got: %v", result)
				t.Errorf("want: %v", tt.expected)
				return
			}

			for i, want := range tt.expected {
				if result[i] != want {
					t.Errorf("parseReasonFilters(%q)[%d] = %q, want %q", tt.search, i, result[i], want)
				}
			}
		})
	}
}

func TestParseNotificationFiltersWithReasons(t *testing.T) {
	tests := []struct {
		name            string
		search          string
		wantReasonCount int
	}{
		{
			name:            "no reason filter",
			search:          "",
			wantReasonCount: 0,
		},
		{
			name:            "single reason filter",
			search:          "reason:author",
			wantReasonCount: 1,
		},
		{
			name:            "reason:participating expands to 6 reasons",
			search:          "reason:participating",
			wantReasonCount: 6,
		},
		{
			name:            "combined with other filters",
			search:          "is:unread repo:owner/repo reason:mention",
			wantReasonCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := parseNotificationFilters(tt.search, false)
			if len(filters.ReasonFilters) != tt.wantReasonCount {
				t.Errorf("ReasonFilters count = %d, want %d", len(filters.ReasonFilters), tt.wantReasonCount)
			}
		})
	}
}

func TestParseRepoFilters(t *testing.T) {
	tests := []struct {
		name     string
		search   string
		expected []string
	}{
		{
			name:     "no repo filter",
			search:   "is:unread",
			expected: []string{},
		},
		{
			name:     "single repo filter",
			search:   "repo:owner/repo",
			expected: []string{"owner/repo"},
		},
		{
			name:     "multiple repo filters",
			search:   "repo:owner/repo1 repo:other/repo2",
			expected: []string{"owner/repo1", "other/repo2"},
		},
		{
			name:     "repo filter with hyphen",
			search:   "repo:my-org/my-repo",
			expected: []string{"my-org/my-repo"},
		},
		{
			name:     "repo filter mixed with other text",
			search:   "is:unread repo:owner/repo some text",
			expected: []string{"owner/repo"},
		},
		{
			name:     "repo filter with underscore",
			search:   "repo:my_org/my_repo",
			expected: []string{"my_org/my_repo"},
		},
		{
			name:     "repo filter with numbers",
			search:   "repo:org123/repo456",
			expected: []string{"org123/repo456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRepoFilters(tt.search)

			if len(result) != len(tt.expected) {
				t.Errorf("parseRepoFilters(%q) returned %d items, want %d", tt.search, len(result), len(tt.expected))
				return
			}

			for i, want := range tt.expected {
				if result[i] != want {
					t.Errorf("parseRepoFilters(%q)[%d] = %q, want %q", tt.search, i, result[i], want)
				}
			}
		})
	}
}

func TestParseNotificationFiltersEdgeCases(t *testing.T) {
	tests := []struct {
		name                  string
		search                string
		wantReadState         data.NotificationReadState
		wantIsDone            bool
		wantExplicitUnread    bool
		wantIncludeBookmarked bool
	}{
		{
			name:                  "whitespace only preserves defaults",
			search:                "   ",
			wantReadState:         data.NotificationStateUnread,
			wantIsDone:            false,
			wantExplicitUnread:    false,
			wantIncludeBookmarked: true,
		},
		{
			name:                  "mixed case is:UNREAD not recognized (case sensitive)",
			search:                "is:UNREAD",
			wantReadState:         data.NotificationStateUnread,
			wantIsDone:            false,
			wantExplicitUnread:    false, // Not recognized due to case
			wantIncludeBookmarked: true,  // Default when not recognized
		},
		{
			name:                  "mixed case is:Read not recognized (case sensitive)",
			search:                "is:Read",
			wantReadState:         data.NotificationStateUnread, // Default when not recognized
			wantIsDone:            false,
			wantExplicitUnread:    false,
			wantIncludeBookmarked: true, // Default when not recognized
		},
		{
			name:                  "mixed case is:ALL not recognized (case sensitive)",
			search:                "is:ALL",
			wantReadState:         data.NotificationStateUnread, // Default when not recognized
			wantIsDone:            false,
			wantExplicitUnread:    false,
			wantIncludeBookmarked: true, // Default when not recognized
		},
		{
			name:                  "mixed case is:Done not recognized (case sensitive)",
			search:                "is:Done",
			wantReadState:         data.NotificationStateUnread,
			wantIsDone:            false, // Not recognized due to case
			wantExplicitUnread:    false,
			wantIncludeBookmarked: true,
		},
		{
			name:                  "multiple spaces between filters",
			search:                "is:unread    repo:owner/repo",
			wantReadState:         data.NotificationStateUnread,
			wantIsDone:            false,
			wantExplicitUnread:    true,
			wantIncludeBookmarked: false,
		},
		{
			name:                  "filter at end of string",
			search:                "some text is:read",
			wantReadState:         data.NotificationStateRead,
			wantIsDone:            false,
			wantExplicitUnread:    false,
			wantIncludeBookmarked: false,
		},
		{
			name:                  "is:done with is:all",
			search:                "is:done is:all",
			wantReadState:         data.NotificationStateAll,
			wantIsDone:            true,
			wantExplicitUnread:    false,
			wantIncludeBookmarked: false,
		},
		{
			name:                  "duplicate is:unread",
			search:                "is:unread is:unread",
			wantReadState:         data.NotificationStateUnread,
			wantIsDone:            false,
			wantExplicitUnread:    true,
			wantIncludeBookmarked: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := parseNotificationFilters(tt.search, false)

			if filters.ReadState != tt.wantReadState {
				t.Errorf("ReadState = %v, want %v", filters.ReadState, tt.wantReadState)
			}
			if filters.IsDone != tt.wantIsDone {
				t.Errorf("IsDone = %v, want %v", filters.IsDone, tt.wantIsDone)
			}
			if filters.ExplicitUnread != tt.wantExplicitUnread {
				t.Errorf("ExplicitUnread = %v, want %v", filters.ExplicitUnread, tt.wantExplicitUnread)
			}
			if filters.IncludeBookmarked != tt.wantIncludeBookmarked {
				t.Errorf("IncludeBookmarked = %v, want %v", filters.IncludeBookmarked, tt.wantIncludeBookmarked)
			}
		})
	}
}

func TestParseReasonFiltersEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		search   string
		expected []string
	}{
		{
			name:     "reason with mixed case passes through as-is",
			search:   "reason:AUTHOR",
			expected: []string{"AUTHOR"}, // Case is preserved, not normalized to lowercase
		},
		{
			name:   "reason:participating with other filters",
			search: "is:unread reason:participating repo:owner/repo",
			expected: []string{
				data.ReasonAuthor,
				data.ReasonComment,
				data.ReasonMention,
				data.ReasonReviewRequested,
				data.ReasonAssign,
				data.ReasonStateChange,
			},
		},
		{
			name:     "duplicate reason filters",
			search:   "reason:author reason:author",
			expected: []string{data.ReasonAuthor, data.ReasonAuthor},
		},
		{
			name:     "empty reason value",
			search:   "reason:",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseReasonFilters(tt.search)

			if len(result) != len(tt.expected) {
				t.Errorf("parseReasonFilters(%q) returned %d items, want %d", tt.search, len(result), len(tt.expected))
				t.Errorf("got: %v", result)
				t.Errorf("want: %v", tt.expected)
				return
			}

			for i, want := range tt.expected {
				if result[i] != want {
					t.Errorf("parseReasonFilters(%q)[%d] = %q, want %q", tt.search, i, result[i], want)
				}
			}
		})
	}
}

func TestParseNotificationFiltersIncludeRead(t *testing.T) {
	tests := []struct {
		name                  string
		search                string
		includeRead           bool
		wantReadState         data.NotificationReadState
		wantIncludeBookmarked bool
		wantExplicitUnread    bool
	}{
		{
			name:                  "includeRead true with empty search defaults to all",
			search:                "",
			includeRead:           true,
			wantReadState:         data.NotificationStateAll,
			wantIncludeBookmarked: false,
			wantExplicitUnread:    false,
		},
		{
			name:                  "includeRead false with empty search defaults to unread",
			search:                "",
			includeRead:           false,
			wantReadState:         data.NotificationStateUnread,
			wantIncludeBookmarked: true,
			wantExplicitUnread:    false,
		},
		{
			name:                  "includeRead true with explicit is:unread overrides to unread",
			search:                "is:unread",
			includeRead:           true,
			wantReadState:         data.NotificationStateUnread,
			wantIncludeBookmarked: false,
			wantExplicitUnread:    true,
		},
		{
			name:                  "includeRead true with explicit is:read overrides to read",
			search:                "is:read",
			includeRead:           true,
			wantReadState:         data.NotificationStateRead,
			wantIncludeBookmarked: false,
			wantExplicitUnread:    false,
		},
		{
			name:                  "includeRead true with random text defaults to all",
			search:                "some random text",
			includeRead:           true,
			wantReadState:         data.NotificationStateAll,
			wantIncludeBookmarked: false,
			wantExplicitUnread:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := parseNotificationFilters(tt.search, tt.includeRead)

			if filters.ReadState != tt.wantReadState {
				t.Errorf("ReadState = %v, want %v", filters.ReadState, tt.wantReadState)
			}
			if filters.IncludeBookmarked != tt.wantIncludeBookmarked {
				t.Errorf("IncludeBookmarked = %v, want %v", filters.IncludeBookmarked, tt.wantIncludeBookmarked)
			}
			if filters.ExplicitUnread != tt.wantExplicitUnread {
				t.Errorf("ExplicitUnread = %v, want %v", filters.ExplicitUnread, tt.wantExplicitUnread)
			}
		})
	}
}
