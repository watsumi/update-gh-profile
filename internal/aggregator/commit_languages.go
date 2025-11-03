package aggregator

import (
	"log"
	"sort"
	"strings"
)

// AggregateCommitLanguages aggregates top 5 languages used per commit
//
// Preconditions:
//   - commitLanguages is in the format map[string]map[string]map[string]int{repository: {commitSHA: {language: count}}}
//     or map[string]map[string]int{commitSHA: {language: count}}
//   - excludedLanguages is a slice of language names to exclude (can be empty)
//
// Postconditions:
// - Returns a map in the format map[string]int{language: count} containing only top 5 (excluding excluded languages)
// - Sorted by usage count in descending order
//
// Invariants:
// - Sorted by usage count in descending order, top 5 are returned
// - Excluded languages are excluded from aggregation
func AggregateCommitLanguages(commitLanguages map[string]map[string]int, excludedLanguages []string) map[string]int {
	log.Printf("Starting aggregation of language usage per commit: %d commits", len(commitLanguages))

	// Convert exclusion list to map for case-insensitive comparison
	excludedMap := make(map[string]bool)
	for _, lang := range excludedLanguages {
		// Trim whitespace and convert to lowercase for case-insensitive comparison
		normalized := strings.TrimSpace(strings.ToLower(lang))
		if normalized != "" {
			excludedMap[normalized] = true
		}
	}
	if len(excludedMap) > 0 {
		log.Printf("Excluded languages (normalized): %v", excludedMap)
	}

	// Map to aggregate usage count per language
	languageCounts := make(map[string]int)

	// Aggregate language usage for each commit
	for commitSHA, langs := range commitLanguages {
		// Handle short SHA
		shaDisplay := commitSHA
		if len(commitSHA) > 7 {
			shaDisplay = commitSHA[:7]
		}
		log.Printf("  Commit %s: %d languages used", shaDisplay, len(langs))
		for lang, count := range langs {
			// Skip excluded languages (case-insensitive comparison)
			normalized := strings.ToLower(strings.TrimSpace(lang))
			if excludedMap[normalized] {
				log.Printf("    Excluding language: %s (normalized: %s)", lang, normalized)
				continue
			}
			languageCounts[lang] += count
		}
	}

	log.Printf("Language usage count aggregation completed: %d languages", len(languageCounts))

	// Sort by usage count and extract top 5
	top5 := extractTop5Languages(languageCounts)

	log.Printf("Top 5 languages by commit aggregation completed: %d languages", len(top5))
	return top5
}

// extractTop5Languages extracts top 5 languages from usage count per language
//
// Preconditions:
// - languageCounts is in the format map[string]int{language: count}
//
// Postconditions:
// - Returns a map containing only top 5, sorted by usage count in descending order
//
// Invariants:
// - When usage counts are the same, sorted by language name in dictionary order
func extractTop5Languages(languageCounts map[string]int) map[string]int {
	if len(languageCounts) == 0 {
		return make(map[string]int)
	}

	// Create a slice of language and count pairs
	type langCount struct {
		lang  string
		count int
	}
	var langList []langCount
	for lang, count := range languageCounts {
		langList = append(langList, langCount{lang: lang, count: count})
	}

	// Sort by usage count descending (ascending by language name when counts are equal)
	sort.Slice(langList, func(i, j int) bool {
		if langList[i].count != langList[j].count {
			return langList[i].count > langList[j].count
		}
		return langList[i].lang < langList[j].lang
	})

	// Extract top 5
	maxCount := 5
	if len(langList) < maxCount {
		maxCount = len(langList)
	}

	result := make(map[string]int)
	for i := 0; i < maxCount; i++ {
		result[langList[i].lang] = langList[i].count
	}

	return result
}
