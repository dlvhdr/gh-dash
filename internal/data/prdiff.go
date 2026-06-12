package data

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	gh "github.com/cli/go-gh/v2/pkg/api"
)

func FetchPullRequestDiff(prNumber int, repoName string) (string, error) {
	host, repoPath, err := splitDiffRepoName(repoName)
	if err != nil {
		return "", err
	}

	client, err := gh.NewRESTClient(gh.ClientOptions{
		Host: host,
		Headers: map[string]string{
			"Accept": "application/vnd.github.v3.diff",
		},
	})
	if err != nil {
		return "", err
	}

	resp, err := client.Request(
		http.MethodGet,
		fmt.Sprintf("repos/%s/pulls/%d", repoPath, prNumber),
		nil,
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	diff, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(diff), nil
}

func splitDiffRepoName(repoName string) (string, string, error) {
	parts := strings.Split(repoName, "/")
	switch len(parts) {
	case 2:
		return "", repoName, nil
	case 3:
		return parts[0], strings.Join(parts[1:], "/"), nil
	default:
		return "", "", fmt.Errorf("invalid repo name %q", repoName)
	}
}
