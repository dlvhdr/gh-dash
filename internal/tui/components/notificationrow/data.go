package notificationrow

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

type Data struct {
	Notification     data.NotificationData
	NewCommentsCount int    // Number of new comments since last read
	SubjectState     string // State of the PR/Issue (OPEN, CLOSED, MERGED)
	IsDraft          bool   // Whether PR is a draft
	Actor            string // Username of the user who triggered the notification
}

func (d Data) GetTitle() string {
	return d.Notification.Subject.Title
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
		// GitHub's API returns subject.url=null for CheckSuite notifications,
		// so we can't link to the specific commit checks page
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
