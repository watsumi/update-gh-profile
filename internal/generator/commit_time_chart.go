package generator

import (
	"fmt"
	"strings"

	"github.com/watsumi/update-gh-profile/internal/aggregator"
)

// GenerateCommitTimeChart コミットが多い時間帯を表示する SVG グラフを生成する
//
// Preconditions:
// - timeDistribution が map[int]int{時間帯: コミット数} の形式であること
//
// Postconditions:
// - 返される文字列は有効な SVG 形式である
// - SVG には時間帯ごとのコミット数が表示される
//
// Invariants:
// - 24時間すべての時間帯が表示される（データがない時間帯は0として表示）
func GenerateCommitTimeChart(timeDistribution map[int]int) (string, error) {
	if len(timeDistribution) == 0 {
		return generateEmptyChart("Commit Time Distribution", "No data available"), nil
	}

	// 時間帯順でソート
	sortedPairs := aggregator.SortCommitTimeDistributionByHour(timeDistribution)

	// 24時間すべてのデータを確保（データがない時間帯は0）
	hourlyData := make(map[int]int)
	for i := 0; i < 24; i++ {
		hourlyData[i] = 0
	}
	for _, pair := range sortedPairs {
		if pair.Hour >= 0 && pair.Hour <= 23 {
			hourlyData[pair.Hour] = pair.Count
		}
	}

	// 最大コミット数を取得（色の濃さを決定するため）
	maxCommits := 0
	for _, count := range hourlyData {
		if count > maxCommits {
			maxCommits = count
		}
	}
	if maxCommits == 0 {
		maxCommits = 1 // 0除算を防ぐ
	}

	// SVG のサイズを設定
	width := DefaultSVGWidth
	height := 200
	padding := 20
	chartWidth := width - padding*2
	chartHeight := height - padding*2 - 60 // タイトルとラベルのスペース

	// SVG を構築
	var svg strings.Builder

	// ヘッダー
	svg.WriteString(fmt.Sprintf(SVGHeader, width, height, width, height))

	// スタイル定義
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

	// 背景（ボーダー付き）
	svg.WriteString(fmt.Sprintf(`  <rect width="%d" height="%d" fill="%s" rx="10" stroke="#30363d" stroke-width="1"/>
`, width, height, DefaultBackgroundColor))

	// タイトル（装飾付き）
	svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="20" font-weight="700" fill="%s" text-anchor="middle">🕐 Commit Time Distribution (UTC)</text>
`, width/2, 37, AccentColor))

	// ヒートマップ形式で表示
	barWidth := float64(chartWidth) / 24.0
	barHeight := float64(chartHeight)
	startX := float64(padding)

	for hour := 0; hour < 24; hour++ {
		count := hourlyData[hour]
		x := startX + float64(hour)*barWidth

		// コミット数の比率に基づいて色の濃さを決定
		intensity := float64(count) / float64(maxCommits)
		if intensity > 1.0 {
			intensity = 1.0
		}

		// 色を計算（コミット数が多いほど濃い）
		baseColor := "#58a6ff"
		if intensity > 0.8 {
			baseColor = "#1f6feb" // 最も濃い
		} else if intensity > 0.6 {
			baseColor = "#388bfd" // 濃い
		} else if intensity > 0.4 {
			baseColor = "#58a6ff" // 中程度
		} else if intensity > 0.2 {
			baseColor = "#79c0ff" // 薄い
		} else if intensity > 0 {
			baseColor = "#b1ddff" // 最も薄い
		} else {
			baseColor = "#21262d" // データなし
		}

		// バーを描画
		barHeightScaled := barHeight * intensity
		if barHeightScaled < 5 && count > 0 {
			barHeightScaled = 5 // 最小の高さを確保
		}

		y := float64(height-padding) - barHeightScaled

		svg.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" rx="3" filter="url(#barGlow)" opacity="0.9"/>
`, x+1, y, barWidth-2, barHeightScaled, baseColor))

		// 時間帯ラベル（6時間ごと）
		if hour%6 == 0 {
			svg.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="10" fill="%s" text-anchor="middle">%02d:00</text>
`, x+barWidth/2, height-padding+15, DefaultTextColor, hour))
		}

		// コミット数が0より大きい場合は数値を表示（小さなテキスト）
		if count > 0 {
			textY := y - 3
			if textY < float64(padding+20) {
				textY = y + 12 // バーの上に表示するスペースがない場合は下に
			}
			svg.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%.1f" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="8" fill="%s" text-anchor="middle" opacity="0.8">%d</text>
`, x+barWidth/2, textY, DefaultTextColor, count))
		}
	}

	// 凡例（コミット数が多い順に色の説明）
	legendY := height - padding - chartHeight - 25
	svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="11" fill="%s">High</text>
`, padding, legendY, DefaultTextColor))

	// カラーバーを表示
	for i := 0; i < 5; i++ {
		color := []string{"#1f6feb", "#388bfd", "#58a6ff", "#79c0ff", "#b1ddff"}[i]
		x := padding + 40 + (i * 25)
		svg.WriteString(fmt.Sprintf(`  <rect x="%d" y="%d" width="20" height="8" fill="%s" rx="1"/>
`, x, legendY-8, color))
	}

	svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="11" fill="%s">Low</text>
`, padding+40+125, legendY, DefaultTextColor))

	// フッター
	svg.WriteString(SVGFooter)

	return svg.String(), nil
}
