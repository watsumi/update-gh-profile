package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/watsumi/update-gh-profile/internal/aggregator"
	"github.com/watsumi/update-gh-profile/internal/generator"
	"github.com/watsumi/update-gh-profile/internal/git"
	"github.com/watsumi/update-gh-profile/internal/logger"
	"github.com/watsumi/update-gh-profile/internal/readme"
	"github.com/watsumi/update-gh-profile/internal/repository"

	"github.com/google/go-github/v76/github"
)

// Config workflow configuration
type Config struct {
	RepoPath          string          // Repository path (location of README.md)
	SVGOutputDir      string          // Output directory for SVG files
	Timezone          string          // Timezone (e.g., "Asia/Tokyo", "UTC")
	CommitMessage     string          // Git commit message
	MaxRepositories   int             // Maximum number of repositories to process (0 = all)
	ExcludeForks      bool            // Whether to exclude forked repositories
	ExcludedLanguages []string        // List of language names to exclude from ranking
	LogLevel          logger.LogLevel // Log level
}

// Run executes the main workflow
//
// Preconditions:
// - ctx is a valid context.Context
// - client is an initialized GitHub API client
// - config is a valid Config struct
//
// Postconditions:
// - README.md is updated
// - SVG files are generated and saved
// - Git commit and push are executed if there are changes
//
// Invariants:
// - Errors are handled appropriately when they occur
func Run(ctx context.Context, token string, config Config) error {
	// Configure logger
	if config.LogLevel != 0 {
		logger.DefaultLogger.SetLevel(config.LogLevel)
	}

	logger.Info("Starting workflow")

	// Validate token (already passed, but verify)
	if token == "" {
		logger.Error("GITHUB_TOKEN is not set")
		return fmt.Errorf("GITHUB_TOKEN is not set")
	}

	// Fetch authenticated user information via GraphQL (using generated types)
	username, userID, err := repository.FetchViewerGenerated(ctx, token)
	if err != nil {
		logger.LogError(err, "Failed to fetch authenticated user information")
		return fmt.Errorf("failed to fetch authenticated user information: %w", err)
	}
	logger.Info("Authenticated user: %s", username)

	// 1-2. Fetch and aggregate data using GraphQL
	fmt.Println("\n📊 Fetching and aggregating repository data...")
	logger.Info("Fetching data")

	languageTotals, commitHistories, timeDistributions, allCommitLanguages, totalCommits, totalPRs, repos, err := AggregateGraphQLData(
		ctx, token, username, userID, config.ExcludeForks)
	if err != nil {
		logger.LogError(err, "Failed to fetch and aggregate GraphQL data")
		return fmt.Errorf("failed to fetch and aggregate GraphQL data: %w", err)
	}

	if len(languageTotals) == 0 {
		logger.Warning("No repository data found")
		return fmt.Errorf("no repository data found")
	}

	logger.Info("Data fetch completed: languages=%d, commit histories=%d, total commits=%d, total PRs=%d",
		len(languageTotals), len(commitHistories), totalCommits, totalPRs)
	fmt.Printf("✅ Data fetched (languages: %d, commit histories: %d repositories)\n",
		len(languageTotals), len(commitHistories))

	// 3. Aggregate data and generate rankings
	fmt.Println("\n📈 Aggregating data and generating rankings...")

	// Language ranking (all languages, excluding specified ones)
	var rankedLanguages []aggregator.LanguageStat
	if len(languageTotals) > 0 {
		rankedLanguages = aggregator.RankLanguages(languageTotals)
		// Filter excluded languages (before any percentage filtering)
		// This ensures excluded languages are not included in the pie chart
		if len(config.ExcludedLanguages) > 0 {
			rankedLanguages = aggregator.FilterExcludedLanguages(rankedLanguages, config.ExcludedLanguages)
		}
		// Note: Removed FilterMinorLanguages to show all languages in pie chart
	}

	// Aggregate commit history
	logger.Info("Aggregating commit history...")
	aggregatedHistoryMap := aggregator.AggregateCommitHistory(commitHistories)
	aggregatedHistory := aggregator.SortCommitHistoryByDate(aggregatedHistoryMap)
	logger.Info("Commit history aggregation completed: %d days", len(aggregatedHistory))

	// Aggregate commit time distribution
	logger.Info("Aggregating commit time distribution...")
	aggregatedTimeDistMap := aggregator.AggregateCommitTimeDistribution(timeDistributions)
	aggregatedTimeDist := aggregator.SortCommitTimeDistributionByHour(aggregatedTimeDistMap)
	logger.Info("Commit time distribution aggregation completed: %d time slots", len(aggregatedTimeDist))

	// Top 5 languages by commit (excluding excluded languages)
	top5Languages := aggregator.AggregateCommitLanguages(allCommitLanguages, config.ExcludedLanguages)

	// Summary statistics
	var reposForSummary []*github.Repository
	if len(repos) > 0 {
		reposForSummary = repos
	}
	summaryStats := aggregator.AggregateSummaryStats(reposForSummary, totalCommits, totalPRs)

	// 4. Generate SVG charts
	fmt.Println("\n🎨 Generating SVG charts...")

	// Determine SVG output directory (same directory as README.md)
	var svgOutputDir string
	if config.SVGOutputDir == "" || config.SVGOutputDir == "." {
		// Use GITHUB_WORKSPACE in GitHub Actions environment (same path as README.md)
		if config.RepoPath == "" || config.RepoPath == "." {
			if workspace := os.Getenv("GITHUB_WORKSPACE"); workspace != "" {
				svgOutputDir = workspace
			} else {
				svgOutputDir = "."
			}
		} else {
			svgOutputDir = config.RepoPath
		}
	} else {
		svgOutputDir = config.SVGOutputDir
	}

	// Create output directory
	err = os.MkdirAll(svgOutputDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	svgs := make(map[string]string)

	// Language ranking SVG
	if len(rankedLanguages) > 0 {
		langSVG, err := generator.GenerateLanguageChart(rankedLanguages, 10)
		if err == nil {
			langPath := filepath.Join(svgOutputDir, "language_chart.svg")
			err = generator.SaveSVG(langSVG, langPath)
			if err != nil {
				logger.LogError(err, "Failed to save language ranking SVG")
			} else {
				svgs["language_chart.svg"] = langPath
				logger.Info("Generated language ranking SVG: %s", langPath)
				fmt.Printf("  ✅ Generated language ranking SVG: %s\n", langPath)
			}
		}
	}

	// Commit history SVG
	if len(aggregatedHistory) > 0 {
		// Convert DateCommitPair slice to map[string]int
		historyMap := make(map[string]int)
		for _, pair := range aggregatedHistory {
			historyMap[pair.Date] = pair.Count
		}
		historySVG, err := generator.GenerateCommitHistoryChart(historyMap)
		if err == nil {
			historyPath := filepath.Join(svgOutputDir, "commit_history_chart.svg")
			err = generator.SaveSVG(historySVG, historyPath)
			if err == nil {
				svgs["commit_history_chart.svg"] = historyPath
				fmt.Printf("  ✅ Generated commit history SVG: %s\n", historyPath)
			}
		}
	}

	// Commit time distribution SVG
	if len(aggregatedTimeDist) > 0 {
		// Convert HourCommitPair slice to map[int]int
		timeDistMap := make(map[int]int)
		for _, pair := range aggregatedTimeDist {
			timeDistMap[pair.Hour] = pair.Count
		}
		timeSVG, err := generator.GenerateCommitTimeChart(timeDistMap)
		if err == nil {
			timePath := filepath.Join(svgOutputDir, "commit_time_chart.svg")
			err = generator.SaveSVG(timeSVG, timePath)
			if err == nil {
				svgs["commit_time_chart.svg"] = timePath
				fmt.Printf("  ✅ Generated commit time distribution SVG: %s\n", timePath)
			}
		}
	}

	// Top 5 languages by commit SVG
	if len(top5Languages) > 0 {
		commitLangSVG, err := generator.GenerateCommitLanguagesChart(top5Languages)
		if err == nil {
			commitLangPath := filepath.Join(svgOutputDir, "commit_languages_chart.svg")
			err = generator.SaveSVG(commitLangSVG, commitLangPath)
			if err == nil {
				svgs["commit_languages_chart.svg"] = commitLangPath
				fmt.Printf("  ✅ Generated top 5 languages by commit SVG: %s\n", commitLangPath)
			}
		}
	}

	// Summary card SVG
	if summaryStats.RepositoryCount > 0 {
		summarySVG, err := generator.GenerateSummaryCard(summaryStats)
		if err == nil {
			summaryPath := filepath.Join(svgOutputDir, "summary_card.svg")
			err = generator.SaveSVG(summarySVG, summaryPath)
			if err == nil {
				svgs["summary_card.svg"] = summaryPath
				fmt.Printf("  ✅ Generated summary card SVG: %s\n", summaryPath)
			}
		}
	}

	// 5. Update README.md
	fmt.Println("\n📝 Updating README.md...")

	// Determine README.md path (align with RepoPath and Git operation path)
	var readmeBasePath string
	if config.RepoPath == "" || config.RepoPath == "." {
		// Use GITHUB_WORKSPACE in GitHub Actions environment
		if workspace := os.Getenv("GITHUB_WORKSPACE"); workspace != "" {
			readmeBasePath = workspace
		} else {
			readmeBasePath = "."
		}
	} else {
		readmeBasePath = config.RepoPath
	}
	readmePath := filepath.Join(readmeBasePath, "README.md")

	// Create README if it doesn't exist
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		err = os.WriteFile(readmePath, []byte("# GitHub Profile\n\n"), 0644)
		if err != nil {
			return fmt.Errorf("failed to create README.md: %w", err)
		}
		fmt.Printf("  ℹ️  Created README.md\n")
	}

	// Embed SVG charts
	svgSections := map[string]string{
		"LANGUAGE_STATS":   "language_chart.svg",
		"COMMIT_HISTORY":   "commit_history_chart.svg",
		"COMMIT_TIME":      "commit_time_chart.svg",
		"COMMIT_LANGUAGES": "commit_languages_chart.svg",
		"SUMMARY_STATS":    "summary_card.svg",
	}

	for sectionTag, svgFile := range svgSections {
		if svgPath, ok := svgs[svgFile]; ok {
			// Convert to relative path (using README.md base path)
			relPath, err := filepath.Rel(readmeBasePath, svgPath)
			if err != nil {
				relPath = svgFile
			}

			err = readme.EmbedSVGWithCustomPath(readmePath, relPath, sectionTag, "")
			if err != nil {
				logger.LogErrorWithContext(err, sectionTag, "Failed to update section")
				fmt.Printf("  ⚠️  Failed to update section %s: %v\n", sectionTag, err)
			} else {
				logger.Info("Updated section %s", sectionTag)
				fmt.Printf("  ✅ Updated section %s\n", sectionTag)
			}
		}
	}

	// 6. Git commit and push
	fmt.Println("\n🔀 Executing Git operations...")

	repoPath := config.RepoPath
	// Use GITHUB_WORKSPACE in GitHub Actions environment (when RepoPath is empty or ".")
	if repoPath == "" || repoPath == "." {
		if workspace := os.Getenv("GITHUB_WORKSPACE"); workspace != "" {
			repoPath = workspace
			logger.Info("Using GITHUB_WORKSPACE: %s", workspace)
		} else {
			if repoPath == "" {
				repoPath = "."
			}
			logger.Info("GITHUB_WORKSPACE not set, using current directory: %s", repoPath)
		}
	}

	// Convert to absolute path for logging
	absRepoPath, err := filepath.Abs(repoPath)
	if err == nil {
		logger.Info("Repository path (absolute): %s", absRepoPath)
	}

	// Check if it's a Git repository
	if !git.IsGitRepository(repoPath) {
		// Debug info: check for .git directory
		absPath, _ := filepath.Abs(repoPath)
		gitDir := filepath.Join(absPath, ".git")
		if _, err := os.Stat(gitDir); err != nil {
			logger.Warning(".git directory not found: %s (error: %v)", gitDir, err)
		} else {
			logger.Warning(".git directory exists but not recognized as Git repository: %s", gitDir)
		}
		logger.Warning("Not a Git repository, skipping commit and push (path: %s)", repoPath)
		fmt.Printf("  ℹ️  Not a Git repository, skipping commit and push (path: %s)\n", repoPath)
		return nil
	}

	// Check for changes
	hasChanges, err := git.HasChanges(repoPath)
	if err != nil {
		logger.LogError(err, "Failed to check for changes")
		return fmt.Errorf("failed to check for changes: %w", err)
	}

	if !hasChanges {
		logger.Info("No changes, skipping commit and push")
		fmt.Println("  ℹ️  No changes, skipping commit and push")
		return nil
	}

	// Commit message
	commitMsg := config.CommitMessage
	if commitMsg == "" {
		commitMsg = "chore: update GitHub profile metrics"
	}

	// Commit and push
	// In GitHub Actions environment, credentials are automatically configured, so token is not needed (pass empty string)
	logger.Info("Executing Git commit and push...")
	err = git.CommitAndPush(repoPath, commitMsg, nil, "origin", "", "")
	if err != nil {
		logger.LogError(err, "Failed to commit and push")
		return fmt.Errorf("failed to commit and push: %w", err)
	}

	logger.Info("Git commit and push completed")
	fmt.Println("  ✅ Git commit and push completed")

	logger.Info("All processing completed")
	fmt.Println("\n✅ All processing completed!")

	return nil
}
