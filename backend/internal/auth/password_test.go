package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	t.Run("returns a valid bcrypt hash with the configured cost", func(t *testing.T) {
		hash, err := HashPassword("secureP@ss1", 12)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)

		// Verify it is a valid bcrypt hash with cost 12.
		cost, err := bcrypt.Cost([]byte(hash))
		require.NoError(t, err)
		assert.Equal(t, 12, cost)
	})

	t.Run("different calls produce different hashes (salted)", func(t *testing.T) {
		h1, err := HashPassword("samePassword", 10)
		require.NoError(t, err)
		h2, err := HashPassword("samePassword", 10)
		require.NoError(t, err)
		assert.NotEqual(t, h1, h2, "bcrypt should produce unique hashes due to random salt")
	})
}

func TestCheckPassword(t *testing.T) {
	t.Run("correct password returns nil", func(t *testing.T) {
		hash, err := HashPassword("correctPassword", 10)
		require.NoError(t, err)

		err = CheckPassword("correctPassword", hash)
		assert.NoError(t, err)
	})

	t.Run("wrong password returns bcrypt.ErrMismatchedHashAndPassword", func(t *testing.T) {
		hash, err := HashPassword("correctPassword", 10)
		require.NoError(t, err)

		err = CheckPassword("wrongPassword", hash)
		assert.ErrorIs(t, err, bcrypt.ErrMismatchedHashAndPassword)
	})
}
