package readme

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// UpdateSection README.md の指定されたセクションを新しいコンテンツで置き換える
// タグが存在しない場合は自動的に追加する
//
// Preconditions:
// - readmePath が有効な README.md ファイルパスであること
// - startTag と endTag が有効なコメントタグ文字列であること
// - newContent が置き換えに使用する新しいコンテンツであること
//
// Postconditions:
// - README.md の startTag と endTag の間が newContent で置き換えられる
// - タグが存在しない場合はファイル末尾に追加される
// - 既存のコンテンツ（タグ以外）は保持される
//
// Invariants:
// - ファイルが存在しない場合はエラーが返される
func UpdateSection(readmePath, startTag, endTag, newContent string) error {
	// README ファイルを読み込む
	content, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("README ファイルの読み込みに失敗しました: %w", err)
	}

	readmeContent := string(content)

	// タグ間のコンテンツを置き換え
	updatedContent, err := ReplaceSectionOrAppend(readmeContent, startTag, endTag, newContent)
	if err != nil {
		return fmt.Errorf("セクションの置き換えに失敗しました: %w", err)
	}

	// 更新されたコンテンツをファイルに書き込む
	err = os.WriteFile(readmePath, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("README ファイルの書き込みに失敗しました: %w", err)
	}

	return nil
}

// ReplaceSectionOrAppend テキスト内の指定されたセクション（startTag と endTag の間）を新しいコンテンツで置き換える
// タグが存在しない場合はファイル末尾に自動追加する
//
// Preconditions:
// - content が読み込まれたファイルのコンテンツであること
// - startTag と endTag が有効なコメントタグ文字列であること
//
// Postconditions:
// - startTag と endTag の間のコンテンツが newContent で置き換えられる
// - タグが存在しない場合はファイル末尾に追加される
// - タグ自体は保持される
//
// Invariants:
// - 既存のコンテンツ（タグ以外）は保持される
func ReplaceSectionOrAppend(content, startTag, endTag, newContent string) (string, error) {
	// startTag と endTag が存在するか確認
	startIndex := strings.Index(content, startTag)
	endIndex := strings.Index(content, endTag)

	// タグが存在しない場合はファイル末尾に追加
	if startIndex == -1 || endIndex == -1 {
		// ファイル末尾に改行がない場合は追加
		result := content
		if !strings.HasSuffix(result, "\n") && !strings.HasSuffix(result, "\n\n") {
			result += "\n"
		}
		// タグとコンテンツを追加
		result += "\n" + startTag + "\n"
		if newContent != "" {
			result += newContent + "\n"
		}
		result += endTag + "\n"
		return result, nil
	}

	// endTag の位置が startTag より前にある場合はエラー
	if endIndex < startIndex {
		return "", fmt.Errorf("終了タグが開始タグより前にあります")
	}

	// 既存の ReplaceSection のロジックを使用
	return ReplaceSection(content, startTag, endTag, newContent)
}

// ReplaceSection テキスト内の指定されたセクション（startTag と endTag の間）を新しいコンテンツで置き換える
//
// Preconditions:
// - content が読み込まれたファイルのコンテンツであること
// - startTag と endTag が有効なコメントタグ文字列であること
//
// Postconditions:
// - startTag と endTag の間のコンテンツが newContent で置き換えられる
// - タグ自体は保持される
//
// Invariants:
// - タグが見つからない場合はエラーが返される
// - 既存のコンテンツ（タグ以外）は保持される
func ReplaceSection(content, startTag, endTag, newContent string) (string, error) {
	// startTag と endTag が存在するか確認
	startIndex := strings.Index(content, startTag)
	if startIndex == -1 {
		return "", fmt.Errorf("開始タグが見つかりません: %s", startTag)
	}

	endIndex := strings.Index(content, endTag)
	if endIndex == -1 {
		return "", fmt.Errorf("終了タグが見つかりません: %s", endTag)
	}

	// endTag の位置が startTag より前にある場合はエラー
	if endIndex < startIndex {
		return "", fmt.Errorf("終了タグが開始タグより前にあります")
	}

	// startTag の終了位置を取得（タグの終端まで）
	startTagEnd := startIndex + len(startTag)

	// endTag の開始位置（タグの開始位置）

	// 置き換え: [開始タグの終端] + [新しいコンテンツ] + [終了タグ]
	before := content[:startTagEnd]
	after := content[endIndex:]

	// 改行の調整
	result := before

	// 新しいコンテンツが空の場合でも改行を保持
	if newContent == "" {
		// タグの後に改行がない場合は追加
		if !strings.HasSuffix(before, "\n") {
			result += "\n"
		}
	} else {
		// タグの後に改行がない場合は追加
		if !strings.HasSuffix(before, "\n") {
			result += "\n"
		}
		result += newContent
		// endTag の前に改行が必要な場合は追加
		if !strings.HasPrefix(after, "\n") && !strings.HasSuffix(newContent, "\n") {
			result += "\n"
		}
	}

	result += after

	return result, nil
}

