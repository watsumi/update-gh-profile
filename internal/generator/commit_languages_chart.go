package generator

import (
	"fmt"
	"sort"
	"strings"
)

// GenerateCommitLanguagesChart コミットごとの使用言語Top5を表示する SVG を生成する
//
// Preconditions:
// - commitLanguages が map[string]int{言語名: 使用回数} の形式であること
//
// Postconditions:
// - 返される文字列は有効な SVG 形式である
// - SVG にはTop5の言語とその使用回数が表示される
//
// Invariants:
// - 上位5つの言語のみが表示される
func GenerateCommitLanguagesChart(commitLanguages map[string]int) (string, error) {
	if len(commitLanguages) == 0 {
		return generateEmptyChart("Top 5 Languages by Commit", "No data available"), nil
	}

	// 使用回数でソートしてTop5を抽出
	type langCount struct {
		lang  string
		count int
	}
	var langList []langCount
	for lang, count := range commitLanguages {
		langList = append(langList, langCount{lang: lang, count: count})
	}

	// 使用回数降順でソート
	sort.Slice(langList, func(i, j int) bool {
		return langList[i].count > langList[j].count
	})

	// Top5を取得
	maxItems := 5
	if len(langList) > maxItems {
		langList = langList[:maxItems]
	}

	if len(langList) == 0 {
		return generateEmptyChart("Top 5 Languages by Commit", "No data available"), nil
	}

	// SVG のサイズを設定
	width := DefaultSVGWidth
	height := 280
	padding := 20
	chartWidth := width - padding*2

	// 最大使用回数を取得
	maxCount := langList[0].count
	if maxCount == 0 {
		maxCount = 1 // 0除算を防ぐ
	}

	// SVG を構築
	var svg strings.Builder

	// ヘッダー
	svg.WriteString(fmt.Sprintf(SVGHeader, width, height, width, height))

	// スタイル定義（カラーパレット）
	colors := []string{"#58a6ff", "#1f6feb", "#6e40c9", "#ff7b72", "#f85149"}
	svg.WriteString(`  <defs>
`)

	for i, color := range colors {
		svg.WriteString(fmt.Sprintf(`    <linearGradient id="grad%d" x1="0%%" y1="0%%" x2="0%%" y2="100%%">
      <stop offset="0%%" style="stop-color:%s;stop-opacity:1" />
      <stop offset="100%%" style="stop-color:%s;stop-opacity:0.7" />
    </linearGradient>
`, i, color, color))
	}

	svg.WriteString(`  </defs>

`)

	// 背景
	svg.WriteString(fmt.Sprintf(`  <rect width="%d" height="%d" fill="%s" rx="8"/>
`, width, height, DefaultBackgroundColor))

	// タイトル
	svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="18" font-weight="600" fill="%s" text-anchor="middle">Top 5 Languages by Commit</text>
`, width/2, 35, DefaultTextColor))

	// 棒グラフを表示
	barHeight := 30
	barSpacing := 45
	startY := 70
	barMaxWidth := chartWidth - 200 // 言語名と数値のスペースを確保

	for i, item := range langList {
		yPos := startY + (i * barSpacing)

		// バーの幅を計算
		barWidth := int(float64(barMaxWidth) * float64(item.count) / float64(maxCount))

		// 言語名
		svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="14" fill="%s">%s</text>
`, padding, yPos+5, DefaultTextColor, escapeXML(item.lang)))

		// バーの背景
		barX := 140
		svg.WriteString(fmt.Sprintf(`  <rect x="%d" y="%d" width="%d" height="%d" fill="#21262d" rx="4"/>
`, barX, yPos-12, barMaxWidth, barHeight))

		// バー（グラデーション）
		if barWidth > 0 {
			colorIndex := i % len(colors)
			svg.WriteString(fmt.Sprintf(`  <rect x="%d" y="%d" width="%d" height="%d" fill="url(#grad%d)" rx="4"/>
`, barX, yPos-12, barWidth, barHeight, colorIndex))
		}

		// 使用回数（バーの右側）
		countText := fmt.Sprintf("%d files", item.count)
		textX := barX + barMaxWidth + 10
		svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="12" fill="%s">%s</text>
`, textX, yPos+5, DefaultTextColor, countText))
	}

	// フッター
	svg.WriteString(SVGFooter)

	return svg.String(), nil
}
