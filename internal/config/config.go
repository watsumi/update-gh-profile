package config

import (
	"errors"
	"fmt"
	"log"
	"os"
)

// Config アプリケーションの設定を保持する構造体
// Go言語では、構造体（struct）を使ってデータをまとめます
type Config struct {
	// GitHubToken GitHub API の認証トークン
	// 文字列型（string）で定義
	GitHubToken string

	// TargetUsername 対象となるGitHubユーザー名（空の場合は認証ユーザー）
	// 文字列型（string）で定義
	TargetUsername string
}

// Load 環境変数から設定を読み込む
// Goでは、大文字で始まる関数は外部パッケージから呼び出せます（公開関数）
func Load() (*Config, error) {
	// &Config{} は構造体のポインタを作成します
	// Goでは、構造体を返す場合、ポインタで返すのが一般的です
	cfg := &Config{}

	// 環境変数 GITHUB_TOKEN を読み込む
	// os.Getenv は環境変数の値を返します（存在しない場合は空文字列）
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		// errors.New で新しいエラーを作成
		return nil, errors.New("GITHUB_TOKEN 環境変数が設定されていません")
	}
	cfg.GitHubToken = token

	// 環境変数 GITHUB_USERNAME を読み込む（オプション）
	// 空の場合は、後で認証ユーザーを取得します
	username := os.Getenv("GITHUB_USERNAME")
	cfg.TargetUsername = username

	// ログ出力: 設定読み込み成功（INFOレベル相当）
	log.Printf("設定を読み込みました: 対象ユーザー=%s", func() string {
		if username == "" {
			return "認証ユーザー（デフォルト）"
		}
		return username
	}())

	return cfg, nil
}

// GetTargetUser 対象ユーザー名を取得する
// 設定されていない場合は空文字列を返す（デフォルトは認証ユーザー）
func (c *Config) GetTargetUser() string {
	if c.TargetUsername == "" {
		return "" // 空文字列は認証ユーザーを意味する
	}
	return c.TargetUsername
}

// Validate 設定値の検証を行う
func (c *Config) Validate() error {
	if c.GitHubToken == "" {
		return errors.New("GitHubToken が設定されていません")
	}

	// トークンが空でないことを確認（最小限の検証）
	if len(c.GitHubToken) < 10 {
		return fmt.Errorf("GitHubToken が短すぎます（長さ: %d）", len(c.GitHubToken))
	}

	return nil
}
