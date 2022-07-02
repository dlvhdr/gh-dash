package data

import (
	"fmt"
	"log"
	"time"

	"github.com/cli/go-gh"
	graphql "github.com/cli/shurcooL-graphql"
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
	BaseRefName    string
	HeadRepository struct {
		Name string
	}
	HeadRef struct {
		Name string
	}
	Repository struct {
		NameWithOwner string
	}
	Comments      Comments `graphql:"comments(last: 5, orderBy: { field: UPDATED_AT, direction: DESC })"`
	LatestReviews Reviews  `graphql:"latestReviews(last: 3)"`
	IsDraft       bool
	Commits       Commits `graphql:"commits(last: 1)"`
}

type CheckRun struct {
	Name       graphql.String
	Status     graphql.String
	Conclusion graphql.String
	CheckSuite struct {
		Creator struct {
			Login graphql.String
		}
		WorkflowRun struct {
			Workflow struct {
				Name graphql.String
			}
		}
	}
}

type StatusContext struct {
	Context graphql.String
	State   graphql.String
	Creator struct {
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

type Comment struct {
	Author struct {
		Login string
	}
	Body      string
	UpdatedAt time.Time
}

type Comments struct {
	Nodes      []Comment
	TotalCount int
}

type Review struct {
	Author struct {
		Login string
	}
	Body      string
	State     string
	UpdatedAt time.Time
}

type Reviews struct {
	Nodes []Review
}

func (data PullRequestData) GetRepoNameWithOwner() string {
	return data.Repository.NameWithOwner
}

func (data PullRequestData) GetNumber() int {
	return data.Number
}

func (data PullRequestData) GetUrl() string {
	return data.Url
}

func (data PullRequestData) GetUpdatedAt() time.Time {
	return data.UpdatedAt
}

func makePullRequestsQuery(query string) string {
	return fmt.Sprintf("is:pr %s", query)
}

func FetchPullRequests(query string, limit int) ([]PullRequestData, error) {
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
		"query": graphql.String(makePullRequestsQuery(query)),
		"limit": graphql.Int(limit),
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
