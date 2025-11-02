package generator

import (
	"fmt"
	"os"
	"path/filepath"
)

// SaveSVG SVG コンテンツをファイルに保存する
//
// Preconditions:
// - svgContent が有効な SVG 文字列であること
// - filePath が有効なファイルパスであること
//
// Postconditions:
// - 指定されたパスに SVG ファイルが作成される
// - ファイルのエンコーディングは UTF-8 である
//
// Invariants:
// - ディレクトリが存在しない場合は自動作成される
// - 既存のファイルがある場合は上書きされる
func SaveSVG(svgContent, filePath string) error {
	if svgContent == "" {
		return fmt.Errorf("SVG コンテンツが空です")
	}

	if filePath == "" {
		return fmt.Errorf("ファイルパスが空です")
	}

	// ディレクトリが存在しない場合は作成
	dir := filepath.Dir(filePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
		}
	}

	// SVG ファイルを UTF-8 エンコーディングで保存
	// os.WriteFile は UTF-8 で書き込むため、明示的なエンコーディング指定は不要
	err := os.WriteFile(filePath, []byte(svgContent), 0644)
	if err != nil {
		return fmt.Errorf("SVG ファイルの保存に失敗しました: %w", err)
	}

	return nil
}

// SaveMultipleSVGs 複数の SVG を一度に保存する
//
// Preconditions:
// - svgs が map[ファイル名]SVGコンテンツ の形式であること
// - outputDir が有効なディレクトリパスであること
//
// Postconditions:
// - 各 SVG ファイルが outputDir に保存される
// - すべてのファイルが正常に保存される、またはエラーが返される
//
// Invariants:
// - すべてのファイルが同じディレクトリに保存される
func SaveMultipleSVGs(svgs map[string]string, outputDir string) error {
	if len(svgs) == 0 {
		return fmt.Errorf("保存する SVG ファイルがありません")
	}

	if outputDir == "" {
		outputDir = "."
	}

	// ディレクトリが存在しない場合は作成
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
	}

	// 各 SVG を保存
	for filename, content := range svgs {
		// .svg 拡張子がない場合は追加
		if filepath.Ext(filename) != ".svg" {
			filename = filename + ".svg"
		}

		filepath := filepath.Join(outputDir, filename)

		if err := SaveSVG(content, filepath); err != nil {
			return fmt.Errorf("ファイル %s の保存に失敗しました: %w", filename, err)
		}
	}

	return nil
}
