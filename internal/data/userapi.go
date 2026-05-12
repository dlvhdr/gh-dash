package data

import (
	"sync"

	"charm.land/log/v2"
	gh "github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
)

var (
	repoUserCache = make(map[string][]User)
	userCacheMu   sync.RWMutex
)

type User struct {
	Login string `json:"login"`
	Name  string `json:"name"`
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
func FetchRepoUsers(owner, repoName string) ([]User, error) {
	// Check cache first
	repo := owner + "/" + repoName
	if cachedUsers, ok := CachedRepoUsers(repo); ok {
		log.Debug(
			"Using cached repo users",
			"owner",
			owner,
			"repoName",
			repoName,
			"len(cachedUsers)",
			len(cachedUsers),
		)
		return cachedUsers, nil
	}

	log.Debug("Fetching repo users", "owner", owner, "repoName", repoName)

	// Initialize client if needed
	if client == nil {
		var err error
		client, err = gh.DefaultGraphQLClient()
		if err != nil {
			return nil, err
		}
	}

	// Query only publicly available mentionable users
	// This includes anyone who has interacted with the repo (issues, PRs, comments)
	var result MentionableUsersResponse
	variables := map[string]any{
		"owner": graphql.String(owner),
		"name":  graphql.String(repoName),
		"limit": graphql.Int(100),
	}

	err := client.Query("GetMentionableUsers", &result, variables)
	if err != nil {
		return nil, err
	}

	users := make([]User, 0)
	for _, user := range result.Repository.MentionableUsers.Nodes {
		if user.Login != "" {
			users = append(users, user)
		}
	}

	userCacheMu.Lock()
	defer userCacheMu.Unlock()

	repoUserCache[repo] = users
	log.Debug(
		"Successfully fetched repo users",
		"owner",
		owner,
		"repoName",
		repoName,
		"len",
		len(users),
	)
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
