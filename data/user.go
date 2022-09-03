package data

import (
	"github.com/cli/go-gh"
)

func CurrentLoginName() (string, error) {
	client, err := gh.GQLClient(nil)
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
