package erro

type AlreadyInStateError struct {
	err error
}

func (e AlreadyInStateError) Error() string {
	return e.err.Error()
}

// NewAlreadyInStateError ...
func NewAlreadyInStateError(err error) AlreadyInStateError {
	return AlreadyInStateError{
		err: err,
	}
}
