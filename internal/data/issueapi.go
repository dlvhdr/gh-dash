package data

import (
	"fmt"
	"net/url"
	"time"

	"github.com/charmbracelet/log"
	gh "github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	"github.com/shurcooL/githubv4"

	"github.com/dlvhdr/gh-dash/v4/internal/provider"
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
	Color string
	Name  string
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
	// Use GitLab provider if configured
	if provider.IsGitLab() {
		return fetchIssuesFromGitLab(query, limit, pageInfo)
	}

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

// fetchIssuesFromGitLab fetches issues from GitLab and converts them to the internal format
func fetchIssuesFromGitLab(query string, limit int, pageInfo *PageInfo) (IssuesResponse, error) {
	p := provider.GetProvider()

	var providerPageInfo *provider.PageInfo
	if pageInfo != nil {
		providerPageInfo = &provider.PageInfo{
			HasNextPage: pageInfo.HasNextPage,
			StartCursor: pageInfo.StartCursor,
			EndCursor:   pageInfo.EndCursor,
		}
	}

	resp, err := p.FetchIssues(query, limit, providerPageInfo)
	if err != nil {
		return IssuesResponse{}, err
	}

	issues := make([]IssueData, len(resp.Issues))
	for i, issue := range resp.Issues {
		issues[i] = convertProviderIssueToData(issue)
	}

	return IssuesResponse{
		Issues:     issues,
		TotalCount: resp.TotalCount,
		PageInfo: PageInfo{
			HasNextPage: resp.PageInfo.HasNextPage,
			StartCursor: resp.PageInfo.StartCursor,
			EndCursor:   resp.PageInfo.EndCursor,
		},
	}, nil
}

// convertProviderIssueToData converts provider.IssueData to data.IssueData
func convertProviderIssueToData(issue provider.IssueData) IssueData {
	assignees := make([]Assignee, len(issue.Assignees.Nodes))
	for i, a := range issue.Assignees.Nodes {
		assignees[i] = Assignee{Login: a.Login}
	}

	labels := make([]Label, len(issue.Labels.Nodes))
	for i, l := range issue.Labels.Nodes {
		labels[i] = Label{Name: l.Name, Color: l.Color}
	}

	comments := make([]IssueComment, len(issue.Comments.Nodes))
	for i, c := range issue.Comments.Nodes {
		comments[i] = IssueComment{
			Author:    struct{ Login string }{Login: c.Author.Login},
			Body:      c.Body,
			UpdatedAt: c.UpdatedAt,
		}
	}

	return IssueData{
		Number: issue.Number,
		Title:  issue.Title,
		Body:   issue.Body,
		State:  issue.State,
		Author: struct{ Login string }{Login: issue.Author.Login},
		AuthorAssociation: issue.AuthorAssociation,
		UpdatedAt:         issue.UpdatedAt,
		CreatedAt:         issue.CreatedAt,
		Url:               issue.Url,
		Repository:        Repository{NameWithOwner: issue.Repository.NameWithOwner, IsArchived: issue.Repository.IsArchived},
		Assignees:         Assignees{Nodes: assignees},
		Comments:          IssueComments{TotalCount: issue.Comments.TotalCount, Nodes: comments},
		Reactions:         IssueReactions{TotalCount: issue.Reactions.TotalCount},
		Labels:            IssueLabels{Nodes: labels},
	}
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

// FetchIssueComments fetches comments for a single issue (GitLab only)
func FetchIssueComments(issueUrl string) ([]IssueComment, error) {
	if !provider.IsGitLab() {
		return nil, nil // GitHub fetches comments in the main query
	}

	p := provider.GetProvider()
	providerComments, err := p.FetchIssueComments(issueUrl)
	if err != nil {
		return nil, err
	}

	comments := make([]IssueComment, len(providerComments))
	for i, c := range providerComments {
		comments[i] = IssueComment{
			Author:    struct{ Login string }{Login: c.Author.Login},
			Body:      c.Body,
			UpdatedAt: c.UpdatedAt,
		}
	}

	return comments, nil
}
