package readme

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReplaceSection(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		startTag   string
		endTag     string
		newContent string
		want       string
		wantError  bool
	}{
		{
			name: "正常系: 基本的な置き換え",
			content: `# Title

<!-- START_SECTION -->
old content
<!-- END_SECTION -->

rest of content`,
			startTag:   "<!-- START_SECTION -->",
			endTag:     "<!-- END_SECTION -->",
			newContent: "new content",
			want: `# Title

<!-- START_SECTION -->
new content
<!-- END_SECTION -->

rest of content`,
			wantError: false,
		},
		{
			name: "正常系: 複数行のコンテンツ",
			content: `<!-- START_SECTION -->
old line 1
old line 2
<!-- END_SECTION -->`,
			startTag:   "<!-- START_SECTION -->",
			endTag:     "<!-- END_SECTION -->",
			newContent: "new line 1\nnew line 2",
			want: `<!-- START_SECTION -->
new line 1
new line 2
<!-- END_SECTION -->`,
			wantError: false,
		},
		{
			name: "正常系: 空のコンテンツ",
			content: `<!-- START_SECTION -->
old content
<!-- END_SECTION -->`,
			startTag:   "<!-- START_SECTION -->",
			endTag:     "<!-- END_SECTION -->",
			newContent: "",
			want: `<!-- START_SECTION -->
<!-- END_SECTION -->`,
			wantError: false,
		},
		{
			name: "エラー: 開始タグが見つからない",
			content: `some content
<!-- END_SECTION -->`,
			startTag:   "<!-- START_SECTION -->",
			endTag:     "<!-- END_SECTION -->",
			newContent: "new content",
			want:       "",
			wantError:  true,
		},
		{
			name: "エラー: 終了タグが見つからない",
			content: `<!-- START_SECTION -->
some content`,
			startTag:   "<!-- START_SECTION -->",
			endTag:     "<!-- END_SECTION -->",
			newContent: "new content",
			want:       "",
			wantError:  true,
		},
		{
			name: "エラー: 終了タグが開始タグより前",
			content: `<!-- END_SECTION -->
<!-- START_SECTION -->
some content`,
			startTag:   "<!-- START_SECTION -->",
			endTag:     "<!-- END_SECTION -->",
			newContent: "new content",
			want:       "",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ReplaceSection(tt.content, tt.startTag, tt.endTag, tt.newContent)

			if tt.wantError {
				if err == nil {
					t.Errorf("ReplaceSection() はエラーを返すべきでしたが、nil を返しました")
				}
			} else {
				if err != nil {
					t.Errorf("ReplaceSection() エラー = %v, エラーを期待していませんでした", err)
					return
				}

				if strings.TrimSpace(result) != strings.TrimSpace(tt.want) {
					t.Errorf("ReplaceSection() = %q, want %q", result, tt.want)
				}
			}
		})
	}
}

func TestFindSection(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		startTag  string
		endTag    string
		want      string
		wantError bool
	}{
		{
			name: "正常系: セクションの抽出",
			content: `# Title

<!-- START_SECTION -->
extracted content
<!-- END_SECTION -->

rest`,
			startTag:  "<!-- START_SECTION -->",
			endTag:    "<!-- END_SECTION -->",
			want:      "extracted content",
			wantError: false,
		},
		{
			name: "正常系: 複数行のセクション",
			content: `<!-- START_SECTION -->
line 1
line 2
line 3
<!-- END_SECTION -->`,
			startTag:  "<!-- START_SECTION -->",
			endTag:    "<!-- END_SECTION -->",
			want:      "line 1\nline 2\nline 3",
			wantError: false,
		},
		{
			name: "エラー: 開始タグが見つからない",
			content: `some content
<!-- END_SECTION -->`,
			startTag:  "<!-- START_SECTION -->",
			endTag:    "<!-- END_SECTION -->",
			want:      "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FindSection(tt.content, tt.startTag, tt.endTag)

			if tt.wantError {
				if err == nil {
					t.Errorf("FindSection() はエラーを返すべきでしたが、nil を返しました")
				}
			} else {
				if err != nil {
					t.Errorf("FindSection() エラー = %v, エラーを期待していませんでした", err)
					return
				}

				if strings.TrimSpace(result) != strings.TrimSpace(tt.want) {
					t.Errorf("FindSection() = %q, want %q", result, tt.want)
				}
			}
		})
	}
}

