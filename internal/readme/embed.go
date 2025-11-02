package readme

import (
	"fmt"
	"path/filepath"
	"strings"
)

// EmbedSVGs README.md の指定セクションに SVG グラフを埋め込む
//
// Preconditions:
// - readmePath が有効な README.md ファイルパスであること
// - svgFiles が SVG ファイルパスのスライスであること
// - sectionTag が更新するセクションのタグ名であること（例: "LANGUAGE_STATS"）
//
// Postconditions:
// - README.md の指定セクションが SVG グラフの Markdown 記法で更新される
// - SVG ファイルは相対パスで埋め込まれる
//
// Invariants:
// - グラフの順序は svgFiles の順序に従う
// - Markdown 記法は適切に使用される
func EmbedSVGs(readmePath string, svgFiles []string, sectionTag string) error {
	if len(svgFiles) == 0 {
		return fmt.Errorf("埋め込む SVG ファイルがありません")
	}

	// セクションタグを正規化
	startTag, endTag := NormalizeTags(sectionTag)

	// SVG グラフの Markdown 記法を生成
	markdown := GenerateSVGMarkdown(svgFiles)

	// セクションを更新
	err := UpdateSection(readmePath, startTag, endTag, markdown)
	if err != nil {
		return fmt.Errorf("SVG グラフの埋め込みに失敗しました: %w", err)
	}

	return nil
}

// GenerateSVGMarkdown 複数の SVG ファイルから Markdown 記法の画像埋め込みコードを生成する
//
// Preconditions:
// - svgFiles が SVG ファイルパスのスライスであること
//
// Postconditions:
// - Markdown 記法の画像埋め込みコードが返される
// - 各 SVG は改行で区切られる
//
// Invariants:
// - 相対パスが使用される
// - Markdown 記法は標準的な形式である
func GenerateSVGMarkdown(svgFiles []string) string {
	var markdown strings.Builder

	for i, svgFile := range svgFiles {
		// ファイル名を取得（パスの最後の部分）
		fileName := filepath.Base(svgFile)

		// 画像の説明文を生成（ファイル名から拡張子を除く）
		description := strings.TrimSuffix(fileName, ".svg")
		description = strings.ReplaceAll(description, "_", " ")
		description = strings.Title(description) // 各単語の最初を大文字に

		// Markdown 記法の画像埋め込みを生成
		// 形式: ![説明](相対パス)
		markdown.WriteString(fmt.Sprintf("![%s](%s)", description, fileName))

		// 最後のファイル以外は改行を追加
		if i < len(svgFiles)-1 {
			markdown.WriteString("\n\n")
		}
	}

	return markdown.String()
}

// EmbedSVGWithCustomPath カスタムパスを使用して SVG グラフを埋め込む
//
// Preconditions:
// - readmePath が有効な README.md ファイルパスであること
// - svgFilePath が SVG ファイルパスであること
// - sectionTag が更新するセクションのタグ名であること
// - description が画像の説明文であること（省略可能）
//
// Postconditions:
// - README.md の指定セクションが SVG グラフの Markdown 記法で更新される
//
// Invariants:
// - パスは相対パスまたは絶対パスが使用可能
func EmbedSVGWithCustomPath(readmePath, svgFilePath, sectionTag, description string) error {
	// セクションタグを正規化
	startTag, endTag := NormalizeTags(sectionTag)

	// 説明文がない場合はファイル名から生成
	if description == "" {
		fileName := filepath.Base(svgFilePath)
		description = strings.TrimSuffix(fileName, ".svg")
		description = strings.ReplaceAll(description, "_", " ")
		description = strings.Title(description)
	}

	// Markdown 記法を生成
	markdown := fmt.Sprintf("![%s](%s)", description, svgFilePath)

	// セクションを更新
	err := UpdateSection(readmePath, startTag, endTag, markdown)
	if err != nil {
		return fmt.Errorf("SVG グラフの埋め込みに失敗しました: %w", err)
	}

	return nil
}

// EmbedMultipleSVGSections 複数の SVG を異なるセクションに埋め込む
//
// Preconditions:
// - readmePath が有効な README.md ファイルパスであること
// - svgSections が map[セクションタグ名]SVGファイルパス の形式であること
//
// Postconditions:
// - 各セクションが対応する SVG グラフで更新される
// - すべてのセクションが正常に更新される、またはエラーが返される
//
// Invariants:
// - セクションの順序は保証されない（map の特性）
func EmbedMultipleSVGSections(readmePath string, svgSections map[string]string) error {
	if len(svgSections) == 0 {
		return fmt.Errorf("埋め込む SVG セクションがありません")
	}

	// 各セクションを更新
	for sectionTag, svgFile := range svgSections {
		err := EmbedSVGWithCustomPath(readmePath, svgFile, sectionTag, "")
		if err != nil {
			return fmt.Errorf("セクション %s の更新に失敗しました: %w", sectionTag, err)
		}
	}

	return nil
}
