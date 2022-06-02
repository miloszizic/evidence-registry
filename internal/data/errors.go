package data

import "fmt"

const (
	ErrCodeUnknown ErrorCode = iota
	ErrCodeNotFound
	ErrCodeConflict
	ErrCodeExists
	ErrCodeInvalid
	ErrCodeInvalidCredentials
)

type ErrorCode uint
type Error struct {
	org  error
	msg  string
	code ErrorCode
}

// Error returns the error message, when orig is not nil, it returns error message with origin.
func (e *Error) Error() string {
	if e.org != nil {
		return fmt.Sprintf("%s: %s", e.msg, e.org)
	}
	return e.msg
}

// Code returns the error code.
func (e *Error) Code() ErrorCode {
	return e.code
}

// Unwrap returns the original error.
func (e *Error) Unwrap() error {
	return e.org
}

// WrapErrorf wraps the error with the given code and message.
func WrapErrorf(orig error, code ErrorCode, format string, args ...interface{}) error {
	return &Error{
		org:  orig,
		msg:  fmt.Sprintf(format, args...),
		code: code,
	}
}

// NewErrorf creates a new error with the given code and message.
func NewErrorf(code ErrorCode, format string, args ...interface{}) error {
	return &Error{
		msg:  fmt.Sprintf(format, args...),
		code: code,
	}
}
