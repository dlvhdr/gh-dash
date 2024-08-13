package git

import (
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

	branches := make([]Branch, len(bNames))
	for i, b := range bNames {
		var updatedAt *time.Time
		commits, err := gitm.Log(dir, b, gitm.LogOptions{MaxCount: 1})
		if err == nil && len(commits) > 0 {
			updatedAt = &commits[0].Committer.When
		}
		branches[i] = Branch{Name: b, LastUpdatedAt: updatedAt}
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
