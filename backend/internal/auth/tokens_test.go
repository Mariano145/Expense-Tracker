package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRefreshToken(t *testing.T) {
	t.Run("returns a non-empty hex token of expected length", func(t *testing.T) {
		token, err := GenerateRefreshToken()
		require.NoError(t, err)
		assert.Len(t, token, 64, "32 random bytes → 64 hex characters")
	})

	t.Run("successive calls produce different tokens", func(t *testing.T) {
		t1, err := GenerateRefreshToken()
		require.NoError(t, err)
		t2, err := GenerateRefreshToken()
		require.NoError(t, err)
		assert.NotEqual(t, t1, t2)
	})
}

func TestHashRefreshToken(t *testing.T) {
	t.Run("returns a valid SHA-256 hex digest", func(t *testing.T) {
		hash := HashRefreshToken("some-refresh-token-value")
		assert.Len(t, hash, 64, "SHA-256 produces 32 bytes → 64 hex chars")

		// Verify it matches a known SHA-256 computation.
		expected := sha256.Sum256([]byte("some-refresh-token-value"))
		assert.Equal(t, hex.EncodeToString(expected[:]), hash)
	})

	t.Run("different inputs produce different hashes", func(t *testing.T) {
		h1 := HashRefreshToken("token-a")
		h2 := HashRefreshToken("token-b")
		assert.NotEqual(t, h1, h2)
	})
}
