package provider

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

// GitLabProvider implements the Provider interface for GitLab
type GitLabProvider struct {
	host        string
	labelColors map[string]map[string]string // project -> label name -> color
}

// NewGitLabProvider creates a new GitLab provider
func NewGitLabProvider(host string) *GitLabProvider {
	// Normalize host - remove protocol if present
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimSuffix(host, "/")

	return &GitLabProvider{
		host:        host,
		labelColors: make(map[string]map[string]string),
	}
}

func (g *GitLabProvider) GetType() ProviderType {
	return GitLab
}

func (g *GitLabProvider) GetHost() string {
	return g.host
}

func (g *GitLabProvider) GetCLICommand() string {
	return "glab"
}

// runGlab runs a glab command and returns the output
// Note: glab uses the configured host from 'glab config'. Make sure glab is configured
// with: glab auth login --hostname <your-gitlab-host>
func (g *GitLabProvider) runGlab(args ...string) ([]byte, error) {
	log.Debug("Running glab command", "args", args, "host", g.host)

	cmd := exec.Command("glab", args...)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Error("glab command failed", "stderr", string(exitErr.Stderr), "args", args)
			return nil, fmt.Errorf("glab command failed: %s", string(exitErr.Stderr))
		}
		return nil, err
	}
	return output, nil
}

// glabLabel represents a GitLab label with color information
type glabLabel struct {
	Name      string `json:"name"`
	Color     string `json:"color"`
	TextColor string `json:"text_color"`
}

// fetchLabelColors fetches and caches label colors for a project
func (g *GitLabProvider) fetchLabelColors(project string) {
	if project == "" {
		return
	}

	// Check if already cached
	if _, ok := g.labelColors[project]; ok {
		return
	}

	output, err := g.runGlab("label", "list", "--repo", project, "--output", "json", "--per-page", "100")
	if err != nil {
		log.Debug("Failed to fetch labels for color cache", "project", project, "err", err)
		return
	}

	var labels []glabLabel
	if err := json.Unmarshal(output, &labels); err != nil {
		log.Debug("Failed to parse labels", "err", err)
		return
	}

	g.labelColors[project] = make(map[string]string)
	for _, l := range labels {
		// Remove # prefix if present
		color := strings.TrimPrefix(l.Color, "#")
		g.labelColors[project][l.Name] = color
	}

	log.Debug("Cached label colors", "project", project, "count", len(labels))
}

// getLabelColor returns the color for a label, or empty string if not found
func (g *GitLabProvider) getLabelColor(project, labelName string) string {
	if projectLabels, ok := g.labelColors[project]; ok {
		if color, ok := projectLabels[labelName]; ok {
			return color
		}
	}
	return ""
}

// GitLab API response structures
type glabMergeRequest struct {
	IID          int       `json:"iid"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	State        string    `json:"state"`
	Draft        bool      `json:"draft"`
	WebURL       string    `json:"web_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	SourceBranch string    `json:"source_branch"`
	TargetBranch string    `json:"target_branch"`
	Author       struct {
		Username string `json:"username"`
	} `json:"author"`
	Assignees []struct {
		Username string `json:"username"`
	} `json:"assignees"`
	Reviewers []struct {
		Username string `json:"username"`
	} `json:"reviewers"`
	Labels               []string `json:"labels"`
	UserNotesCount       int      `json:"user_notes_count"`
	MergeStatus          string   `json:"merge_status"`
	HasConflicts         bool     `json:"has_conflicts"`
	BlockingDiscussions  int      `json:"blocking_discussions_resolved_count"`
	ChangesCount         string   `json:"changes_count"`
	DiffRefs             struct{} `json:"diff_refs"`
	ProjectID            int      `json:"project_id"`
	SourceProjectID      int      `json:"source_project_id"`
	TargetProjectID      int      `json:"target_project_id"`
	References           struct {
		Full string `json:"full"`
	} `json:"references"`
	Pipeline *struct {
		Status string `json:"status"`
	} `json:"pipeline"`
}

type glabIssue struct {
	IID         int       `json:"iid"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	State       string    `json:"state"`
	WebURL      string    `json:"web_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Author      struct {
		Username string `json:"username"`
	} `json:"author"`
	Assignees []struct {
		Username string `json:"username"`
	} `json:"assignees"`
	Labels         []string `json:"labels"`
	UserNotesCount int      `json:"user_notes_count"`
	Upvotes        int      `json:"upvotes"`
	Downvotes      int      `json:"downvotes"`
	References     struct {
		Full string `json:"full"`
	} `json:"references"`
}

type glabProject struct {
	PathWithNamespace string `json:"path_with_namespace"`
	Archived          bool   `json:"archived"`
}

