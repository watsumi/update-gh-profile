package readme

import (
	"fmt"
	"strings"
	"time"
)

// FormatTimestamp タイムスタンプをフォーマットする
//
// Preconditions:
// - t が有効な time.Time であること
// - format が有効なフォーマット文字列であること（省略可能、デフォルトは RFC3339）
//
// Postconditions:
// - フォーマットされたタイムスタンプ文字列が返される
//
// Invariants:
// - タイムゾーン情報が含まれる
func FormatTimestamp(t time.Time, format string) string {
	if format == "" {
		// デフォルトフォーマット: RFC3339 (例: 2024-01-15T10:30:00Z09:00)
		return t.Format(time.RFC3339)
	}

	return t.Format(format)
}

// FormatTimestampWithTimezone タイムゾーンを指定してタイムスタンプをフォーマットする
//
// Preconditions:
// - t が有効な time.Time であること
// - timezone が有効なタイムゾーン名であること（例: "Asia/Tokyo", "UTC"）
//
// Postconditions:
// - 指定されたタイムゾーンでフォーマットされたタイムスタンプ文字列が返される
//
// Invariants:
// - タイムゾーン情報が正しく適用される
func FormatTimestampWithTimezone(t time.Time, timezone string) (string, error) {
	// タイムゾーンを読み込む
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", fmt.Errorf("タイムゾーンの読み込みに失敗しました: %w", err)
	}

	// 指定されたタイムゾーンに変換
	localTime := t.In(loc)

	// RFC3339 フォーマットで返す
	return localTime.Format(time.RFC3339), nil
}

// AddUpdateTimestamp README.md の指定セクションに更新日時を追加する
//
// Preconditions:
// - readmePath が有効な README.md ファイルパスであること
// - sectionTag が更新するセクションのタグ名であること
// - timestamp が有効な time.Time であること（省略可能、デフォルトは現在時刻）
//
// Postconditions:
// - README.md の指定セクションに更新日時が追加される
// - 既存のコンテンツの後に追加される
//
// Invariants:
// - タイムスタンプは明確な形式で表示される
func AddUpdateTimestamp(readmePath, sectionTag string, timestamp time.Time, timezone string) error {
	// セクションタグを正規化
	startTag, endTag := NormalizeTags(sectionTag)

	// README ファイルを読み込む
	content, err := ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("README ファイルの読み込みに失敗しました: %w", err)
	}

	// 既存のセクションコンテンツを取得（エラーは無視して空文字列を扱う）
	sectionContent, _ := FindSection(content, startTag, endTag)

	// タイムスタンプをフォーマット
	var timestampStr string
	if timezone != "" {
		timestampStr, err = FormatTimestampWithTimezone(timestamp, timezone)
		if err != nil {
			return fmt.Errorf("タイムスタンプのフォーマットに失敗しました: %w", err)
		}
	} else {
		timestampStr = FormatTimestamp(timestamp, "")
	}

	// 更新日時のマークダウンを生成
	timestampMarkdown := fmt.Sprintf("\n\n*最終更新: %s*", timestampStr)

	// 新しいコンテンツを作成（既存コンテンツ + タイムスタンプ）
	newContent := sectionContent
	if sectionContent != "" {
		newContent += timestampMarkdown
	} else {
		newContent = strings.TrimPrefix(timestampMarkdown, "\n\n")
	}

	// セクションを更新（タグがない場合は自動追加される）
	err = UpdateSection(readmePath, startTag, endTag, newContent)
	if err != nil {
		return fmt.Errorf("セクションの更新に失敗しました: %w", err)
	}

	return nil
}

// GenerateTimestampMarkdown タイムスタンプの Markdown 形式を生成する
//
// Preconditions:
// - timestamp が有効な time.Time であること
//
// Postconditions:
// - Markdown 形式のタイムスタンプ文字列が返される
//
// Invariants:
// - フォーマットは一貫している
func GenerateTimestampMarkdown(timestamp time.Time, timezone string) (string, error) {
	var timestampStr string
	var err error

	if timezone != "" {
		timestampStr, err = FormatTimestampWithTimezone(timestamp, timezone)
		if err != nil {
			return "", fmt.Errorf("タイムスタンプのフォーマットに失敗しました: %w", err)
		}
	} else {
		timestampStr = FormatTimestamp(timestamp, "")
	}

	return fmt.Sprintf("*最終更新: %s*", timestampStr), nil
}

// GetCurrentTimestamp 現在のタイムスタンプを取得する
//
// Preconditions:
// - なし
//
// Postconditions:
// - 現在の時刻が time.Time として返される
//
// Invariants:
// - UTC タイムゾーンで返される
func GetCurrentTimestamp() time.Time {
	return time.Now().UTC()
}
