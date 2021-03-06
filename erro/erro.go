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

type NotOpenError struct {
	err error
}

func (e NotOpenError) Error() string {
	return e.err.Error()
}

// NewNotOpenError ...
func NewNotOpenError(err error) NotOpenError {
	return NotOpenError{
		err: err,
	}
}

type NoReviewerError struct {
	err error
}

func (e NoReviewerError) Error() string {
	return e.err.Error()
}

// NewNoReviewerError ...
func NewNoReviewerError(err error) NoReviewerError {
	return NoReviewerError{
		err: err,
	}
}
