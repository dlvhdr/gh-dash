package reposection

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/git"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tasks"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

// newTestRepoModel creates a minimal Model with a single PR in m.Prs so that
// the UpdatePRMsg handler can find and update it.
func newTestRepoModel(prNumber int) Model {
	ctx := &context.ProgramContext{
		StartTask: func(task context.Task) tea.Cmd {
			return func() tea.Msg { return nil }
		},
	}
	m := Model{
		BaseModel: section.BaseModel{
			Ctx: ctx,
		},
		// repo must be non-nil; updateBranchesWithPrs ranges over repo.Branches.
		repo: &git.Repo{Branches: []git.Branch{}},
		Prs: []data.PullRequestData{
			{Number: prNumber},
		},
	}
	return m
}

// TestUpdatePRMsg_AutoMergeEnabled_SetsFlag verifies that when an UpdatePRMsg
// with AutoMergeEnabled=true is processed, the matching PR's AutoMergeEnabled
// flag is set to true on data.PullRequestData.
func TestUpdatePRMsg_AutoMergeEnabled_SetsFlag(t *testing.T) {
	m := newTestRepoModel(42)

	require.False(t, m.Prs[0].AutoMergeEnabled, "AutoMergeEnabled should start false")

	autoMerge := true
	msg := tasks.UpdatePRMsg{
		PrNumber:         42,
		AutoMergeEnabled: &autoMerge,
	}

	result, _ := m.Update(msg)
	updated := result.(*Model)

	assert.True(t, updated.Prs[0].AutoMergeEnabled,
		"AutoMergeEnabled should be set to true after processing UpdatePRMsg")
	assert.Nil(t, updated.Prs[0].AutoMergeRequest,
		"AutoMergeRequest should remain nil (only real API data should populate it)")
}

// TestUpdatePRMsg_IsMerged_SetsState verifies that when an UpdatePRMsg with
// IsMerged=true is processed, the PR state is set to "MERGED".
func TestUpdatePRMsg_IsMerged_SetsState(t *testing.T) {
	m := newTestRepoModel(7)
	m.Prs[0].State = "OPEN"

	isMerged := true
	msg := tasks.UpdatePRMsg{
		PrNumber: 7,
		IsMerged: &isMerged,
	}

	result, _ := m.Update(msg)
	updated := result.(*Model)

	assert.Equal(t, "MERGED", updated.Prs[0].State)
	assert.Empty(t, updated.Prs[0].Mergeable)
}

// TestUpdatePRMsg_UnknownPR_NoChange verifies that a message for an unknown
// PR number does not mutate any existing PR.
func TestUpdatePRMsg_UnknownPR_NoChange(t *testing.T) {
	m := newTestRepoModel(1)
	m.Prs[0].State = "OPEN"

	autoMerge := true
	msg := tasks.UpdatePRMsg{
		PrNumber:         999,
		AutoMergeEnabled: &autoMerge,
	}

	result, _ := m.Update(msg)
	updated := result.(*Model)

	assert.False(t, updated.Prs[0].AutoMergeEnabled,
		"unrelated PR should not have AutoMergeEnabled set")
	assert.Equal(t, "OPEN", updated.Prs[0].State,
		"unrelated PR state should be unchanged")
}
