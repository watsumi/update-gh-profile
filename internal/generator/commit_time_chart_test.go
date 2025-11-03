package generator

import (
	"strings"
	"testing"
)

func TestGenerateCommitTimeChart(t *testing.T) {
	tests := []struct {
		name             string
		timeDistribution map[int]int
		wantContains     []string
		wantNotContains  []string
	}{
		{
			name: "Normal case: multiple time slot data",
			timeDistribution: map[int]int{
				9:  10,
				10: 15,
				14: 8,
				22: 5,
			},
			wantContains: []string{
				"Commit Time Distribution",
				"UTC",
				":00",
				"<svg",
				"<rect",
			},
			wantNotContains: []string{},
		},
		{
			name:             "Empty data",
			timeDistribution: map[int]int{},
			wantContains: []string{
				"Commit Time Distribution",
				"No data available",
			},
			wantNotContains: []string{},
		},
		{
			name: "When data exists for all time slots",
			timeDistribution: func() map[int]int {
				data := make(map[int]int)
				for i := 0; i < 24; i++ {
					data[i] = i * 2 // Set different values for each time slot
				}
				return data
			}(),
			wantContains: []string{
				"Commit Time Distribution",
				":00",
				"High",
				"Low",
			},
			wantNotContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svg, err := GenerateCommitTimeChart(tt.timeDistribution)
			if err != nil {
				t.Errorf("GenerateCommitTimeChart() error = %v", err)
				return
			}

			// Basic SVG format validation
			if !strings.HasPrefix(svg, "<?xml") {
				t.Errorf("GenerateCommitTimeChart() SVG should start with <?xml")
			}

			if !strings.Contains(svg, "<svg") {
				t.Errorf("GenerateCommitTimeChart() SVG should contain <svg> tag")
			}

			// Check if expected strings are contained
			for _, want := range tt.wantContains {
				if !strings.Contains(svg, want) {
					t.Errorf("GenerateCommitTimeChart() should contain %q", want)
				}
			}

			// Check that strings that should not be contained are not present
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(svg, notWant) {
					t.Errorf("GenerateCommitTimeChart() should not contain %q", notWant)
				}
			}
		})
	}
}