func TestNormalizeTags(t *testing.T) {
	tests := []struct {
		name      string
		tagName   string
		wantStart string
		wantEnd   string
	}{
		{
			name:      "基本的なタグ名",
			tagName:   "LANGUAGE_STATS",
			wantStart: "<!-- START_LANGUAGE_STATS -->",
			wantEnd:   "<!-- END_LANGUAGE_STATS -->",
		},
		{
			name:      "小文字のタグ名（大文字に変換される）",
			tagName:   "language_stats",
			wantStart: "<!-- START_LANGUAGE_STATS -->",
			wantEnd:   "<!-- END_LANGUAGE_STATS -->",
		},
		{
			name:      "既にHTMLコメント形式",
			tagName:   "<!-- START_SECTION -->",
			wantStart: "<!-- START_SECTION -->",
			wantEnd:   "<!-- END_SECTION -->",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := NormalizeTags(tt.tagName)
			if start != tt.wantStart {
				t.Errorf("NormalizeTags() start = %q, want %q", start, tt.wantStart)
			}
			if end != tt.wantEnd {
				t.Errorf("NormalizeTags() end = %q, want %q", end, tt.wantEnd)
			}
		})
	}
}

func TestValidateTags(t *testing.T) {
	testDir := "test_readme"
	defer func() {
		os.RemoveAll(testDir)
	}()

	// テスト用のREADMEファイルを作成
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("テストディレクトリの作成に失敗しました: %v", err)
	}

	tests := []struct {
		name      string
		content   string
		startTag  string
		endTag    string
		wantError bool
	}{
		{
			name: "正常系: 正しいタグ",
			content: `# README

<!-- START_SECTION -->
content
<!-- END_SECTION -->`,
			startTag:  "<!-- START_SECTION -->",
			endTag:    "<!-- END_SECTION -->",
			wantError: false,
		},
		{
			name: "エラー: 開始タグが見つからない",
			content: `# README

<!-- END_SECTION -->`,
			startTag:  "<!-- START_SECTION -->",
			endTag:    "<!-- END_SECTION -->",
			wantError: true,
		},
		{
			name: "エラー: 終了タグが見つからない",
			content: `# README

<!-- START_SECTION -->
content`,
			startTag:  "<!-- START_SECTION -->",
			endTag:    "<!-- END_SECTION -->",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(testDir, "README_"+tt.name+".md")
			err := os.WriteFile(testFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("テストファイルの作成に失敗しました: %v", err)
			}

			err = ValidateTags(testFile, tt.startTag, tt.endTag)

			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateTags() はエラーを返すべきでしたが、nil を返しました")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateTags() エラー = %v, エラーを期待していませんでした", err)
				}
			}
		})
	}
}

func TestUpdateSection(t *testing.T) {
	testDir := "test_update"
	defer func() {
		os.RemoveAll(testDir)
	}()

	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("テストディレクトリの作成に失敗しました: %v", err)
	}

	testFile := filepath.Join(testDir, "README.md")
	initialContent := `# Test README

<!-- START_SECTION -->
old content
<!-- END_SECTION -->

rest of content`

	err = os.WriteFile(testFile, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("テストファイルの作成に失敗しました: %v", err)
	}

	// セクションを更新
	newContent := "new content"
	err = UpdateSection(testFile, "<!-- START_SECTION -->", "<!-- END_SECTION -->", newContent)
	if err != nil {
		t.Fatalf("UpdateSection() エラー = %v", err)
	}

	// ファイルを読み込んで確認
	updatedContent, err := ReadFile(testFile)
	if err != nil {
		t.Fatalf("ファイルの読み込みに失敗しました: %v", err)
	}

	if !strings.Contains(updatedContent, newContent) {
		t.Errorf("UpdateSection() 更新されたコンテンツが含まれていません")
	}

	if !strings.Contains(updatedContent, "<!-- START_SECTION -->") {
		t.Errorf("UpdateSection() 開始タグが保持されていません")
	}

	if !strings.Contains(updatedContent, "<!-- END_SECTION -->") {
		t.Errorf("UpdateSection() 終了タグが保持されていません")
	}

	if strings.Contains(updatedContent, "old content") {
		t.Errorf("UpdateSection() 古いコンテンツが残っています")
	}
}
