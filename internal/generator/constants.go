package generator

// Constants for SVG generation
const (
	// DefaultSVGWidth Default SVG width in pixels
	DefaultSVGWidth = 495

	// DefaultSVGHeight Default SVG height in pixels
	DefaultSVGHeight = 320

	// MaxLanguageItems Maximum number of items to display in language ranking
	MaxLanguageItems = 10

	// SVGHeader SVG header
	SVGHeader = `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">
`

	// SVGFooter SVG footer
	SVGFooter = `</svg>`

	// DefaultBackgroundColor Default background color
	DefaultBackgroundColor = "#0d1117"

	// DefaultTextColor Default text color
	DefaultTextColor = "#c9d1d9"

	// AccentColor Main accent color
	AccentColor = "#58a6ff"

	// AccentColorDark Dark accent color
	AccentColorDark = "#1f6feb"

	// SecondaryColor Secondary color
	SecondaryColor = "#7c3aed"

	// SuccessColor Success color
	SuccessColor = "#56d364"

	// WarningColor Warning color
	WarningColor = "#f85149"
)
