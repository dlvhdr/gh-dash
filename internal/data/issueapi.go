package data

import (
	"fmt"
	"net/url"
	"time"

	"charm.land/log/v2"
	gh "github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	"github.com/shurcooL/githubv4"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

type IssueData struct {
	Number int
	Title  string
	Body   string
	State  string
	Author struct {
		Login string
	}
	AuthorAssociation string
	UpdatedAt         time.Time
	CreatedAt         time.Time
	Url               string
	Repository        Repository
	Assignees         Assignees      `graphql:"assignees(first: 3)"`
	Comments          IssueComments  `graphql:"comments(last: 15)"`
	Reactions         IssueReactions `graphql:"reactions(first: 1)"`
	Labels            IssueLabels    `graphql:"labels(first: 20)"`
	IssueType         *IssueType
	// LinkedPrs is the open and merged PRs that will close the issue when merged, or that closed it.
	LinkedPrs LinkedPullRequests `graphql:"closedByPullRequestsReferences(first: 10)"`
}

type LinkedPullRequests struct {
	TotalCount int
	Nodes      []LinkedPullRequest
}

type LinkedPullRequest struct {
	State   string
	IsDraft bool
}

// State sums up the state of the linked PRs: OPEN if any of them is open and not a draft, then DRAFT if any is an
// open draft, and otherwise MERGED.
func (l LinkedPullRequests) State() string {
	state := "MERGED"
	for _, pr := range l.Nodes {
		if pr.State != "OPEN" {
			continue
		}
		if !pr.IsDraft {
			return "OPEN"
		}
		state = "DRAFT"
	}
	return state
}

// IssueType is the type assigned to an issue (Bug, Feature, Task, or an org's own types). It's nil for an issue
// with no type, and for repos where issue types aren't enabled.
type IssueType struct {
	Name string
	// Color is one of GitHub's IssueTypeColor values: GRAY, BLUE, GREEN, YELLOW, ORANGE, RED, PINK, or PURPLE.
	Color string
}

type IssueComments struct {
	Nodes      []IssueComment
	TotalCount int
}

type IssueComment struct {
	Author struct {
		Login string
	}
	Body      string
	UpdatedAt time.Time
}

type IssueReactions struct {
	TotalCount int
}

type Label struct {
	Color       string
	Name        string
	Description string
}

type IssueLabels struct {
	Nodes []Label
}

func (data IssueData) GetAuthor(theme theme.Theme, showAuthorIcons bool) string {
	author := data.Author.Login
	if showAuthorIcons {
		author += fmt.Sprintf(" %s", GetAuthorRoleIcon(data.AuthorAssociation, theme))
	}
	return author
}

func (data IssueData) GetTitle() string {
	return data.Title
}

func (data IssueData) GetRepoNameWithOwner() string {
	return data.Repository.NameWithOwner
}

func (data IssueData) GetRepoNameAndOwner() (owner, repoName string) {
	return data.Repository.Owner.Login, data.Repository.Name
}

func (data IssueData) GetNumber() int {
	return data.Number
}

func (data IssueData) GetUrl() string {
	return data.Url
}

func (data IssueData) GetUpdatedAt() time.Time {
	return data.UpdatedAt
}

func (data IssueData) GetCreatedAt() time.Time {
	return data.CreatedAt
}

func makeIssuesQuery(query string) string {
	return fmt.Sprintf("is:issue archived:false %s sort:updated", query)
}

func FetchIssues(query string, limit int, pageInfo *PageInfo) (IssuesResponse, error) {
	var err error
	if client == nil {
		client, err = gh.DefaultGraphQLClient()
	}

	if err != nil {
		return IssuesResponse{}, err
	}

	var queryResult struct {
		Search struct {
			Nodes []struct {
				Issue IssueData `graphql:"... on Issue"`
			}
			IssueCount int
			PageInfo   PageInfo
		} `graphql:"search(type: ISSUE, first: $limit, after: $endCursor, query: $query)"`
	}
	var endCursor *string
	if pageInfo != nil {
		endCursor = &pageInfo.EndCursor
	}
	variables := map[string]any{
		"query":     graphql.String(makeIssuesQuery(query)),
		"limit":     graphql.Int(limit),
		"endCursor": (*graphql.String)(endCursor),
	}
	log.Debug("Fetching issues", "query", query, "limit", limit, "endCursor", endCursor)
	err = client.Query("SearchIssues", &queryResult, variables)
	if err != nil {
		return IssuesResponse{}, err
	}
	log.Info("Successfully fetched issues", "query", query, "count", queryResult.Search.IssueCount)

	issues := make([]IssueData, 0, len(queryResult.Search.Nodes))
	for _, node := range queryResult.Search.Nodes {
		issues = append(issues, node.Issue)
	}

	return IssuesResponse{
		Issues:     issues,
		TotalCount: queryResult.Search.IssueCount,
		PageInfo:   queryResult.Search.PageInfo,
	}, nil
}

type IssuesResponse struct {
	Issues     []IssueData
	TotalCount int
	PageInfo   PageInfo
}

// FetchIssue fetches a single issue by its GitHub URL
func FetchIssue(issueUrl string) (IssueData, error) {
	var err error
	if client == nil {
		client, err = gh.DefaultGraphQLClient()
		if err != nil {
			return IssueData{}, err
		}
	}

	var queryResult struct {
		Resource struct {
			Issue IssueData `graphql:"... on Issue"`
		} `graphql:"resource(url: $url)"`
	}
	parsedUrl, err := url.Parse(issueUrl)
	if err != nil {
		return IssueData{}, err
	}
	variables := map[string]any{
		"url": githubv4.URI{URL: parsedUrl},
	}
	log.Debug("Fetching Issue", "url", issueUrl)
	err = client.Query("FetchIssue", &queryResult, variables)
	if err != nil {
		return IssueData{}, err
	}
	log.Info("Successfully fetched Issue", "url", issueUrl)

	return queryResult.Resource.Issue, nil
}
