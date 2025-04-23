package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Helper to create a temporary .env file.
// Returns the full path to the created file.
func createTempDotEnv(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()                      // Create a temporary directory
	path := filepath.Join(dir, ".env_test") // Use a distinct name like .env_test
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write temp .env file at %s: %v", path, err)
	}
	return path
}

func TestLoad(t *testing.T) {
	// --- Test Defaults (but token is required) ---
	t.Run("DefaultsWithToken", func(t *testing.T) {
		t.Setenv("GHSURF_GITHUB_TOKEN", "dummy-token-for-defaults-test")

		// Call Load with an empty string to skip .env loading
		cfg, err := Load("") // <-- Pass empty string

		if err != nil {
			t.Fatalf("Load(\"\") returned unexpected error: %v", err)
		}
		if cfg == nil {
			t.Fatal("Load(\"\") returned nil config on success")
		}

		expected := &Config{
			Port:        "8080",
			LogLevel:    "INFO",
			GithubToken: "dummy-token-for-defaults-test",
		}

		if !reflect.DeepEqual(cfg, expected) {
			t.Errorf("Expected default config %+v, but got %+v", expected, cfg)
		}
	})

	// --- Test Environment Variables ---
	t.Run("EnvVars", func(t *testing.T) {
		t.Setenv("PORT", "9090")
		t.Setenv("LOG_LEVEL", "DEBUG")
		t.Setenv("GHSURF_GITHUB_TOKEN", "test-token-env")

		// Call Load with an empty string to skip .env loading
		cfg, err := Load("") // <-- Pass empty string

		if err != nil {
			t.Fatalf("Load(\"\") returned unexpected error: %v", err)
		}
		if cfg == nil {
			t.Fatal("Load(\"\") returned nil config on success")
		}

		expected := &Config{
			Port:        "9090",
			LogLevel:    "DEBUG",
			GithubToken: "test-token-env",
		}

		if !reflect.DeepEqual(cfg, expected) {
			t.Errorf("Expected config from env vars %+v, but got %+v", expected, cfg)
		}
	})

	// --- Test .env File Loading ---
	t.Run("DotEnv", func(t *testing.T) {
		dotEnvContent := `
PORT=7070
LOG_LEVEL=WARN
GHSURF_GITHUB_TOKEN=test-token-dotenv # Token is set in .env
`
		// Create the temp file and get its path
		tempEnvPath := createTempDotEnv(t, dotEnvContent)

		// Call Load with the path to the temporary file
		cfg, err := Load(tempEnvPath) // <-- Pass temp file path

		if err != nil {
			t.Fatalf("Load(%s) returned unexpected error: %v", tempEnvPath, err)
		}
		if cfg == nil {
			t.Fatalf("Load(%s) returned nil config on success", tempEnvPath)
		}

		expected := &Config{
			Port:        "7070",
			LogLevel:    "WARN",
			GithubToken: "test-token-dotenv",
		}

		if !reflect.DeepEqual(cfg, expected) {
			t.Errorf("Expected config from .env %+v, but got %+v", expected, cfg)
		}
	})

	// --- Test Env Vars Override .env ---
	t.Run("EnvOverridesDotEnv", func(t *testing.T) {
		// Set env vars that should override .env
		t.Setenv("PORT", "9090")
		t.Setenv("LOG_LEVEL", "DEBUG")
		t.Setenv("GHSURF_GITHUB_TOKEN", "test-token-env")

		dotEnvContent := `
PORT=7070
LOG_LEVEL=WARN
GHSURF_GITHUB_TOKEN=test-token-dotenv # This should be overridden
`
		// Create the temp file and get its path
		tempEnvPath := createTempDotEnv(t, dotEnvContent)

		// Call Load with the path to the temporary file
		cfg, err := Load(tempEnvPath) // <-- Pass temp file path

		if err != nil {
			t.Fatalf("Load(%s) returned unexpected error: %v", tempEnvPath, err)
		}
		if cfg == nil {
			t.Fatalf("Load(%s) returned nil config on success", tempEnvPath)
		}

		// Expect values from Env Vars where they exist
		expected := &Config{
			Port:        "9090",           // From Env
			LogLevel:    "DEBUG",          // From Env
			GithubToken: "test-token-env", // From Env
		}

		if !reflect.DeepEqual(cfg, expected) {
			t.Errorf("Expected config with env override %+v, but got %+v", expected, cfg)
		}
	})

	// --- Test Missing Token Error ---
	t.Run("Error_MissingToken", func(t *testing.T) {
		t.Setenv("GHSURF_GITHUB_TOKEN", "") // Ensure token is not set
		t.Setenv("PORT", "")
		t.Setenv("LOG_LEVEL", "")

		// Call Load with an empty string (no .env file involved)
		cfg, err := Load("") // <-- Pass empty string

		if err == nil {
			t.Fatal("Load(\"\") did not return an error when GHSURF_GITHUB_TOKEN was missing")
		}
		expectedErrMsg := "GHSURF_GITHUB_TOKEN environment variable is required"
		if !strings.Contains(err.Error(), expectedErrMsg) {
			t.Errorf("Expected error message to contain '%s', but got '%v'", expectedErrMsg, err)
		}
		if cfg != nil {
			t.Errorf("Expected config to be nil when Load() returns an error, but got %+v", cfg)
		}
	})
}
