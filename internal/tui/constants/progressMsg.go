package constants

import tea "charm.land/bubbletea/v2"

type TaskFinishedMsg struct {
	TaskId      string
	SectionId   int
	SectionType string
	Err         error
	Msg         tea.Msg
}

type ClearTaskMsg struct {
	TaskId string
}
