package branch

import (
	"time"

	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/git"
)

type BranchData struct {
	Data git.Branch
	PR   *data.PullRequestData
}

func (b BranchData) GetRepoNameWithOwner() string {
	return b.Data.Remotes[0]
}

func (b BranchData) GetTitle() string {
	return b.Data.Name
}

func (b BranchData) GetNumber() int {
	if b.PR == nil {
		return 0
	}
	return b.PR.Number
}

func (b BranchData) GetUrl() string {
	if b.PR == nil {
		return ""
	}
	return b.PR.Url
}

func (b BranchData) GetUpdatedAt() time.Time {
	return *b.Data.LastUpdatedAt
}
