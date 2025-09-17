package constants

import tea "github.com/charmbracelet/bubbletea"

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
