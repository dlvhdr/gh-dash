package data

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"charm.land/log/v2"
	gh "github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	checks "github.com/dlvhdr/x/gh-checks"
	"github.com/shurcooL/githubv4"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

type AutoMergeRequest struct {
	EnabledAt   time.Time `graphql:"enabledAt"`
	MergeMethod string    `graphql:"mergeMethod"`
	EnabledBy   struct {
		Login string `graphql:"login"`
	} `graphql:"enabledBy"`
}

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
	ID     string
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
	AutoMergeRequest *AutoMergeRequest
	// AutoMergeEnabled is a local UI flag set when the user enables auto-merge
	// via the TUI.  It is NOT fetched from the API; it mirrors the same field
	// on prrow.Data so that the branch section can show the auto-merge icon
	// immediately without waiting for a full refresh.
	AutoMergeEnabled bool
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

type CheckSuiteNode struct {
	Status     graphql.String
	Conclusion graphql.String

	App struct {
		Name graphql.String
	}

	WorkflowRun struct {
		Workflow struct {
			Name graphql.String
		}
	}
}

type CheckSuites struct {
	TotalCount graphql.Int
	Nodes      []CheckSuiteNode
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
			// CheckSuites are fetched separately from StatusCheckRollup because
			// workflows awaiting approval (conclusion ACTION_REQUIRED) and workflows
			// still queued have no CheckRun objects yet, so they don’t appear in
			// StatusCheckRollup.contexts.
			CheckSuites CheckSuites `graphql:"checkSuites(last: 20)"`
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
	clientMu     sync.Mutex
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
	clientMu.Lock()
	if client == nil {
		if config.IsFeatureEnabled(config.FF_MOCK_DATA) {
			log.Info("using mock data", "server", "https://localhost:3000")
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
			client, err = gh.NewGraphQLClient(
				gh.ClientOptions{Host: "localhost:3000", AuthToken: "fake-token"},
			)
		} else {
			client, err = gh.DefaultGraphQLClient()
		}
	}
	clientMu.Unlock()

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
	clientMu.Lock()
	if client == nil {
		client, err = gh.DefaultGraphQLClient()
		if err != nil {
			clientMu.Unlock()
			return EnrichedPullRequestData{}, err
		}
	}
	clientMu.Unlock()

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

// PRMergeStatus represents the outcome of a merge action.
type PRMergeStatus struct {
	State        string // OPEN, CLOSED, MERGED
	HasAutoMerge bool
}

// mergeMethodForRepo returns the GraphQL merge method string to use based on
// which methods the repository allows, in priority order: MERGE > SQUASH > REBASE.
//
// NOTE: Future enhancement — when multiple merge methods are allowed, consider:
//   - Prompting the user to select from the available options before merging
//   - Splitting MergePR into separate MergePR / SquashPR / RebasePR task functions
//     so the user can invoke them with explicit keybindings
func mergeMethodForRepo(repo Repository) (graphql.String, error) {
	switch {
	case repo.AllowMergeCommit:
		return "MERGE", nil
	case repo.AllowSquashMerge:
		return "SQUASH", nil
	case repo.AllowRebaseMerge:
		return "REBASE", nil
	default:
		return "", fmt.Errorf("repository has no allowed merge methods")
	}
}

// enableAutoMerge calls the enablePullRequestAutoMerge GraphQL mutation and
// returns the resulting PRMergeStatus.
func enableAutoMerge(prNodeID string, mergeMethod graphql.String) (PRMergeStatus, error) {
	var mutation struct {
		EnablePullRequestAutoMerge struct {
			PullRequest struct {
				State            string `graphql:"state"`
				AutoMergeRequest *struct {
					EnabledAt time.Time `graphql:"enabledAt"`
				} `graphql:"autoMergeRequest"`
			} `graphql:"pullRequest"`
		} `graphql:"enablePullRequestAutoMerge(input: $input)"`
	}
	type EnablePullRequestAutoMergeInput struct {
		PullRequestID string         `json:"pullRequestId"`
		MergeMethod   graphql.String `json:"mergeMethod"`
	}
	variables := map[string]any{
		"input": EnablePullRequestAutoMergeInput{
			PullRequestID: prNodeID,
			MergeMethod:   mergeMethod,
		},
	}
	if err := client.Mutate("EnablePullRequestAutoMerge", &mutation, variables); err != nil {
		return PRMergeStatus{}, err
	}
	hasAutoMerge := mutation.EnablePullRequestAutoMerge.PullRequest.AutoMergeRequest != nil
	log.Info("Auto-merge enabled for PR", "nodeID", prNodeID, "hasAutoMerge", hasAutoMerge)
	return PRMergeStatus{
		State:        mutation.EnablePullRequestAutoMerge.PullRequest.State,
		HasAutoMerge: hasAutoMerge,
	}, nil
}

// MergePullRequest performs the appropriate merge action via a single GraphQL
// mutation, returning the outcome directly from the mutation response without
// a follow-up query.
//
// The mutation chosen depends on the PR's current mergeStateStatus
// https://docs.github.com/en/graphql/reference/enums#mergestatestatus
//   - CLEAN, UNSTABLE, or HAS_HOOKS → mergePullRequest
//   - BLOCKED                       → enablePullRequestAutoMerge
//   - BEHIND                        → enablePullRequestAutoMerge
//   - DIRTY                         → error
//   - UNKNOWN                       → log warning, attempt enablePullRequestAutoMerge
//
// Note: draft PRs are not handled here; callers should check IsDraft before
// invoking this function.
func MergePullRequest(prNodeID string, mergeStateStatus MergeStateStatus, repo Repository) (PRMergeStatus, error) {
	clientMu.Lock()
	if client == nil {
		var err error
		if config.IsFeatureEnabled(config.FF_MOCK_DATA) {
			log.Info("using mock data", "server", "https://localhost:3000")
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
			client, err = gh.NewGraphQLClient(gh.ClientOptions{Host: "localhost:3000", AuthToken: "fake-token"})
		} else {
			client, err = gh.DefaultGraphQLClient()
		}
		if err != nil {
			clientMu.Unlock()
			return PRMergeStatus{}, err
		}
	}
	clientMu.Unlock()

	log.Debug("Performing PR merge via GraphQL", "nodeID", prNodeID, "mergeStateStatus", mergeStateStatus)

	switch mergeStateStatus {
	case "CLEAN", "UNSTABLE", "HAS_HOOKS":
		mergeMethod, err := mergeMethodForRepo(repo)
		if err != nil {
			return PRMergeStatus{}, err
		}

		var mutation struct {
			MergePullRequest struct {
				PullRequest struct {
					State string `graphql:"state"`
				} `graphql:"pullRequest"`
			} `graphql:"mergePullRequest(input: $input)"`
		}
		type MergePullRequestInput struct {
			PullRequestID string         `json:"pullRequestId"`
			MergeMethod   graphql.String `json:"mergeMethod"`
		}
		variables := map[string]any{
			"input": MergePullRequestInput{
				PullRequestID: prNodeID,
				MergeMethod:   mergeMethod,
			},
		}
		if err := client.Mutate("MergePullRequest", &mutation, variables); err != nil {
			return PRMergeStatus{}, err
		}
		log.Info("PR merged directly", "nodeID", prNodeID, "state", mutation.MergePullRequest.PullRequest.State)
		return PRMergeStatus{State: mutation.MergePullRequest.PullRequest.State}, nil

	case "DIRTY":
		return PRMergeStatus{}, fmt.Errorf("PR has merge conflicts, please resolve locally")

	case "BLOCKED", "BEHIND":
		mergeMethod, err := mergeMethodForRepo(repo)
		if err != nil {
			return PRMergeStatus{}, err
		}
		return enableAutoMerge(prNodeID, mergeMethod)

	case "UNKNOWN":
		// UNKNOWN is returned by GitHub when the merge state cannot be
		// determined (e.g. checks are still pending or the state is
		// temporarily unavailable).  We optimistically attempt to enable
		// auto-merge so that the PR merges automatically once the
		// blocking condition clears.  The warning is intentionally
		// log-only; surfacing it to the user is left as a future
		// enhancement (see: https://github.com/dlvhdr/gh-dash/discussions/546).
		log.Warn("Unknown merge state status, attempting auto-merge", "status", mergeStateStatus)
		mergeMethod, err := mergeMethodForRepo(repo)
		if err != nil {
			return PRMergeStatus{}, err
		}
		return enableAutoMerge(prNodeID, mergeMethod)

	default:
		// Any future MergeStateStatus values not listed above fall here.
		// We apply the same optimistic enableAutoMerge strategy rather
		// than hard-failing, so that newly-introduced GitHub statuses
		// don't silently break the merge workflow.  As with UNKNOWN,
		// the warning is log-only for now; a user-visible notification
		// is a future enhancement.
		log.Warn("Unrecognised merge state status, attempting auto-merge", "status", mergeStateStatus)
		mergeMethod, err := mergeMethodForRepo(repo)
		if err != nil {
			return PRMergeStatus{}, err
		}
		return enableAutoMerge(prNodeID, mergeMethod)
	}
}
