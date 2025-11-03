package generator

import (
	"strings"
	"testing"
)

func TestGenerateCommitLanguagesChart(t *testing.T) {
	tests := []struct {
		name            string
		commitLanguages map[string]int
		wantContains    []string
		wantNotContains []string
	}{
		{
			name: "正常系: Top5言語",
			commitLanguages: map[string]int{
				"Go":         100,
				"Python":     80,
				"TypeScript": 60,
				"JavaScript": 40,
				"Rust":       20,
				"C++":        10, // 6番目（表示されない）
			},
			wantContains: []string{
				"Top 5 Languages by Commit",
				"Go",
				"Python",
				"TypeScript",
				"JavaScript",
				"Rust",
				"100 files",
				"80 files",
				"<svg",
			},
			wantNotContains: []string{
				"C++", // Top5に含まれないため表示されない
			},
		},
		{
			name:            "空のデータ",
			commitLanguages: map[string]int{},
			wantContains: []string{
				"Top 5 Languages by Commit",
				"No data available",
			},
			wantNotContains: []string{},
		},
		{
			name: "言語数が5未満",
			commitLanguages: map[string]int{
				"Go":     50,
				"Python": 30,
			},
			wantContains: []string{
				"Go",
				"Python",
				"50 files",
				"30 files",
			},
			wantNotContains: []string{},
		},
		{
			name: "特殊文字を含む言語名",
			commitLanguages: map[string]int{
				"C++": 50,
				"C#":  30,
			},
			wantContains: []string{
				"C++",
				"C#",
			},
			wantNotContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svg, err := GenerateCommitLanguagesChart(tt.commitLanguages)
			if err != nil {
				t.Errorf("GenerateCommitLanguagesChart() error = %v", err)
				return
			}

			// SVG形式の基本的な検証
			if !strings.HasPrefix(svg, "<?xml") {
				t.Errorf("GenerateCommitLanguagesChart() SVG should start with <?xml")
			}

			if !strings.Contains(svg, "<svg") {
				t.Errorf("GenerateCommitLanguagesChart() SVG should contain <svg> tag")
			}

			// 期待される文字列が含まれているか確認
			for _, want := range tt.wantContains {
				if !strings.Contains(svg, want) {
					t.Errorf("GenerateCommitLanguagesChart() should contain %q", want)
				}
			}

			// 含まれていないことを期待する文字列の確認
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(svg, notWant) {
					t.Errorf("GenerateCommitLanguagesChart() should not contain %q", notWant)
				}
			}
		})
	}
}
