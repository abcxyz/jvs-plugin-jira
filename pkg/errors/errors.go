package errors

// Error is a concrete error implementation.
type Error string

// Error satisfies the error interface.
func (e Error) Error() string {
	return string(e)
}

const (
	ErrInvalidJustification = Error("invalid justification")
	ErrInternal             = Error("internal error, unable to perform jira validation")
)
