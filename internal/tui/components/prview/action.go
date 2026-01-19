package prview

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/keys"
)

// PRActionType identifies the type of action requested by a key press in the PR view.
type PRActionType int

const (
	PRActionNone PRActionType = iota
	PRActionApprove
	PRActionAssign
	PRActionUnassign
	PRActionComment
	PRActionDiff
	PRActionCheckout
	PRActionClose
	PRActionReady
	PRActionReopen
	PRActionMerge
	PRActionUpdate
	PRActionSummaryViewMore
)

// PRAction represents an action to be performed on a PR.
type PRAction struct {
	Type PRActionType
}

// MsgToAction converts a tea.Msg to a PRAction if it matches a known key binding.
func MsgToAction(msg tea.Msg) *PRAction {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	switch {
	case key.Matches(keyMsg, keys.PRKeys.Approve):
		return &PRAction{Type: PRActionApprove}
	case key.Matches(keyMsg, keys.PRKeys.Assign):
		return &PRAction{Type: PRActionAssign}
	case key.Matches(keyMsg, keys.PRKeys.Unassign):
		return &PRAction{Type: PRActionUnassign}
	case key.Matches(keyMsg, keys.PRKeys.Comment):
		return &PRAction{Type: PRActionComment}
	case key.Matches(keyMsg, keys.PRKeys.Diff):
		return &PRAction{Type: PRActionDiff}
	case key.Matches(keyMsg, keys.PRKeys.Checkout):
		return &PRAction{Type: PRActionCheckout}
	case key.Matches(keyMsg, keys.PRKeys.Close):
		return &PRAction{Type: PRActionClose}
	case key.Matches(keyMsg, keys.PRKeys.Ready):
		return &PRAction{Type: PRActionReady}
	case key.Matches(keyMsg, keys.PRKeys.Reopen):
		return &PRAction{Type: PRActionReopen}
	case key.Matches(keyMsg, keys.PRKeys.Merge):
		return &PRAction{Type: PRActionMerge}
	case key.Matches(keyMsg, keys.PRKeys.Update):
		return &PRAction{Type: PRActionUpdate}
	case key.Matches(keyMsg, keys.PRKeys.SummaryViewMore):
		return &PRAction{Type: PRActionSummaryViewMore}
	}

	return nil
}
