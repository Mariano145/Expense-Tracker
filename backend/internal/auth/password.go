// Package auth provides cryptographic primitives for authentication:
// JWT token management, password hashing (bcrypt), and refresh token generation.
package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword returns a bcrypt hash of the given password with the specified cost.
func HashPassword(password string, cost int) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword compares a plaintext password with a bcrypt hash.
// Returns nil on match, bcrypt.ErrMismatchedHashAndPassword on mismatch.
func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
