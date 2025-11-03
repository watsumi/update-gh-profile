package generator

import (
	"fmt"
	"strings"

	"github.com/watsumi/update-gh-profile/internal/aggregator"
)

// GenerateCommitTimeChart generates an SVG chart showing commit time distribution
//
// Preconditions:
// - timeDistribution is in the format map[int]int{time slot: commit count}
//
// Postconditions:
// - Returns a valid SVG string
// - SVG displays commit count per time slot
//
// Invariants:
// - All 24 hours are displayed (time slots with no data are shown as 0)
func GenerateCommitTimeChart(timeDistribution map[int]int) (string, error) {
	if len(timeDistribution) == 0 {
		return generateEmptyChart("Commit Time Distribution", "No data available"), nil
	}

	// Sort by time slot
	sortedPairs := aggregator.SortCommitTimeDistributionByHour(timeDistribution)

	// Ensure data for all 24 hours (time slots with no data are 0)
	hourlyData := make(map[int]int)
	for i := 0; i < 24; i++ {
		hourlyData[i] = 0
	}
	for _, pair := range sortedPairs {
		if pair.Hour >= 0 && pair.Hour <= 23 {
			hourlyData[pair.Hour] = pair.Count
		}
	}

	// Get maximum commit count (to determine color intensity)
	maxCommits := 0
	for _, count := range hourlyData {
		if count > maxCommits {
			maxCommits = count
		}
	}
	if maxCommits == 0 {
		maxCommits = 1 // Prevent division by zero
	}

	// Set SVG size
	width := DefaultSVGWidth
	height := 200
	padding := 20
	chartWidth := width - padding*2
	chartHeight := height - padding*2 - 60 // Space for title and labels

	// Build SVG
	var svg strings.Builder

	// Header
	svg.WriteString(fmt.Sprintf(SVGHeader, width, height, width, height))

	// Style definitions
	svg.WriteString(`  <defs>
    <linearGradient id="timeGrad" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" style="stop-color:#58a6ff;stop-opacity:1" />
      <stop offset="50%" style="stop-color:#7c3aed;stop-opacity:1" />
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
	svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="20" font-weight="700" fill="%s" text-anchor="middle">🕐 Commit Time Distribution (UTC)</text>
`, width/2, 37, AccentColor))

	// Display in heatmap format
	barWidth := float64(chartWidth) / 24.0
	barHeight := float64(chartHeight)
	startX := float64(padding)

	for hour := 0; hour < 24; hour++ {
		count := hourlyData[hour]
		x := startX + float64(hour)*barWidth

		// Determine color intensity based on commit count ratio
		intensity := float64(count) / float64(maxCommits)
		if intensity > 1.0 {
			intensity = 1.0
		}

		// Calculate color (higher commit count = darker color)
		baseColor := "#58a6ff"
		if intensity > 0.8 {
			baseColor = "#1f6feb" // Darkest
		} else if intensity > 0.6 {
			baseColor = "#388bfd" // Dark
		} else if intensity > 0.4 {
			baseColor = "#58a6ff" // Medium
		} else if intensity > 0.2 {
			baseColor = "#79c0ff" // Light
		} else if intensity > 0 {
			baseColor = "#b1ddff" // Lightest
		} else {
			baseColor = "#21262d" // No data
		}

		// Draw bar
		barHeightScaled := barHeight * intensity
		if barHeightScaled < 5 && count > 0 {
			barHeightScaled = 5 // Ensure minimum height
		}

		y := float64(height-padding) - barHeightScaled

		svg.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" rx="3" filter="url(#barGlow)" opacity="0.9"/>
`, x+1, y, barWidth-2, barHeightScaled, baseColor))

		// Time slot label (every 6 hours)
		if hour%6 == 0 {
			svg.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="10" fill="%s" text-anchor="middle">%02d:00</text>
`, x+barWidth/2, height-padding+15, DefaultTextColor, hour))
		}

		// Display count if greater than 0 (small text)
		if count > 0 {
			textY := y - 3
			if textY < float64(padding+20) {
				textY = y + 12 // Display below if there's no space above the bar
			}
			svg.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%.1f" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="8" fill="%s" text-anchor="middle" opacity="0.8">%d</text>
`, x+barWidth/2, textY, DefaultTextColor, count))
		}
	}

	// Legend (color explanation sorted by commit count)
	legendY := height - padding - chartHeight - 25
	svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="11" fill="%s">High</text>
`, padding, legendY, DefaultTextColor))

	// Display color bar
	for i := 0; i < 5; i++ {
		color := []string{"#1f6feb", "#388bfd", "#58a6ff", "#79c0ff", "#b1ddff"}[i]
		x := padding + 40 + (i * 25)
		svg.WriteString(fmt.Sprintf(`  <rect x="%d" y="%d" width="20" height="8" fill="%s" rx="1"/>
`, x, legendY-8, color))
	}

	svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="11" fill="%s">Low</text>
`, padding+40+125, legendY, DefaultTextColor))

	// Footer
	svg.WriteString(SVGFooter)

	return svg.String(), nil
}
