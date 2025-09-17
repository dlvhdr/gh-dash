package constants

type ErrMsg struct {
	Err error
}

func (e ErrMsg) Error() string { return e.Err.Error() }
