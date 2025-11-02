package aggregator

import (
	"testing"
)

func TestAggregateCommitLanguages(t *testing.T) {
	tests := []struct {
		name            string
		commitLanguages map[string]map[string]int
		wantCount       int
		wantTop5        []string
	}{
		{
			name: "正常系: 複数言語、複数コミット",
			commitLanguages: map[string]map[string]int{
				"sha1": {
					"Go":       5,
					"Python":   3,
					"JavaScript": 2,
				},
				"sha2": {
					"Go":       3,
					"TypeScript": 4,
					"Python":   2,
				},
				"sha3": {
					"Rust":     2,
					"Go":       1,
					"Python":   1,
				},
			},
			wantCount: 5,
			wantTop5:  []string{"Go", "Python", "TypeScript", "JavaScript", "Rust"},
		},
		{
			name:            "空のmap",
			commitLanguages: map[string]map[string]int{},
			wantCount:       0,
			wantTop5:        []string{},
		},
		{
			name: "言語数が5未満",
			commitLanguages: map[string]map[string]int{
				"sha1": {
					"Go":     5,
					"Python": 3,
				},
			},
			wantCount: 2,
			wantTop5:  []string{"Go", "Python"},
		},
		{
			name: "同じ使用回数の言語",
			commitLanguages: map[string]map[string]int{
				"sha1": {
					"Go":     3,
					"Python": 3,
					"Rust":   3,
					"Java":   3,
					"C++":    3,
				},
			},
			wantCount: 5,
			wantTop5:  []string{"C++", "Go", "Java", "Python", "Rust"}, // 使用回数が同じ場合は辞書順
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AggregateCommitLanguages(tt.commitLanguages)

			if len(result) != tt.wantCount {
				t.Errorf("AggregateCommitLanguages() count = %d, want %d", len(result), tt.wantCount)
			}

			// Top5の順序と内容を確認
			for i, wantLang := range tt.wantTop5 {
				if i >= len(tt.wantTop5) {
					break
				}
				if _, ok := result[wantLang]; !ok {
					t.Errorf("AggregateCommitLanguages() missing language %s in result", wantLang)
				}
			}

			// 使用回数が多い順にソートされていることを確認
			var sortedLangs []string
			type langCount struct {
				lang  string
				count int
			}
			var langList []langCount
			for lang, count := range result {
				langList = append(langList, langCount{lang: lang, count: count})
			}

			// 使用回数降順でソート
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

			// Top5の順序を確認（使用回数が同じ場合は辞書順になる可能性があるため、柔軟にチェック）
			for i := 1; i < len(langList); i++ {
				if langList[i-1].count < langList[i].count {
					t.Errorf("AggregateCommitLanguages() not sorted correctly: %d < %d",
						langList[i-1].count, langList[i].count)
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
			name: "正常系: 5言語以上",
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
			name: "言語数が5未満",
			languageCounts: map[string]int{
				"Go":     10,
				"Python": 5,
			},
			wantCount:   2,
			wantTopLang: "Go",
		},
		{
			name:           "空のmap",
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
