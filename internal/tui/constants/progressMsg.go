package constants

import tea "charm.land/bubbletea/v2"

type TaskFinishedMsg struct {
	TaskId      string
	SectionId   int
	SectionType string
	Err         error
	Msg         tea.Msg
	// FinishedText overrides the FinishedText set on the Task at creation time
	// when non-empty. Only applied on successful completion (Err == nil).
	// Use this when the task's outcome determines the displayed message
	// (e.g. "merged" vs "auto-merge enabled" vs "queued") and the final text
	// cannot be known until the task completes.
	FinishedText string
}

type ClearTaskMsg struct {
	TaskId string
}
