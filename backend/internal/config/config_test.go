package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadEnvFile(t *testing.T) {
	t.Run("parses KEY=VALUE, skips comments and blank lines, strips quotes", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".env")
		content := "" +
			"# comment line\n" +
			"\n" +
			"FOO=bar\n" +
			"EMPTY=\n" +
			"  SPACED = value \n" +
			"QUOTED=\"double quoted\"\n" +
			"SINGLE='single quoted'\n" +
			"URL=postgres://user:pass@localhost:5432/db?sslmode=disable\n"
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

		values, err := loadEnvFile(path)
		require.NoError(t, err)

		assert.Equal(t, "bar", values["FOO"])
		assert.Equal(t, "", values["EMPTY"])
		assert.Equal(t, "value", values["SPACED"], "keys and values are trimmed")
		assert.Equal(t, "double quoted", values["QUOTED"])
		assert.Equal(t, "single quoted", values["SINGLE"])
		assert.Equal(t, "postgres://user:pass@localhost:5432/db?sslmode=disable", values["URL"], "value splits on first = only")
	})

	t.Run("missing file returns ErrNoEnvFile sentinel without failing", func(t *testing.T) {
		_, err := loadEnvFile(filepath.Join(t.TempDir(), ".env"))
		require.ErrorIs(t, err, errNoEnvFile)
	})

	t.Run("malformed line without = is skipped", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".env")
		require.NoError(t, os.WriteFile(path, []byte("JUST_A_KEY\nGOOD=value\n"), 0o600))

		values, err := loadEnvFile(path)
		require.NoError(t, err)
		assert.Equal(t, "value", values["GOOD"])
		_, ok := values["JUST_A_KEY"]
		assert.False(t, ok)
	})
}

func TestLoadEnvFileIntoEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	require.NoError(t, os.WriteFile(path, []byte("CONFIG_TEST_VAR=from_file\nCONFIG_TEST_PRESENT=from_file\n"), 0o600))

	t.Setenv("CONFIG_TEST_PRESENT", "from_environment")

	require.NoError(t, loadEnvFileIntoEnv(path))
	t.Cleanup(func() { os.Unsetenv("CONFIG_TEST_VAR") })

	// File fills unset variables...
	assert.Equal(t, "from_file", os.Getenv("CONFIG_TEST_VAR"))
	// ...but never overwrites variables already set in the real environment.
	assert.Equal(t, "from_environment", os.Getenv("CONFIG_TEST_PRESENT"))
}

func TestLoadReadsDotEnvFromWorkingDir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".env"),
		[]byte("DATABASE_URL=postgres://x@localhost/db\nJWT_SECRET=0123456789abcdef0123456789abcdef\nPORT=9090\n"), 0o600))
	t.Chdir(dir)

	// Clear real env so the file is the only source.
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("PORT", "")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "postgres://x@localhost/db", cfg.DatabaseURL)
	assert.Equal(t, "0123456789abcdef0123456789abcdef", cfg.JWTSecret)
	assert.Equal(t, "9090", cfg.Port)
	assert.Equal(t, 15*time.Minute, cfg.JWTAccessExpiry, "defaults still apply for unset values")
}
