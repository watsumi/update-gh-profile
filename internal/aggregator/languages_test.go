package aggregator

import (
	"testing"
)

func TestRankLanguages(t *testing.T) {
	tests := []struct {
		name           string
		languageTotals map[string]int
		wantCount      int
		wantFirst      string
	}{
		{
			name: "Normal case: multiple languages",
			languageTotals: map[string]int{
				"Go":         1000,
				"Python":     500,
				"JavaScript": 200,
			},
			wantCount: 3,
			wantFirst: "Go",
		},
		{
			name:           "Empty map",
			languageTotals: map[string]int{},
			wantCount:      0,
		},
		{
			name: "Single language",
			languageTotals: map[string]int{
				"Go": 1000,
			},
			wantCount: 1,
			wantFirst: "Go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ranked := RankLanguages(tt.languageTotals)

			if len(ranked) != tt.wantCount {
				t.Errorf("RankLanguages() count = %d, want %d", len(ranked), tt.wantCount)
			}

			if tt.wantCount > 0 && ranked[0].Language != tt.wantFirst {
				t.Errorf("RankLanguages() first = %s, want %s", ranked[0].Language, tt.wantFirst)
			}

			// Verify total percentage (accounting for rounding errors)
			if len(ranked) > 0 {
				totalPercentage := 0.0
				for _, lang := range ranked {
					totalPercentage += lang.Percentage
				}
				if totalPercentage < 99.9 || totalPercentage > 100.1 {
					t.Errorf("RankLanguages() total percentage = %.2f%%, want ~100%%", totalPercentage)
				}
			}

			// Verify sort order (descending)
			for i := 1; i < len(ranked); i++ {
				if ranked[i-1].Bytes < ranked[i].Bytes {
					t.Errorf("RankLanguages() not sorted correctly: %d < %d", ranked[i-1].Bytes, ranked[i].Bytes)
				}
			}
		})
	}
}

func TestFilterMinorLanguages(t *testing.T) {
	ranked := []LanguageStat{
		{Language: "Go", Bytes: 1000, Percentage: 60.0},
		{Language: "Python", Bytes: 300, Percentage: 18.0},
		{Language: "JavaScript", Bytes: 200, Percentage: 12.0},
		{Language: "Rust", Bytes: 100, Percentage: 6.0},
		{Language: "C", Bytes: 70, Percentage: 4.0},
	}

	tests := []struct {
		name      string
		threshold float64
		wantCount int
	}{
		{"No threshold (0%)", 0.0, 5},
		{"Threshold 5%", 5.0, 4},
		{"Threshold 10%", 10.0, 3},
		{"Threshold 50%", 50.0, 1},
		{"Threshold 100%", 100.0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := FilterMinorLanguages(ranked, tt.threshold)

			if len(filtered) != tt.wantCount {
				t.Errorf("FilterMinorLanguages() count = %d, want %d", len(filtered), tt.wantCount)
			}

			// Verify that order is preserved
			for i := 0; i < len(filtered); i++ {
				found := false
				for j, orig := range ranked {
					if orig.Language == filtered[i].Language {
						found = true
						// Verify that order hasn't changed (check original index)
						if i != j {
							t.Errorf("FilterMinorLanguages() order changed: expected %s at index %d, got at %d",
								filtered[i].Language, j, i)
						}
						break
					}
				}
				if !found {
					t.Errorf("FilterMinorLanguages() language %s not found in original", filtered[i].Language)
				}
			}

			// Verify that all elements are above threshold
			for _, lang := range filtered {
				if lang.Percentage < tt.threshold {
					t.Errorf("FilterMinorLanguages() language %s has percentage %.2f%% < threshold %.2f%%",
						lang.Language, lang.Percentage, tt.threshold)
				}
			}
		})
	}
}
