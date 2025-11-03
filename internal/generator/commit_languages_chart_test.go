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
			name: "Normal case: Top 5 languages",
			commitLanguages: map[string]int{
				"Go":         100,
				"Python":     80,
				"TypeScript": 60,
				"JavaScript": 40,
				"Rust":       20,
				"C++":        10, // 6th place (not displayed)
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
				"C++", // Not displayed because not in Top 5
			},
		},
		{
			name:            "Empty data",
			commitLanguages: map[string]int{},
			wantContains: []string{
				"Top 5 Languages by Commit",
				"No data available",
			},
			wantNotContains: []string{},
		},
		{
			name: "Number of languages less than 5",
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
			name: "Language names with special characters",
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

			// Basic SVG format validation
			if !strings.HasPrefix(svg, "<?xml") {
				t.Errorf("GenerateCommitLanguagesChart() SVG should start with <?xml")
			}

			if !strings.Contains(svg, "<svg") {
				t.Errorf("GenerateCommitLanguagesChart() SVG should contain <svg> tag")
			}

			// Check if expected strings are contained
			for _, want := range tt.wantContains {
				if !strings.Contains(svg, want) {
					t.Errorf("GenerateCommitLanguagesChart() should contain %q", want)
				}
			}

			// Check that strings that should not be contained are not present
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(svg, notWant) {
					t.Errorf("GenerateCommitLanguagesChart() should not contain %q", notWant)
				}
			}
		})
	}
}
