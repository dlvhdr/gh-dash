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
			filters := parseNotificationFilters(tt.search)

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
