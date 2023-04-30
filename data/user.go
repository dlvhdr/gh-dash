package data

import (
	gh "github.com/cli/go-gh/v2/pkg/api"
)

func CurrentLoginName() (string, error) {
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
