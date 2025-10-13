package data

import (
	"github.com/charmbracelet/log"
	gh "github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
)

type VersionResponse struct {
	Repository struct {
		LatestRelease struct {
			TagName string
		}
	} `graphql:"repository(owner: $owner, name: $name)"`
}

func FetchLatestVersion() (VersionResponse, error) {
	var queryResult VersionResponse
	var err error
	if client == nil {
		client, err = gh.DefaultGraphQLClient()
	}
	if err != nil {
		return VersionResponse{}, err
	}

	variables := map[string]any{
		"owner": graphql.String("dlvhdr"),
		"name":  graphql.String("gh-dash"),
	}

	log.Debug("Fetching latest version")
	err = client.Query("LatestVersion", &queryResult, variables)
	if err != nil {
		return VersionResponse{}, err
	}
	log.Debug("Successfully fetched latest version", "version",
		queryResult.Repository.LatestRelease.TagName)

	return queryResult, nil
}
