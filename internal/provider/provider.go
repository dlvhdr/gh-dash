package provider

import (
	"time"
)

// ProviderType represents the type of Git hosting provider
type ProviderType string

const (
	GitHub ProviderType = "github"
	GitLab ProviderType = "gitlab"
)

// Provider is the interface that abstracts GitHub and GitLab APIs
type Provider interface {
	// GetType returns the provider type
	GetType() ProviderType

	// GetHost returns the hostname (e.g., "gitlab.krone.at" or "github.com")
	GetHost() string

	// FetchPullRequests fetches pull/merge requests
	FetchPullRequests(query string, limit int, pageInfo *PageInfo) (PullRequestsResponse, error)

	// FetchPullRequest fetches a single pull/merge request by URL
	FetchPullRequest(prUrl string) (EnrichedPullRequestData, error)

	// FetchIssues fetches issues
	FetchIssues(query string, limit int, pageInfo *PageInfo) (IssuesResponse, error)

	// FetchIssueComments fetches comments for a single issue (GitLab only)
	FetchIssueComments(issueUrl string) ([]IssueComment, error)

	// GetCurrentUser returns the current authenticated user
	GetCurrentUser() (string, error)

	// GetCLICommand returns the CLI command name ("gh" or "glab")
	GetCLICommand() string
}

// PageInfo for pagination
type PageInfo struct {
	HasNextPage bool
	StartCursor string
	EndCursor   string
}

// Repository represents a repository
type Repository struct {
	NameWithOwner string
	IsArchived    bool
}

// Author represents an author
type Author struct {
	Login string
}

// Assignee represents an assignee
type Assignee struct {
	Login string
}

// Assignees represents a list of assignees
type Assignees struct {
	Nodes []Assignee
}

// Label represents a label
type Label struct {
	Color string
	Name  string
}

// Labels represents a list of labels
type Labels struct {
	Nodes []Label
}

// Comment represents a comment
type Comment struct {
	Author    Author
	Body      string
	UpdatedAt time.Time
}

// Comments represents comment counts
type Comments struct {
	TotalCount int
}

// CommentsWithBody includes full comment data
type CommentsWithBody struct {
	TotalCount int
	Nodes      []Comment
}

// Review represents a review
type Review struct {
	Author    Author
	Body      string
	State     string
	UpdatedAt time.Time
}

// Reviews represents reviews
type Reviews struct {
	TotalCount int
	Nodes      []Review
}

// ReviewThreads represents review threads
type ReviewThreads struct {
	TotalCount int
}

// ReviewRequests represents review requests
type ReviewRequests struct {
	TotalCount int
	Nodes      []ReviewRequestNode
}

// ReviewRequestNode represents a single review request
type ReviewRequestNode struct {
	AsCodeOwner       bool
	RequestedReviewer struct {
		User      struct{ Login string }
		Team      struct{ Slug, Name string }
		Bot       struct{ Login string }
		Mannequin struct{ Login string }
	}
}

// GetReviewerDisplayName returns the display name for a reviewer
func (r ReviewRequestNode) GetReviewerDisplayName() string {
	if r.RequestedReviewer.User.Login != "" {
		return r.RequestedReviewer.User.Login
	}
	if r.RequestedReviewer.Team.Slug != "" {
		return r.RequestedReviewer.Team.Slug
	}
	if r.RequestedReviewer.Bot.Login != "" {
		return r.RequestedReviewer.Bot.Login
	}
	if r.RequestedReviewer.Mannequin.Login != "" {
		return r.RequestedReviewer.Mannequin.Login
	}
	return ""
}

// GetReviewerType returns the type of reviewer
func (r ReviewRequestNode) GetReviewerType() string {
	if r.RequestedReviewer.User.Login != "" {
		return "User"
	}
	if r.RequestedReviewer.Team.Slug != "" {
		return "Team"
	}
	if r.RequestedReviewer.Bot.Login != "" {
		return "Bot"
	}
	if r.RequestedReviewer.Mannequin.Login != "" {
		return "Mannequin"
	}
	return ""
}

// IsTeam returns true if the reviewer is a team
func (r ReviewRequestNode) IsTeam() bool {
	return r.RequestedReviewer.Team.Slug != ""
}

// ChangedFile represents a changed file
type ChangedFile struct {
	Additions  int
	Deletions  int
	Path       string
	ChangeType string
}

// ChangedFiles represents changed files
type ChangedFiles struct {
	TotalCount int
	Nodes      []ChangedFile
}

// Commits represents commits
type Commits struct {
	TotalCount int
}

// PullRequestData represents a pull/merge request
type PullRequestData struct {
	Number            int
	Title             string
	Body              string
	Author            Author
	AuthorAssociation string
	UpdatedAt         time.Time
	CreatedAt         time.Time
	Url               string
	State             string
	Mergeable         string
	ReviewDecision    string
	Additions         int
	Deletions         int
	HeadRefName       string
	BaseRefName       string
	HeadRepository    struct{ Name string }
	HeadRef           struct{ Name string }
	Repository        Repository
	Assignees         Assignees
	Comments          Comments
	ReviewThreads     ReviewThreads
	Reviews           Reviews
	ReviewRequests    ReviewRequests
	Files             ChangedFiles
	IsDraft           bool
	Commits           Commits
	Labels            Labels
	MergeStateStatus  string
	CIStatus          string // "success", "failure", "pending", ""
}

// PullRequestsResponse represents a list of pull requests
type PullRequestsResponse struct {
	Prs        []PullRequestData
	TotalCount int
	PageInfo   PageInfo
}

// EnrichedPullRequestData represents enriched pull request data
type EnrichedPullRequestData struct {
	Url            string
	Number         int
	Repository     Repository
	Comments       CommentsWithBody
	ReviewThreads  ReviewThreads
	ReviewRequests ReviewRequests
	Reviews        Reviews
	Commits        []CommitData
	ChangedFiles   []ChangedFile
}

// CommitData represents commit data
type CommitData struct {
	AbbreviatedOid  string
	CommittedDate   time.Time
	MessageHeadline string
	Author          struct {
		Name string
		User struct{ Login string }
	}
	StatusCheckRollup string
}

// IssueData represents an issue
type IssueData struct {
	Number            int
	Title             string
	Body              string
	State             string
	Author            Author
	AuthorAssociation string
	UpdatedAt         time.Time
	CreatedAt         time.Time
	Url               string
	Repository        Repository
	Assignees         Assignees
	Comments          IssueComments
	Reactions         IssueReactions
	Labels            Labels
}

// IssueComments represents issue comments
type IssueComments struct {
	Nodes      []IssueComment
	TotalCount int
}

// IssueComment represents an issue comment
type IssueComment struct {
	Author    Author
	Body      string
	UpdatedAt time.Time
}

// IssueReactions represents issue reactions
type IssueReactions struct {
	TotalCount int
}

// IssuesResponse represents a list of issues
type IssuesResponse struct {
	Issues     []IssueData
	TotalCount int
	PageInfo   PageInfo
}
