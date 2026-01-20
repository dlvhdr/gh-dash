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
	PageInfo      PageInfo
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

func FetchNotifications(limit int, repoFilters []string, readState NotificationReadState, pageInfo *PageInfo) (NotificationsResponse, error) {
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

	// Determine page number from PageInfo (EndCursor stores the current page as string)
	page := 1
	if pageInfo != nil && pageInfo.EndCursor != "" {
		fmt.Sscanf(pageInfo.EndCursor, "%d", &page)
	}

	if len(repoFilters) == 0 {
		// No repo filter, fetch all notifications
		path := fmt.Sprintf("notifications?per_page=%d&page=%d%s", limit, page, allParam)
		log.Debug("Fetching notifications", "limit", limit, "page", page, "readState", readState)
		err = client.Get(path, &allNotifications)
		if err != nil {
			return NotificationsResponse{}, err
		}
	} else {
		// Fetch notifications for each repo and combine
		for _, repo := range repoFilters {
			var repoNotifications []NotificationData
			path := fmt.Sprintf("repos/%s/notifications?per_page=%d&page=%d%s", repo, limit, page, allParam)
			log.Debug("Fetching notifications for repo", "repo", repo, "limit", limit, "page", page, "readState", readState)
			err = client.Get(path, &repoNotifications)
			if err != nil {
				log.Warn("Failed to fetch notifications for repo", "repo", repo, "err", err)
				continue
			}
			allNotifications = append(allNotifications, repoNotifications...)
		}
	}

	// Determine if there's a next page BEFORE filtering (based on raw API response count).
	// If the API returned a full page, there are likely more notifications on the server.
	// We must check this before filtering because the caller needs accurate pagination info
	// to fetch additional pages when many notifications are filtered out locally.
	rawCount := len(allNotifications)
	hasNextPage := rawCount >= limit
	nextPage := fmt.Sprintf("%d", page+1)

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

	log.Info("Successfully fetched notifications", "rawCount", rawCount, "filteredCount", len(allNotifications), "page", page, "hasNextPage", hasNextPage, "readState", readState)

	return NotificationsResponse{
		Notifications: allNotifications,
		TotalCount:    len(allNotifications),
		PageInfo: PageInfo{
			HasNextPage: hasNextPage,
			EndCursor:   nextPage,
		},
	}, nil
}

// FetchNotificationByThreadId fetches a single notification by its thread ID.
// This is useful for fetching bookmarked or session-marked-read notifications
// that may not appear in the regular notifications list.
func FetchNotificationByThreadId(threadId string) (*NotificationData, error) {
	client, err := getRESTClient()
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("notifications/threads/%s", threadId)
	log.Debug("Fetching notification by thread ID", "threadId", threadId)

	var notification NotificationData
	err = client.Get(path, &notification)
	if err != nil {
		return nil, err
	}

	return &notification, nil
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

// WorkflowRun represents a GitHub Actions workflow run
type WorkflowRun struct {
	Id         int64     `json:"id"`
	Name       string    `json:"name"`
	HtmlUrl    string    `json:"html_url"`
	HeadBranch string    `json:"head_branch"`
	UpdatedAt  time.Time `json:"updated_at"`
	Conclusion string    `json:"conclusion"` // success, failure, cancelled, etc.
}

// WorkflowRunsResponse represents the response from the workflow runs API
type WorkflowRunsResponse struct {
	TotalCount   int           `json:"total_count"`
	WorkflowRuns []WorkflowRun `json:"workflow_runs"`
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

// FindBestWorkflowRunMatch finds the workflow run closest in time to the notification.
// Returns nil if no suitable match is found within the time window.
// Exported for testing.
func FindBestWorkflowRunMatch(runs []WorkflowRun, notificationUpdatedAt time.Time) *WorkflowRun {
	if len(runs) == 0 {
		return nil
	}

	var bestMatch *WorkflowRun
	bestDiff := time.Hour * 24 * 365 // Start with a large value

	for i := range runs {
		run := &runs[i]
		diff := notificationUpdatedAt.Sub(run.UpdatedAt)
		if diff < 0 {
			diff = -diff
		}

		// Prefer runs that are close in time (within a reasonable window)
		if diff < bestDiff && diff < time.Hour {
			bestDiff = diff
			bestMatch = run
		}
	}

	// If no close match, just return the most recent run
	if bestMatch == nil {
		bestMatch = &runs[0]
	}

	return bestMatch
}

// FetchRecentWorkflowRun fetches recent workflow runs for a repo and finds the best match
// based on the notification's updated_at timestamp. Returns the HTML URL of the matching run.
// The title parameter is the notification subject title (e.g., "CI / build (push)")
// which may help identify the correct workflow run.
func FetchRecentWorkflowRun(repo string, notificationUpdatedAt time.Time, title string) (string, error) {
	client, err := getRESTClient()
	if err != nil {
		return "", err
	}

	// Fetch recent workflow runs (limit to 20 for performance)
	path := fmt.Sprintf("repos/%s/actions/runs?per_page=20", repo)
	log.Debug("Fetching workflow runs", "repo", repo, "path", path)

	var response WorkflowRunsResponse
	err = client.Get(path, &response)
	if err != nil {
		log.Debug("Failed to fetch workflow runs", "repo", repo, "err", err)
		return "", err
	}

	if len(response.WorkflowRuns) == 0 {
		return "", nil
	}

	bestMatch := FindBestWorkflowRunMatch(response.WorkflowRuns, notificationUpdatedAt)
	if bestMatch != nil {
		log.Debug("Found matching workflow run", "id", bestMatch.Id, "name", bestMatch.Name, "url", bestMatch.HtmlUrl)
		return bestMatch.HtmlUrl, nil
	}

	return "", nil
}
