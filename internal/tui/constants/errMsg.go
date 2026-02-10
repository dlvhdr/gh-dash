package constants

type ErrMsg struct {
	Err error
}

func (e ErrMsg) Error() string { return e.Err.Error() }

// ExecFinishedMsg is sent when an external process (tea.ExecProcess) finishes.
type ExecFinishedMsg struct{}
