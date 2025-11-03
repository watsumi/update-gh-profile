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
			name: "Normal case: multiple date data",
			commitHistory: map[string]int{
				"2024-01-01": 10,
				"2024-01-02": 15,
				"2024-01-03": 8,
				"2024-01-04": 20,
				"2024-01-05": 12,
			},
			wantContains: []string{
				"Commit History",
				"01/01", // MM/DD format label
				"<svg",
				"<rect", // Changed to bar chart
			},
			wantNotContains: []string{},
		},
		{
			name:          "Empty data",
			commitHistory: map[string]int{},
			wantContains: []string{
				"Commit History",
				"No data available",
			},
			wantNotContains: []string{},
		},
		{
			name: "Single date",
			commitHistory: map[string]int{
				"2024-01-01": 10,
			},
			wantContains: []string{
				"Commit History",
				"01/01", // MM/DD format label
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

			// Basic SVG format validation
			if !strings.HasPrefix(svg, "<?xml") {
				t.Errorf("GenerateCommitHistoryChart() SVG should start with <?xml")
			}

			if !strings.Contains(svg, "<svg") {
				t.Errorf("GenerateCommitHistoryChart() SVG should contain <svg> tag")
			}

			// Check if expected strings are contained
			for _, want := range tt.wantContains {
				if !strings.Contains(svg, want) {
					t.Errorf("GenerateCommitHistoryChart() should contain %q", want)
				}
			}

			// Check that strings that should not be contained are not present
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(svg, notWant) {
					t.Errorf("GenerateCommitHistoryChart() should not contain %q", notWant)
				}
			}
		})
	}
}
