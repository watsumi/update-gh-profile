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

	// スタイル定義
	svg.WriteString(`  <defs>
    <linearGradient id="areaGrad" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" style="stop-color:#58a6ff;stop-opacity:0.3" />
      <stop offset="100%" style="stop-color:#58a6ff;stop-opacity:0" />
    </linearGradient>
  </defs>

`)

	// 背景
	svg.WriteString(fmt.Sprintf(`  <rect width="%d" height="%d" fill="%s" rx="8"/>
`, width, height, DefaultBackgroundColor))

	// タイトル
	svg.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="Segoe UI, system-ui, -apple-system, sans-serif" font-size="18" font-weight="600" fill="%s" text-anchor="middle">Commit History</text>
`, width/2, 30, DefaultTextColor))

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

	// データポイントの位置を計算
	type Point struct {
		X     float64
		Y     float64
		Date  string
		Count int
	}
	points := make([]Point, len(sortedPairs))
	xStep := float64(chartWidth) / float64(len(sortedPairs)-1)

	for i, pair := range sortedPairs {
		x := float64(padding) + float64(i)*xStep
		// Y座標は下から上（コミット数が多いほど上）
		yRatio := float64(pair.Count) / float64(maxValue)
		y := float64(padding+chartHeight) - (float64(chartHeight) * yRatio)

		points[i] = Point{
			X:     x,
			Y:     y,
			Date:  pair.Date,
			Count: pair.Count,
		}
	}

	// エリアグラフのパス（グラデーション塗りつぶし）
	var areaPath strings.Builder
	areaPath.WriteString(fmt.Sprintf("M %d %d ", padding, padding+chartHeight))
	for i, p := range points {
		if i == 0 {
			areaPath.WriteString(fmt.Sprintf("L %.1f %.1f ", p.X, p.Y))
		} else {
			areaPath.WriteString(fmt.Sprintf("L %.1f %.1f ", p.X, p.Y))
		}
	}
	areaPath.WriteString(fmt.Sprintf("L %d %d Z", width-padding, padding+chartHeight))

	svg.WriteString(fmt.Sprintf(`  <path d="%s" fill="url(#areaGrad)" opacity="0.5"/>
`, areaPath.String()))

	// 折れ線グラフのパス
	var linePath strings.Builder
	for i, p := range points {
		if i == 0 {
			linePath.WriteString(fmt.Sprintf("M %.1f %.1f ", p.X, p.Y))
		} else {
			linePath.WriteString(fmt.Sprintf("L %.1f %.1f ", p.X, p.Y))
		}
	}

	svg.WriteString(fmt.Sprintf(`  <path d="%s" fill="none" stroke="#58a6ff" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
`, linePath.String()))

	// データポイント（円）
	for _, p := range points {
		svg.WriteString(fmt.Sprintf(`  <circle cx="%.1f" cy="%.1f" r="3" fill="#58a6ff" stroke="%s" stroke-width="1"/>
`, p.X, p.Y, DefaultBackgroundColor))
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
