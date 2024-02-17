package data

import (
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	gh "github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
)

type IssueData struct {
	Number int
	Title  string
	Body   string
	State  string
	Author struct {
		Login string
	}
	UpdatedAt  time.Time
	Url        string
	Repository Repository
	Assignees  Assignees      `graphql:"assignees(first: 3)"`
	Comments   IssueComments  `graphql:"comments(first: 15)"`
	Reactions  IssueReactions `graphql:"reactions(first: 1)"`
	Labels     IssueLabels    `graphql:"labels(first: 3)"`
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

func makeIssuesQuery(query string) string {
	return fmt.Sprintf("is:issue %s sort:updated", query)
}

func FetchIssues(query string, limit int, pageInfo *PageInfo) (IssuesResponse, error) {
	var err error
	client, err := gh.DefaultGraphQLClient()
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
	variables := map[string]interface{}{
		"query":     graphql.String(makeIssuesQuery(query)),
		"limit":     graphql.Int(limit),
		"endCursor": (*graphql.String)(endCursor),
	}
	log.Debug("Fetching issues", "query", query, "limit", limit, "endCursor", endCursor)
	err = client.Query("SearchIssues", &queryResult, variables)
	if err != nil {
		return IssuesResponse{}, err
	}
	log.Debug("Successfully fetched issues", "query", query, "count", queryResult.Search.IssueCount)

	issues := make([]IssueData, 0, len(queryResult.Search.Nodes))
	for _, node := range queryResult.Search.Nodes {
		if node.Issue.Repository.IsArchived {
			continue
		}
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
