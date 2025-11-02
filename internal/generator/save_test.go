package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveSVG(t *testing.T) {
	tests := []struct {
		name       string
		svgContent string
		filepath   string
		wantError  bool
	}{
		{
			name:       "正常系: 基本的な保存",
			svgContent: `<?xml version="1.0" encoding="UTF-8"?><svg xmlns="http://www.w3.org/2000/svg"><text>Test</text></svg>`,
			filepath:   "test_output.svg",
			wantError:  false,
		},
		{
			name:       "正常系: サブディレクトリへの保存",
			svgContent: `<?xml version="1.0" encoding="UTF-8"?><svg xmlns="http://www.w3.org/2000/svg"><text>Test</text></svg>`,
			filepath:   "test_output/subdir/test.svg",
			wantError:  false,
		},
		{
			name:       "エラー: 空のコンテンツ",
			svgContent: "",
			filepath:   "test_empty.svg",
			wantError:  true,
		},
		{
			name:       "エラー: 空のファイルパス",
			svgContent: `<?xml version="1.0" encoding="UTF-8"?><svg></svg>`,
			filepath:   "",
			wantError:  true,
		},
	}

	// テスト用の一時ディレクトリを作成
	testDir := "test_output"
	defer func() {
		// テスト後にクリーンアップ
		os.RemoveAll(testDir)
		os.Remove("test_output.svg")
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用のファイルパスを準備
			var testPath string
			if tt.filepath != "" {
				if filepath.IsAbs(tt.filepath) {
					testPath = tt.filepath
				} else {
					testPath = filepath.Join(testDir, tt.filepath)
				}
			}

			err := SaveSVG(tt.svgContent, testPath)

			if tt.wantError {
				if err == nil {
					t.Errorf("SaveSVG() はエラーを返すべきでしたが、nil を返しました")
				}
			} else {
				if err != nil {
					t.Errorf("SaveSVG() エラー = %v, エラーを期待していませんでした", err)
					return
				}

				// ファイルが実際に作成されたか確認
				if testPath != "" {
					if _, err := os.Stat(testPath); os.IsNotExist(err) {
						t.Errorf("SaveSVG() ファイルが作成されませんでした: %s", testPath)
					}
				}
			}
		})
	}
}

func TestSaveMultipleSVGs(t *testing.T) {
	testDir := "test_multiple_output"
	defer func() {
		os.RemoveAll(testDir)
	}()

	tests := []struct {
		name      string
		svgs      map[string]string
		outputDir string
		wantError bool
	}{
		{
			name: "正常系: 複数のSVGを保存",
			svgs: map[string]string{
				"chart1":     `<?xml version="1.0" encoding="UTF-8"?><svg><text>Chart 1</text></svg>`,
				"chart2":     `<?xml version="1.0" encoding="UTF-8"?><svg><text>Chart 2</text></svg>`,
				"chart3.svg": `<?xml version="1.0" encoding="UTF-8"?><svg><text>Chart 3</text></svg>`,
			},
			outputDir: testDir,
			wantError: false,
		},
		{
			name:      "エラー: 空のマップ",
			svgs:      map[string]string{},
			outputDir: testDir,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SaveMultipleSVGs(tt.svgs, tt.outputDir)

			if tt.wantError {
				if err == nil {
					t.Errorf("SaveMultipleSVGs() はエラーを返すべきでしたが、nil を返しました")
				}
			} else {
				if err != nil {
					t.Errorf("SaveMultipleSVGs() エラー = %v, エラーを期待していませんでした", err)
					return
				}

				// すべてのファイルが作成されたか確認
				for filename := range tt.svgs {
					expectedExt := ".svg"
					if filepath.Ext(filename) != "" {
						expectedExt = ""
					}
					filePath := filepath.Join(tt.outputDir, filename+expectedExt)
					if _, err := os.Stat(filePath); os.IsNotExist(err) {
						t.Errorf("SaveMultipleSVGs() ファイルが作成されませんでした: %s", filePath)
					}
				}
			}
		})
	}
}
