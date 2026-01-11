package notificationrow

import (
	"testing"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

func TestExtractNumberFromUrl(t *testing.T) {
	tests := []struct {
		name     string
		apiUrl   string
		expected string
	}{
		{
			name:     "empty URL returns empty string",
			apiUrl:   "",
			expected: "",
		},
		{
			name:     "PR API URL extracts number",
			apiUrl:   "https://api.github.com/repos/owner/repo/pulls/123",
			expected: "123",
		},
		{
			name:     "Issue API URL extracts number",
			apiUrl:   "https://api.github.com/repos/owner/repo/issues/456",
			expected: "456",
		},
		{
			name:     "Discussion API URL extracts number",
			apiUrl:   "https://api.github.com/repos/owner/repo/discussions/789",
			expected: "789",
		},
		{
			name:     "URL with no slashes returns empty",
			apiUrl:   "no-slashes",
			expected: "",
		},
		{
			name:     "URL ending with slash returns empty",
			apiUrl:   "https://api.github.com/repos/owner/repo/",
			expected: "",
		},
		{
			name:     "single segment after slash",
			apiUrl:   "/123",
			expected: "123",
		},
		{
			name:     "large number",
			apiUrl:   "https://api.github.com/repos/owner/repo/pulls/999999",
			expected: "999999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractNumberFromUrl(tt.apiUrl)
			if result != tt.expected {
				t.Errorf("extractNumberFromUrl(%q) = %q, want %q", tt.apiUrl, result, tt.expected)
			}
		})
	}
}

