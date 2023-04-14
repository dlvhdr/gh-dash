package common

import (
	"fmt"
	"strings"
)

// support [user|org]/* matching for repositories
// and local path mapping to [partial path prefix]/*
// prioritize full repo mapping if it exists
func GetRepoLocalPath(repoName string, cfgPaths map[string]string) string {
	exactMatchPath, ok := cfgPaths[repoName]
	// prioritize full repo to path mapping in config
	if ok {
		return exactMatchPath
	}

	var repoPath string

	owner, repo, repoValid := func() (string, string, bool) {
		repoParts := strings.Split(repoName, "/")
		// return repo owner, repo, and indicate properly owner/repo format
		return repoParts[0], repoParts[len(repoParts)-1], len(repoParts) == 2
	}()

	if repoValid {
		// match config:repoPath values of {owner}/* as map key
		wildcardPath, wildcarded := cfgPaths[fmt.Sprintf("%s/*", owner)]

		if wildcarded {
			// adjust wildcard match to wildcard path - ~/somepath/* to ~/somepath/{repo}
			repoPath = fmt.Sprintf("%s/%s", strings.TrimSuffix(wildcardPath, "/*"), repo)
		}
	}

	return repoPath
}
