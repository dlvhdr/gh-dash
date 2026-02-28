package prrow

import (
	"time"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

type Data struct {
	Primary    *data.PullRequestData
	Enriched   data.EnrichedPullRequestData
	IsEnriched bool
}

func (data Data) GetTitle() string {
	return data.Primary.Title
}

func (data Data) GetRepoNameWithOwner() string {
	return data.Primary.Repository.NameWithOwner
}

func (data Data) GetNumber() int {
	return data.Primary.Number
}

func (data Data) GetUrl() string {
	return data.Primary.Url
}

func (data Data) GetUpdatedAt() time.Time {
	return data.Primary.UpdatedAt
}

func (data Data) GetCreatedAt() time.Time {
	return data.Primary.CreatedAt
}

// GetPrimaryPRData returns the underlying PullRequestData, satisfying the
// primaryPRDataProvider interface consumed by tasks.extractPullRequestData
// (see internal/tui/components/tasks/pr.go). The method is on a pointer
// receiver, so only *Data satisfies primaryPRDataProvider — a caller holding
// a plain Data value (non-pointer) will not match the interface case and
// extractPullRequestData will return nil. All prrow rows are stored and
// passed as *Data throughout the codebase, so this is the expected usage.
func (d *Data) GetPrimaryPRData() *data.PullRequestData {
	return d.Primary
}
