package errors

// ErrorCode represents a string-based error code
type ErrorCode string

const (
	// ErrCodeInternal is the error code for internal errors
	ErrCodeInternal ErrorCode = "ERR_INTERNAL"

	// ErrCodeOpsGenerateDiff is the error code for errors generating a diff
	ErrCodeOpsGenerateDiff ErrorCode = "ERR_OPS_GENERATE_DIFF"

	// ErrCodeOpsApplyChangeSet is the error code for errors applying a change set
	ErrCodeOpsApplyChangeSet ErrorCode = "ERR_OPS_APPLY_CHANGE_SET"
)

// String returns the string representation of the error code
func (e ErrorCode) String() string {
	return string(e)
}