func (g *GitLabProvider) FetchPullRequests(query string, limit int, pageInfo *PageInfo) (PullRequestsResponse, error) {
	// Parse the query to extract project and filters
	// GitLab uses different query syntax than GitHub
	// Common filters: state, labels, author, assignee, reviewer

	// Parse query for common filters
	filters := g.parseQuery(query)

	// Fetch label colors for the project (if specified)
	if project, ok := filters["repo"]; ok {
		g.fetchLabelColors(project)
	}

	args := []string{"mr", "list", "--output", "json", "--per-page", strconv.Itoa(limit)}

	if project, ok := filters["repo"]; ok {
		args = append(args, "--repo", project)
	}

	if state, ok := filters["state"]; ok {
		args = append(args, "--state", state)
	} else if strings.Contains(query, "is:open") {
		args = append(args, "--state", "opened")
	} else if strings.Contains(query, "is:closed") {
		args = append(args, "--state", "closed")
	} else if strings.Contains(query, "is:merged") {
		args = append(args, "--state", "merged")
	}

	if author, ok := filters["author"]; ok {
		if author == "@me" {
			args = append(args, "--author", "@me")
		} else {
			args = append(args, "--author", author)
		}
	}

	if assignee, ok := filters["assignee"]; ok {
		if assignee == "@me" {
			args = append(args, "--assignee", "@me")
		} else {
			args = append(args, "--assignee", assignee)
		}
	}

	if reviewer, ok := filters["reviewer"]; ok {
		if reviewer == "@me" {
			args = append(args, "--reviewer", "@me")
		} else {
			args = append(args, "--reviewer", reviewer)
		}
	}

	if labels, ok := filters["label"]; ok {
		args = append(args, "--label", labels)
	}

	if strings.Contains(query, "draft:true") {
		args = append(args, "--draft")
	}

	if strings.Contains(query, "review-requested:@me") {
		args = append(args, "--reviewer", "@me")
	}

	output, err := g.runGlab(args...)
	if err != nil {
		return PullRequestsResponse{}, err
	}

	var mrs []glabMergeRequest
	if err := json.Unmarshal(output, &mrs); err != nil {
		log.Error("Failed to parse MR list", "err", err, "output", string(output))
		return PullRequestsResponse{}, fmt.Errorf("failed to parse MR list: %w", err)
	}

	prs := make([]PullRequestData, 0, len(mrs))
	for _, mr := range mrs {
		pr := g.convertMRtoPR(mr)
		prs = append(prs, pr)
	}

	log.Info("Successfully fetched MRs from GitLab", "count", len(prs))

	return PullRequestsResponse{
		Prs:        prs,
		TotalCount: len(prs),
		PageInfo:   PageInfo{HasNextPage: len(prs) == limit},
	}, nil
}

func (g *GitLabProvider) convertMRtoPR(mr glabMergeRequest) PullRequestData {
	assignees := make([]Assignee, len(mr.Assignees))
	for i, a := range mr.Assignees {
		assignees[i] = Assignee{Login: a.Username}
	}

	// Extract project path from full reference (e.g., "group/project!123")
	projectPath := ""
	if mr.References.Full != "" {
		parts := strings.Split(mr.References.Full, "!")
		if len(parts) > 0 {
			projectPath = parts[0]
		}
	}

	// Get label colors from cache
	labels := make([]Label, len(mr.Labels))
	for i, l := range mr.Labels {
		labels[i] = Label{Name: l, Color: g.getLabelColor(projectPath, l)}
	}

	reviewRequests := make([]ReviewRequestNode, len(mr.Reviewers))
	for i, r := range mr.Reviewers {
		reviewRequests[i] = ReviewRequestNode{}
		reviewRequests[i].RequestedReviewer.User.Login = r.Username
	}

	state := mr.State
	if state == "opened" {
		state = "OPEN"
	} else if state == "merged" {
		state = "MERGED"
	} else if state == "closed" {
		state = "CLOSED"
	}

	mergeable := "UNKNOWN"
	if mr.MergeStatus == "can_be_merged" {
		mergeable = "MERGEABLE"
	} else if mr.HasConflicts {
		mergeable = "CONFLICTING"
	}

	ciStatus := ""
	if mr.Pipeline != nil {
		ciStatus = mr.Pipeline.Status
	}

	return PullRequestData{
		Number:            mr.IID,
		Title:             mr.Title,
		Body:              mr.Description,
		Author:            Author{Login: mr.Author.Username},
		AuthorAssociation: "",
		UpdatedAt:         mr.UpdatedAt,
		CreatedAt:         mr.CreatedAt,
		Url:               mr.WebURL,
		State:             state,
		Mergeable:         mergeable,
		ReviewDecision:    "",
		Additions:         0,
		Deletions:         0,
		HeadRefName:       mr.SourceBranch,
		BaseRefName:       mr.TargetBranch,
		HeadRepository:    struct{ Name string }{Name: ""},
		HeadRef:           struct{ Name string }{Name: mr.SourceBranch},
		Repository:        Repository{NameWithOwner: projectPath, IsArchived: false},
		Assignees:         Assignees{Nodes: assignees},
		Comments:          Comments{TotalCount: mr.UserNotesCount},
		ReviewThreads:     ReviewThreads{TotalCount: 0},
		Reviews:           Reviews{TotalCount: 0, Nodes: nil},
		ReviewRequests:    ReviewRequests{TotalCount: len(reviewRequests), Nodes: reviewRequests},
		Files:             ChangedFiles{TotalCount: 0, Nodes: nil},
		IsDraft:           mr.Draft,
		Commits:           Commits{TotalCount: 0},
		Labels:            Labels{Nodes: labels},
		MergeStateStatus:  mr.MergeStatus,
		CIStatus:          ciStatus,
	}
}

