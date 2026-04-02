package model

import (
	"errors"
	"fmt"
)

// Common domain errors used across the application

var (
	// NotFoundError is returned when a requested entity does not exist
	ErrNotFound = errors.New("entity not found")

	// ForbiddenError is returned when the user lacks permission
	ErrForbidden = errors.New("forbidden")

	// UnauthorizedError is returned when authentication is required
	ErrUnauthorized = errors.New("unauthorized")

	// BadRequestError is returned for invalid input
	ErrBadRequest = errors.New("bad request")

	// InternalError is returned for unexpected server errors
	ErrInternal = errors.New("internal server error")
)

// NotFoundError represents a resource that was not found
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

func (e *NotFoundError) Is(target error) bool {
	return target == ErrNotFound
}

// NewNotFoundError creates a new NotFoundError
func NewNotFoundError(resource, id string) *NotFoundError {
	return &NotFoundError{Resource: resource, ID: id}
}

// ForbiddenError represents an action that is not allowed
type ForbiddenError struct {
	Action string
	Reason string
}

func (e *ForbiddenError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("forbidden: %s - %s", e.Action, e.Reason)
	}
	return fmt.Sprintf("forbidden: %s", e.Action)
}

func (e *ForbiddenError) Is(target error) bool {
	return target == ErrForbidden
}

// NewForbiddenError creates a new ForbiddenError
func NewForbiddenError(action, reason string) *ForbiddenError {
	return &ForbiddenError{Action: action, Reason: reason}
}

// BadRequestError represents invalid input
type BadRequestError struct {
	Field   string
	Message string
}

func (e *BadRequestError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("bad request: %s - %s", e.Field, e.Message)
	}
	return fmt.Sprintf("bad request: %s", e.Message)
}

func (e *BadRequestError) Is(target error) bool {
	return target == ErrBadRequest
}

// NewBadRequestError creates a new BadRequestError
func NewBadRequestError(field, message string) *BadRequestError {
	return &BadRequestError{Field: field, Message: message}
}

// UnauthorizedError represents an authentication failure
type UnauthorizedError struct {
	Reason string
}

func (e *UnauthorizedError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("unauthorized: %s", e.Reason)
	}
	return "unauthorized"
}

func (e *UnauthorizedError) Is(target error) bool {
	return target == ErrUnauthorized
}

// NewUnauthorizedError creates a new UnauthorizedError
func NewUnauthorizedError(reason string) *UnauthorizedError {
	return &UnauthorizedError{Reason: reason}
}

// InternalError represents an unexpected server error
type InternalError struct {
	Operation string
	Err       error
}

func (e *InternalError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("internal error during %s: %v", e.Operation, e.Err)
	}
	return fmt.Sprintf("internal error during %s", e.Operation)
}

func (e *InternalError) Unwrap() error {
	return e.Err
}

func (e *InternalError) Is(target error) bool {
	return target == ErrInternal
}

// NewInternalError creates a new InternalError
func NewInternalError(operation string, err error) *InternalError {
	return &InternalError{Operation: operation, Err: err}
}

// IsNotFound checks if an error is a NotFoundError
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsForbidden checks if an error is a ForbiddenError
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsUnauthorized checks if an error is an UnauthorizedError
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsBadRequest checks if an error is a BadRequestError
func IsBadRequest(err error) bool {
	return errors.Is(err, ErrBadRequest)
}

// IsInternal checks if an error is an InternalError
func IsInternal(err error) bool {
	return errors.Is(err, ErrInternal)
}
