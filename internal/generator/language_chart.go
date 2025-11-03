package generator

import (
	"fmt"
	"strings"

	"github.com/watsumi/update-gh-profile/internal/aggregator"
)

// GenerateLanguageChart 言語ランキングデータから SVG グラフを生成する
//
// Preconditions:
// - rankedLanguages がランキング済み言語スライスであること
// - maxItems が正の整数であること
//
// Postconditions:
// - 返される文字列は有効な SVG 形式である
// - SVG には言語名、使用量、パーセンテージが含まれる
// - 上位 maxItems 件の言語のみが表示される
//
// Invariants:
// - SVG は適切なサイズとスタイリングを持つ
// - テキストは読みやすく表示される
func GenerateLanguageChart(rankedLanguages []aggregator.LanguageStat, maxItems int) (string, error) {
	if maxItems <= 0 {
		maxItems = MaxLanguageItems
	}

	if len(rankedLanguages) == 0 {
		return generateEmptyChart("Language Ranking", "No data available"), nil
	}

	// 表示する言語数を決定
	displayCount := maxItems
	if len(rankedLanguages) < maxItems {
		displayCount = len(rankedLanguages)
	}

	// SVG の高さを動的に調整（1項目あたり30ピクセル + 余白）
	itemHeight := 30
	padding := 20
	titleHeight := 40
	chartHeight := titleHeight + (displayCount * itemHeight) + padding
	width := DefaultSVGWidth

	// SVG を構築
	var svg strings.Builder

	// ヘッダー
	svg.WriteString(fmt.Sprintf(SVGHeader, width, chartHeight, width, chartHeight))

	// スタイル定義（モダンなグラデーションとシャドウ）
	svg.WriteString(`  <defs>
    <linearGradient id="grad" x1="0%" y1="0%" x2="100%" y2="0%">
      <stop offset="0%" style="stop-color:#58a6ff;stop-opacity:1" />
      <stop offset="50%" style="stop-color:#7c3aed;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#1f6feb;stop-opacity:1" />
    </linearGradient>
    <filter id="shadow">
      <feGaussianBlur in="SourceAlpha" stdDeviation="3"/>
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

	// 背景（ボーダー付き）
	svg.WriteString(fmt.Sprintf(`  <rect width="%d" height="%d" fill="%s" rx="10" stroke="#30363d" stroke-width="1"/>
`, width, chartHeight, DefaultBackgroundColor))

	// タイトル（装飾付き）
	svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="20" font-weight="700" fill="%s" text-anchor="middle" filter="url(#shadow)">🗂️ Language Ranking</text>
`, width/2, 32, AccentColor))

	// ランキング項目を表示
	yPos := titleHeight + 10
	maxPercentage := rankedLanguages[0].Percentage // 最大パーセンテージ（バーの幅を計算するため）

	for i := 0; i < displayCount; i++ {
		lang := rankedLanguages[i]
		barWidth := int(float64(width-80) * lang.Percentage / maxPercentage)

		// ランキング番号
		svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="14" fill="%s" font-weight="600">%d.</text>
`, 15, yPos, DefaultTextColor, i+1))

		// 言語名
		svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="14" fill="%s">%s</text>
`, 40, yPos, DefaultTextColor, escapeXML(lang.Language)))

		// バーグラフ
		barX := 140
		barY := yPos - 12
		barHeight := 18

		// バーの背景（グロー効果付き）
		svg.WriteString(fmt.Sprintf(`  <rect x="%d" y="%d" width="%d" height="%d" fill="#161b22" rx="6" stroke="#30363d" stroke-width="1"/>
`, barX, barY, width-150, barHeight))

		// バー（グラデーション + グロー効果）
		if barWidth > 0 {
			svg.WriteString(fmt.Sprintf(`  <rect x="%d" y="%d" width="%d" height="%d" fill="url(#grad)" rx="6" opacity="0.9" filter="url(#shadow)"/>
`, barX, barY, barWidth, barHeight))
		}

		// パーセンテージ（バーの右側）
		percentageText := fmt.Sprintf("%.1f%%", lang.Percentage)
		textX := barX + (width - 150) + 10
		svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="12" fill="%s">%s</text>
`, textX, yPos, DefaultTextColor, percentageText))

		yPos += itemHeight
	}

	// フッター
	svg.WriteString(SVGFooter)

	return svg.String(), nil
}

// generateEmptyChart 空のデータ用のチャートを生成する
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

// escapeXML XML特殊文字をエスケープする
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