// FindSection 指定されたタグ間のコンテンツを抽出する
//
// Preconditions:
// - content が読み込まれたファイルのコンテンツであること
// - startTag と endTag が有効なコメントタグ文字列であること
//
// Postconditions:
// - startTag と endTag の間のコンテンツが返される（タグ自体は含まない）
// - タグが見つからない場合はエラーが返される
//
// Invariants:
// - タグの順序が正しいことが確認される
func FindSection(content, startTag, endTag string) (string, error) {
	startIndex := strings.Index(content, startTag)
	if startIndex == -1 {
		return "", fmt.Errorf("開始タグが見つかりません: %s", startTag)
	}

	endIndex := strings.Index(content, endTag)
	if endIndex == -1 {
		return "", fmt.Errorf("終了タグが見つかりません: %s", endTag)
	}

	if endIndex < startIndex {
		return "", fmt.Errorf("終了タグが開始タグより前にあります")
	}

	// 開始タグの終了位置から終了タグの開始位置までを抽出
	startContentIndex := startIndex + len(startTag)
	sectionContent := content[startContentIndex:endIndex]

	// 前後の空白や改行をトリム
	sectionContent = strings.TrimSpace(sectionContent)

	return sectionContent, nil
}

// ValidateTags README.md 内に指定されたタグが正しく存在するか検証する
//
// Preconditions:
// - readmePath が有効な README.md ファイルパスであること
// - startTag と endTag が有効なコメントタグ文字列であること
//
// Postconditions:
// - タグが存在し、順序が正しい場合は nil が返される
// - タグが見つからない、または順序が不正な場合はエラーが返される
//
// Invariants:
// - ファイルが読み込めることが前提
func ValidateTags(readmePath, startTag, endTag string) error {
	content, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("README ファイルの読み込みに失敗しました: %w", err)
	}

	readmeContent := string(content)

	// タグの存在確認
	startIndex := strings.Index(readmeContent, startTag)
	if startIndex == -1 {
		return fmt.Errorf("開始タグが見つかりません: %s", startTag)
	}

	endIndex := strings.Index(readmeContent, endTag)
	if endIndex == -1 {
		return fmt.Errorf("終了タグが見つかりません: %s", endTag)
	}

	// 順序の確認
	if endIndex < startIndex {
		return fmt.Errorf("終了タグが開始タグより前にあります")
	}

	return nil
}

// ReadFile README.md ファイルを読み込む
//
// Preconditions:
// - filePath が有効なファイルパスであること
//
// Postconditions:
// - ファイルの内容が文字列として返される
// - ファイルが存在しない場合はエラーが返される
//
// Invariants:
// - ファイルは UTF-8 エンコーディングであることを前提とする
func ReadFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("ファイルの読み込みに失敗しました: %w", err)
	}

	return string(content), nil
}

// NormalizeTags コメントタグを正規化する（大文字小文字を統一、空白を整理）
//
// Preconditions:
// - tag がコメントタグ文字列であること（例: "START_LANGUAGE_STATS"）
//
// Postconditions:
// - HTML コメント形式のタグが返される（例: "<!-- START_LANGUAGE_STATS -->"）
//
// Invariants:
// - 返されるタグは有効な HTML コメント形式である
func NormalizeTags(tagName string) (startTag, endTag string) {
	// 既に HTML コメント形式の場合
	if strings.HasPrefix(tagName, "<!--") {
		// タグ名部分を抽出（"<!-- " と " -->" を削除）
		tagContent := strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(tagName, "-->"), "<!--"))
		if strings.HasPrefix(tagContent, "START") {
			// START タグの場合、END タグを生成
			endContent := strings.TrimSpace(strings.Replace(tagContent, "START", "END", 1))
			return tagName, fmt.Sprintf("<!-- %s -->", endContent)
		}
		// その他の場合はそのまま使用
		return tagName, fmt.Sprintf("<!-- END%s -->", strings.TrimPrefix(tagContent, "START"))
	}

	// タグ名から HTML コメント形式を生成（大文字に変換）
	startTag = fmt.Sprintf("<!-- START_%s -->", strings.ToUpper(tagName))
	endTag = fmt.Sprintf("<!-- END_%s -->", strings.ToUpper(tagName))

	return startTag, endTag
}

// EscapeRegexSpecialChars 正規表現の特殊文字をエスケープする（内部使用）
func escapeRegexSpecialChars(s string) string {
	re := regexp.MustCompile(`[.*+?^${}()|[\]\\]`)
	return re.ReplaceAllStringFunc(s, func(matched string) string {
		return "\\" + matched
	})
}
