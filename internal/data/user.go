package data

import (
	gh "github.com/cli/go-gh/v2/pkg/api"

	"github.com/dlvhdr/gh-dash/v4/internal/provider"
)

func CurrentLoginName() (string, error) {
	// Use GitLab provider if configured
	if provider.IsGitLab() {
		return provider.GetProvider().GetCurrentUser()
	}

	client, err := gh.DefaultGraphQLClient()
	if err != nil {
		return "", nil
	}

	var query struct {
		Viewer struct {
			Login string
		}
	}
	err = client.Query("UserCurrent", &query, nil)
	return query.Viewer.Login, err
}
