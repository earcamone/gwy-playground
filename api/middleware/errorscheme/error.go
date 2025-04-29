package errorscheme

type AppError struct {
	Code     int
	Message  string
	Internal error
}

// Error makes AppError compliant with native Go errors, which, to be honest,
// is more of a good practice than actually something useful in this case :P
func (e *AppError) Error() string {
	return e.Message
}

// NewAppError creates a new request error
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:     code,
		Message:  message,
		Internal: err,
	}
}