func TestGetNumber(t *testing.T) {
	tests := []struct {
		name     string
		data     Data
		expected int
	}{
		{
			name: "PullRequest returns extracted number",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Type: "PullRequest",
						Url:  "https://api.github.com/repos/owner/repo/pulls/123",
					},
				},
			},
			expected: 123,
		},
		{
			name: "Issue returns extracted number",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Type: "Issue",
						Url:  "https://api.github.com/repos/owner/repo/issues/456",
					},
				},
			},
			expected: 456,
		},
		{
			name: "Discussion returns 0",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Type: "Discussion",
						Url:  "https://api.github.com/repos/owner/repo/discussions/789",
					},
				},
			},
			expected: 0,
		},
		{
			name: "Release returns 0",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Type: "Release",
						Url:  "https://api.github.com/repos/owner/repo/releases/v1.0.0",
					},
				},
			},
			expected: 0,
		},
		{
			name: "empty URL returns 0",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Type: "PullRequest",
						Url:  "",
					},
				},
			},
			expected: 0,
		},
		{
			name: "non-numeric segment returns 0",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Type: "PullRequest",
						Url:  "https://api.github.com/repos/owner/repo/pulls/abc",
					},
				},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.data.GetNumber()
			if result != tt.expected {
				t.Errorf("GetNumber() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestGetUrl(t *testing.T) {
	tests := []struct {
		name     string
		data     Data
		expected string
	}{
		{
			name: "PullRequest constructs correct URL",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Type: "PullRequest",
						Url:  "https://api.github.com/repos/owner/repo/pulls/123",
					},
					Repository: data.NotificationRepository{
						FullName: "owner/repo",
					},
				},
			},
			expected: "https://github.com/owner/repo/pull/123",
		},
		{
			name: "Issue constructs correct URL",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Type: "Issue",
						Url:  "https://api.github.com/repos/owner/repo/issues/456",
					},
					Repository: data.NotificationRepository{
						FullName: "owner/repo",
					},
				},
			},
			expected: "https://github.com/owner/repo/issues/456",
		},
		{
			name: "Discussion with number constructs correct URL",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Type: "Discussion",
						Url:  "https://api.github.com/repos/owner/repo/discussions/789",
					},
					Repository: data.NotificationRepository{
						FullName: "owner/repo",
					},
				},
			},
			expected: "https://github.com/owner/repo/discussions/789",
		},
		{
			name: "Discussion with empty URL falls back to discussions page",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Type: "Discussion",
						Url:  "",
					},
					Repository: data.NotificationRepository{
						FullName: "owner/repo",
					},
				},
			},
			expected: "https://github.com/owner/repo/discussions",
		},
		{
			name: "Release constructs releases URL",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Type: "Release",
						Url:  "https://api.github.com/repos/owner/repo/releases/12345",
					},
					Repository: data.NotificationRepository{
						FullName: "owner/repo",
					},
				},
			},
			expected: "https://github.com/owner/repo/releases",
		},
		{
			name: "Commit constructs commits URL",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Type: "Commit",
						Url:  "https://api.github.com/repos/owner/repo/commits/abc123",
					},
					Repository: data.NotificationRepository{
						FullName: "owner/repo",
					},
				},
			},
			expected: "https://github.com/owner/repo/commits",
		},
		{
			name: "CheckSuite links to actions page when ResolvedUrl not set",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Type: "CheckSuite",
						Url:  "", // GitHub API returns null for CheckSuite subject.url
					},
					Repository: data.NotificationRepository{
						FullName: "owner/repo",
					},
				},
			},
			expected: "https://github.com/owner/repo/actions",
		},
		{
			name: "CheckSuite uses ResolvedUrl when available",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Type: "CheckSuite",
						Url:  "", // GitHub API returns null for CheckSuite subject.url
					},
					Repository: data.NotificationRepository{
						FullName: "owner/repo",
					},
				},
				ResolvedUrl: "https://github.com/owner/repo/actions/runs/12345678",
			},
			expected: "https://github.com/owner/repo/actions/runs/12345678",
		},
		{
			name: "unknown type falls back to repo URL",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Type: "UnknownType",
						Url:  "https://api.github.com/repos/owner/repo/something",
					},
					Repository: data.NotificationRepository{
						FullName: "owner/repo",
					},
				},
			},
			expected: "https://github.com/owner/repo",
		},
		{
			name: "handles org/repo with special characters",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Type: "PullRequest",
						Url:  "https://api.github.com/repos/my-org/my-repo/pulls/1",
					},
					Repository: data.NotificationRepository{
						FullName: "my-org/my-repo",
					},
				},
			},
			expected: "https://github.com/my-org/my-repo/pull/1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.data.GetUrl()
			if result != tt.expected {
				t.Errorf("GetUrl() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestIsUnread(t *testing.T) {
	tests := []struct {
		name     string
		data     Data
		expected bool
	}{
		{
			name: "unread notification returns true",
			data: Data{
				Notification: data.NotificationData{
					Unread: true,
				},
			},
			expected: true,
		},
		{
			name: "read notification returns false",
			data: Data{
				Notification: data.NotificationData{
					Unread: false,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.data.IsUnread()
			if result != tt.expected {
				t.Errorf("IsUnread() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetReason(t *testing.T) {
	tests := []struct {
		name     string
		data     Data
		expected string
	}{
		{
			name: "returns comment reason",
			data: Data{
				Notification: data.NotificationData{
					Reason: "comment",
				},
			},
			expected: "comment",
		},
		{
			name: "returns mention reason",
			data: Data{
				Notification: data.NotificationData{
					Reason: "mention",
				},
			},
			expected: "mention",
		},
		{
			name: "returns subscribed reason",
			data: Data{
				Notification: data.NotificationData{
					Reason: "subscribed",
				},
			},
			expected: "subscribed",
		},
		{
			name: "returns review_requested reason",
			data: Data{
				Notification: data.NotificationData{
					Reason: "review_requested",
				},
			},
			expected: "review_requested",
		},
		{
			name: "returns empty string when not set",
			data: Data{
				Notification: data.NotificationData{
					Reason: "",
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.data.GetReason()
			if result != tt.expected {
				t.Errorf("GetReason() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetLatestCommentUrl(t *testing.T) {
	tests := []struct {
		name     string
		data     Data
		expected string
	}{
		{
			name: "returns latest comment URL when set",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						LatestCommentUrl: "https://api.github.com/repos/owner/repo/issues/comments/123456",
					},
				},
			},
			expected: "https://api.github.com/repos/owner/repo/issues/comments/123456",
		},
		{
			name: "returns empty string when not set",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						LatestCommentUrl: "",
					},
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.data.GetLatestCommentUrl()
			if result != tt.expected {
				t.Errorf("GetLatestCommentUrl() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetTitle(t *testing.T) {
	tests := []struct {
		name     string
		data     Data
		expected string
	}{
		{
			name: "returns subject title",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Title: "Fix bug in authentication",
					},
				},
			},
			expected: "Fix bug in authentication",
		},
		{
			name: "returns empty string when not set",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Title: "",
					},
				},
			},
			expected: "",
		},
		{
			name: "strips trailing carriage return",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Title: "Set creation URL of worker settings\r",
					},
				},
			},
			expected: "Set creation URL of worker settings",
		},
		{
			name: "strips embedded carriage returns",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Title: "Line one\r\nLine two",
					},
				},
			},
			expected: "Line one Line two",
		},
		{
			name: "trims whitespace",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Title: "  Title with spaces  ",
					},
				},
			},
			expected: "Title with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.data.GetTitle()
			if result != tt.expected {
				t.Errorf("GetTitle() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetRepoNameWithOwner(t *testing.T) {
	tests := []struct {
		name     string
		data     Data
		expected string
	}{
		{
			name: "returns full repository name",
			data: Data{
				Notification: data.NotificationData{
					Repository: data.NotificationRepository{
						FullName: "owner/repo",
					},
				},
			},
			expected: "owner/repo",
		},
		{
			name: "returns org/repo format",
			data: Data{
				Notification: data.NotificationData{
					Repository: data.NotificationRepository{
						FullName: "my-organization/my-repository",
					},
				},
			},
			expected: "my-organization/my-repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.data.GetRepoNameWithOwner()
			if result != tt.expected {
				t.Errorf("GetRepoNameWithOwner() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetSubjectType(t *testing.T) {
	tests := []struct {
		name     string
		data     Data
		expected string
	}{
		{
			name: "returns PullRequest type",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Type: "PullRequest",
					},
				},
			},
			expected: "PullRequest",
		},
		{
			name: "returns Issue type",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Type: "Issue",
					},
				},
			},
			expected: "Issue",
		},
		{
			name: "returns Discussion type",
			data: Data{
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Type: "Discussion",
					},
				},
			},
			expected: "Discussion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.data.GetSubjectType()
			if result != tt.expected {
				t.Errorf("GetSubjectType() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetId(t *testing.T) {
	tests := []struct {
		name     string
		data     Data
		expected string
	}{
		{
			name: "returns notification ID",
			data: Data{
				Notification: data.NotificationData{
					Id: "123456789",
				},
			},
			expected: "123456789",
		},
		{
			name: "returns empty string when not set",
			data: Data{
				Notification: data.NotificationData{
					Id: "",
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.data.GetId()
			if result != tt.expected {
				t.Errorf("GetId() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGenerateActivityDescription(t *testing.T) {
	tests := []struct {
		name        string
		reason      string
		subjectType string
		actor       string
		expected    string
	}{
		// comment reason
		{
			name:        "comment on PR with actor",
			reason:      "comment",
			subjectType: "PullRequest",
			actor:       "octocat",
			expected:    "@octocat commented on this pull request",
		},
		{
			name:        "comment on Issue with actor",
			reason:      "comment",
			subjectType: "Issue",
			actor:       "octocat",
			expected:    "@octocat commented on this issue",
		},
		{
			name:        "comment on other type with actor",
			reason:      "comment",
			subjectType: "Discussion",
			actor:       "octocat",
			expected:    "@octocat commented",
		},
		{
			name:        "comment without actor",
			reason:      "comment",
			subjectType: "PullRequest",
			actor:       "",
			expected:    "New comment",
		},

		// review_requested reason
		{
			name:        "review requested with actor",
			reason:      "review_requested",
			subjectType: "PullRequest",
			actor:       "reviewer",
			expected:    "@reviewer requested your review",
		},
		{
			name:        "review requested without actor",
			reason:      "review_requested",
			subjectType: "PullRequest",
			actor:       "",
			expected:    "Review requested",
		},

		// mention reason
		{
			name:        "mention with actor",
			reason:      "mention",
			subjectType: "Issue",
			actor:       "commenter",
			expected:    "@commenter mentioned you",
		},
		{
			name:        "mention without actor",
			reason:      "mention",
			subjectType: "Issue",
			actor:       "",
			expected:    "You were mentioned",
		},

		// author reason
		{
			name:        "author reason",
			reason:      "author",
			subjectType: "PullRequest",
			actor:       "someone",
			expected:    "Activity on your thread",
		},

		// assign reason
		{
			name:        "assign reason",
			reason:      "assign",
			subjectType: "Issue",
			actor:       "assigner",
			expected:    "You were assigned",
		},

		// state_change reason
		{
			name:        "state change on PR",
			reason:      "state_change",
			subjectType: "PullRequest",
			actor:       "",
			expected:    "Pull request state changed",
		},
		{
			name:        "state change on Issue",
			reason:      "state_change",
			subjectType: "Issue",
			actor:       "",
			expected:    "Issue state changed",
		},
		{
			name:        "state change on other type",
			reason:      "state_change",
			subjectType: "Discussion",
			actor:       "",
			expected:    "State changed",
		},

		// ci_activity reason
		{
			name:        "CI activity",
			reason:      "ci_activity",
			subjectType: "CheckSuite",
			actor:       "",
			expected:    "CI activity",
		},

		// subscribed reason
		{
			name:        "subscribed PR with actor",
			reason:      "subscribed",
			subjectType: "PullRequest",
			actor:       "contributor",
			expected:    "@contributor commented on this pull request",
		},
		{
			name:        "subscribed Issue with actor",
			reason:      "subscribed",
			subjectType: "Issue",
			actor:       "contributor",
			expected:    "@contributor commented on this issue",
		},
		{
			name:        "subscribed other type with actor",
			reason:      "subscribed",
			subjectType: "Discussion",
			actor:       "contributor",
			expected:    "Activity on subscribed thread",
		},
		{
			name:        "subscribed without actor",
			reason:      "subscribed",
			subjectType: "PullRequest",
			actor:       "",
			expected:    "Activity on subscribed thread",
		},

		// team_mention reason
		{
			name:        "team mention",
			reason:      "team_mention",
			subjectType: "Issue",
			actor:       "",
			expected:    "Your team was mentioned",
		},

		// security_alert reason
		{
			name:        "security alert",
			reason:      "security_alert",
			subjectType: "RepositoryVulnerabilityAlert",
			actor:       "",
			expected:    "Security vulnerability detected",
		},

		// unknown reason
		{
			name:        "unknown reason with actor",
			reason:      "unknown_reason",
			subjectType: "PullRequest",
			actor:       "someone",
			expected:    "@someone triggered this notification",
		},
		{
			name:        "unknown reason without actor",
			reason:      "unknown_reason",
			subjectType: "PullRequest",
			actor:       "",
			expected:    "",
		},
		{
			name:        "empty reason without actor",
			reason:      "",
			subjectType: "PullRequest",
			actor:       "",
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateActivityDescription(tt.reason, tt.subjectType, tt.actor)
			if result != tt.expected {
				t.Errorf("GenerateActivityDescription(%q, %q, %q) = %q, want %q",
					tt.reason, tt.subjectType, tt.actor, result, tt.expected)
			}
		})
	}
}
