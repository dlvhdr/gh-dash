package data

import (
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	gh "github.com/cli/go-gh/v2/pkg/api"
)

// Notification subject types from GitHub API
const (
	SubjectTypePullRequest = "PullRequest"
	SubjectTypeIssue       = "Issue"
	SubjectTypeDiscussion  = "Discussion"
	SubjectTypeRelease     = "Release"
	SubjectTypeCommit      = "Commit"
	SubjectTypeCheckSuite  = "CheckSuite"
)

// Notification reasons from GitHub API
const (
	ReasonSubscribed      = "subscribed"
	ReasonReviewRequested = "review_requested"
	ReasonMention         = "mention"
	ReasonAuthor          = "author"
	ReasonComment         = "comment"
	ReasonAssign          = "assign"
	ReasonStateChange     = "state_change"
	ReasonCIActivity      = "ci_activity"
	ReasonTeamMention     = "team_mention"
	ReasonSecurityAlert   = "security_alert"
)

var restClient *gh.RESTClient

type NotificationSubject struct {
	Title            string `json:"title"`
	Url              string `json:"url"`
	LatestCommentUrl string `json:"latest_comment_url"`
	Type             string `json:"type"` // PullRequest, Issue, Discussion, Release, etc.
}

type NotificationRepository struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	Owner    struct {
		Login string `json:"login"`
	} `json:"owner"`
	HtmlUrl string `json:"html_url"`
}

type NotificationData struct {
	Id           string                 `json:"id"`
	Unread       bool                   `json:"unread"`
	Reason       string                 `json:"reason"` // subscribed, review_requested, mention, etc.
	UpdatedAt    time.Time              `json:"updated_at"`
	LastReadAt   *time.Time             `json:"last_read_at"`
	Subject      NotificationSubject    `json:"subject"`
	Repository   NotificationRepository `json:"repository"`
	Url          string                 `json:"url"`
	Subscription string                 `json:"subscription_url"`
}

func (n NotificationData) GetTitle() string {
	return n.Subject.Title
}

func (n NotificationData) GetRepoNameWithOwner() string {
	return n.Repository.FullName
}

func (n NotificationData) GetNumber() int {
	// Notifications don't have a number, return 0
	return 0
}

func (n NotificationData) GetUrl() string {
	// Convert API URL to HTML URL
	// API URL: https://api.github.com/repos/owner/repo/pulls/123
	// HTML URL: https://github.com/owner/repo/pull/123
	return fmt.Sprintf("https://github.com/%s", n.Repository.FullName)
}

func (n NotificationData) GetUpdatedAt() time.Time {
	return n.UpdatedAt
}

func (n NotificationData) GetCreatedAt() time.Time {
	// Notifications don't have a created_at, use updated_at
	return n.UpdatedAt
}

type NotificationsResponse struct {
	Notifications []NotificationData
	TotalCount    int
}

func getRESTClient() (*gh.RESTClient, error) {
	if restClient != nil {
		return restClient, nil
	}
	var err error
	restClient, err = gh.DefaultRESTClient()
	return restClient, err
}

// NotificationReadState represents the read state filter for notifications
type NotificationReadState string

const (
	NotificationStateUnread NotificationReadState = "unread" // Only unread (default)
	NotificationStateRead   NotificationReadState = "read"   // Only read
	NotificationStateAll    NotificationReadState = "all"    // Both read and unread
)

