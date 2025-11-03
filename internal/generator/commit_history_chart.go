package generator

import (
	"fmt"
	"strings"

	"github.com/watsumi/update-gh-profile/internal/aggregator"
)

// GenerateCommitHistoryChart コミット合計の推移を表示する SVG グラフを生成する
//
// Preconditions:
// - commitHistory が map[string]int{日付: コミット数} の形式であること
//
// Postconditions:
// - 返される文字列は有効な SVG 形式である
// - SVG には日付ごとのコミット数の推移が表示される
//
// Invariants:
// - SVG は適切なサイズとスタイリングを持つ
func GenerateCommitHistoryChart(commitHistory map[string]int) (string, error) {
	if len(commitHistory) == 0 {
		return generateEmptyChart("Commit History", "No data available"), nil
	}

	// 日付順でソート
	sortedPairs := aggregator.SortCommitHistoryByDate(commitHistory)

	if len(sortedPairs) == 0 {
		return generateEmptyChart("Commit History", "No data available"), nil
	}

	// SVG のサイズを設定
	width := DefaultSVGWidth
	height := DefaultSVGHeight
	padding := 60
	chartWidth := width - padding*2
	chartHeight := height - padding*2

	// 最大コミット数を取得（Y軸のスケール計算用）
	maxCommits := 0
	for _, pair := range sortedPairs {
		if pair.Count > maxCommits {
			maxCommits = pair.Count
		}
	}

	// 最大値を切り上げ（見やすくするため）
	maxValue := maxCommits
	if maxValue == 0 {
		maxValue = 1 // 0除算を防ぐ
	}
	// 最大値を10の倍数に切り上げ
	if maxValue%10 != 0 {
		maxValue = ((maxValue / 10) + 1) * 10
	}

	// SVG を構築
	var svg strings.Builder

	// ヘッダー
	svg.WriteString(fmt.Sprintf(SVGHeader, width, height, width, height))

	// スタイル定義（棒グラフ用のグラデーション）
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

	// 背景（ボーダー付き）
	svg.WriteString(fmt.Sprintf(`  <rect width="%d" height="%d" fill="%s" rx="10" stroke="#30363d" stroke-width="1"/>
`, width, height, DefaultBackgroundColor))

	// タイトル（装飾付き）
	svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="20" font-weight="700" fill="%s" text-anchor="middle">📈 Commit History</text>
`, width/2, 32, AccentColor))

	// Y軸のグリッド線とラベル
	gridLines := 5
	for i := 0; i <= gridLines; i++ {
		y := padding + (chartHeight * i / gridLines)
		value := maxValue - (maxValue * i / gridLines)

		// グリッド線
		if i < gridLines {
			svg.WriteString(fmt.Sprintf(`  <line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#21262d" stroke-width="1"/>
`, padding, y, width-padding, y))
		}

		// Y軸ラベル
		svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="11" fill="%s" text-anchor="end">%d</text>
`, padding-10, y+4, DefaultTextColor, value))
	}

	// 棒グラフの配置を計算
	barSpacing := float64(chartWidth) / float64(len(sortedPairs))
	barWidth := barSpacing * 0.6 // 棒の幅（60%にして間隔を確保）

	// 各データポイントを棒グラフとして描画
	for i, pair := range sortedPairs {
		// 棒の中心位置を計算
		barCenterX := float64(padding) + float64(i)*barSpacing + barSpacing/2
		barX := barCenterX - barWidth/2

		// Y座標は下から上（コミット数が多いほど上）
		yRatio := float64(pair.Count) / float64(maxValue)
		barY := float64(padding+chartHeight) - (float64(chartHeight) * yRatio)
		barHeight := float64(padding+chartHeight) - barY

		// 棒のグラデーション（グラデーション効果）
		svg.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="url(#barGrad)" rx="4" filter="url(#barGlow)" opacity="0.9"/>
`, barX, barY, barWidth, barHeight))

		// 棒のハイライト（上部に明るい線を追加）
		svg.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="%.1f" height="3" fill="#79c0ff" rx="1" opacity="0.6"/>
`, barX, barY, barWidth))
	}

	// X軸の日付ラベルのためにポイント情報を保持
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

	// X軸の日付ラベル（一定間隔で表示）
	labelInterval := len(sortedPairs) / 6 // 最大6つのラベル
	if labelInterval < 1 {
		labelInterval = 1
	}

	for i := 0; i < len(sortedPairs); i += labelInterval {
		if i < len(points) {
			p := points[i]
			// 日付フォーマット（YYYY-MM-DD → MM/DD）
			dateParts := strings.Split(p.Date, "-")
			dateLabel := dateParts[1] + "/" + dateParts[2]

			svg.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="10" fill="%s" text-anchor="middle">%s</text>
`, p.X, height-padding+20, DefaultTextColor, dateLabel))
		}
	}

	// フッター
	svg.WriteString(SVGFooter)

	return svg.String(), nil
}
