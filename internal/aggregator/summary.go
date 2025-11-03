package aggregator

import (
	"log"

	"github.com/google/go-github/v76/github"
)

// AggregateSummaryStats aggregates summary statistics
//
// Preconditions:
// - repositories is a slice of repository structs
// - totalCommits is the total number of commits across all repositories
// - totalPRs is the total number of pull requests across all repositories
//
// Postconditions:
// - Returns a struct containing total stars, repository count, total commits, and total PRs
//
// Invariants:
// - Sums values from all repositories
// - Fork repositories are excluded (assumes already excluded in repositories)
func AggregateSummaryStats(repositories []*github.Repository, totalCommits, totalPRs int) SummaryStats {
	log.Printf("Starting summary statistics aggregation: %d repositories", len(repositories))

	var stats SummaryStats

	// Count repositories (forks are already excluded in the assumption, but double-check)
	stats.RepositoryCount = 0
	totalStars := 0

	for _, repo := range repositories {
		// Skip fork repositories (assumes already excluded, but double-check)
		if repo.GetFork() {
			continue
		}

		stats.RepositoryCount++
		totalStars += repo.GetStargazersCount()
	}

	stats.TotalStars = totalStars
	stats.TotalCommits = totalCommits
	stats.TotalPullRequests = totalPRs

	log.Printf("Summary statistics aggregation completed:")
	log.Printf("  - Total stars: %d", stats.TotalStars)
	log.Printf("  - Repository count: %d", stats.RepositoryCount)
	log.Printf("  - Total commits: %d", stats.TotalCommits)
	log.Printf("  - Total pull requests: %d", stats.TotalPullRequests)

	return stats
}
