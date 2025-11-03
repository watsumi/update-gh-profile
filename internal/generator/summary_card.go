package generator

import (
	"fmt"
	"strings"

	"github.com/watsumi/update-gh-profile/internal/aggregator"
)

// GenerateSummaryCard generates an SVG summary card displaying stars, repositories, commits, and PRs
//
// Preconditions:
// - stats is a valid SummaryStats struct
//
// Postconditions:
// - Returns a valid SVG string
// - SVG displays 4 metrics (stars, repositories, commits, PRs)
//
// Invariants:
// - All metrics are displayed in card format
// - Icons and values are properly positioned
func GenerateSummaryCard(stats aggregator.SummaryStats) (string, error) {
	// Set SVG size
	width := DefaultSVGWidth
	height := 140
	padding := 20
	cardSpacing := 15
	cardWidth := (width - padding*2 - cardSpacing*3) / 4 // Arrange 4 cards

	// Build SVG
	var svg strings.Builder

	// Header
	svg.WriteString(fmt.Sprintf(SVGHeader, width, height, width, height))

	// Style definitions
	svg.WriteString(`  <defs>
    <filter id="cardShadow">
      <feGaussianBlur in="SourceAlpha" stdDeviation="4"/>
      <feOffset dx="0" dy="2" result="offsetblur"/>
      <feComponentTransfer>
        <feFuncA type="linear" slope="0.3"/>
      </feComponentTransfer>
      <feMerge>
        <feMergeNode/>
        <feMergeNode in="SourceGraphic"/>
      </feMerge>
    </filter>
    <filter id="iconShadow">
      <feGaussianBlur in="SourceAlpha" stdDeviation="2"/>
      <feOffset dx="0" dy="1" result="offsetblur"/>
      <feComponentTransfer>
        <feFuncA type="linear" slope="0.3"/>
      </feComponentTransfer>
      <feMerge>
        <feMergeNode/>
        <feMergeNode in="SourceGraphic"/>
      </feMerge>
    </filter>
    <linearGradient id="cardGrad" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#161b22;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#0d1117;stop-opacity:1" />
    </linearGradient>
`)

	// Gradient definitions for each card
	for i := 0; i < 4; i++ {
		svg.WriteString(fmt.Sprintf(`    <linearGradient id="cardGrad%d" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
      <stop offset="0%%" style="stop-color:%s;stop-opacity:0.15" />
      <stop offset="100%%" style="stop-color:%s;stop-opacity:0.05" />
    </linearGradient>
`, i, []string{"#ffd700", "#58a6ff", "#56d364", "#a371f7"}[i], []string{"#ffd700", "#58a6ff", "#56d364", "#a371f7"}[i]))
	}

	svg.WriteString(`  </defs>

`)

	// Background (gradient + border)
	svg.WriteString(fmt.Sprintf(`  <rect width="%d" height="%d" fill="url(#cardGrad)" rx="12" stroke="#30363d" stroke-width="1"/>
`, width, height))

	// Title (optional, cards are readable without it)
	// svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="18" font-weight="600" fill="%s" text-anchor="middle">Statistics Summary</text>
	// `, width/2, 30, DefaultTextColor))

	// Metric definitions
	type metric struct {
		label string
		value int
		icon  string
		color string
	}

	metrics := []metric{
		{
			label: "Stars",
			value: stats.TotalStars,
			icon:  "⭐",
			color: "#ffd700",
		},
		{
			label: "Repos",
			value: stats.RepositoryCount,
			icon:  "📦",
			color: "#58a6ff",
		},
		{
			label: "Commits",
			value: stats.TotalCommits,
			icon:  "💾",
			color: "#56d364",
		},
		{
			label: "PRs",
			value: stats.TotalPullRequests,
			icon:  "🔀",
			color: "#a371f7",
		},
	}

	// Draw cards for each metric
	startX := padding
	cardY := 40
	iconSize := 32
	iconY := cardY + iconSize - 10
	valueY := cardY + iconSize + 35
	labelY := cardY + iconSize + 55

	for i, m := range metrics {
		cardX := startX + i*(cardWidth+cardSpacing)

		// Card background (gradient + shadow)
		svg.WriteString(fmt.Sprintf(`  <rect x="%d" y="%d" width="%d" height="%d" fill="url(#cardGrad%d)" rx="8" stroke="%s" stroke-width="1.5" opacity="0.8" filter="url(#cardShadow)"/>
`, cardX, cardY, cardWidth, height-cardY-padding, i, m.color))

		// Icon (large + glow effect)
		iconX := cardX + cardWidth/2
		svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI Emoji, Apple Color Emoji, sans-serif" font-size="%d" text-anchor="middle" filter="url(#iconShadow)">%s</text>
`, iconX, iconY, iconSize+2, m.icon))

		// Value (large font)
		valueText := formatNumber(m.value)
		svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="20" font-weight="600" fill="%s" text-anchor="middle">%s</text>
`, iconX, valueY, DefaultTextColor, valueText))

		// Label
		svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="11" fill="%s" text-anchor="middle" opacity="0.7">%s</text>
`, iconX, labelY, DefaultTextColor, m.label))
	}

	// Footer
	svg.WriteString(SVGFooter)

	return svg.String(), nil
}

// formatNumber formats a number into a string with comma separators
// Examples: 1234 -> "1,234", 1000000 -> "1M"
func formatNumber(n int) string {
	if n < 0 {
		return "0"
	}

	// Millions
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000.0)
	}

	// Thousands
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000.0)
	}

	// 3-digit comma separation
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}

	// Insert comma every 3 digits
	result := ""
	for i, r := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(r)
	}

	return result
}
