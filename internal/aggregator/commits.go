package aggregator

import (
	"log"
	"sort"
)

// AggregateCommitHistory aggregates commit counts by date
//
// Preconditions:
// - commitHistories is in the format map[string]map[string]int{repository name: {date: commit count}}
//
// Postconditions:
// - Returns a map in the format map[string]int{date: total commit count}
// - Dates are recorded in YYYY-MM-DD format
//
// Invariants:
// - Commit counts per date from all repositories are summed
func AggregateCommitHistory(commitHistories map[string]map[string]int) map[string]int {
	log.Printf("Starting commit history aggregation: %d repositories", len(commitHistories))

	// Map to store total commit counts per date
	aggregated := make(map[string]int)

	// Aggregate commit history for each repository
	for repoName, history := range commitHistories {
		log.Printf("  %s: aggregating commit history for %d days", repoName, len(history))
		for date, count := range history {
			aggregated[date] += count
		}
	}

	log.Printf("Commit history aggregation completed: %d days", len(aggregated))
	return aggregated
}

// SortCommitHistoryByDate sorts commit history by date
//
// Preconditions:
// - commitHistory is in the format map[string]int{date: commit count}
//
// Postconditions:
// - Returns a slice of date and commit count pairs sorted by date (ascending)
//
// Invariants:
// - Dates are recorded in YYYY-MM-DD format
// - Sort order is based on string comparison of dates (correctly sorted due to YYYY-MM-DD format)
func SortCommitHistoryByDate(commitHistory map[string]int) []DateCommitPair {
	if len(commitHistory) == 0 {
		return []DateCommitPair{}
	}

	// Create DateCommitPair slice
	var pairs []DateCommitPair
	for date, count := range commitHistory {
		pairs = append(pairs, DateCommitPair{
			Date:  date,
			Count: count,
		})
	}

	// Sort by date (ascending)
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Date < pairs[j].Date
	})

	return pairs
}

// DateCommitPair date and commit count pair
type DateCommitPair struct {
	Date  string // Date (YYYY-MM-DD format)
	Count int    // Commit count
}

// AggregateCommitTimeDistribution aggregates commit counts by time slot
//
// Preconditions:
// - timeDistributions is in the format map[string]map[int]int{repository name: {time slot: commit count}}
//
// Postconditions:
// - Returns a map in the format map[int]int{time slot: total commit count}
// - Time slots are recorded in the range 0-23
//
// Invariants:
// - Commit counts per time slot from all repositories are summed
func AggregateCommitTimeDistribution(timeDistributions map[string]map[int]int) map[int]int {
	log.Printf("Starting commit time distribution aggregation: %d repositories", len(timeDistributions))

	// Map to store total commit counts per time slot (0-23 hours)
	aggregated := make(map[int]int)

	// Aggregate time distribution for each repository
	for repoName, distribution := range timeDistributions {
		log.Printf("  %s: aggregating data for %d time slots", repoName, len(distribution))
		for hour, count := range distribution {
			// Verify time slot is within 0-23 range
			if hour < 0 || hour > 23 {
				log.Printf("Warning: time slot %d for repository %s is out of range. Skipping", hour, repoName)
				continue
			}
			aggregated[hour] += count
		}
	}

	log.Printf("Commit time distribution aggregation completed: %d time slots", len(aggregated))
	return aggregated
}

// SortCommitTimeDistributionByHour sorts commit time distribution by time slot
//
// Preconditions:
// - timeDistribution is in the format map[int]int{time slot: commit count}
//
// Postconditions:
// - Returns a slice of time slot and commit count pairs sorted by time slot (ascending, 0-23 hours)
//
// Invariants:
// - Time slots are recorded in the range 0-23
func SortCommitTimeDistributionByHour(timeDistribution map[int]int) []HourCommitPair {
	if len(timeDistribution) == 0 {
		return []HourCommitPair{}
	}

	// Create HourCommitPair slice
	var pairs []HourCommitPair
	for hour, count := range timeDistribution {
		// Range check
		if hour < 0 || hour > 23 {
			continue
		}
		pairs = append(pairs, HourCommitPair{
			Hour:  hour,
			Count: count,
		})
	}

	// Sort by time slot (ascending)
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Hour < pairs[j].Hour
	})

	return pairs
}

// HourCommitPair time slot and commit count pair
type HourCommitPair struct {
	Hour  int // Time slot (0-23 hours)
	Count int // Commit count
}
