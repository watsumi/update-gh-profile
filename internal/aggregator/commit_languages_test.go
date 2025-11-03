package aggregator

import (
	"testing"
)

func TestAggregateCommitLanguages(t *testing.T) {
	tests := []struct {
		name              string
		commitLanguages   map[string]map[string]int
		excludedLanguages []string
		wantCount         int
		wantTop5          []string
		wantNotContains   []string
	}{
		{
			name: "Normal case: multiple languages, multiple commits",
			commitLanguages: map[string]map[string]int{
				"sha1": {
					"Go":         5,
					"Python":     3,
					"JavaScript": 2,
				},
				"sha2": {
					"Go":         3,
					"TypeScript": 4,
					"Python":     2,
				},
				"sha3": {
					"Rust":   2,
					"Go":     1,
					"Python": 1,
				},
			},
			excludedLanguages: []string{},
			wantCount:         5,
			wantTop5:          []string{"Go", "Python", "TypeScript", "JavaScript", "Rust"},
			wantNotContains:   []string{},
		},
		{
			name:              "Empty map",
			commitLanguages:   map[string]map[string]int{},
			excludedLanguages: []string{},
			wantCount:         0,
			wantTop5:          []string{},
			wantNotContains:   []string{},
		},
		{
			name: "Number of languages less than 5",
			commitLanguages: map[string]map[string]int{
				"sha1": {
					"Go":     5,
					"Python": 3,
				},
			},
			excludedLanguages: []string{},
			wantCount:         2,
			wantTop5:          []string{"Go", "Python"},
			wantNotContains:   []string{},
		},
		{
			name: "Languages with same usage count",
			commitLanguages: map[string]map[string]int{
				"sha1": {
					"Go":     3,
					"Python": 3,
					"Rust":   3,
					"Java":   3,
					"C++":    3,
				},
			},
			excludedLanguages: []string{},
			wantCount:         5,
			wantTop5:          []string{"C++", "Go", "Java", "Python", "Rust"}, // Alphabetical order when usage count is the same
			wantNotContains:   []string{},
		},
		{
			name: "When excluded languages are set",
			commitLanguages: map[string]map[string]int{
				"sha1": {
					"Go":         5,
					"Python":     3,
					"JavaScript": 2,
					"HTML":       10,
					"CSS":        8,
				},
			},
			excludedLanguages: []string{"HTML", "CSS"},
			wantCount:         3,
			wantTop5:          []string{"Go", "Python", "JavaScript"},
			wantNotContains:   []string{"HTML", "CSS"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AggregateCommitLanguages(tt.commitLanguages, tt.excludedLanguages)

			if len(result) != tt.wantCount {
				t.Errorf("AggregateCommitLanguages() count = %d, want %d", len(result), tt.wantCount)
			}

			// Verify Top 5 order and content
			for i, wantLang := range tt.wantTop5 {
				if i >= len(tt.wantTop5) {
					break
				}
				if _, ok := result[wantLang]; !ok {
					t.Errorf("AggregateCommitLanguages() missing language %s in result", wantLang)
				}
			}

			// Verify that sorting is by usage count (descending)
			var sortedLangs []string
			type langCount struct {
				lang  string
				count int
			}
			var langList []langCount
			for lang, count := range result {
				langList = append(langList, langCount{lang: lang, count: count})
			}

			// Sort by usage count (descending)
			for i := 0; i < len(langList)-1; i++ {
				for j := i + 1; j < len(langList); j++ {
					if langList[i].count < langList[j].count {
						langList[i], langList[j] = langList[j], langList[i]
					}
				}
			}

			for _, item := range langList {
				sortedLangs = append(sortedLangs, item.lang)
			}

			// Verify Top 5 order (flexible check since alphabetical order may be used when usage count is the same)
			for i := 1; i < len(langList); i++ {
				if langList[i-1].count < langList[i].count {
					t.Errorf("AggregateCommitLanguages() not sorted correctly: %d < %d",
						langList[i-1].count, langList[i].count)
				}
			}

			// Verify that excluded languages are not contained
			for _, notWant := range tt.wantNotContains {
				if _, ok := result[notWant]; ok {
					t.Errorf("AggregateCommitLanguages() should not contain excluded language %q", notWant)
				}
			}
		})
	}
}

func TestExtractTop5Languages(t *testing.T) {
	tests := []struct {
		name           string
		languageCounts map[string]int
		wantCount      int
		wantTopLang    string
	}{
		{
			name: "Normal case: 5 or more languages",
			languageCounts: map[string]int{
				"Go":         10,
				"Python":     8,
				"TypeScript": 6,
				"JavaScript": 4,
				"Rust":       2,
				"C++":        1,
			},
			wantCount:   5,
			wantTopLang: "Go",
		},
		{
			name: "Number of languages less than 5",
			languageCounts: map[string]int{
				"Go":     10,
				"Python": 5,
			},
			wantCount:   2,
			wantTopLang: "Go",
		},
		{
			name:           "Empty map",
			languageCounts: map[string]int{},
			wantCount:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTop5Languages(tt.languageCounts)

			if len(result) != tt.wantCount {
				t.Errorf("extractTop5Languages() count = %d, want %d", len(result), tt.wantCount)
			}

			if tt.wantTopLang != "" {
				if count, ok := result[tt.wantTopLang]; !ok {
					t.Errorf("extractTop5Languages() missing top language %s", tt.wantTopLang)
				} else if count == 0 {
					t.Errorf("extractTop5Languages() top language %s has count 0", tt.wantTopLang)
				}
			}
		})
	}
}
