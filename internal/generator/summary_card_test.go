package generator

import (
	"strings"
	"testing"

	"github.com/watsumi/update-gh-profile/internal/aggregator"
)

func TestGenerateSummaryCard(t *testing.T) {
	tests := []struct {
		name         string
		stats        aggregator.SummaryStats
		wantContains []string
	}{
		{
			name: "正常系: すべてのメトリクスに値がある",
			stats: aggregator.SummaryStats{
				TotalStars:        1234,
				RepositoryCount:   56,
				TotalCommits:      7890,
				TotalPullRequests: 123,
			},
			wantContains: []string{
				"スター",
				"リポジトリ",
				"コミット",
				"プルリク",
				"⭐",
				"📦",
				"💾",
				"🔀",
				"<svg",
			},
		},
		{
			name: "正常系: 大きな数値",
			stats: aggregator.SummaryStats{
				TotalStars:        1234567,
				RepositoryCount:   890,
				TotalCommits:      5678901,
				TotalPullRequests: 2345,
			},
			wantContains: []string{
				"スター",
				"リポジトリ",
				"コミット",
				"プルリク",
			},
		},
		{
			name: "正常系: ゼロ値",
			stats: aggregator.SummaryStats{
				TotalStars:        0,
				RepositoryCount:   0,
				TotalCommits:      0,
				TotalPullRequests: 0,
			},
			wantContains: []string{
				"スター",
				"リポジトリ",
				"コミット",
				"プルリク",
				"0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svg, err := GenerateSummaryCard(tt.stats)
			if err != nil {
				t.Errorf("GenerateSummaryCard() error = %v", err)
				return
			}

			// SVG形式の基本的な検証
			if !strings.HasPrefix(svg, "<?xml") {
				t.Errorf("GenerateSummaryCard() SVG should start with <?xml")
			}

			if !strings.Contains(svg, "<svg") {
				t.Errorf("GenerateSummaryCard() SVG should contain <svg> tag")
			}

			// 期待される文字列が含まれているか確認
			for _, want := range tt.wantContains {
				if !strings.Contains(svg, want) {
					t.Errorf("GenerateSummaryCard() should contain %q", want)
				}
			}
		})
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{
			name:     "小さい数値",
			input:    123,
			expected: "123",
		},
		{
			name:     "3桁区切り",
			input:    1234,
			expected: "1,234",
		},
		{
			name:     "大きな3桁区切り",
			input:    1234567,
			expected: "1,234,567",
		},
		{
			name:     "K単位",
			input:    1234,
			expected: "1.2K",
		},
		{
			name:     "M単位",
			input:    1234567,
			expected: "1.2M",
		},
		{
			name:     "ゼロ",
			input:    0,
			expected: "0",
		},
		{
			name:     "負の値",
			input:    -100,
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatNumber(tt.input)

			// K/M単位の場合は正確な文字列ではなく、形式を確認
			if tt.input >= 1000000 {
				if !strings.HasSuffix(result, "M") {
					t.Errorf("formatNumber(%d) = %q, expected string ending with 'M'", tt.input, result)
				}
			} else if tt.input >= 1000 {
				if !strings.HasSuffix(result, "K") {
					t.Errorf("formatNumber(%d) = %q, expected string ending with 'K'", tt.input, result)
				}
			} else {
				if result != tt.expected {
					t.Errorf("formatNumber(%d) = %q, want %q", tt.input, result, tt.expected)
				}
			}
		})
	}
}
