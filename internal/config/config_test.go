package config

import (
	"os"
	"testing"
)

// TestLoad 設定読み込みのテスト
// Goのテストファイルは *_test.go という命名規則に従います
// テスト関数は Test で始まり、*testing.T を受け取ります
func TestLoad(t *testing.T) {
	// テストケース1: 正常なケース
	// 環境変数を設定
	os.Setenv("GITHUB_TOKEN", "test_token_12345")

	// 設定を読み込む
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() エラー = %v, エラーが発生しないことを期待", err)
	}

	// 設定値が正しいか確認
	if cfg.GitHubToken != "test_token_12345" {
		t.Errorf("GitHubToken = %v, 期待値 = test_token_12345", cfg.GitHubToken)
	}

	// テストケース2: トークンが設定されていない場合
	os.Unsetenv("GITHUB_TOKEN")
	_, err = Load()
	if err == nil {
		t.Error("Load() エラー = nil, エラーが発生することを期待")
	}
}

// TestValidate 設定値検証のテスト
func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "正常な設定",
			config: &Config{
				GitHubToken: "valid_token_12345",
			},
			wantErr: false,
		},
		{
			name: "トークンが空",
			config: &Config{
				GitHubToken: "",
			},
			wantErr: true,
		},
		{
			name: "トークンが短すぎる",
			config: &Config{
				GitHubToken: "short",
			},
			wantErr: true,
		},
	}

	// テーブル駆動テスト（Table-Driven Tests）
	// Go言語でよく使われるテストパターン
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() エラー = %v, 期待値 = %v", err, tt.wantErr)
			}
		})
	}
}