func (g *GitLabProvider) FetchPullRequest(prUrl string) (EnrichedPullRequestData, error) {
	// Parse project and MR IID from URL
	// Format: https://gitlab.example.com/group/project/-/merge_requests/123
	parsedURL, err := url.Parse(prUrl)
	if err != nil {
		return EnrichedPullRequestData{}, err
	}

	path := parsedURL.Path
	// Extract project and MR ID
	re := regexp.MustCompile(`(.+?)/-/merge_requests/(\d+)`)
	matches := re.FindStringSubmatch(path)
	if len(matches) < 3 {
		return EnrichedPullRequestData{}, fmt.Errorf("invalid MR URL format: %s", prUrl)
	}

	project := strings.TrimPrefix(matches[1], "/")
	mrID := matches[2]

	// Get MR details
	output, err := g.runGlab("mr", "view", mrID, "--repo", project, "--output", "json")
	if err != nil {
		return EnrichedPullRequestData{}, err
	}

	var mr glabMergeRequest
	if err := json.Unmarshal(output, &mr); err != nil {
		return EnrichedPullRequestData{}, fmt.Errorf("failed to parse MR: %w", err)
	}

	// Get MR notes/comments
	notesOutput, err := g.runGlab("mr", "note", "list", mrID, "--repo", project, "--output", "json")
	comments := CommentsWithBody{TotalCount: mr.UserNotesCount}
	if err == nil && len(notesOutput) > 0 {
		var notes []struct {
			Body      string    `json:"body"`
			Author    struct{ Username string } `json:"author"`
			CreatedAt time.Time `json:"created_at"`
		}
		if json.Unmarshal(notesOutput, &notes) == nil {
			for _, note := range notes {
				comments.Nodes = append(comments.Nodes, Comment{
					Author:    Author{Login: note.Author.Username},
					Body:      note.Body,
					UpdatedAt: note.CreatedAt,
				})
			}
		}
	}

	reviewRequests := make([]ReviewRequestNode, len(mr.Reviewers))
	for i, r := range mr.Reviewers {
		reviewRequests[i] = ReviewRequestNode{}
		reviewRequests[i].RequestedReviewer.User.Login = r.Username
	}

	return EnrichedPullRequestData{
		Url:            mr.WebURL,
		Number:         mr.IID,
		Repository:     Repository{NameWithOwner: project, IsArchived: false},
		Comments:       comments,
		ReviewThreads:  ReviewThreads{TotalCount: 0},
		ReviewRequests: ReviewRequests{TotalCount: len(reviewRequests), Nodes: reviewRequests},
		Reviews:        Reviews{TotalCount: 0},
		Commits:        nil,
	}, nil
}

