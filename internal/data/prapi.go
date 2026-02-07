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
	checks "github.com/dlvhdr/x/gh-checks"
	"github.com/shurcooL/githubv4"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

type SuggestedReviewer struct {
	IsAuthor    bool
	IsCommenter bool
	Reviewer    struct {
		Login string
	}
}

type EnrichedPullRequestData struct {
	Url     string
	Number  int
	Title   string
	Body    string
	State   string
	IsDraft bool
	Author  struct {
		Login string
	}
	AuthorAssociation string
	UpdatedAt         time.Time
	CreatedAt         time.Time
	Mergeable         string
	ReviewDecision    string
	Additions         int
	Deletions         int
	HeadRefName       string
	BaseRefName       string
	HeadRepository    struct {
		Name string
	}
	HeadRef struct {
		Name string
	}
	Labels             PRLabels  `graphql:"labels(first: 6)"`
	Assignees          Assignees `graphql:"assignees(first: 3)"`
	Repository         Repository
	Commits            LastCommitWithStatusChecks `graphql:"commits(last: 1)"`
	AllCommits         AllCommits                 `graphql:"allCommits: commits(last: 100)"`
	Comments           CommentsWithBody           `graphql:"comments(last: 50, orderBy: { field: UPDATED_AT, direction: DESC })"`
	ReviewThreads      ReviewThreadsWithComments  `graphql:"reviewThreads(last: 50)"`
	ReviewRequests     ReviewRequests             `graphql:"reviewRequests(last: 100)"`
	Reviews            Reviews                    `graphql:"reviews(last: 100)"`
	SuggestedReviewers []SuggestedReviewer
	Files              ChangedFiles `graphql:"files(first: 5)"`
}

type PullRequestData struct {
	Number int
	Title  string
	Body   string
	Author struct {
		Login string
	}
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
	HeadRepository    struct {
		Name string
	}
	HeadRef struct {
		Name string
	}
	Repository       Repository
	Assignees        Assignees      `graphql:"assignees(first: 3)"`
	Comments         Comments       `graphql:"comments"`
	ReviewThreads    ReviewThreads  `graphql:"reviewThreads"`
	Reviews          Reviews        `graphql:"reviews(last: 3)"`
	ReviewRequests   ReviewRequests `graphql:"reviewRequests(last: 5)"`
	Files            ChangedFiles   `graphql:"files(first: 5)"`
	IsDraft          bool
	IsInMergeQueue   bool
	Commits          Commits          `graphql:"commits(last: 1)"`
	Labels           PRLabels         `graphql:"labels(first: 6)"`
	MergeStateStatus MergeStateStatus `graphql:"mergeStateStatus"`
}

type CheckRun struct {
	Name       graphql.String
	Status     graphql.String
	Conclusion checks.CheckRunState
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

type StatusCheckRollupStats struct {
	State    checks.CommitState
	Contexts struct {
		TotalCount                 graphql.Int
		CheckRunCount              graphql.Int
		CheckRunCountsByState      []ContextCountByState
		StatusContextCount         graphql.Int
		StatusContextCountsByState []ContextCountByState
	} `graphql:"contexts(last: 1)"`
}

type AllCommits struct {
	Nodes []struct {
		Commit struct {
			AbbreviatedOid  string
			CommittedDate   time.Time
			MessageHeadline string
			Author          struct {
				Name string
				User struct {
					Login string
				}
			}
			StatusCheckRollup StatusCheckRollupStats
		}
	}
}

type LastCommitWithStatusChecks struct {
	Nodes []struct {
		Commit struct {
			Deployments struct {
				Nodes []struct {
					Task        graphql.String
					Description graphql.String
				}
			} `graphql:"deployments(last: 10)"`
			CommitUrl         graphql.String
			StatusCheckRollup struct {
				State    graphql.String
				Contexts struct {
					TotalCount                 graphql.Int
					CheckRunCount              graphql.Int
					CheckRunCountsByState      []ContextCountByState
					StatusContextCount         graphql.Int
					StatusContextCountsByState []ContextCountByState
					Nodes                      []struct {
						Typename      graphql.String `graphql:"__typename"`
						CheckRun      CheckRun       `graphql:"... on CheckRun"`
						StatusContext StatusContext  `graphql:"... on StatusContext"`
					}
				} `graphql:"contexts(last: 100)"`
			}
		}
	}
	TotalCount int
}

type CommentsWithBody struct {
	TotalCount graphql.Int
	Nodes      []Comment
}

type ContextCountByState = struct {
	Count graphql.Int
	State checks.CheckRunState
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
			CommitUrl         graphql.String
			StatusCheckRollup struct {
				State graphql.String
			}
		}
	}
	TotalCount int
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
	TotalCount int
}

type ReviewThreads struct {
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
	TotalCount int
	Nodes      []Review
}

type ReviewThreadsWithComments struct {
	Nodes []struct {
		Id           string
		IsOutdated   bool
		OriginalLine int
		StartLine    int
		Line         int
		Path         string
		Comments     ReviewComments `graphql:"comments(first: 20)"`
	}
}

type ChangedFile struct {
	Additions  int
	Deletions  int
	Path       string
	ChangeType string
}

type ChangedFiles struct {
	TotalCount int
	Nodes      []ChangedFile
}

type RequestedReviewerUser struct {
	Login string `graphql:"login"`
}

type RequestedReviewerTeam struct {
	Slug string `graphql:"slug"`
	Name string `graphql:"name"`
}

type RequestedReviewerBot struct {
	Login string `graphql:"login"`
}

