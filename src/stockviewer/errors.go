package stockviewer

import (
	"errors"
	"fmt"
)

var (
	ErrStockNotFound      = errors.New("stock not found")
	ErrInvalidFilter      = errors.New("invalid filter parameters")
	ErrSyncInProgress     = errors.New("sync already in progress")
	ErrExternalAPIFailure = errors.New("external API failure")
	ErrDatabaseConnection = errors.New("database connection error")
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type StorageError struct {
	Operation string
	Err       error
}

func (e StorageError) Error() string {
	return fmt.Sprintf("storage error during %s: %v", e.Operation, e.Err)
}

func (e StorageError) Unwrap() error {
	return e.Err
}

type ExternalAPIError struct {
	Service    string
	StatusCode int
	Message    string
	Err        error
}

func (e ExternalAPIError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("external API error from %s (status %d): %s", e.Service, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("external API error from %s: %s", e.Service, e.Message)
}

func (e ExternalAPIError) Unwrap() error {
	return e.Err
}

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}
