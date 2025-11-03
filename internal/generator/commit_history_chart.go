package generator

import (
	"fmt"
	"strings"

	"github.com/watsumi/update-gh-profile/internal/aggregator"
)

// GenerateCommitHistoryChart generates an SVG chart showing commit trend over time
//
// Preconditions:
// - commitHistory is in the format map[string]int{date: commit count}
//
// Postconditions:
// - Returns a valid SVG string
// - SVG displays commit count trend per date
//
// Invariants:
// - SVG has appropriate size and styling
func GenerateCommitHistoryChart(commitHistory map[string]int) (string, error) {
	if len(commitHistory) == 0 {
		return generateEmptyChart("Commit History", "No data available"), nil
	}

	// Sort by date
	sortedPairs := aggregator.SortCommitHistoryByDate(commitHistory)

	if len(sortedPairs) == 0 {
		return generateEmptyChart("Commit History", "No data available"), nil
	}

	// Set SVG size
	width := DefaultSVGWidth
	height := DefaultSVGHeight
	padding := 60
	chartWidth := width - padding*2
	chartHeight := height - padding*2

	// Get maximum commit count (for Y-axis scale calculation)
	maxCommits := 0
	for _, pair := range sortedPairs {
		if pair.Count > maxCommits {
			maxCommits = pair.Count
		}
	}

	// Round up maximum value (for better visibility)
	maxValue := maxCommits
	if maxValue == 0 {
		maxValue = 1 // Prevent division by zero
	}
	// Round up to nearest multiple of 10
	if maxValue%10 != 0 {
		maxValue = ((maxValue / 10) + 1) * 10
	}

	// Build SVG
	var svg strings.Builder

	// Header
	svg.WriteString(fmt.Sprintf(SVGHeader, width, height, width, height))

	// Style definitions (gradient for bar chart)
	svg.WriteString(`  <defs>
    <linearGradient id="barGrad" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" style="stop-color:#58a6ff;stop-opacity:1" />
      <stop offset="50%" style="stop-color:#7c3aed;stop-opacity:0.9" />
      <stop offset="100%" style="stop-color:#1f6feb;stop-opacity:0.8" />
    </linearGradient>
    <filter id="barGlow">
      <feGaussianBlur stdDeviation="2" result="coloredBlur"/>
      <feMerge>
        <feMergeNode in="coloredBlur"/>
        <feMergeNode in="SourceGraphic"/>
      </feMerge>
    </filter>
  </defs>

`)

	// Background (with border)
	svg.WriteString(fmt.Sprintf(`  <rect width="%d" height="%d" fill="%s" rx="10" stroke="#30363d" stroke-width="1"/>
`, width, height, DefaultBackgroundColor))

	// Title (decorated)
	svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="20" font-weight="700" fill="%s" text-anchor="middle">📈 Commit History</text>
`, width/2, 32, AccentColor))

	// Y-axis grid lines and labels
	gridLines := 5
	for i := 0; i <= gridLines; i++ {
		y := padding + (chartHeight * i / gridLines)
		value := maxValue - (maxValue * i / gridLines)

		// Grid line
		if i < gridLines {
			svg.WriteString(fmt.Sprintf(`  <line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#21262d" stroke-width="1"/>
`, padding, y, width-padding, y))
		}

		// Y-axis label
		svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="11" fill="%s" text-anchor="end">%d</text>
`, padding-10, y+4, DefaultTextColor, value))
	}

	// Calculate bar chart layout
	barSpacing := float64(chartWidth) / float64(len(sortedPairs))
	barWidth := barSpacing * 0.6 // Bar width (60% to ensure spacing)

	// Draw each data point as a bar
	for i, pair := range sortedPairs {
		// Calculate bar center position
		barCenterX := float64(padding) + float64(i)*barSpacing + barSpacing/2
		barX := barCenterX - barWidth/2

		// Y coordinate from bottom to top (higher commit count = higher)
		yRatio := float64(pair.Count) / float64(maxValue)
		barY := float64(padding+chartHeight) - (float64(chartHeight) * yRatio)
		barHeight := float64(padding+chartHeight) - barY

		// Bar gradient (gradient effect)
		svg.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="url(#barGrad)" rx="4" filter="url(#barGlow)" opacity="0.9"/>
`, barX, barY, barWidth, barHeight))

		// Bar highlight (add bright line at top)
		svg.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="%.1f" height="3" fill="#79c0ff" rx="1" opacity="0.6"/>
`, barX, barY, barWidth))
	}

	// Keep point information for X-axis date labels
	type Point struct {
		X    float64
		Date string
	}
	points := make([]Point, len(sortedPairs))
	for i, pair := range sortedPairs {
		barCenterX := float64(padding) + float64(i)*barSpacing + barSpacing/2
		points[i] = Point{
			X:    barCenterX,
			Date: pair.Date,
		}
	}

	// X-axis date labels (displayed at regular intervals)
	labelInterval := len(sortedPairs) / 6 // Maximum 6 labels
	if labelInterval < 1 {
		labelInterval = 1
	}

	for i := 0; i < len(sortedPairs); i += labelInterval {
		if i < len(points) {
			p := points[i]
			// Date format (YYYY-MM-DD → MM/DD)
			dateParts := strings.Split(p.Date, "-")
			dateLabel := dateParts[1] + "/" + dateParts[2]

			svg.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="10" fill="%s" text-anchor="middle">%s</text>
`, p.X, height-padding+20, DefaultTextColor, dateLabel))
		}
	}

	// Footer
	svg.WriteString(SVGFooter)

	return svg.String(), nil
}
