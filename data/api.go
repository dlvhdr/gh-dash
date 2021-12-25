package data

import (
	"log"
	"time"

	"github.com/cli/go-gh"
	graphql "github.com/cli/shurcooL-graphql"
)

const (
	Limit = 20
)

type PullRequestData struct {
	Number int
	Title  string
	Author struct {
		Login string
	}
	UpdatedAt      time.Time
	Url            string
	State          string
	Mergeable      string
	ReviewDecision string
	Additions      int
	Deletions      int
	HeadRefName    string
	HeadRepository struct {
		Name string
	}
	IsDraft bool
	Commits Commits `graphql:"commits(last: 1)"`
}

type Commits struct {
	Nodes []struct {
		Commit struct {
			StatusCheckRollup struct {
				Contexts struct {
					Nodes []struct {
						CheckRun struct {
							Status     graphql.String
							Conclusion graphql.String
						} `graphql:"... on CheckRun"`
					}
				} `graphql:"contexts(last: 20)"`
			}
		}
	}
}

func FetchRepoPullRequests(query string) ([]PullRequestData, error) {
	var err error
	client, err := gh.GQLClient(nil)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	var queryResult struct {
		Search struct {
			Nodes []struct {
				PullRequest PullRequestData `graphql:"... on PullRequest"`
			}
		} `graphql:"search(type: ISSUE, first: $limit, query: $query)"`
	}
	variables := map[string]interface{}{
		"query": graphql.String(query),
		"limit": graphql.Int(Limit),
	}
	err = client.Query("SearchPullRequests", &queryResult, variables)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	prs := make([]PullRequestData, 0, len(queryResult.Search.Nodes))
	for _, node := range queryResult.Search.Nodes {
		prs = append(prs, node.PullRequest)
	}
	return prs, nil
}
