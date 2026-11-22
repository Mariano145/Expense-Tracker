package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAccessToken(t *testing.T) {
	secret := "test-secret-that-is-at-least-32-chars"

	t.Run("generates a parseable token with correct claims", func(t *testing.T) {
		token, err := GenerateAccessToken("user-123", "test@example.com", secret, 15*time.Minute)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Parse it back and verify claims.
		parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		require.NoError(t, err)
		require.True(t, parsed.Valid)

		claims, ok := parsed.Claims.(jwt.MapClaims)
		require.True(t, ok)
		assert.Equal(t, "user-123", claims["sub"])
		assert.Equal(t, "test@example.com", claims["email"])
	})
}

func TestParseAccessToken(t *testing.T) {
	secret := "test-secret-that-is-at-least-32-chars"

	t.Run("valid token returns userID and email", func(t *testing.T) {
		token, err := GenerateAccessToken("user-456", "alice@example.com", secret, 15*time.Minute)
		require.NoError(t, err)

		userID, email, err := ParseAccessToken(token, secret)
		require.NoError(t, err)
		assert.Equal(t, "user-456", userID)
		assert.Equal(t, "alice@example.com", email)
	})

	t.Run("expired token returns error", func(t *testing.T) {
		// Generate a token that is already expired (negative duration).
		token, err := GenerateAccessToken("user-789", "bob@example.com", secret, -1*time.Minute)
		require.NoError(t, err)

		_, _, err = ParseAccessToken(token, secret)
		require.Error(t, err)
		assert.ErrorIs(t, err, jwt.ErrTokenExpired)
	})

	t.Run("wrong secret returns signature error", func(t *testing.T) {
		token, err := GenerateAccessToken("user-1", "x@example.com", secret, 15*time.Minute)
		require.NoError(t, err)

		_, _, err = ParseAccessToken(token, "different-secret-entirely-wrong!!")
		require.Error(t, err)
	})

	t.Run("malformed token returns error", func(t *testing.T) {
		_, _, err := ParseAccessToken("not-a-valid-jwt", secret)
		require.Error(t, err)
	})
}
