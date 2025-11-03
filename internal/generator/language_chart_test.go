package generator

import (
	"strings"
	"testing"

	"github.com/watsumi/update-gh-profile/internal/aggregator"
)

func TestGenerateLanguageChart(t *testing.T) {
	tests := []struct {
		name            string
		rankedLanguages []aggregator.LanguageStat
		maxItems        int
		wantContains    []string
		wantNotContains []string
	}{
		{
			name: "正常系: 複数言語",
			rankedLanguages: []aggregator.LanguageStat{
				{Language: "Go", Bytes: 1000, Percentage: 50.0},
				{Language: "Python", Bytes: 500, Percentage: 25.0},
				{Language: "JavaScript", Bytes: 300, Percentage: 15.0},
				{Language: "TypeScript", Bytes: 200, Percentage: 10.0},
			},
			maxItems: 5,
			wantContains: []string{
				"Language Ranking",
				"Go",
				"Python",
				"JavaScript",
				"TypeScript",
				"50.0%",
				"25.0%",
			},
			wantNotContains: []string{},
		},
		{
			name:            "空のデータ",
			rankedLanguages: []aggregator.LanguageStat{},
			maxItems:        5,
			wantContains: []string{
				"Language Ranking",
				"No data available",
			},
			wantNotContains: []string{},
		},
		{
			name: "maxItems未満の言語数",
			rankedLanguages: []aggregator.LanguageStat{
				{Language: "Go", Bytes: 1000, Percentage: 60.0},
				{Language: "Python", Bytes: 400, Percentage: 40.0},
			},
			maxItems: 10,
			wantContains: []string{
				"Go",
				"Python",
				"60.0%",
				"40.0%",
			},
			wantNotContains: []string{},
		},
		{
			name: "特殊文字を含む言語名",
			rankedLanguages: []aggregator.LanguageStat{
				{Language: "C++", Bytes: 1000, Percentage: 50.0},
				{Language: "C#", Bytes: 500, Percentage: 25.0},
			},
			maxItems: 5,
			wantContains: []string{
				"C++",
				"C#",
			},
			wantNotContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svg, err := GenerateLanguageChart(tt.rankedLanguages, tt.maxItems)
			if err != nil {
				t.Errorf("GenerateLanguageChart() error = %v", err)
				return
			}

			// SVG形式の基本的な検証
			if !strings.HasPrefix(svg, "<?xml") {
				t.Errorf("GenerateLanguageChart() SVG should start with <?xml")
			}

			if !strings.Contains(svg, "<svg") {
				t.Errorf("GenerateLanguageChart() SVG should contain <svg> tag")
			}

			// 期待される文字列が含まれているか確認
			for _, want := range tt.wantContains {
				if !strings.Contains(svg, want) {
					t.Errorf("GenerateLanguageChart() should contain %q", want)
				}
			}

			// 含まれていないことを期待する文字列の確認
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(svg, notWant) {
					t.Errorf("GenerateLanguageChart() should not contain %q", notWant)
				}
			}
		})
	}
}

func TestEscapeXML(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:            "アンパサンド",
			input:           "Go & Python",
			wantContains:    []string{"Go", "&amp;", "Python"}, // &amp;が含まれていることを確認
			wantNotContains: []string{},
		},
		{
			name:            "不等号",
			input:           "A < B > C",
			wantContains:    []string{"&lt;", "&gt;"},
			wantNotContains: []string{"<", ">"}, // エスケープされていない < と > が含まれていないこと
		},
		{
			name:            "引用符",
			input:           `"Hello"`,
			wantContains:    []string{"&quot;"},
			wantNotContains: []string{`"`}, // エスケープされていない " が含まれていないこと
		},
		{
			name:            "通常の文字",
			input:           "Go",
			wantContains:    []string{"Go"},
			wantNotContains: []string{},
		},
		{
			name:            "C++（特殊文字なし）",
			input:           "C++",
			wantContains:    []string{"C++"},
			wantNotContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			escaped := escapeXML(tt.input)

			// 期待されるエスケープシーケンスが含まれていることを確認
			for _, want := range tt.wantContains {
				if !strings.Contains(escaped, want) {
					t.Errorf("escapeXML(%q) should contain %q, got %q", tt.input, want, escaped)
				}
			}

			// エスケープされていない文字が含まれていないことを確認
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(escaped, notWant) {
					t.Errorf("escapeXML(%q) should not contain %q, got %q", tt.input, notWant, escaped)
				}
			}
		})
	}
}
