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

type SponsorsResponse struct {
	User struct {
		Sponsors struct {
			Nodes []struct {
				Typename string `graphql:"__typename"`
				User     struct {
					Login string
					Url   string
				} `graphql:"... on User"`
				Organization struct {
					Name string
					Url  string
				} `graphql:"... on Organization"`
			}
		} `graphql:"sponsors(first: 100)"`
	} `graphql:"user(login: $login)"`
}

func FetchSponsors() (SponsorsResponse, error) {
	var queryResult SponsorsResponse
	var err error
	if client == nil {
		client, err = gh.DefaultGraphQLClient()
	}
	if err != nil {
		return SponsorsResponse{}, err
	}

	variables := map[string]any{
		"login": graphql.String("dlvhdr"),
	}

	log.Debug("Fetching sponsors")
	err = client.Query("Sponsors", &queryResult, variables)
	if err != nil {
		return SponsorsResponse{}, err
	}
	log.Debug("Successfully fetched sponsors")

	return queryResult, nil
}
