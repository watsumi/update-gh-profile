package generator

import (
	"fmt"
	"strings"

	"github.com/watsumi/update-gh-profile/internal/aggregator"
)

// GenerateSummaryCard スター数、リポジトリ数、コミット数、PR数を表示するサマリーカードの SVG を生成する
//
// Preconditions:
// - stats が有効な SummaryStats 構造体であること
//
// Postconditions:
// - 返される文字列は有効な SVG 形式である
// - SVG には4つのメトリクス（スター、リポジトリ、コミット、PR）が表示される
//
// Invariants:
// - すべてのメトリクスがカード形式で表示される
// - アイコンと数値が適切に配置される
func GenerateSummaryCard(stats aggregator.SummaryStats) (string, error) {
	// SVG のサイズを設定
	width := DefaultSVGWidth
	height := 140
	padding := 20
	cardSpacing := 15
	cardWidth := (width - padding*2 - cardSpacing*3) / 4 // 4つのカードを配置

	// SVG を構築
	var svg strings.Builder

	// ヘッダー
	svg.WriteString(fmt.Sprintf(SVGHeader, width, height, width, height))

	// スタイル定義
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

	// 各カード用のグラデーション定義
	for i := 0; i < 4; i++ {
		svg.WriteString(fmt.Sprintf(`    <linearGradient id="cardGrad%d" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
      <stop offset="0%%" style="stop-color:%s;stop-opacity:0.15" />
      <stop offset="100%%" style="stop-color:%s;stop-opacity:0.05" />
    </linearGradient>
`, i, []string{"#ffd700", "#58a6ff", "#56d364", "#a371f7"}[i], []string{"#ffd700", "#58a6ff", "#56d364", "#a371f7"}[i]))
	}

	svg.WriteString(`  </defs>

`)

	// 背景（グラデーション + ボーダー）
	svg.WriteString(fmt.Sprintf(`  <rect width="%d" height="%d" fill="url(#cardGrad)" rx="12" stroke="#30363d" stroke-width="1"/>
`, width, height, DefaultBackgroundColor))

	// タイトル（省略可能、カードだけでも見やすい）
	// svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="18" font-weight="600" fill="%s" text-anchor="middle">統計サマリー</text>
	// `, width/2, 30, DefaultTextColor))

	// メトリクス定義
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

	// 各メトリクスのカードを描画
	startX := padding
	cardY := 40
	iconSize := 32
	iconY := cardY + iconSize - 10
	valueY := cardY + iconSize + 35
	labelY := cardY + iconSize + 55

	for i, m := range metrics {
		cardX := startX + i*(cardWidth+cardSpacing)

		// カードの背景（グラデーション + シャドウ）
		svg.WriteString(fmt.Sprintf(`  <rect x="%d" y="%d" width="%d" height="%d" fill="url(#cardGrad%d)" rx="8" stroke="%s" stroke-width="1.5" opacity="0.8" filter="url(#cardShadow)"/>
`, cardX, cardY, cardWidth, height-cardY-padding, i, m.color))

		// アイコン（大きめ + グロー効果）
		iconX := cardX + cardWidth/2
		svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI Emoji, Apple Color Emoji, sans-serif" font-size="%d" text-anchor="middle" filter="url(#iconShadow)">%s</text>
`, iconX, iconY, iconSize+2, m.icon))

		// 数値（大きなフォント）
		valueText := formatNumber(m.value)
		svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="20" font-weight="600" fill="%s" text-anchor="middle">%s</text>
`, iconX, valueY, DefaultTextColor, valueText))

		// ラベル
		svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="11" fill="%s" text-anchor="middle" opacity="0.7">%s</text>
`, iconX, labelY, DefaultTextColor, m.label))
	}

	// フッター
	svg.WriteString(SVGFooter)

	return svg.String(), nil
}

// formatNumber 数値を3桁区切りの文字列にフォーマットする
// 例: 1234 -> "1,234", 1000000 -> "1M"
func formatNumber(n int) string {
	if n < 0 {
		return "0"
	}

	// 百万単位
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000.0)
	}

	// 千単位
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000.0)
	}

	// 3桁区切り
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}

	// 3桁ごとにカンマを挿入
	result := ""
	for i, r := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(r)
	}

	return result
}
