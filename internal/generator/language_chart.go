package generator

import (
	"fmt"
	"math"
	"strings"

	"github.com/watsumi/update-gh-profile/internal/aggregator"
)

// GenerateLanguageChart generates a pie chart SVG from language ranking data
//
// Preconditions:
// - rankedLanguages is a slice of ranked languages
// - maxItems is a positive integer (not used for pie chart, kept for compatibility)
//
// Postconditions:
// - Returns a valid SVG string
// - SVG contains a pie chart with all languages and their percentages
//
// Invariants:
// - SVG has appropriate size and styling
// - Text is displayed in a readable format
func GenerateLanguageChart(rankedLanguages []aggregator.LanguageStat, maxItems int) (string, error) {
	if len(rankedLanguages) == 0 {
		return generateEmptyChart("Language Distribution", "No data available"), nil
	}

	width := DefaultSVGWidth
	height := DefaultSVGHeight
	padding := 20
	titleHeight := 40

	// Pie chart settings
	// Center X is positioned at 75% of width to leave space for legend on the left
	centerX := float64(width) * 0.75
	centerY := float64(titleHeight) + (float64(height-titleHeight-padding) / 2.0)
	radius := 90.0 // Radius of the pie chart

	// Color palette for pie chart slices
	colors := []string{
		"#58a6ff", "#7c3aed", "#1f6feb", "#56d364", "#ff7b72",
		"#a5a5ff", "#f85149", "#79c0ff", "#ffa657", "#ffd33d",
		"#9ecbff", "#bf87ff", "#ffbe6b", "#85e89d", "#ffab70",
	}

	// SVG builder
	var svg strings.Builder

	// Header
	svg.WriteString(fmt.Sprintf(SVGHeader, width, height, width, height))

	// Style definitions (gradient and shadow)
	svg.WriteString(`  <defs>
    <filter id="shadow">
      <feGaussianBlur in="SourceAlpha" stdDeviation="2"/>
      <feOffset dx="0" dy="2" result="offsetblur"/>
      <feComponentTransfer>
        <feFuncA type="linear" slope="0.3"/>
      </feComponentTransfer>
      <feMerge>
        <feMergeNode/>
        <feMergeNode in="SourceGraphic"/>
      </feMerge>
    </filter>
  </defs>

`)

	// Background (with border)
	svg.WriteString(fmt.Sprintf(`  <rect width="%d" height="%d" fill="%s" rx="10" stroke="#30363d" stroke-width="1"/>
`, width, height, DefaultBackgroundColor))

	// Title
	svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="20" font-weight="700" fill="%s" text-anchor="middle" filter="url(#shadow)">🗂️ Language Distribution</text>
`, width/2, 32, AccentColor))

	// Calculate total percentage for normalization (in case sum is not 100%)
	totalPercentage := 0.0
	for _, lang := range rankedLanguages {
		totalPercentage += lang.Percentage
	}

	// Draw pie chart slices
	currentAngle := -90.0 // Start from top (-90 degrees in SVG)
	for i, lang := range rankedLanguages {
		color := colors[i%len(colors)]
		percentage := lang.Percentage
		if totalPercentage > 0 {
			percentage = (lang.Percentage / totalPercentage) * 100.0
		}

		// Calculate slice angle (convert percentage to degrees)
		angle := (percentage / 100.0) * 360.0

		// Only draw if percentage is significant (> 0.1%)
		if percentage > 0.1 {
			// Calculate end angle
			endAngle := currentAngle + angle

			// Convert angles to radians for calculations
			startRad := currentAngle * math.Pi / 180.0
			endRad := endAngle * math.Pi / 180.0

			// Calculate arc endpoints
			x1 := centerX + radius*math.Cos(startRad)
			y1 := centerY + radius*math.Sin(startRad)
			x2 := centerX + radius*math.Cos(endRad)
			y2 := centerY + radius*math.Sin(endRad)

			// Large arc flag (1 if angle > 180 degrees, 0 otherwise)
			largeArcFlag := 0
			if angle > 180.0 {
				largeArcFlag = 1
			}

			// Create path for pie slice
			path := fmt.Sprintf("M %.1f %.1f L %.1f %.1f A %.1f %.1f 0 %d 1 %.1f %.1f Z",
				centerX, centerY, x1, y1, radius, radius, largeArcFlag, x2, y2)

			// Draw slice
			svg.WriteString(fmt.Sprintf(`  <path d="%s" fill="%s" stroke="%s" stroke-width="2" opacity="0.9" filter="url(#shadow)"/>
`, path, color, DefaultBackgroundColor))

			currentAngle = endAngle
		}
	}

	// Draw legend and labels on the left side
	legendX := padding // Start legend near the left padding
	legendY := titleHeight + 25
	legendItemHeight := 22
	legendRightX := (width / 2) - padding

	// Draw legend items for all languages
	for i, lang := range rankedLanguages {
		if i >= 15 { // Limit to 15 items to fit in the chart
			break
		}

		color := colors[i%len(colors)]
		y := legendY + (i * legendItemHeight)

		// Color square
		svg.WriteString(fmt.Sprintf(`  <rect x="%d" y="%d" width="12" height="12" fill="%s" rx="2"/>
`, legendX, y-8, color))

		// Language name
		langText := escapeXML(lang.Language)
		if len(langText) > 15 {
			langText = langText[:12] + "..."
		}
		svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="11" fill="%s">%s</text>
`, legendX+18, y, DefaultTextColor, langText))

		// Percentage
		percentage := lang.Percentage
		if totalPercentage > 0 {
			percentage = (lang.Percentage / totalPercentage) * 100.0
		}
		percentageText := fmt.Sprintf("%.1f%%", percentage)
		svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="11" fill="%s" font-weight="600" text-anchor="end">%s</text>
`, legendRightX, y, AccentColor, percentageText))
	}

	// If there are more than 15 languages, show count
	if len(rankedLanguages) > 15 {
		remaining := len(rankedLanguages) - 15
		svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="10" fill="%s" font-style="italic">+%d more languages</text>
`, legendX, legendY+(15*legendItemHeight), DefaultTextColor, remaining))
	}

	// Footer
	svg.WriteString(SVGFooter)

	return svg.String(), nil
}

// generateEmptyChart generates a chart for empty data
func generateEmptyChart(title, message string) string {
	width := DefaultSVGWidth
	height := 200

	var svg strings.Builder
	svg.WriteString(fmt.Sprintf(SVGHeader, width, height, width, height))
	svg.WriteString(fmt.Sprintf(`  <rect width="%d" height="%d" fill="%s" rx="8"/>
`, width, height, DefaultBackgroundColor))
	svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="18" font-weight="600" fill="%s">%s</text>
`, width/2, 60, DefaultTextColor, title))
	svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="14" fill="%s">%s</text>
`, width/2, 100, DefaultTextColor, message))
	svg.WriteString(SVGFooter)

	return svg.String()
}

// escapeXML escapes XML special characters
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
