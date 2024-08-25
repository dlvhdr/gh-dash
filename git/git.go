package git

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	gitm "github.com/aymanbagabas/git-module"

	"github.com/dlvhdr/gh-dash/v4/utils"
)

// Extends git.Repository
type Repo struct {
	gitm.Repository
	Origin   string
	Remotes  []string
	Branches []Branch
}

type Branch struct {
	Name          string
	LastUpdatedAt *time.Time
	LastCommitMsg *string
	CommitsAhead  int
	CommitsBehind int
	IsCheckedOut  bool
	Remotes       []string
}

func GetOriginUrl(dir string) (string, error) {
	repo, err := gitm.Open(dir)
	if err != nil {
		return "", err
	}
	remotes, err := repo.Remotes()
	if err != nil {
		return "", err
	}

	for _, remote := range remotes {
		if remote != "origin" {
			continue
		}

		urls, err := gitm.RemoteGetURL(dir, remote)
		if err != nil || len(urls) == 0 {
			return "", err
		}
		return urls[0], nil
	}

	return "", errors.New("no origin remote found")
}

func GetRepo(dir string) (*Repo, error) {
	repo, err := gitm.Open(dir)
	if err != nil {
		return nil, err
	}
	err = repo.Fetch(gitm.FetchOptions{CommandOptions: gitm.CommandOptions{Args: []string{"--all"}}})
	if err != nil {
		return nil, err
	}

	bNames, err := repo.Branches()
	if err != nil {
		return nil, err
	}

	headRev, err := repo.RevParse("HEAD")
	if err != nil {
		return nil, err
	}

	branches := make([]Branch, len(bNames))
	for i, b := range bNames {
		var updatedAt *time.Time
		var lastCommitMsg *string
		isHead := false
		commits, err := gitm.Log(dir, b, gitm.LogOptions{MaxCount: 1})
		if err == nil && len(commits) > 0 {
			updatedAt = &commits[0].Committer.When
			isHead = commits[0].ID.Equal(headRev)
			lastCommitMsg = utils.StringPtr(commits[0].Summary())
		}
		commitsAhead, err := repo.RevListCount([]string{fmt.Sprintf("origin/%s..%s", b, b)})
		if err != nil {
			commitsAhead = 0
		}
		commitsBehind, err := repo.RevListCount([]string{fmt.Sprintf("%s..origin/%s", b, b)})
		if err != nil {
			commitsBehind = 0
		}
		remotes, err := repo.RemoteGetURL(b)
		if err != nil {
			commitsBehind = 0
		}
		branches[i] = Branch{
			Name:          b,
			LastUpdatedAt: updatedAt,
			IsCheckedOut:  isHead,
			Remotes:       remotes,
			LastCommitMsg: lastCommitMsg,
			CommitsAhead:  int(commitsAhead),
			CommitsBehind: int(commitsBehind),
		}
	}
	sort.Slice(branches, func(i, j int) bool {
		if branches[j].LastUpdatedAt == nil || branches[i].LastUpdatedAt == nil {
			return false
		}
		return branches[i].LastUpdatedAt.After(*branches[j].LastUpdatedAt)
	})

	remotes, err := repo.Remotes()
	if err != nil {
		return nil, err
	}

	origin, err := gitm.RemoteGetURL(dir, "origin", gitm.RemoteGetURLOptions{All: true})
	if err != nil {
		return nil, err
	}

	return &Repo{Repository: *repo, Origin: origin[0], Remotes: remotes, Branches: branches}, nil
}

func GetRepoShortName(url string) string {
	r, _ := strings.CutPrefix(url, "https://github.com/")
	r, _ = strings.CutSuffix(r, ".git")
	return r
}
