package constants

// func StartTask() tea.Msg {
// 	return spinner.TickMsg{}
// }

type TaskFinishedMsg struct {
	TaskId string
	Err    error
}

type ClearTaskMsg struct {
	TaskId string
}
