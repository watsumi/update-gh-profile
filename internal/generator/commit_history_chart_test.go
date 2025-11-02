package generator

import (
	"strings"
	"testing"
)

func TestGenerateCommitHistoryChart(t *testing.T) {
	tests := []struct {
		name            string
		commitHistory   map[string]int
		wantContains    []string
		wantNotContains []string
	}{
		{
			name: "正常系: 複数の日付データ",
			commitHistory: map[string]int{
				"2024-01-01": 10,
				"2024-01-02": 15,
				"2024-01-03": 8,
				"2024-01-04": 20,
				"2024-01-05": 12,
			},
			wantContains: []string{
				"コミット推移",
				"01/01", // MM/DD形式のラベル
				"<svg",
				"<path",
				"<circle",
			},
			wantNotContains: []string{},
		},
		{
			name:          "空のデータ",
			commitHistory: map[string]int{},
			wantContains: []string{
				"コミット推移",
				"データがありません",
			},
			wantNotContains: []string{},
		},
		{
			name: "単一日付",
			commitHistory: map[string]int{
				"2024-01-01": 10,
			},
			wantContains: []string{
				"コミット推移",
				"01/01", // MM/DD形式のラベル
				"<svg",
			},
			wantNotContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svg, err := GenerateCommitHistoryChart(tt.commitHistory)
			if err != nil {
				t.Errorf("GenerateCommitHistoryChart() error = %v", err)
				return
			}

			// SVG形式の基本的な検証
			if !strings.HasPrefix(svg, "<?xml") {
				t.Errorf("GenerateCommitHistoryChart() SVG should start with <?xml")
			}

			if !strings.Contains(svg, "<svg") {
				t.Errorf("GenerateCommitHistoryChart() SVG should contain <svg> tag")
			}

			// 期待される文字列が含まれているか確認
			for _, want := range tt.wantContains {
				if !strings.Contains(svg, want) {
					t.Errorf("GenerateCommitHistoryChart() should contain %q", want)
				}
			}

			// 含まれていないことを期待する文字列の確認
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(svg, notWant) {
					t.Errorf("GenerateCommitHistoryChart() should not contain %q", notWant)
				}
			}
		})
	}
}
