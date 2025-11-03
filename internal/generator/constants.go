package generator

// SVG 生成関連の定数
const (
	// DefaultSVGWidth デフォルトの SVG 幅（ピクセル）
	DefaultSVGWidth = 495

	// DefaultSVGHeight デフォルトの SVG 高さ（ピクセル）
	DefaultSVGHeight = 320

	// MaxLanguageItems 言語ランキングで表示する最大項目数
	MaxLanguageItems = 10

	// SVGHeader SVG ヘッダー
	SVGHeader = `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">
`

	// SVGFooter SVG フッター
	SVGFooter = `</svg>`

	// DefaultBackgroundColor デフォルトの背景色
	DefaultBackgroundColor = "#0d1117"

	// DefaultTextColor デフォルトのテキスト色
	DefaultTextColor = "#c9d1d9"

	// AccentColor アクセントカラー（メイン）
	AccentColor = "#58a6ff"

	// AccentColorDark アクセントカラー（ダーク）
	AccentColorDark = "#1f6feb"

	// SecondaryColor セカンダリカラー
	SecondaryColor = "#7c3aed"

	// SuccessColor 成功色
	SuccessColor = "#56d364"

	// WarningColor 警告色
	WarningColor = "#f85149"
)
