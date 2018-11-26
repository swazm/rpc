package rpc

import (
	"fmt"
	"net/http"
)

// Error represents an error that occurred while handling a request.
type Error struct {
	Code    int
	Info    string
	Message string
}

// NewError creates a new Error instance.
func NewError(code int, message string) *Error {
	he := &Error{Code: code, Message: http.StatusText(code), Info: message}
	return he
}

// Error makes it compatible with `error` interface.
func (he *Error) Error() string {
	return fmt.Sprintf("code=%d, message=%v", he.Code, he.Message)
}
