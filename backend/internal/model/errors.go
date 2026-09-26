package model

import "errors"

// Sentinel errors for the authentication domain.
// The service layer returns these so the handler layer can map them to HTTP status codes.
var (
	// ErrEmailExists is returned when a registration attempt uses an email that is already registered.
	ErrEmailExists = errors.New("email already registered")

	// ErrUserNotFound is returned when a user cannot be found by ID or email.
	ErrUserNotFound = errors.New("user not found")

	// ErrInvalidCredentials is returned when login fails due to wrong email or password.
	// The message is intentionally vague to avoid revealing which field is incorrect.
	ErrInvalidCredentials = errors.New("invalid email or password")

	// ErrTokenExpired is returned when a refresh token has passed its expiry time.
	ErrTokenExpired = errors.New("token has expired")

	// ErrTokenRevoked is returned when a refresh token has been explicitly revoked (e.g. logout).
	ErrTokenRevoked = errors.New("token has been revoked")

	// ErrTokenNotFound is returned when a refresh token hash does not match any record.
	ErrTokenNotFound = errors.New("token not found")

	// ErrTokenReuse is returned when a previously rotated refresh token is presented again.
	// This indicates potential token theft — the entire token family is revoked.
	ErrTokenReuse = errors.New("token reuse detected — family revoked")
)
