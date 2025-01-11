package data

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/charmbracelet/log"
	gh "github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	"github.com/shurcooL/githubv4"

	"github.com/dlvhdr/gh-dash/v4/config"
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
	Repository       Repository
	Assignees        Assignees     `graphql:"assignees(first: 3)"`
	Comments         Comments      `graphql:"comments(last: 5, orderBy: { field: UPDATED_AT, direction: DESC })"`
	LatestReviews    Reviews       `graphql:"latestReviews(last: 3)"`
	ReviewThreads    ReviewThreads `graphql:"reviewThreads(last: 20)"`
	IsDraft          bool
	Commits          Commits          `graphql:"commits(last: 1)"`
	Labels           PRLabels         `graphql:"labels(first: 3)"`
	MergeStateStatus MergeStateStatus `graphql:"mergeStateStatus"`
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

type ReviewComment struct {
	Author struct {
		Login string
	}
	Body      string
	UpdatedAt time.Time
	StartLine int
	Line      int
}

type ReviewComments struct {
	Nodes      []ReviewComment
	TotalCount int
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

type ReviewThreads struct {
	Nodes []struct {
		Id           string
		IsOutdated   bool
		OriginalLine int
		StartLine    int
		Line         int
		Path         string
		Comments     ReviewComments `graphql:"comments(first: 10)"`
	}
}

type PRLabel struct {
	Color string
	Name  string
}

type PRLabels struct {
	Nodes []Label
}

type MergeStateStatus string

type PageInfo struct {
	HasNextPage bool
	StartCursor string
	EndCursor   string
}

func (data PullRequestData) GetTitle() string {
	return data.Title
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
	return fmt.Sprintf("is:pr %s sort:updated", query)
}

type PullRequestsResponse struct {
	Prs        []PullRequestData
	TotalCount int
	PageInfo   PageInfo
}

var client *gh.GraphQLClient

func FetchPullRequests(query string, limit int, pageInfo *PageInfo) (PullRequestsResponse, error) {
	var err error
	if client == nil {
		if config.IsFeatureEnabled(config.FF_MOCK_DATA) {
			log.Debug("using mock data", "server", "https://localhost:3000")
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
			client, err = gh.NewGraphQLClient(gh.ClientOptions{Host: "localhost:3000", AuthToken: "fake-token"})
		} else {
			client, err = gh.DefaultGraphQLClient()
		}
	}

	if err != nil {
		return PullRequestsResponse{}, err
	}

	var queryResult struct {
		Search struct {
			Nodes []struct {
				PullRequest PullRequestData `graphql:"... on PullRequest"`
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
		"query":     graphql.String(makePullRequestsQuery(query)),
		"limit":     graphql.Int(limit),
		"endCursor": (*graphql.String)(endCursor),
	}
	log.Debug("Fetching PRs", "query", query, "limit", limit, "endCursor", endCursor)
	err = client.Query("SearchPullRequests", &queryResult, variables)
	if err != nil {
		return PullRequestsResponse{}, err
	}
	log.Debug("Successfully fetched PRs", "count", queryResult.Search.IssueCount)

	prs := make([]PullRequestData, 0, len(queryResult.Search.Nodes))
	for _, node := range queryResult.Search.Nodes {
		if node.PullRequest.Repository.IsArchived {
			continue
		}
		prs = append(prs, node.PullRequest)
	}

	return PullRequestsResponse{
		Prs:        prs,
		TotalCount: queryResult.Search.IssueCount,
		PageInfo:   queryResult.Search.PageInfo,
	}, nil
}

func FetchPullRequest(prUrl string) (PullRequestData, error) {
	var err error
	client, err := gh.DefaultGraphQLClient()

	if err != nil {
		return PullRequestData{}, err
	}

	var queryResult struct {
		Resource struct {
			PullRequest PullRequestData `graphql:"... on PullRequest"`
		} `graphql:"resource(url: $url)"`
	}
	parsedUrl, err := url.Parse(prUrl)
	if err != nil {
		return PullRequestData{}, err
	}
	variables := map[string]interface{}{
		"url": githubv4.URI{URL: parsedUrl},
	}
	log.Debug("Fetching PR", "url", prUrl)
	err = client.Query("FetchPullRequest", &queryResult, variables)
	if err != nil {
		return PullRequestData{}, err
	}
	log.Debug("Successfully fetched PR", "url", prUrl)

	return queryResult.Resource.PullRequest, nil
}
