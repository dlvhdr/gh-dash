package reposection

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/git"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tasks"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

// newTestRepoModel creates a Model via the real NewModel constructor so that
// all internal sub-models are properly wired.  A single git branch and a
// matching PR are injected so that UpdatePRMsg can find the PR and
// updateBranchesWithPrs can populate m.Branches.
func newTestRepoModel(prNumber int) Model {
	thm := *theme.DefaultTheme
	s := context.InitStyles(thm)
	cfg := &config.Config{
		// A non-nil ThemeConfig is required because ToTableRow dereferences
		// Config.Theme.Ui.Table fields.
		Theme: &config.ThemeConfig{},
	}
	ctx := &context.ProgramContext{
		Config: cfg,
		Theme:  thm,
		Styles: s,
		StartTask: func(task context.Task) tea.Cmd {
			return func() tea.Msg { return nil }
		},
	}

	t := config.RepoView
	m := NewModel(0, ctx, config.PrsSectionConfig{Type: &t}, time.Time{})

	const branchName = "feature/test"
	// Inject a real git.Repo so updateBranchesWithPrs has something to iterate.
	m.repo = &git.Repo{
		Branches: []git.Branch{
			{Name: branchName},
		},
	}
	m.Prs = []data.PullRequestData{
		{Number: prNumber, HeadRefName: branchName},
	}
	return m
}

// TestUpdatePRMsg_AutoMergeEnabled_SetsFlag verifies that when an UpdatePRMsg
// with AutoMergeEnabled=true is processed:
//   - the PR number is recorded in the autoMergeEnabledPRs map, and
//   - the corresponding Branch in m.Branches has AutoMergeEnabled set to true.
func TestUpdatePRMsg_AutoMergeEnabled_SetsFlag(t *testing.T) {
	m := newTestRepoModel(42)

	require.False(t, m.autoMergeEnabledPRs[42], "autoMergeEnabledPRs should start empty")

	autoMerge := true
	msg := tasks.UpdatePRMsg{
		PrNumber:         42,
		AutoMergeEnabled: &autoMerge,
	}

	result, _ := m.Update(msg)
	updated := result.(*Model)

	assert.True(t, updated.autoMergeEnabledPRs[42],
		"autoMergeEnabledPRs[42] should be true after processing UpdatePRMsg")

	// The flag must also be visible on the corresponding Branch so that
	// renderState / RenderState can show the auto-merge icon immediately.
	require.Len(t, updated.Branches, 1, "expected one branch after update")
	assert.True(t, updated.Branches[0].AutoMergeEnabled,
		"Branch.AutoMergeEnabled should be true after processing UpdatePRMsg")

	// AutoMergeRequest must not be touched — only real API data should set it.
	assert.Nil(t, updated.Branches[0].PR.AutoMergeRequest,
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
// PR number does not mutate any existing PR or branch.
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

	assert.False(t, updated.autoMergeEnabledPRs[1],
		"unrelated PR should not have AutoMergeEnabled set in the map")
	assert.Equal(t, "OPEN", updated.Prs[0].State,
		"unrelated PR state should be unchanged")
	require.Len(t, updated.Branches, 1)
	assert.False(t, updated.Branches[0].AutoMergeEnabled,
		"unrelated branch should not have AutoMergeEnabled set")
}
