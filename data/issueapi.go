package data

import (
	"fmt"
	"log"
	"time"

	"github.com/cli/go-gh"
	graphql "github.com/cli/shurcooL-graphql"
)

type IssueData struct {
	Number int
	Title  string
	Body   string
	Author struct {
		Login string
	}
	UpdatedAt  time.Time
	Url        string
	Repository struct {
		Name string
	}
	Assignees Assignees      `graphql:"assignees(first: 3)"`
	Comments  IssueComments  `graphql:"comments(first: 15)"`
	Reactions IssueReactions `graphql:"reactions(first: 1)"`
}

type Assignees struct {
	Nodes []Assignee
}

type Assignee struct {
	Login string
}

type IssueComments struct {
	TotalCount int
}

type IssueReactions struct {
	TotalCount int
}

func (data IssueData) GetUrl() string {
	return data.Url
}

func (data IssueData) GetUpdatedAt() time.Time {
	return data.UpdatedAt
}

func makeIssuesQuery(query string) string {
	return fmt.Sprintf("is:issue %s", query)
}

func FetchIssues(query string, limit int) ([]IssueData, error) {
	var err error
	client, err := gh.GQLClient(nil)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	var queryResult struct {
		Search struct {
			Nodes []struct {
				Issue IssueData `graphql:"... on Issue"`
			}
		} `graphql:"search(type: ISSUE, first: $limit, query: $query)"`
	}
	variables := map[string]interface{}{
		"query": graphql.String(makeIssuesQuery(query)),
		"limit": graphql.Int(limit),
	}
	err = client.Query("SearchIssues", &queryResult, variables)
	if err != nil {
		return nil, err
	}

	issues := make([]IssueData, 0, len(queryResult.Search.Nodes))
	for _, node := range queryResult.Search.Nodes {
		issues = append(issues, node.Issue)
	}
	return issues, nil
}
