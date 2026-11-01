package domain

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrForbidden    = errors.New("forbidden")
	ErrUnauthorized = errors.New("unauthorized")
	ErrTimeout      = errors.New("timeout")
	ErrValidation   = errors.New("validation failed")
)

type FieldError struct {
	Field  string
	Reason string
}

func (e *FieldError) Error() string {
	return ErrValidation.Error() + ": " + e.Field + ": " + e.Reason
}

func (e *FieldError) Unwrap() error {
	return ErrValidation
}
