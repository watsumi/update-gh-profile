package aggregator

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/google/go-github/v76/github"
)

// AggregateLanguages aggregates language data from all repositories
//
// Preconditions:
// - repositories is a slice of repository structs
// - languageData is in the format map[string]map[string]int{repository: {language: bytes}}
//
// Postconditions:
// - Returns a map in the format map[string]int{language: totalBytes}
// - Forked repository data is excluded (assumes repositories already filtered)
//
// Invariants:
// - Data for the same language is summed
// - Forked repositories are excluded
func AggregateLanguages(repositories []*github.Repository, languageData map[string]map[string]int) map[string]int {
	log.Printf("Starting language data aggregation: %d repositories", len(repositories))

	// Map to aggregate total bytes per language
	languageTotals := make(map[string]int)
	// Map to count repositories using each language
	languageRepoCounts := make(map[string]map[string]bool)

	for _, repo := range repositories {
		// Skip forked repositories (assumes already filtered, but check just in case)
		if repo.GetFork() {
			continue
		}

		// Generate repository name (owner/repo format)
		repoKey := fmt.Sprintf("%s/%s", repo.GetOwner().GetLogin(), repo.GetName())

		// Get language data for this repository
		langs, ok := languageData[repoKey]
		if !ok {
			continue // Skip if no language data
		}

		// Sum bytes for each language
		for lang, bytes := range langs {
			languageTotals[lang] += bytes

			// Count repositories using each language
			if languageRepoCounts[lang] == nil {
				languageRepoCounts[lang] = make(map[string]bool)
			}
			languageRepoCounts[lang][repoKey] = true
		}
	}

	log.Printf("Language data aggregation completed: %d languages", len(languageTotals))
	return languageTotals
}

// RankLanguages ranks language data
//
// Preconditions:
// - languageTotals is in the format map[string]int{language: totalBytes}
//
// Postconditions:
// - Returns a slice of LanguageStat structs, sorted by bytes in descending order
// - Each LanguageStat contains a percentage
//
// Invariants:
// - Total percentage equals 100% (excluding rounding errors)
func RankLanguages(languageTotals map[string]int) []LanguageStat {
	if len(languageTotals) == 0 {
		return []LanguageStat{}
	}

	// Calculate total bytes
	totalBytes := 0
	for _, bytes := range languageTotals {
		totalBytes += bytes
	}

	if totalBytes == 0 {
		log.Printf("Warning: total bytes is 0")
		return []LanguageStat{}
	}

	// Create LanguageStat slice
	var ranked []LanguageStat
	for lang, bytes := range languageTotals {
		percentage := float64(bytes) / float64(totalBytes) * 100.0
		ranked = append(ranked, LanguageStat{
			Language:   lang,
			Bytes:      bytes,
			Percentage: percentage,
		})
	}

	// Sort by bytes in descending order
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Bytes > ranked[j].Bytes
	})

	log.Printf("Language ranking generation completed: %d languages (total bytes: %d)", len(ranked), totalBytes)
	return ranked
}

// FilterMinorLanguages excludes languages below the threshold
//
// Preconditions:
// - rankedLanguages is a slice of ranked languages
// - threshold is between 0 and 100
//
// Postconditions:
// - Returns a slice containing only languages with percentage above or equal to threshold
//
// Invariants:
// - Original slice order is preserved
func FilterMinorLanguages(rankedLanguages []LanguageStat, threshold float64) []LanguageStat {
	if threshold < 0 || threshold > 100 {
		log.Printf("Warning: threshold is out of range (%f). Including all languages", threshold)
		return rankedLanguages
	}

	var filtered []LanguageStat
	for _, lang := range rankedLanguages {
		if lang.Percentage >= threshold {
			filtered = append(filtered, lang)
		}
	}

	log.Printf("Filtering by threshold (%.2f%%) completed: %d languages → %d languages", threshold, len(rankedLanguages), len(filtered))
	return filtered
}

// FilterExcludedLanguages excludes specified languages
//
// Preconditions:
// - rankedLanguages is a slice of ranked languages
// - excludedLanguages is a slice of language names to exclude (can be empty)
//
// Postconditions:
// - Returns a slice containing only languages not in the exclusion list
//
// Invariants:
// - Original slice order is preserved
// - Case-insensitive comparison
func FilterExcludedLanguages(rankedLanguages []LanguageStat, excludedLanguages []string) []LanguageStat {
	if len(excludedLanguages) == 0 {
		return rankedLanguages
	}

	// Convert exclusion list to map for case-insensitive comparison
	excludedMap := make(map[string]bool)
	for _, lang := range excludedLanguages {
		// Trim whitespace and convert to lowercase for case-insensitive comparison
		normalized := strings.TrimSpace(strings.ToLower(lang))
		if normalized != "" {
			excludedMap[normalized] = true
		}
	}

	var filtered []LanguageStat
	for _, lang := range rankedLanguages {
		// Normalize language name for comparison
		normalized := strings.ToLower(lang.Language)
		if !excludedMap[normalized] {
			filtered = append(filtered, lang)
		}
	}

	log.Printf("Filtering by excluded languages completed: %d languages → %d languages (excluded: %v)", len(rankedLanguages), len(filtered), excludedLanguages)
	return filtered
}