type RequestedReviewerMannequin struct {
	Login string `graphql:"login"`
}

type ReviewRequestNode struct {
	AsCodeOwner       bool `graphql:"asCodeOwner"`
	RequestedReviewer struct {
		User      RequestedReviewerUser      `graphql:"... on User"`
		Team      RequestedReviewerTeam      `graphql:"... on Team"`
		Bot       RequestedReviewerBot       `graphql:"... on Bot"`
		Mannequin RequestedReviewerMannequin `graphql:"... on Mannequin"`
	} `graphql:"requestedReviewer"`
}

type ReviewRequests struct {
	TotalCount int
	Nodes      []ReviewRequestNode
}

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

func (r ReviewRequestNode) IsTeam() bool {
	return r.RequestedReviewer.Team.Slug != ""
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

func (data PullRequestData) GetAuthor(theme theme.Theme, showAuthorIcon bool) string {
	author := data.Author.Login
	if showAuthorIcon {
		author += fmt.Sprintf(" %s", GetAuthorRoleIcon(data.AuthorAssociation, theme))
	}
	return author
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

func (data PullRequestData) GetCreatedAt() time.Time {
	return data.CreatedAt
}

// ToPullRequestData converts EnrichedPullRequestData to PullRequestData
// This is useful when we fetch a single PR and need basic PR fields
func (e EnrichedPullRequestData) ToPullRequestData() PullRequestData {
	return PullRequestData{
		Number:            e.Number,
		Title:             e.Title,
		Body:              e.Body,
		Author:            e.Author,
		AuthorAssociation: e.AuthorAssociation,
		UpdatedAt:         e.UpdatedAt,
		CreatedAt:         e.CreatedAt,
		Url:               e.Url,
		State:             e.State,
		Mergeable:         e.Mergeable,
		ReviewDecision:    e.ReviewDecision,
		Additions:         e.Additions,
		Deletions:         e.Deletions,
		HeadRefName:       e.HeadRefName,
		BaseRefName:       e.BaseRefName,
		HeadRepository:    e.HeadRepository,
		HeadRef:           e.HeadRef,
		Repository:        e.Repository,
		Assignees:         e.Assignees,
		IsDraft:           e.IsDraft,
		Labels:            e.Labels,
		Files:             e.Files,
		// Note: Comments, ReviewThreads, Reviews, ReviewRequests, Commits
		// have different types in EnrichedPullRequestData vs PullRequestData
		// We leave them as zero values since the enriched data will be used instead
	}
}

func makePullRequestsQuery(query string) string {
	return fmt.Sprintf("is:pr archived:false %s sort:updated", query)
}

type PullRequestsResponse struct {
	Prs        []PullRequestData
	TotalCount int
	PageInfo   PageInfo
}

var (
	client       *gh.GraphQLClient
	cachedClient *gh.GraphQLClient
)

func SetClient(c *gh.GraphQLClient) {
	client = c
	cachedClient = c
}

// ClearEnrichmentCache clears the cached GraphQL client used for fetching
// enriched PR/Issue data. Call this when refreshing to ensure fresh data.
func ClearEnrichmentCache() {
	cachedClient = nil
}

// IsEnrichmentCacheCleared returns true if the enrichment cache is cleared.
// This is primarily for testing purposes.
func IsEnrichmentCacheCleared() bool {
	return cachedClient == nil
}

func FetchPullRequests(query string, limit int, pageInfo *PageInfo) (PullRequestsResponse, error) {
	var err error
	if client == nil {
		if config.IsFeatureEnabled(config.FF_MOCK_DATA) {
			log.Info("using mock data", "server", "https://localhost:3000")
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
	variables := map[string]any{
		"query":     graphql.String(makePullRequestsQuery(query)),
		"limit":     graphql.Int(limit),
		"endCursor": (*graphql.String)(endCursor),
	}
	log.Debug("Fetching PRs", "query", query, "limit", limit, "endCursor", endCursor)
	err = client.Query("SearchPullRequests", &queryResult, variables)
	if err != nil {
		return PullRequestsResponse{}, err
	}
	log.Info("Successfully fetched PRs", "count", queryResult.Search.IssueCount)

	prs := make([]PullRequestData, 0, len(queryResult.Search.Nodes))
	for _, node := range queryResult.Search.Nodes {
		prs = append(prs, node.PullRequest)
	}

	return PullRequestsResponse{
		Prs:        prs,
		TotalCount: queryResult.Search.IssueCount,
		PageInfo:   queryResult.Search.PageInfo,
	}, nil
}

func FetchPullRequest(prUrl string) (EnrichedPullRequestData, error) {
	var err error
	if client == nil {
		client, err = gh.DefaultGraphQLClient()
		if err != nil {
			return EnrichedPullRequestData{}, err
		}
	}

	var queryResult struct {
		Resource struct {
			PullRequest EnrichedPullRequestData `graphql:"... on PullRequest"`
		} `graphql:"resource(url: $url)"`
	}
	parsedUrl, err := url.Parse(prUrl)
	if err != nil {
		return EnrichedPullRequestData{}, err
	}
	variables := map[string]any{
		"url": githubv4.URI{URL: parsedUrl},
	}
	log.Debug("Fetching PR", "url", prUrl)
	err = client.Query("FetchPullRequest", &queryResult, variables)
	if err != nil {
		return EnrichedPullRequestData{}, err
	}
	log.Info("Successfully fetched PR", "url", prUrl)

	return queryResult.Resource.PullRequest, nil
}
