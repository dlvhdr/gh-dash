package provider

// GitHubProvider implements the Provider interface for GitHub
// GitHub uses the gh CLI and the existing data package functions
type GitHubProvider struct {
	host string
}

// NewGitHubProvider creates a new GitHub provider
func NewGitHubProvider() *GitHubProvider {
	return &GitHubProvider{
		host: "github.com",
	}
}

func (g *GitHubProvider) GetType() ProviderType {
	return GitHub
}

func (g *GitHubProvider) GetHost() string {
	return g.host
}

func (g *GitHubProvider) GetCLICommand() string {
	return "gh"
}

// FetchPullRequests is not implemented for GitHub provider
// The data package handles GitHub directly
func (g *GitHubProvider) FetchPullRequests(query string, limit int, pageInfo *PageInfo) (PullRequestsResponse, error) {
	// This should not be called - data package handles GitHub directly
	panic("GitHubProvider.FetchPullRequests should not be called directly")
}

// FetchPullRequest is not implemented for GitHub provider
// The data package handles GitHub directly
func (g *GitHubProvider) FetchPullRequest(prUrl string) (EnrichedPullRequestData, error) {
	// This should not be called - data package handles GitHub directly
	panic("GitHubProvider.FetchPullRequest should not be called directly")
}

// FetchIssues is not implemented for GitHub provider
// The data package handles GitHub directly
func (g *GitHubProvider) FetchIssues(query string, limit int, pageInfo *PageInfo) (IssuesResponse, error) {
	// This should not be called - data package handles GitHub directly
	panic("GitHubProvider.FetchIssues should not be called directly")
}

// GetCurrentUser is not implemented for GitHub provider
// The data package handles GitHub directly
func (g *GitHubProvider) GetCurrentUser() (string, error) {
	// This should not be called - data package handles GitHub directly
	panic("GitHubProvider.GetCurrentUser should not be called directly")
}

// FetchIssueComments is not implemented for GitHub provider
// GitHub fetches comments as part of the GraphQL query
func (g *GitHubProvider) FetchIssueComments(issueUrl string) ([]IssueComment, error) {
	// This should not be called - GitHub fetches comments in the main query
	panic("GitHubProvider.FetchIssueComments should not be called directly")
}
