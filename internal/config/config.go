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
	// すべてのリポジトリを読み取る権限が必要
	GitHubToken string
}

// Load 環境変数から設定を読み込む
// Goでは、大文字で始まる関数は外部パッケージから呼び出せます（公開関数）
func Load() (*Config, error) {
	// &Config{} は構造体のポインタを作成します
	// Goでは、構造体を返す場合、ポインタで返すのが一般的です
	cfg := &Config{}

	// 環境変数 GITHUB_TOKEN を読み込む
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, errors.New("GITHUB_TOKEN 環境変数が設定されていません")
	}
	cfg.GitHubToken = token

	// ログ出力: 設定読み込み成功（INFOレベル相当）
	log.Printf("設定を読み込みました: トークン設定=設定済み（認証ユーザーは自動取得されます）")

	return cfg, nil
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