func FetchNotifications(limit int, repoFilters []string, readState NotificationReadState) (NotificationsResponse, error) {
	client, err := getRESTClient()
	if err != nil {
		return NotificationsResponse{}, err
	}

	var allNotifications []NotificationData

	// Build query params
	// all=true returns both read and unread notifications
	includeAll := readState == NotificationStateRead || readState == NotificationStateAll
	allParam := ""
	if includeAll {
		allParam = "&all=true"
	}

	if len(repoFilters) == 0 {
		// No repo filter, fetch all notifications
		path := fmt.Sprintf("notifications?per_page=%d%s", limit, allParam)
		log.Debug("Fetching notifications", "limit", limit, "readState", readState)
		err = client.Get(path, &allNotifications)
		if err != nil {
			return NotificationsResponse{}, err
		}
	} else {
		// Fetch notifications for each repo and combine
		for _, repo := range repoFilters {
			var repoNotifications []NotificationData
			path := fmt.Sprintf("repos/%s/notifications?per_page=%d%s", repo, limit, allParam)
			log.Debug("Fetching notifications for repo", "repo", repo, "limit", limit, "readState", readState)
			err = client.Get(path, &repoNotifications)
			if err != nil {
				log.Warn("Failed to fetch notifications for repo", "repo", repo, "err", err)
				continue
			}
			allNotifications = append(allNotifications, repoNotifications...)
		}
	}

	// Filter by read state if needed
	switch readState {
	case NotificationStateRead:
		// Keep only read notifications
		filtered := make([]NotificationData, 0)
		for _, n := range allNotifications {
			if !n.Unread {
				filtered = append(filtered, n)
			}
		}
		allNotifications = filtered
	case NotificationStateUnread:
		// Keep only unread notifications (API default, but filter just in case)
		filtered := make([]NotificationData, 0)
		for _, n := range allNotifications {
			if n.Unread {
				filtered = append(filtered, n)
			}
		}
		allNotifications = filtered
	case NotificationStateAll:
		// Keep all, no filtering needed
	}

	log.Info("Successfully fetched notifications", "count", len(allNotifications), "readState", readState)

	return NotificationsResponse{
		Notifications: allNotifications,
		TotalCount:    len(allNotifications),
	}, nil
}

func MarkNotificationDone(threadId string) error {
	client, err := getRESTClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("notifications/threads/%s", threadId)
	log.Debug("Marking notification as done", "threadId", threadId)

	// DELETE marks as done
	err = client.Delete(path, nil)
	if err != nil {
		return err
	}
	log.Info("Successfully marked notification as done", "threadId", threadId)
	return nil
}

func MarkNotificationRead(threadId string) error {
	client, err := getRESTClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("notifications/threads/%s", threadId)
	log.Debug("Marking notification as read", "threadId", threadId)

	// PATCH marks as read - returns 205 Reset Content with no body
	// The REST client may return an error trying to parse the empty response
	err = client.Patch(path, nil, nil)
	if err != nil && err.Error() != "unexpected end of JSON input" {
		return err
	}
	log.Info("Successfully marked notification as read", "threadId", threadId)
	return nil
}

func UnsubscribeFromThread(threadId string) error {
	client, err := getRESTClient()
	if err != nil {
		return err
	}

	log.Debug("Unsubscribing from notification thread", "threadId", threadId)

	// DELETE /notifications/threads/{thread_id}/subscription
	// Mutes all future notifications for a conversation until you comment on the thread or get an @mention
	path := fmt.Sprintf("notifications/threads/%s/subscription", threadId)
	err = client.Delete(path, nil)
	if err != nil && err.Error() != "unexpected end of JSON input" {
		return err
	}
	log.Info("Successfully unsubscribed from thread", "threadId", threadId)
	return nil
}

func MarkAllNotificationsRead() error {
	client, err := getRESTClient()
	if err != nil {
		return err
	}

	log.Debug("Marking all notifications as read")

	// PUT /notifications marks all as read - returns 205 Reset Content with no body
	// The REST client may return an error trying to parse the empty response
	err = client.Put("notifications", nil, nil)
	if err != nil && err.Error() != "unexpected end of JSON input" {
		return err
	}
	log.Info("Successfully marked all notifications as read")
	return nil
}

// CommentResponse represents a GitHub comment with author info
type CommentResponse struct {
	User struct {
		Login string `json:"login"`
	} `json:"user"`
}

// FetchCommentAuthor fetches the author of a comment from its API URL
// apiUrl is like: https://api.github.com/repos/owner/repo/issues/comments/123456
func FetchCommentAuthor(apiUrl string) (string, error) {
	if apiUrl == "" {
		return "", nil
	}

	client, err := getRESTClient()
	if err != nil {
		return "", err
	}

	// Extract the path from the full URL
	const apiPrefix = "https://api.github.com/"
	path := apiUrl
	if len(apiUrl) > len(apiPrefix) && apiUrl[:len(apiPrefix)] == apiPrefix {
		path = apiUrl[len(apiPrefix):]
	}

	var response CommentResponse
	err = client.Get(path, &response)
	if err != nil {
		log.Debug("Failed to fetch comment author", "url", apiUrl, "err", err)
		return "", err
	}

	return response.User.Login, nil
}
