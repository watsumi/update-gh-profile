package aggregator

// LanguageStat language statistics
type LanguageStat struct {
	Language        string  // Language name
	Bytes           int     // Total bytes
	Percentage      float64 // Percentage of total
	RepositoryCount int     // Number of repositories where used
}

// SummaryStats summary statistics
type SummaryStats struct {
	TotalStars        int // Total stars
	RepositoryCount   int // Repository count
	TotalCommits      int // Total commits
	TotalPullRequests int // Total pull requests
}

// AggregatedMetrics aggregated metrics
type AggregatedMetrics struct {
	Languages              []LanguageStat // Ranked language slice
	TotalBytes             int            // Total bytes for all languages
	RepositoryCount        int            // Number of target repositories
	CommitHistory          map[string]int // Commit count per date
	CommitTimeDistribution map[int]int    // Commit count per time slot
	CommitLanguages        map[string]int // Top 5 languages by commit
	SummaryStats           SummaryStats   // Summary statistics
}
