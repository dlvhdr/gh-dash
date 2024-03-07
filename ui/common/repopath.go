package common

import (
	"fmt"
	"strings"
)

// GetRepoLocalPath returns the local path for a given repo name.
// It will return the path if it exists in the config, or if the
// repo name matches a wildcard path in the config.
// It will prioritize exact repo name matches over wildcard matches.
// If the second return value is true, the first return value is guaranteed
// to be a valid path.
// If the second return value is false, the first return value is undefined.
// For a given config of:
//
//	{
//	  "user/repo": "/path/to/user/repo",
//	  "user_2/*":  "/path/to/user_2/*",
//	}
//
// GetRepoLocalPath("user/repo", config) will return: "/path/to/user/repo", true
// GetRepoLocalPath("user_2/some_repo", config) will return: "/path/to/user_2/some_repo", true
// GetRepoLocalPath("user/other_repo", config) will return: "", false
func GetRepoLocalPath(repoName string, cfgPaths map[string]string) (string, bool) {
	exactMatchPath, ok := cfgPaths[repoName]
	// prioritize full repo to path mapping in config
	if ok {
		return exactMatchPath, true
	}

	owner, repo, repoValid := func() (string, string, bool) {
		repoParts := strings.Split(repoName, "/")
		// return repo owner, repo, and indicate properly owner/repo format
		return repoParts[0], repoParts[len(repoParts)-1], len(repoParts) == 2
	}()

	if !repoValid {
		return "", false
	}

	// match config:repoPath values of {owner}/* as map key
	wildcardPath, wildcardFound := cfgPaths[fmt.Sprintf("%s/*", owner)]

	if wildcardFound {
		// adjust wildcard match to wildcard path - ~/somepath/* to ~/somepath/{repo}
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(wildcardPath, "/*"), repo), true
	}

	if template, ok := cfgPaths[":owner/:repo"]; ok {
		return strings.ReplaceAll(strings.ReplaceAll(template, ":owner", owner), ":repo", repo), true
	}

	return "", false
}
