package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/watsumi/update-gh-profile/internal/config"
	"github.com/watsumi/update-gh-profile/internal/logger"
	"github.com/watsumi/update-gh-profile/internal/workflow"
)

func main() {
	// Parse command line arguments
	var (
		excludeForksStr     = flag.String("exclude-forks", "true", "Whether to exclude forked repositories (true/false)")
		excludeLanguagesStr = flag.String("exclude-languages", "", "Language names to exclude from ranking (comma-separated, e.g., JSON,Markdown,Text)")
	)
	flag.Parse()

	fmt.Println("update-gh-profile: GitHub profile auto-update tool")
	fmt.Println("Initialization complete")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error: failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Printf("Error: failed to validate configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ GitHub Token is set")

	// Create context
	ctx := context.Background()

	// Authenticated user will be automatically fetched via GraphQL

	// Configure fork exclusion
	excludeForks, err := strconv.ParseBool(*excludeForksStr)
	if err != nil {
		fmt.Printf("Warning: invalid exclude-forks value (%s). Using default value true\n", *excludeForksStr)
		excludeForks = true
	}

	// Configure excluded languages (from environment variable or command line argument)
	var excludedLanguages []string
	if excludeLanguagesEnv := os.Getenv("EXCLUDE_LANGUAGES"); excludeLanguagesEnv != "" {
		// Load from environment variable
		excludedLanguages = parseLanguageList(excludeLanguagesEnv)
	} else if *excludeLanguagesStr != "" {
		// Load from command line argument
		excludedLanguages = parseLanguageList(*excludeLanguagesStr)
	}

	fmt.Println("\n✅ GitHub API client initialization successful!")

	// Configure log level (load from environment variable)
	logLevelStr := os.Getenv("LOG_LEVEL")
	if logLevelStr == "" {
		logLevelStr = "INFO"
	}
	logLevel := logger.ParseLogLevel(logLevelStr)

	// Workflow configuration
	// Set RepoPath to empty string to automatically use GITHUB_WORKSPACE in GitHub Actions environment
	workflowConfig := workflow.Config{
		RepoPath:          "",                                     // Empty string = automatically use GITHUB_WORKSPACE in GitHub Actions environment
		SVGOutputDir:      ".",                                    // Output directory for SVG files
		Timezone:          "UTC",                                  // Timezone
		CommitMessage:     "chore: update GitHub profile metrics", // Git commit message
		MaxRepositories:   0,                                      // 0 = all repositories
		ExcludeForks:      excludeForks,
		ExcludedLanguages: excludedLanguages, // List of languages to exclude
		LogLevel:          logLevel,          // Log level
	}

	// Execute workflow
	fmt.Println("\n🚀 Starting main workflow...")
	err = workflow.Run(ctx, cfg.GitHubToken, workflowConfig)
	if err != nil {
		fmt.Printf("Error: failed to execute workflow: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ All processing completed!")
	os.Exit(0)
}

// parseLanguageList converts a comma-separated language name string to a slice
func parseLanguageList(languagesStr string) []string {
	if languagesStr == "" {
		return []string{}
	}

	// Split by comma
	parts := strings.Split(languagesStr, ",")
	languages := make([]string, 0, len(parts))

	for _, part := range parts {
		// Trim whitespace
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			languages = append(languages, trimmed)
		}
	}

	return languages
}
