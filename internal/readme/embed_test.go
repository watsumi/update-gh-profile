package readme

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSVGMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		svgFiles []string
		want     string
	}{
		{
			name:     "正常系: 単一のSVGファイル",
			svgFiles: []string{"language_chart.svg"},
			want:     "![Language Chart](language_chart.svg)",
		},
		{
			name:     "正常系: 複数のSVGファイル",
			svgFiles: []string{"language_chart.svg", "commit_history_chart.svg"},
			want:     "![Language Chart](language_chart.svg)\n\n![Commit History Chart](commit_history_chart.svg)",
		},
		{
			name:     "正常系: アンダースコアを含むファイル名",
			svgFiles: []string{"commit_languages_chart.svg"},
			want:     "![Commit Languages Chart](commit_languages_chart.svg)",
		},
		{
			name:     "正常系: パスを含むファイル名",
			svgFiles: []string{"charts/language_chart.svg"},
			want:     "![Language Chart](language_chart.svg)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSVGMarkdown(tt.svgFiles)
			if result != tt.want {
				t.Errorf("GenerateSVGMarkdown() = %q, want %q", result, tt.want)
			}
		})
	}
}

func TestEmbedSVGs(t *testing.T) {
	testDir := "test_embed"
	defer func() {
		os.RemoveAll(testDir)
	}()

	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("テストディレクトリの作成に失敗しました: %v", err)
	}

	// テスト用のREADMEファイルを作成
	testReadme := filepath.Join(testDir, "README.md")
	initialContent := `# Test README

<!-- START_LANGUAGE_STATS -->
old content
<!-- END_LANGUAGE_STATS -->
`

	err = os.WriteFile(testReadme, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("テストファイルの作成に失敗しました: %v", err)
	}

	// テスト用のSVGファイルパス
	svgFiles := []string{"language_chart.svg", "commit_history_chart.svg"}

	// SVG グラフを埋め込む
	err = EmbedSVGs(testReadme, svgFiles, "LANGUAGE_STATS")
	if err != nil {
		t.Fatalf("EmbedSVGs() エラー = %v", err)
	}

	// ファイルを読み込んで確認
	content, err := ReadFile(testReadme)
	if err != nil {
		t.Fatalf("ファイルの読み込みに失敗しました: %v", err)
	}

	// Markdown 記法が含まれているか確認
	if !strings.Contains(content, "![Language Chart]") {
		t.Errorf("EmbedSVGs() 最初のSVGが埋め込まれていません")
	}

	if !strings.Contains(content, "![Commit History Chart]") {
		t.Errorf("EmbedSVGs() 2番目のSVGが埋め込まれていません")
	}

	// タグが保持されているか確認
	if !strings.Contains(content, "<!-- START_LANGUAGE_STATS -->") {
		t.Errorf("EmbedSVGs() 開始タグが保持されていません")
	}

	if !strings.Contains(content, "<!-- END_LANGUAGE_STATS -->") {
		t.Errorf("EmbedSVGs() 終了タグが保持されていません")
	}
}

func TestEmbedSVGWithCustomPath(t *testing.T) {
	testDir := "test_embed_custom"
	defer func() {
		os.RemoveAll(testDir)
	}()

	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("テストディレクトリの作成に失敗しました: %v", err)
	}

	// テスト用のREADMEファイルを作成
	testReadme := filepath.Join(testDir, "README.md")
	initialContent := `# Test README

<!-- START_COMMIT_HISTORY -->
old content
<!-- END_COMMIT_HISTORY -->
`

	err = os.WriteFile(testReadme, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("テストファイルの作成に失敗しました: %v", err)
	}

	// カスタムパスと説明文で SVG を埋め込む
	customPath := "charts/custom_chart.svg"
	customDescription := "コミット推移グラフ"

	err = EmbedSVGWithCustomPath(testReadme, customPath, "COMMIT_HISTORY", customDescription)
	if err != nil {
		t.Fatalf("EmbedSVGWithCustomPath() エラー = %v", err)
	}

	// ファイルを読み込んで確認
	content, err := ReadFile(testReadme)
	if err != nil {
		t.Fatalf("ファイルの読み込みに失敗しました: %v", err)
	}

	// カスタムパスと説明文が含まれているか確認
	if !strings.Contains(content, customDescription) {
		t.Errorf("EmbedSVGWithCustomPath() カスタム説明文が含まれていません")
	}

	if !strings.Contains(content, customPath) {
		t.Errorf("EmbedSVGWithCustomPath() カスタムパスが含まれていません")
	}
}

func TestEmbedMultipleSVGSections(t *testing.T) {
	testDir := "test_embed_multiple"
	defer func() {
		os.RemoveAll(testDir)
	}()

	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("テストディレクトリの作成に失敗しました: %v", err)
	}

	// テスト用のREADMEファイルを作成
	testReadme := filepath.Join(testDir, "README.md")
	initialContent := `# Test README

<!-- START_LANGUAGE_STATS -->
old content 1
<!-- END_LANGUAGE_STATS -->

<!-- START_COMMIT_HISTORY -->
old content 2
<!-- END_COMMIT_HISTORY -->
`

	err = os.WriteFile(testReadme, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("テストファイルの作成に失敗しました: %v", err)
	}

	// 複数のセクションに SVG を埋め込む
	svgSections := map[string]string{
		"LANGUAGE_STATS": "language_chart.svg",
		"COMMIT_HISTORY": "commit_history_chart.svg",
	}

	err = EmbedMultipleSVGSections(testReadme, svgSections)
	if err != nil {
		t.Fatalf("EmbedMultipleSVGSections() エラー = %v", err)
	}

	// ファイルを読み込んで確認
	content, err := ReadFile(testReadme)
	if err != nil {
		t.Fatalf("ファイルの読み込みに失敗しました: %v", err)
	}

	// 各セクションが更新されているか確認
	if !strings.Contains(content, "language_chart.svg") {
		t.Errorf("EmbedMultipleSVGSections() LANGUAGE_STATS セクションが更新されていません")
	}

	if !strings.Contains(content, "commit_history_chart.svg") {
		t.Errorf("EmbedMultipleSVGSections() COMMIT_HISTORY セクションが更新されていません")
	}
}
