package model

import "time"

// RefreshToken represents a refresh token stored in the database.
// The TokenHash field holds a SHA-256 hash of the original token — plain text
// tokens are never persisted. FamilyID groups rotated tokens together so that
// reuse of an already-rotated token can trigger family-wide revocation.
type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	FamilyID  string
	ExpiresAt time.Time
	Revoked   bool
	CreatedAt time.Time
}
