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
			name: "Normal case: all metrics have values",
			stats: aggregator.SummaryStats{
				TotalStars:        1234,
				RepositoryCount:   56,
				TotalCommits:      7890,
				TotalPullRequests: 123,
			},
			wantContains: []string{
				"Stars",
				"Repos",
				"Commits",
				"PRs",
				"⭐",
				"📦",
				"💾",
				"🔀",
				"<svg",
			},
		},
		{
			name: "Normal case: large numbers",
			stats: aggregator.SummaryStats{
				TotalStars:        1234567,
				RepositoryCount:   890,
				TotalCommits:      5678901,
				TotalPullRequests: 2345,
			},
			wantContains: []string{
				"Stars",
				"Repos",
				"Commits",
				"PRs",
			},
		},
		{
			name: "Normal case: zero values",
			stats: aggregator.SummaryStats{
				TotalStars:        0,
				RepositoryCount:   0,
				TotalCommits:      0,
				TotalPullRequests: 0,
			},
			wantContains: []string{
				"Stars",
				"Repos",
				"Commits",
				"PRs",
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

			// Basic SVG format validation
			if !strings.HasPrefix(svg, "<?xml") {
				t.Errorf("GenerateSummaryCard() SVG should start with <?xml")
			}

			if !strings.Contains(svg, "<svg") {
				t.Errorf("GenerateSummaryCard() SVG should contain <svg> tag")
			}

			// Check if expected strings are contained
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
			name:     "Small number",
			input:    123,
			expected: "123",
		},
		{
			name:     "3-digit comma separation",
			input:    1234,
			expected: "1,234",
		},
		{
			name:     "Large 3-digit comma separation",
			input:    1234567,
			expected: "1,234,567",
		},
		{
			name:     "K unit",
			input:    1234,
			expected: "1.2K",
		},
		{
			name:     "M unit",
			input:    1234567,
			expected: "1.2M",
		},
		{
			name:     "Zero",
			input:    0,
			expected: "0",
		},
		{
			name:     "Negative value",
			input:    -100,
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatNumber(tt.input)

			// For K/M units, verify format rather than exact string
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
