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
	os.Setenv("GITHUB_USERNAME", "testuser")

	// 設定を読み込む
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() エラー = %v, エラーが発生しないことを期待", err)
	}

	// 設定値が正しいか確認
	if cfg.GitHubToken != "test_token_12345" {
		t.Errorf("GitHubToken = %v, 期待値 = test_token_12345", cfg.GitHubToken)
	}
	if cfg.TargetUsername != "testuser" {
		t.Errorf("TargetUsername = %v, 期待値 = testuser", cfg.TargetUsername)
	}

	// テストケース2: トークンが設定されていない場合
	os.Unsetenv("GITHUB_TOKEN")
	_, err = Load()
	if err == nil {
		t.Error("Load() エラー = nil, エラーが発生することを期待")
	}

	// 環境変数をクリーンアップ
	os.Unsetenv("GITHUB_USERNAME")
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
				GitHubToken:    "valid_token_12345",
				TargetUsername: "testuser",
			},
			wantErr: false,
		},
		{
			name: "トークンが空",
			config: &Config{
				GitHubToken:    "",
				TargetUsername: "testuser",
			},
			wantErr: true,
		},
		{
			name: "トークンが短すぎる",
			config: &Config{
				GitHubToken:    "short",
				TargetUsername: "testuser",
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

// TestGetTargetUser 対象ユーザー取得のテスト
func TestGetTargetUser(t *testing.T) {
	tests := []struct {
		name           string
		config         *Config
		expectedResult string
	}{
		{
			name: "ユーザー名が設定されている",
			config: &Config{
				TargetUsername: "testuser",
			},
			expectedResult: "testuser",
		},
		{
			name: "ユーザー名が空（デフォルト）",
			config: &Config{
				TargetUsername: "",
			},
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetTargetUser()
			if result != tt.expectedResult {
				t.Errorf("GetTargetUser() = %v, 期待値 = %v", result, tt.expectedResult)
			}
		})
	}
}
