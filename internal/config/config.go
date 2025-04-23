package config

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	LogLevel    string
	GithubToken string
}

// Load loads configuration from a specific .env file (if provided) and environment variables.
// Pass an empty string for envFilename to skip loading a .env file.
func Load(envFilename string) (*Config, error) { // <-- Added envFilename parameter
	// --- Load .env file if filename is provided ---
	if envFilename != "" {
		err := godotenv.Load(envFilename) // <-- Use the parameter
		if err != nil && !os.IsNotExist(err) {
			// Log a warning if the specified file exists but couldn't be loaded.
			// Don't warn if the file simply doesn't exist (os.IsNotExist).
			log.Printf("Warning: Error loading specified .env file '%s': %v", envFilename, err)
		} else if err == nil {
			log.Printf("Loaded configuration from .env file: %s", envFilename)
		}
	} else {
		log.Println("No .env filename provided, relying solely on environment variables.")
	}

	port := getEnv("PORT", "8080")
	logLevel := getEnv("LOG_LEVEL", "INFO")
	githubToken := getEnv("GHSURF_GITHUB_TOKEN", "")

	if githubToken == "" {
		return nil, errors.New("FATAL: GHSURF_GITHUB_TOKEN environment variable is required") // Corrected error message
	}

	return &Config{
		Port:        port,
		LogLevel:    logLevel,
		GithubToken: githubToken,
	}, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
