package handler

import (
	"regexp"
	"strings"
)

// emailRegexp is a pragmatic subset of the RFC 5322 address spec:
// non-space local part, "@", domain with at least one dot, no trailing dot.
var emailRegexp = regexp.MustCompile(`^[^\s@]+@[^\s@]+(\.[^\s@.]+)+$`)

const minPasswordLength = 8

// validateRegistration returns a client-facing message when the email or
// password fail the spec rules (user-auth.md), or "" when both are valid.
// Runs in the handler layer: bad input must never reach the service
// (avoids wasting a bcrypt hash on a request that will fail anyway).
func validateRegistration(email, password string) string {
	if msg := validateEmail(email); msg != "" {
		return msg
	}
	if len(password) < minPasswordLength {
		return "password must be at least 8 characters"
	}
	return ""
}

// validateLogin checks shape only (not strength): login must not leak
// password policy details for existing accounts.
func validateLogin(email, password string) string {
	if msg := validateEmail(email); msg != "" {
		return msg
	}
	if password == "" {
		return "password is required"
	}
	return ""
}

func validateEmail(email string) string {
	if !emailRegexp.MatchString(strings.TrimSpace(email)) {
		return "invalid email format"
	}
	return ""
}
