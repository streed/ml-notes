package errors

import "errors"

// Common errors used throughout the application
var (
	// Database errors
	ErrNoteNotFound  = errors.New("note not found")
	ErrDatabaseQuery = errors.New("database query failed")

	// Validation errors
	ErrEmptyContent      = errors.New("content cannot be empty")
	ErrInvalidDimensions = errors.New("invalid vector dimensions")
	ErrInvalidBoolean    = errors.New("invalid boolean value (use true/false)")
	ErrUnknownConfigKey  = errors.New("unknown configuration key")
	ErrInvalidNoteID     = errors.New("invalid note ID")

	// Embedding errors
	ErrInvalidEmbeddingLength = errors.New("invalid embedding data length")
	ErrDimensionMismatch      = errors.New("embedding dimension mismatch")
)
