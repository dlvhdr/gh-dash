package git

import (
	"errors"
	"sort"
	"time"

	gitm "github.com/aymanbagabas/git-module"
)

// Extends git.Repository
type Repo struct {
	Origin   string
	Remotes  []string
	Branches []Branch
}

type Branch struct {
	Name          string
	LastUpdatedAt *time.Time
	IsCheckedOut  bool
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
		isHead := false
		commits, err := gitm.Log(dir, b, gitm.LogOptions{MaxCount: 1})
		if err == nil && len(commits) > 0 {
			updatedAt = &commits[0].Committer.When
			isHead = commits[0].ID.Equal(headRev)
		}
		branches[i] = Branch{Name: b, LastUpdatedAt: updatedAt, IsCheckedOut: isHead}
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

	return &Repo{Origin: origin[0], Remotes: remotes, Branches: branches}, nil
}
