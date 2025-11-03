package generator

import (
	"fmt"
	"sort"
	"strings"
)

// GenerateCommitLanguagesChart generates an SVG displaying top 5 languages by commit
//
// Preconditions:
// - commitLanguages is in the format map[string]int{language name: usage count}
//
// Postconditions:
// - Returns a valid SVG string
// - SVG displays top 5 languages and their usage counts
//
// Invariants:
// - Only top 5 languages are displayed
func GenerateCommitLanguagesChart(commitLanguages map[string]int) (string, error) {
	if len(commitLanguages) == 0 {
		return generateEmptyChart("Top 5 Languages by Commit", "No data available"), nil
	}

	// Sort by usage count and extract top 5
	type langCount struct {
		lang  string
		count int
	}
	var langList []langCount
	for lang, count := range commitLanguages {
		langList = append(langList, langCount{lang: lang, count: count})
	}

	// Sort in descending order by usage count
	sort.Slice(langList, func(i, j int) bool {
		return langList[i].count > langList[j].count
	})

	// Get top 5
	maxItems := 5
	if len(langList) > maxItems {
		langList = langList[:maxItems]
	}

	if len(langList) == 0 {
		return generateEmptyChart("Top 5 Languages by Commit", "No data available"), nil
	}

	// Set SVG size
	width := DefaultSVGWidth
	height := 280
	padding := 20
	chartWidth := width - padding*2

	// Get maximum usage count
	maxCount := langList[0].count
	if maxCount == 0 {
		maxCount = 1 // Prevent division by zero
	}

	// Build SVG
	var svg strings.Builder

	// Header
	svg.WriteString(fmt.Sprintf(SVGHeader, width, height, width, height))

	// Style definitions (richer color palette)
	colors := []string{"#58a6ff", "#7c3aed", "#1f6feb", "#56d364", "#ff7b72"}
	svg.WriteString(`  <defs>
`)

	for i, color := range colors {
		svg.WriteString(fmt.Sprintf(`    <linearGradient id="grad%d" x1="0%%" y1="0%%" x2="100%%" y2="0%%">
      <stop offset="0%%" style="stop-color:%s;stop-opacity:1" />
      <stop offset="100%%" style="stop-color:%s;stop-opacity:0.8" />
    </linearGradient>
`, i, color, color))
	}

	svg.WriteString(`    <filter id="barShadow">
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

	// Title (decorated)
	svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="20" font-weight="700" fill="%s" text-anchor="middle">💻 Top 5 Languages by Commit</text>
`, width/2, 37, AccentColor))

	// Display bar chart
	barHeight := 30
	barSpacing := 45
	startY := 70
	barMaxWidth := chartWidth - 200 // Reserve space for language name and count

	for i, item := range langList {
		yPos := startY + (i * barSpacing)

		// Calculate bar width
		barWidth := int(float64(barMaxWidth) * float64(item.count) / float64(maxCount))

		// Language name
		svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="14" fill="%s">%s</text>
`, padding, yPos+5, DefaultTextColor, escapeXML(item.lang)))

		// Bar background
		barX := 140
		svg.WriteString(fmt.Sprintf(`  <rect x="%d" y="%d" width="%d" height="%d" fill="#161b22" rx="6" stroke="#30363d" stroke-width="1"/>
`, barX, yPos-12, barMaxWidth, barHeight))

		// Bar (gradient + shadow)
		if barWidth > 0 {
			colorIndex := i % len(colors)
			svg.WriteString(fmt.Sprintf(`  <rect x="%d" y="%d" width="%d" height="%d" fill="url(#grad%d)" rx="6" filter="url(#barShadow)" opacity="0.95"/>
`, barX, yPos-12, barWidth, barHeight, colorIndex))
		}

		// Usage count (right side of bar)
		countText := fmt.Sprintf("%d files", item.count)
		textX := barX + barMaxWidth + 10
		svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="12" fill="%s">%s</text>
`, textX, yPos+5, DefaultTextColor, countText))
	}

	// Footer
	svg.WriteString(SVGFooter)

	return svg.String(), nil
}
