package client

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

// ErrUnexpectedStatusCode is an error type that is returned when a status code does not match an expected one.
type ErrUnexpectedStatusCode struct {
	expected int
	actual   int
}

func (e *ErrUnexpectedStatusCode) StatusCode() int {
	return e.actual
}
func (e *ErrUnexpectedStatusCode) Error() string {
	return fmt.Sprintf("invalid status code: %d, expected: %d", e.actual, e.expected)
}

func NewErrUnexpectedStatusCode(expected, actual int) *ErrUnexpectedStatusCode {
	return &ErrUnexpectedStatusCode{expected: expected, actual: actual}
}

// ErrInvalidField is an error type that is returned when a field is invalid.
type ErrInvalidField struct {
	field  string
	reason string
}

func (e *ErrInvalidField) Error() string {
	return fmt.Sprintf("%s is invalid: %s", e.field, e.reason)
}

func NewErrInvalidField(field, reason string) *ErrInvalidField {
	return &ErrInvalidField{field: field, reason: reason}
}

// ErrMissingField is an error type that is returned when a required field is missing.
type ErrMissingField struct {
	field string
}

func (e *ErrMissingField) Error() string {
	return fmt.Sprintf("%s is required", e.field)
}

func NewErrMissingField(field string) *ErrMissingField {
	return &ErrMissingField{field: field}
}
