package data

import (
	"fmt"
	"sync"

	"github.com/charmbracelet/log"
	gh "github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
)

var (
	repoUserCache = make(map[string][]User)
	userCacheMu   sync.RWMutex
)

type User struct {
	Login string `json:"login"`
}

type MentionableUsersResponse struct {
	Repository struct {
		MentionableUsers struct {
			Nodes []User
		} `graphql:"mentionableUsers(first: $limit)"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

func CachedRepoUsers(repoNameWithOwner string) ([]User, bool) {
	userCacheMu.RLock()
	defer userCacheMu.RUnlock()
	users, ok := repoUserCache[repoNameWithOwner]
	return users, ok
}

// FetchRepoUsers fetches users that can be mentioned in a repository.
// It uses the publicly available mentionableUsers field which includes
// anyone who can interact with the repository (issue/PR authors, commenters, etc.)
func FetchRepoUsers(repoNameWithOwner string) ([]User, error) {
	// Check cache first
	if cachedUsers, ok := CachedRepoUsers(repoNameWithOwner); ok {
		log.Debug("FetchRepoUsers: cache hit", "repo", repoNameWithOwner, "count", len(cachedUsers))
		return cachedUsers, nil
	}
	log.Debug("FetchRepoUsers: cache miss", "repo", repoNameWithOwner)

	parts := splitRepoName(repoNameWithOwner)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repo name format: %s", repoNameWithOwner)
	}
	owner, name := parts[0], parts[1]

	// Initialize client if needed
	if client == nil {
		var err error
		client, err = gh.DefaultGraphQLClient()
		if err != nil {
			return nil, err
		}
	}

	log.Debug("FetchRepoUsers: executing GraphQL query", "repo", repoNameWithOwner)

	// Query only publicly available mentionable users
	// This includes anyone who has interacted with the repo (issues, PRs, comments)
	var result MentionableUsersResponse
	variables := map[string]any{
		"owner": graphql.String(owner),
		"name":  graphql.String(name),
		"limit": graphql.Int(100),
	}

	err := client.Query("GetMentionableUsers", &result, variables)
	if err != nil {
		log.Error("FetchRepoUsers: GraphQL query failed", "repo", repoNameWithOwner, "err", err)
		return nil, err
	}

	users := result.Repository.MentionableUsers.Nodes

	userCacheMu.Lock()
	defer userCacheMu.Unlock()

	repoUserCache[repoNameWithOwner] = users
	log.Info("FetchRepoUsers: successfully fetched mentionable users",
		"repo", repoNameWithOwner,
		"count", len(users))
	return users, nil
}

func ClearUserCache() {
	userCacheMu.Lock()
	defer userCacheMu.Unlock()
	repoUserCache = make(map[string][]User)
}

func ClearRepoUserCache(repoNameWithOwner string) {
	userCacheMu.Lock()
	defer userCacheMu.Unlock()
	delete(repoUserCache, repoNameWithOwner)
}

func UserLogins(users []User) []string {
	logins := make([]string, len(users))
	for i, user := range users {
		logins[i] = user.Login
	}
	return logins
}

// splitRepoName splits "owner/repo" into ["owner", "repo"]
func splitRepoName(repoNameWithOwner string) []string {
	parts := make([]string, 0, 2)
	start := 0
	for i, c := range repoNameWithOwner {
		if c == '/' {
			if i > start {
				parts = append(parts, repoNameWithOwner[start:i])
			}
			start = i + 1
		}
	}
	if start < len(repoNameWithOwner) {
		parts = append(parts, repoNameWithOwner[start:])
	}
	return parts
}
