// Package common provides shared domain types, errors, and value objects
// used across all bounded contexts in the Payment Control Plane.
package common

import "errors"

// Sentinel errors for domain-level error classification.
// All domain-specific errors should wrap these to enable
// consistent error handling at the interface layer.
var (
	// ErrNotFound indicates the requested resource does not exist.
	ErrNotFound = errors.New("resource not found")

	// ErrConflict indicates a uniqueness or state conflict.
	ErrConflict = errors.New("resource already exists")

	// ErrInvalidInput indicates the provided input fails validation.
	ErrInvalidInput = errors.New("invalid input")

	// ErrInternal indicates an unexpected internal error.
	ErrInternal = errors.New("internal error")

	// ErrUnauthorized indicates missing or invalid authentication.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates insufficient permissions.
	ErrForbidden = errors.New("forbidden")
)
