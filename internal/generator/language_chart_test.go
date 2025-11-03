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
			name: "Normal case: multiple languages",
			rankedLanguages: []aggregator.LanguageStat{
				{Language: "Go", Bytes: 1000, Percentage: 50.0},
				{Language: "Python", Bytes: 500, Percentage: 25.0},
				{Language: "JavaScript", Bytes: 300, Percentage: 15.0},
				{Language: "TypeScript", Bytes: 200, Percentage: 10.0},
			},
			maxItems: 5,
			wantContains: []string{
				"Language Distribution",
				"Go",
				"Python",
				"JavaScript",
				"TypeScript",
				"50.0%",
				"25.0%",
				"<path", // Pie chart path elements
			},
			wantNotContains: []string{},
		},
		{
			name:            "Empty data",
			rankedLanguages: []aggregator.LanguageStat{},
			maxItems:        5,
			wantContains: []string{
				"Language Distribution",
				"No data available",
			},
			wantNotContains: []string{},
		},
		{
			name: "Number of languages less than maxItems",
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
			name: "Language names with special characters",
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

			// Basic SVG format validation
			if !strings.HasPrefix(svg, "<?xml") {
				t.Errorf("GenerateLanguageChart() SVG should start with <?xml")
			}

			if !strings.Contains(svg, "<svg") {
				t.Errorf("GenerateLanguageChart() SVG should contain <svg> tag")
			}

			// Check if expected strings are contained
			for _, want := range tt.wantContains {
				if !strings.Contains(svg, want) {
					t.Errorf("GenerateLanguageChart() should contain %q", want)
				}
			}

			// Check that strings that should not be contained are not present
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
			name:            "Ampersand",
			input:           "Go & Python",
			wantContains:    []string{"Go", "&amp;", "Python"}, // Verify that &amp; is contained
			wantNotContains: []string{},
		},
		{
			name:            "Inequality operators",
			input:           "A < B > C",
			wantContains:    []string{"&lt;", "&gt;"},
			wantNotContains: []string{"<", ">"}, // Verify that unescaped < and > are not contained
		},
		{
			name:            "Quotes",
			input:           `"Hello"`,
			wantContains:    []string{"&quot;"},
			wantNotContains: []string{`"`}, // Verify that unescaped " is not contained
		},
		{
			name:            "Normal characters",
			input:           "Go",
			wantContains:    []string{"Go"},
			wantNotContains: []string{},
		},
		{
			name:            "C++ (no special characters)",
			input:           "C++",
			wantContains:    []string{"C++"},
			wantNotContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			escaped := escapeXML(tt.input)

			// Verify that expected escape sequences are contained
			for _, want := range tt.wantContains {
				if !strings.Contains(escaped, want) {
					t.Errorf("escapeXML(%q) should contain %q, got %q", tt.input, want, escaped)
				}
			}

			// Verify that unescaped characters are not contained
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(escaped, notWant) {
					t.Errorf("escapeXML(%q) should not contain %q, got %q", tt.input, notWant, escaped)
				}
			}
		})
	}
}
