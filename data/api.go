package data

import (
	"fmt"
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
	Body   string
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

type CheckRun struct {
	Title      graphql.String
	Url        graphql.String
	CheckSuite struct {
		App struct {
			Name graphql.String
		}
		Creator struct {
			Login graphql.String
		}
		WorkflowRun struct {
			Workflow struct {
				Name graphql.String
			}
		}
	}
	Summary    graphql.String
	Text       graphql.String
	Name       graphql.String
	Status     graphql.String
	Conclusion graphql.String
}

type StatusContext struct {
	Context     graphql.String
	Description graphql.String
	State       graphql.String
	Creator     struct {
		Login graphql.String
	}
}

type Commits struct {
	Nodes []struct {
		Commit struct {
			Deployments struct {
				Nodes []struct {
					Task        graphql.String
					Description graphql.String
				}
			} `graphql:"deployments(last: 10)"`
			StatusCheckRollup struct {
				Contexts struct {
					TotalCount graphql.Int
					Nodes      []struct {
						Typename      graphql.String `graphql:"__typename"`
						CheckRun      CheckRun       `graphql:"... on CheckRun"`
						StatusContext StatusContext  `graphql:"... on StatusContext"`
					}
				} `graphql:"contexts(last: 20)"`
			}
		}
	}
}

func makeQuery(query string) string {
	return fmt.Sprintf("is:pr %s", query)
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
		"query": graphql.String(makeQuery(query)),
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
