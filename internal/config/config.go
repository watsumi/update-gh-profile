package config

import (
	"errors"
	"fmt"
	"log"
	"os"
)

// Config struct to hold application configuration
// In Go, structs are used to group data together
type Config struct {
	// GitHubToken authentication token for GitHub API
	// Requires permission to read all repositories
	GitHubToken string
}

// Load loads configuration from environment variables
// In Go, functions starting with capital letters can be called from external packages (public functions)
func Load() (*Config, error) {
	// &Config{} creates a pointer to a struct
	// In Go, it's common to return structs by pointer
	cfg := &Config{}

	// Load GITHUB_TOKEN environment variable
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, errors.New("GITHUB_TOKEN environment variable is not set")
	}
	cfg.GitHubToken = token

	// Log output: configuration load success (INFO level equivalent)
	log.Printf("Configuration loaded: token=set (authenticated user will be automatically fetched)")

	return cfg, nil
}

// Validate validates configuration values
func (c *Config) Validate() error {
	if c.GitHubToken == "" {
		return errors.New("GitHubToken is not set")
	}

	// Verify token is not empty (minimal validation)
	if len(c.GitHubToken) < 10 {
		return fmt.Errorf("GitHubToken is too short (length: %d)", len(c.GitHubToken))
	}

	return nil
}
