package prview

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
