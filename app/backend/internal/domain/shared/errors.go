package shared

import (
	"errors"
	"fmt"
)

// ValidationError means a value breaks a business rule. The API returns it as
// 422 validation_failed (docs/design/api.md).
type ValidationError struct {
	Field   string
	Message string
}

func (validationErr *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", validationErr.Field, validationErr.Message)
}

// NewValidationError returns a ValidationError for field.
func NewValidationError(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}

// StateError means an action is not allowed in the current state. The API returns it as
// 409 state_conflict.
type StateError struct {
	Message string
}

func (stateErr *StateError) Error() string {
	return stateErr.Message
}

// NewStateError returns a StateError.
func NewStateError(message string) error {
	return &StateError{Message: message}
}

// ErrNotFound means the entity does not exist. Repositories return errors wrapping it.
var ErrNotFound = errors.New("not found")

// ErrVersionConflict means the entity changed since the caller read it. The API returns it
// as 409 version_conflict.
var ErrVersionConflict = errors.New("version conflict")