func (g *GitLabProvider) FetchIssues(query string, limit int, pageInfo *PageInfo) (IssuesResponse, error) {
	filters := g.parseQuery(query)

	// Fetch label colors for the project (if specified)
	if project, ok := filters["repo"]; ok {
		g.fetchLabelColors(project)
	}

	args := []string{"issue", "list", "--output", "json", "--per-page", strconv.Itoa(limit)}

	if project, ok := filters["repo"]; ok {
		args = append(args, "--repo", project)
	}

	if state, ok := filters["state"]; ok {
		args = append(args, "--state", state)
	} else if strings.Contains(query, "is:open") {
		args = append(args, "--state", "opened")
	} else if strings.Contains(query, "is:closed") {
		args = append(args, "--state", "closed")
	}

	if author, ok := filters["author"]; ok {
		if author == "@me" {
			args = append(args, "--author", "@me")
		} else {
			args = append(args, "--author", author)
		}
	}

	if assignee, ok := filters["assignee"]; ok {
		if assignee == "@me" {
			args = append(args, "--assignee", "@me")
		} else {
			args = append(args, "--assignee", assignee)
		}
	}

	if labels, ok := filters["label"]; ok {
		args = append(args, "--label", labels)
	}

	output, err := g.runGlab(args...)
	if err != nil {
		return IssuesResponse{}, err
	}

	var glabIssues []glabIssue
	if err := json.Unmarshal(output, &glabIssues); err != nil {
		log.Error("Failed to parse issue list", "err", err, "output", string(output))
		return IssuesResponse{}, fmt.Errorf("failed to parse issue list: %w", err)
	}

	issues := make([]IssueData, 0, len(glabIssues))
	for _, issue := range glabIssues {
		issues = append(issues, g.convertGitLabIssue(issue))
	}

	log.Info("Successfully fetched issues from GitLab", "count", len(issues))

	return IssuesResponse{
		Issues:     issues,
		TotalCount: len(issues),
		PageInfo:   PageInfo{HasNextPage: len(issues) == limit},
	}, nil
}

func (g *GitLabProvider) convertGitLabIssue(issue glabIssue) IssueData {
	assignees := make([]Assignee, len(issue.Assignees))
	for i, a := range issue.Assignees {
		assignees[i] = Assignee{Login: a.Username}
	}

	// Extract project path from full reference
	projectPath := ""
	if issue.References.Full != "" {
		parts := strings.Split(issue.References.Full, "#")
		if len(parts) > 0 {
			projectPath = parts[0]
		}
	}

	// Get label colors from cache
	labels := make([]Label, len(issue.Labels))
	for i, l := range issue.Labels {
		labels[i] = Label{Name: l, Color: g.getLabelColor(projectPath, l)}
	}

	state := issue.State
	if state == "opened" {
		state = "OPEN"
	} else if state == "closed" {
		state = "CLOSED"
	}

	return IssueData{
		Number:            issue.IID,
		Title:             issue.Title,
		Body:              issue.Description,
		State:             state,
		Author:            Author{Login: issue.Author.Username},
		AuthorAssociation: "",
		UpdatedAt:         issue.UpdatedAt,
		CreatedAt:         issue.CreatedAt,
		Url:               issue.WebURL,
		Repository:        Repository{NameWithOwner: projectPath, IsArchived: false},
		Assignees:         Assignees{Nodes: assignees},
		Comments:          IssueComments{TotalCount: issue.UserNotesCount},
		Reactions:         IssueReactions{TotalCount: issue.Upvotes + issue.Downvotes},
		Labels:            Labels{Nodes: labels},
	}
}

func (g *GitLabProvider) GetCurrentUser() (string, error) {
	output, err := g.runGlab("auth", "status", "--show-token")
	if err != nil {
		return "", err
	}

	// Parse the auth status output to get username
	// Output format includes "Logged in to ... as USERNAME"
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Logged in to") && strings.Contains(line, " as ") {
			parts := strings.Split(line, " as ")
			if len(parts) >= 2 {
				// Clean up the username
				username := strings.TrimSpace(parts[1])
				// Remove any trailing info in parentheses
				if idx := strings.Index(username, " ("); idx > 0 {
					username = username[:idx]
				}
				return username, nil
			}
		}
	}

	return "", fmt.Errorf("could not determine current user")
}

// parseQuery parses a GitHub-style query into GitLab filters
func (g *GitLabProvider) parseQuery(query string) map[string]string {
	filters := make(map[string]string)

	// Split query into parts
	parts := strings.Fields(query)
	for _, part := range parts {
		if strings.HasPrefix(part, "repo:") {
			filters["repo"] = strings.TrimPrefix(part, "repo:")
		} else if strings.HasPrefix(part, "author:") {
			filters["author"] = strings.TrimPrefix(part, "author:")
		} else if strings.HasPrefix(part, "assignee:") {
			filters["assignee"] = strings.TrimPrefix(part, "assignee:")
		} else if strings.HasPrefix(part, "reviewer:") || strings.HasPrefix(part, "review-requested:") {
			val := strings.TrimPrefix(part, "reviewer:")
			val = strings.TrimPrefix(val, "review-requested:")
			filters["reviewer"] = val
		} else if strings.HasPrefix(part, "label:") {
			filters["label"] = strings.TrimPrefix(part, "label:")
		} else if strings.HasPrefix(part, "state:") {
			filters["state"] = strings.TrimPrefix(part, "state:")
		}
	}

	return filters
}
