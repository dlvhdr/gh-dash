package notificationrow

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

// PR/Issue state constants from GitHub API
const (
	StateOpen   = "OPEN"
	StateClosed = "CLOSED"
	StateMerged = "MERGED"
)

type Data struct {
	Notification        data.NotificationData
	NewCommentsCount    int    // Number of new comments since last read
	SubjectState        string // State of the PR/Issue (OPEN, CLOSED, MERGED)
	IsDraft             bool   // Whether PR is a draft
	Actor               string // Username of the user who triggered the notification
	ActivityDescription string // Human-readable description of the activity (e.g., "@user commented on this PR")
	ResolvedUrl         string // Async-resolved URL (e.g., for CheckSuite -> specific workflow run)
}

func (d Data) GetTitle() string {
	// Sanitize title: remove carriage returns and other control characters
	// that can corrupt terminal rendering (e.g., GitHub sometimes returns
	// titles with trailing \r characters)
	title := d.Notification.Subject.Title
	title = strings.ReplaceAll(title, "\r", "")
	title = strings.ReplaceAll(title, "\n", " ")
	return strings.TrimSpace(title)
}

func (d Data) GetRepoNameWithOwner() string {
	return d.Notification.Repository.FullName
}

func (d Data) GetNumber() int {
	subject := d.Notification.Subject
	if subject.Type == "PullRequest" || subject.Type == "Issue" {
		numStr := extractNumberFromUrl(subject.Url)
		if num, err := strconv.Atoi(numStr); err == nil {
			return num
		}
	}
	return 0
}

func (d Data) GetUrl() string {
	subject := d.Notification.Subject
	repo := d.Notification.Repository.FullName

	switch subject.Type {
	case "PullRequest":
		return fmt.Sprintf("https://github.com/%s/pull/%s", repo, extractNumberFromUrl(subject.Url))
	case "Issue":
		return fmt.Sprintf("https://github.com/%s/issues/%s", repo, extractNumberFromUrl(subject.Url))
	case "Discussion":
		num := extractNumberFromUrl(subject.Url)
		if num != "" {
			return fmt.Sprintf("https://github.com/%s/discussions/%s", repo, num)
		}
		return fmt.Sprintf("https://github.com/%s/discussions", repo)
	case "Release":
		return fmt.Sprintf("https://github.com/%s/releases", repo)
	case "Commit":
		return fmt.Sprintf("https://github.com/%s/commits", repo)
	case "CheckSuite":
		// GitHub's API returns subject.url=null for CheckSuite notifications.
		// ResolvedUrl is populated asynchronously with the specific workflow run URL.
		// Until resolved, we fall back to the repository's actions page.
		if d.ResolvedUrl != "" {
			return d.ResolvedUrl
		}
		return fmt.Sprintf("https://github.com/%s/actions", repo)
	default:
		return fmt.Sprintf("https://github.com/%s", repo)
	}
}

func (d Data) GetUpdatedAt() time.Time {
	return d.Notification.UpdatedAt
}

func (d Data) GetCreatedAt() time.Time {
	return d.Notification.UpdatedAt
}

func (d Data) GetId() string {
	return d.Notification.Id
}

func (d Data) GetSubjectType() string {
	return d.Notification.Subject.Type
}

func (d Data) GetReason() string {
	return d.Notification.Reason
}

func (d Data) IsUnread() bool {
	return d.Notification.Unread
}

func (d Data) GetLatestCommentUrl() string {
	return d.Notification.Subject.LatestCommentUrl
}

// extractNumberFromUrl extracts the last path segment (typically a number) from an API URL
func extractNumberFromUrl(apiUrl string) string {
	if apiUrl == "" {
		return ""
	}
	for i := len(apiUrl) - 1; i >= 0; i-- {
		if apiUrl[i] == '/' {
			return apiUrl[i+1:]
		}
	}
	return ""
}

// GenerateActivityDescription creates a human-readable description of the notification activity
func GenerateActivityDescription(reason, subjectType, actor string) string {
	switch reason {
	case "comment":
		if actor != "" {
			switch subjectType {
			case "PullRequest":
				return fmt.Sprintf("@%s commented on this pull request", actor)
			case "Issue":
				return fmt.Sprintf("@%s commented on this issue", actor)
			default:
				return fmt.Sprintf("@%s commented", actor)
			}
		}
		return "New comment"
	case "review_requested":
		if actor != "" {
			return fmt.Sprintf("@%s requested your review", actor)
		}
		return "Review requested"
	case "mention":
		if actor != "" {
			return fmt.Sprintf("@%s mentioned you", actor)
		}
		return "You were mentioned"
	case "author":
		return "Activity on your thread"
	case "assign":
		return "You were assigned"
	case "state_change":
		switch subjectType {
		case "PullRequest":
			return "Pull request state changed"
		case "Issue":
			return "Issue state changed"
		default:
			return "State changed"
		}
	case "ci_activity":
		return "CI activity"
	case "subscribed":
		if actor != "" {
			switch subjectType {
			case "PullRequest":
				return fmt.Sprintf("@%s commented on this pull request", actor)
			case "Issue":
				return fmt.Sprintf("@%s commented on this issue", actor)
			default:
				return "Activity on subscribed thread"
			}
		}
		return "Activity on subscribed thread"
	case "team_mention":
		return "Your team was mentioned"
	case "security_alert":
		return "Security vulnerability detected"
	default:
		if actor != "" {
			return fmt.Sprintf("@%s triggered this notification", actor)
		}
		return ""
	}
}
