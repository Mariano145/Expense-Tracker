package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

// errNoEnvFile is returned by loadEnvFile when the file does not exist.
// It is not an error condition for Load — .env is optional in production,
// where real environment variables are used instead.
var errNoEnvFile = errors.New("no .env file present")

// loadEnvFile parses a dotenv-style file (KEY=VALUE per line).
// Rules: blank lines and lines starting with '#' are ignored; each line is
// split on the FIRST '=' only (so URLs with '=' survive); surrounding spaces
// are trimmed; one layer of matching single/double quotes is stripped.
// Lines without '=' are skipped silently.
func loadEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errNoEnvFile
		}
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue // malformed line: no assignment
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') ||
				(val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		if key == "" {
			continue
		}
		values[key] = val
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return values, nil
}

// loadEnvFileIntoEnv sets each parsed variable via os.Setenv, but only when
// the variable is not already set to a non-empty value in the process
// environment — real env vars always win over the file.
func loadEnvFileIntoEnv(path string) error {
	values, err := loadEnvFile(path)
	if err != nil {
		return err
	}
	for key, val := range values {
		if os.Getenv(key) == "" {
			if err := os.Setenv(key, val); err != nil {
				return fmt.Errorf("set %s: %w", key, err)
			}
		}
	}
	return nil
}
