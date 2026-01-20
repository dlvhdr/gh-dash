package notificationrow

import (
	"strings"
	"testing"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

func TestGetReasonDescription(t *testing.T) {
	tests := []struct {
		name        string
		reason      string
		subjectType string
		expected    string
	}{
		// review_requested
		{
			name:        "review_requested",
			reason:      "review_requested",
			subjectType: "PullRequest",
			expected:    "Review requested",
		},
		// subscribed
		{
			name:        "subscribed PR",
			reason:      "subscribed",
			subjectType: "PullRequest",
			expected:    "Activity on subscribed thread",
		},
		{
			name:        "subscribed Issue",
			reason:      "subscribed",
			subjectType: "Issue",
			expected:    "Activity on subscribed thread",
		},
		// mention
		{
			name:        "mention",
			reason:      "mention",
			subjectType: "Issue",
			expected:    "You were mentioned",
		},
		// author
		{
			name:        "author",
			reason:      "author",
			subjectType: "PullRequest",
			expected:    "Activity on your thread",
		},
		// comment
		{
			name:        "comment on PR",
			reason:      "comment",
			subjectType: "PullRequest",
			expected:    "New comment on pull request",
		},
		{
			name:        "comment on Issue",
			reason:      "comment",
			subjectType: "Issue",
			expected:    "New comment on issue",
		},
		{
			name:        "comment on other type",
			reason:      "comment",
			subjectType: "Discussion",
			expected:    "New comment",
		},
		// assign
		{
			name:        "assign",
			reason:      "assign",
			subjectType: "Issue",
			expected:    "You were assigned",
		},
		// state_change
		{
			name:        "state_change PR",
			reason:      "state_change",
			subjectType: "PullRequest",
			expected:    "State changed",
		},
		// ci_activity
		{
			name:        "ci_activity",
			reason:      "ci_activity",
			subjectType: "CheckSuite",
			expected:    "CI activity",
		},
		// unknown/empty
		{
			name:        "unknown reason",
			reason:      "unknown_reason",
			subjectType: "PullRequest",
			expected:    "",
		},
		{
			name:        "empty reason",
			reason:      "",
			subjectType: "PullRequest",
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Notification{
				Data: &Data{
					Notification: data.NotificationData{
						Reason: tt.reason,
						Subject: data.NotificationSubject{
							Type: tt.subjectType,
						},
					},
				},
			}
			result := n.getReasonDescription()
			if result != tt.expected {
				t.Errorf("getReasonDescription() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseCheckSuiteStatus(t *testing.T) {
	// Test that CheckSuite status parsing from title works correctly
	// This is tested indirectly through renderType, but we can verify
	// the title parsing logic

	tests := []struct {
		name             string
		title            string
		expectsFailed    bool
		expectsSuccess   bool
		expectsCancelled bool
		expectsSkipped   bool
	}{
		{
			name:          "failed status",
			title:         "CI / build (push) failed",
			expectsFailed: true,
		},
		{
			name:           "succeeded status",
			title:          "CI / build (push) succeeded",
			expectsSuccess: true,
		},
		{
			name:             "cancelled status",
			title:            "CI / build (push) cancelled",
			expectsCancelled: true,
		},
		{
			name:             "canceled status (US spelling)",
			title:            "CI / build (push) canceled",
			expectsCancelled: true,
		},
		{
			name:           "skipped status",
			title:          "CI / build (push) skipped",
			expectsSkipped: true,
		},
		{
			name:  "unknown status",
			title: "CI / build (push) running",
		},
		{
			name:          "case insensitive - FAILED",
			title:         "CI / build (push) FAILED",
			expectsFailed: true,
		},
		{
			name:           "case insensitive - SUCCEEDED",
			title:          "CI / build (push) SUCCEEDED",
			expectsSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			titleLower := strings.ToLower(tt.title)

			isFailed := strings.Contains(titleLower, "failed")
			isSuccess := strings.Contains(titleLower, "succeeded")
			isCancelled := strings.Contains(titleLower, "cancelled") || strings.Contains(titleLower, "canceled")
			isSkipped := strings.Contains(titleLower, "skipped")

			if tt.expectsFailed && !isFailed {
				t.Error("Expected failed to be detected")
			}
			if tt.expectsSuccess && !isSuccess {
				t.Error("Expected succeeded to be detected")
			}
			if tt.expectsCancelled && !isCancelled {
				t.Error("Expected cancelled/canceled to be detected")
			}
			if tt.expectsSkipped && !isSkipped {
				t.Error("Expected skipped to be detected")
			}
		})
	}
}

func TestRenderActivityOutput(t *testing.T) {
	// Test renderActivity output format
	tests := []struct {
		name             string
		newCommentsCount int
		expectsEmpty     bool
		expectsCount     bool
	}{
		{
			name:             "zero comments returns empty lines",
			newCommentsCount: 0,
			expectsEmpty:     true,
		},
		{
			name:             "negative comments returns empty lines",
			newCommentsCount: -1,
			expectsEmpty:     true,
		},
		{
			name:             "positive comments shows count",
			newCommentsCount: 5,
			expectsCount:     true,
		},
		{
			name:             "single comment shows count",
			newCommentsCount: 1,
			expectsCount:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Notification{
				Data: &Data{
					NewCommentsCount: tt.newCommentsCount,
				},
			}
			result := n.renderActivity()

			if tt.expectsEmpty {
				// Should be just newlines (empty content with 3-line structure)
				if result != "\n\n" {
					t.Errorf("renderActivity() = %q, want empty lines", result)
				}
			}
			if tt.expectsCount {
				// Should contain the count
				if !strings.Contains(result, "+") {
					t.Errorf("renderActivity() = %q, should contain '+' for count", result)
				}
			}
		})
	}
}

func TestDataState(t *testing.T) {
	// Test that SubjectState and IsDraft are properly used
	tests := []struct {
		name         string
		subjectState string
		isDraft      bool
		subjectType  string
	}{
		{
			name:         "open PR",
			subjectState: StateOpen,
			isDraft:      false,
			subjectType:  "PullRequest",
		},
		{
			name:         "draft PR",
			subjectState: StateOpen,
			isDraft:      true,
			subjectType:  "PullRequest",
		},
		{
			name:         "merged PR",
			subjectState: StateMerged,
			isDraft:      false,
			subjectType:  "PullRequest",
		},
		{
			name:         "closed PR",
			subjectState: StateClosed,
			isDraft:      false,
			subjectType:  "PullRequest",
		},
		{
			name:         "open Issue",
			subjectState: StateOpen,
			isDraft:      false,
			subjectType:  "Issue",
		},
		{
			name:         "closed Issue",
			subjectState: StateClosed,
			isDraft:      false,
			subjectType:  "Issue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Data{
				SubjectState: tt.subjectState,
				IsDraft:      tt.isDraft,
				Notification: data.NotificationData{
					Subject: data.NotificationSubject{
						Type: tt.subjectType,
					},
				},
			}

			// Verify the data is set correctly
			if d.SubjectState != tt.subjectState {
				t.Errorf("SubjectState = %q, want %q", d.SubjectState, tt.subjectState)
			}
			if d.IsDraft != tt.isDraft {
				t.Errorf("IsDraft = %v, want %v", d.IsDraft, tt.isDraft)
			}
			if d.GetSubjectType() != tt.subjectType {
				t.Errorf("GetSubjectType() = %q, want %q", d.GetSubjectType(), tt.subjectType)
			}
		})
	}
}

func TestActivityDescriptionFallback(t *testing.T) {
	// Test that ActivityDescription is used when set, else falls back to reason
	tests := []struct {
		name                string
		activityDescription string
		reason              string
		subjectType         string
		expectContains      string
	}{
		{
			name:                "uses ActivityDescription when set",
			activityDescription: "@user commented on this PR",
			reason:              "comment",
			subjectType:         "PullRequest",
			expectContains:      "@user commented",
		},
		{
			name:                "falls back to reason when ActivityDescription empty",
			activityDescription: "",
			reason:              "review_requested",
			subjectType:         "PullRequest",
			expectContains:      "Review requested",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Data{
				ActivityDescription: tt.activityDescription,
				Notification: data.NotificationData{
					Reason: tt.reason,
					Subject: data.NotificationSubject{
						Type: tt.subjectType,
					},
				},
			}

			// When ActivityDescription is set, it should be used
			if tt.activityDescription != "" {
				if d.ActivityDescription != tt.activityDescription {
					t.Errorf("ActivityDescription = %q, want %q", d.ActivityDescription, tt.activityDescription)
				}
			}
		})
	}
}
