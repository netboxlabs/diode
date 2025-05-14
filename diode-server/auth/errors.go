package auth

// Error auth service error
type Error struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func (e *Error) Error() string {
	return e.Message
}

// NewAuthError creates a new auth service error
func NewAuthError(message string, statusCode int) *Error {
	return &Error{
		Message:    message,
		StatusCode: statusCode,
	}
}
