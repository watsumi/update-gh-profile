package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/watsumi/update-gh-profile/internal/logger"
	"github.com/watsumi/update-gh-profile/internal/repository"

	"github.com/google/go-github/v76/github"
)

// AggregateGraphQLData aggregates data fetched from GraphQL
func AggregateGraphQLData(ctx context.Context, token string, username, userID string, excludeForks bool) (
	map[string]int, // languageTotals
	map[string]map[string]int, // commitHistories
	map[string]map[int]int, // timeDistributions
	map[string]map[string]int, // allCommitLanguages
	int, // totalCommits
	int, // totalPRs
	[]*github.Repository, // repos (for summary statistics)
	error,
) {
	logger.Info("Fetching repository information in bulk")

	// 1. Fetch repository information via GraphQL (using generated types)
	repoGraphQLData, err := repository.FetchRepositoriesWithGraphQLGenerated(ctx, token, username, excludeForks)
	if err != nil {
		logger.LogError(err, "Failed to fetch repository information via GraphQL")
		return nil, nil, nil, nil, 0, 0, nil, fmt.Errorf("failed to fetch repository information via GraphQL: %w", err)
	}

	logger.Info("Fetched %d repository information items", len(repoGraphQLData))

	// 2. Fetch user details (commit count, PR count, etc.) (using generated types)
	userDetails, err := repository.FetchUserDetailsWithGraphQLGenerated(ctx, token, username)
	if err != nil {
		// Treat temporary errors like 502 Bad Gateway as warnings (not fatal)
		logger.Warning("Failed to fetch user details via GraphQL: %v (continuing)", err)
		userDetails = nil // Explicitly set to nil
	}

	// 3. Fetch commit time distribution (past 1 year)
	since := time.Now().AddDate(-1, 0, 0)
	until := time.Now()
	timeDistribution, err := repository.FetchProductiveTimeWithGraphQL(ctx, token, username, userID, since, until)
	if err != nil {
		logger.LogError(err, "Failed to fetch commit time distribution via GraphQL")
		timeDistribution = make(map[int]int) // Continue with empty map
	}

	// 4. Fetch languages per commit
	commitLanguages, err := repository.FetchCommitLanguagesWithGraphQL(ctx, token, username)
	if err != nil {
		logger.LogError(err, "Failed to fetch commit language information via GraphQL")
		commitLanguages = make(map[string]map[string]int) // Continue with empty map
	}

	// 5. Aggregate data
	languageTotals := make(map[string]int)
	commitHistories := make(map[string]map[string]int)

	// Aggregate language data per repository
	for _, repo := range repoGraphQLData {
		repoKey := fmt.Sprintf("%s/%s", repo.Owner.Login, repo.Name)

		// Aggregate language data
		for _, lang := range repo.Languages.Nodes {
			languageTotals[lang.Name] += lang.Size
		}

		// Aggregate commit history (by date)
		if repo.DefaultBranchRef.Target.History.Nodes != nil {
			history := make(map[string]int)
			for _, commit := range repo.DefaultBranchRef.Target.History.Nodes {
				// Get date (YYYY-MM-DD format)
				date := ""
				if commit.CommittedDate != "" {
					t, err := time.Parse(time.RFC3339, commit.CommittedDate)
					if err == nil {
						date = t.UTC().Format("2006-01-02")
						history[date]++
					}
				} else if commit.Author.Date != "" {
					t, err := time.Parse(time.RFC3339, commit.Author.Date)
					if err == nil {
						date = t.UTC().Format("2006-01-02")
						history[date]++
					}
				}
			}
			if len(history) > 0 {
				commitHistories[repoKey] = history
			}
		}
	}

	// Aggregate commit time distribution per repository
	timeDistributions := make(map[string]map[int]int)
	// Use time distribution fetched from GraphQL as-is
	// (aggregated result from all repositories, so treat as single entry)
	if len(timeDistribution) > 0 {
		timeDistributions["all"] = timeDistribution
	}

	// Aggregate language data per commit
	allCommitLanguages := commitLanguages

	// Calculate total commits and total PRs
	var totalCommits, totalPRs, totalStars int
	if userDetails != nil {
		// Get PR count from user details
		totalPRs = userDetails.PullRequests.TotalCount
		// Sum commit count and star count per repository
		for _, repo := range repoGraphQLData {
			if repo.DefaultBranchRef.Target.History.TotalCount > 0 {
				totalCommits += repo.DefaultBranchRef.Target.History.TotalCount
			}
			totalStars += repo.StargazerCount
		}
	} else {
		// Fallback: aggregate from repository data
		for _, repo := range repoGraphQLData {
			if repo.DefaultBranchRef.Target.History.TotalCount > 0 {
				totalCommits += repo.DefaultBranchRef.Target.History.TotalCount
			}
			totalStars += repo.StargazerCount
		}
	}

	logger.Info("Data aggregation completed: languages=%d, commit histories=%d, time distributions=%d",
		len(languageTotals), len(commitHistories), len(timeDistributions))

	// Create github.Repository structs from GraphQL data (for compatibility with existing code)
	var repos []*github.Repository
	for _, repoData := range repoGraphQLData {
		repo := &github.Repository{
			Name:            github.String(repoData.Name),
			StargazersCount: github.Int(repoData.StargazerCount),
			Owner: &github.User{
				Login: github.String(repoData.Owner.Login),
			},
		}
		repos = append(repos, repo)
	}

	return languageTotals, commitHistories, timeDistributions, allCommitLanguages, totalCommits, totalPRs, repos, nil
}
