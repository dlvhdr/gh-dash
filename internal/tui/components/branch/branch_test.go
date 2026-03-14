package branch

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
)

// newTestBranch creates a minimal Branch with a PR having the given state and
// auto-merge configuration.
func newTestBranch(prState string, autoMergeRequest *data.AutoMergeRequest, autoMergeEnabled bool) Branch {
	return Branch{
		PR: &data.PullRequestData{
			State:            prState,
			AutoMergeRequest: autoMergeRequest,
			AutoMergeEnabled: autoMergeEnabled,
		},
	}
}

// TestRenderState_AutoMerge_ViaAutoMergeRequest verifies that RenderState
// returns the auto-merge icon when AutoMergeRequest is non-nil.
func TestRenderState_AutoMerge_ViaAutoMergeRequest(t *testing.T) {
	b := newTestBranch("OPEN", &data.AutoMergeRequest{}, false)
	state := b.RenderState()
	assert.True(t,
		strings.Contains(state, constants.AutoMergeIcon),
		"RenderState should include AutoMergeIcon when AutoMergeRequest is set; got %q", state)
}

// TestRenderState_AutoMerge_ViaLocalFlag verifies that RenderState returns the
// auto-merge icon when the local AutoMergeEnabled flag is true.
func TestRenderState_AutoMerge_ViaLocalFlag(t *testing.T) {
	b := newTestBranch("OPEN", nil, true)
	state := b.RenderState()
	assert.True(t,
		strings.Contains(state, constants.AutoMergeIcon),
		"RenderState should include AutoMergeIcon when AutoMergeEnabled flag is true; got %q", state)
}

// TestRenderState_Open_NoDraft verifies the normal open PR state.
func TestRenderState_Open_NoDraft(t *testing.T) {
	b := newTestBranch("OPEN", nil, false)
	state := b.RenderState()
	assert.Contains(t, state, constants.OpenIcon)
	assert.NotContains(t, state, constants.AutoMergeIcon)
}

// TestRenderState_Draft verifies the draft PR state.
func TestRenderState_Draft(t *testing.T) {
	b := Branch{
		PR: &data.PullRequestData{
			State:   "OPEN",
			IsDraft: true,
		},
	}
	state := b.RenderState()
	assert.Contains(t, state, constants.DraftIcon)
	assert.Contains(t, state, "Draft")
}

// TestRenderState_Merged verifies the merged PR state.
func TestRenderState_Merged(t *testing.T) {
	b := newTestBranch("MERGED", nil, false)
	state := b.RenderState()
	assert.Contains(t, state, constants.MergedIcon)
}

// TestRenderState_Closed verifies the closed PR state.
func TestRenderState_Closed(t *testing.T) {
	b := newTestBranch("CLOSED", nil, false)
	state := b.RenderState()
	assert.Contains(t, state, constants.ClosedIcon)
}
