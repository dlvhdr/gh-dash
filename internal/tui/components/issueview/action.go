package issueview

// IssueActionType identifies the type of action requested by a key press in the issue view.
type IssueActionType int

const (
	IssueActionNone IssueActionType = iota
	IssueActionLabel
	IssueActionAssign
	IssueActionUnassign
	IssueActionComment
	IssueActionClose
	IssueActionReopen
)

// IssueAction represents an action to be performed on an issue.
type IssueAction struct {
	Type IssueActionType
}
