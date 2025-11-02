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
			name: "正常系: 複数言語",
			languageTotals: map[string]int{
				"Go":         1000,
				"Python":     500,
				"JavaScript": 200,
			},
			wantCount: 3,
			wantFirst: "Go",
		},
		{
			name:           "空のmap",
			languageTotals: map[string]int{},
			wantCount:      0,
		},
		{
			name: "単一言語",
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

			// パーセンテージの合計を確認（丸め誤差を考慮）
			if len(ranked) > 0 {
				totalPercentage := 0.0
				for _, lang := range ranked {
					totalPercentage += lang.Percentage
				}
				if totalPercentage < 99.9 || totalPercentage > 100.1 {
					t.Errorf("RankLanguages() total percentage = %.2f%%, want ~100%%", totalPercentage)
				}
			}

			// ソート順序を確認（降順）
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
		{"閾値なし（0%）", 0.0, 5},
		{"閾値5%", 5.0, 4},
		{"閾値10%", 10.0, 3},
		{"閾値50%", 50.0, 1},
		{"閾値100%", 100.0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := FilterMinorLanguages(ranked, tt.threshold)

			if len(filtered) != tt.wantCount {
				t.Errorf("FilterMinorLanguages() count = %d, want %d", len(filtered), tt.wantCount)
			}

			// 順序が保持されていることを確認
			for i := 0; i < len(filtered); i++ {
				found := false
				for j, orig := range ranked {
					if orig.Language == filtered[i].Language {
						found = true
						// 順序が変わっていないことを確認（元のインデックスをチェック）
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

			// すべての要素が閾値以上であることを確認
			for _, lang := range filtered {
				if lang.Percentage < tt.threshold {
					t.Errorf("FilterMinorLanguages() language %s has percentage %.2f%% < threshold %.2f%%",
						lang.Language, lang.Percentage, tt.threshold)
				}
			}
		})
	}
}
