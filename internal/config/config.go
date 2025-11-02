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
	// GitHubTokenRead リポジトリ情報を読み取るためのGitHub API認証トークン
	// すべてのリポジトリを読み取る権限が必要
	GitHubTokenRead string

	// GitHubTokenWrite README.mdを更新するためのGitHub API認証トークン
	// リポジトリへの書き込み権限が必要（Git push用）
	GitHubTokenWrite string

	// GitHubToken GitHub API の認証トークン（後方互換性のため）
	// GITHUB_TOKEN_READ または GITHUB_TOKEN_WRITE が設定されていない場合に使用
	// 両方のトークンが設定されていない場合は、このトークンが両方に使用される
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

	// 環境変数 GITHUB_TOKEN_READ を読み込む
	tokenRead := os.Getenv("GITHUB_TOKEN_READ")
	// 環境変数 GITHUB_TOKEN_WRITE を読み込む
	tokenWrite := os.Getenv("GITHUB_TOKEN_WRITE")
	// 環境変数 GITHUB_TOKEN を読み込む（後方互換性のため）
	token := os.Getenv("GITHUB_TOKEN")

	// トークンの設定ロジック:
	// 1. GITHUB_TOKEN_READ が設定されている場合はそれを使用
	// 2. 設定されていない場合は GITHUB_TOKEN を使用
	// 3. どちらも設定されていない場合はエラー
	if tokenRead == "" {
		if token == "" {
			return nil, errors.New("GITHUB_TOKEN_READ または GITHUB_TOKEN 環境変数が設定されていません")
		}
		tokenRead = token
	}
	cfg.GitHubTokenRead = tokenRead

	// 1. GITHUB_TOKEN_WRITE が設定されている場合はそれを使用
	// 2. 設定されていない場合は GITHUB_TOKEN を使用
	// 3. どちらも設定されていない場合はエラー
	if tokenWrite == "" {
		if token == "" {
			return nil, errors.New("GITHUB_TOKEN_WRITE または GITHUB_TOKEN 環境変数が設定されていません")
		}
		tokenWrite = token
	}
	cfg.GitHubTokenWrite = tokenWrite

	// 後方互換性のため、GITHUB_TOKEN も設定
	cfg.GitHubToken = token

	// 環境変数 GITHUB_USERNAME を読み込む（オプション）
	// 空の場合は、後で認証ユーザーを取得します
	username := os.Getenv("GITHUB_USERNAME")
	cfg.TargetUsername = username

	// ログ出力: 設定読み込み成功（INFOレベル相当）
	tokenStatus := "共通トークン"
	if tokenRead != token || tokenWrite != token {
		tokenStatus = "個別設定"
	}
	log.Printf("設定を読み込みました: 対象ユーザー=%s, トークン設定=%s", func() string {
		if username == "" {
			return "認証ユーザー（デフォルト）"
		}
		return username
	}(), tokenStatus)

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
	if c.GitHubTokenRead == "" {
		return errors.New("GitHubTokenRead が設定されていません")
	}

	if c.GitHubTokenWrite == "" {
		return errors.New("GitHubTokenWrite が設定されていません")
	}

	// トークンが空でないことを確認（最小限の検証）
	if len(c.GitHubTokenRead) < 10 {
		return fmt.Errorf("GitHubTokenRead が短すぎます（長さ: %d）", len(c.GitHubTokenRead))
	}

	if len(c.GitHubTokenWrite) < 10 {
		return fmt.Errorf("GitHubTokenWrite が短すぎます（長さ: %d）", len(c.GitHubTokenWrite))
	}

	return nil
}
