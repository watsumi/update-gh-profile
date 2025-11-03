package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/watsumi/update-gh-profile/internal/config"
	"github.com/watsumi/update-gh-profile/internal/logger"
	"github.com/watsumi/update-gh-profile/internal/workflow"
)

func main() {
	// コマンドライン引数のパース
	var (
		excludeForksStr = flag.String("exclude-forks", "true", "フォークリポジトリを除外するか（true/false）")
	)
	flag.Parse()

	fmt.Println("update-gh-profile: GitHub プロフィール自動更新ツール")
	fmt.Println("初期化完了")

	// 設定を読み込む
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("エラー: 設定の読み込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Printf("エラー: 設定の検証に失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ GitHub Token が設定されています")

	// コンテキストの作成
	ctx := context.Background()

	// 認証ユーザーはGraphQLで自動的に取得されます

	// フォーク除外の設定
	excludeForks, err := strconv.ParseBool(*excludeForksStr)
	if err != nil {
		fmt.Printf("警告: exclude-forks の値が不正です（%s）。デフォルト値 true を使用します\n", *excludeForksStr)
		excludeForks = true
	}

	fmt.Println("\n✅ GitHub API クライアントの初期化に成功しました！")

	// ログレベルの設定（環境変数から読み込み）
	logLevelStr := os.Getenv("LOG_LEVEL")
	if logLevelStr == "" {
		logLevelStr = "INFO"
	}
	logLevel := logger.ParseLogLevel(logLevelStr)

	// ワークフロー設定
	workflowConfig := workflow.Config{
		RepoPath:        ".",                                    // カレントディレクトリ
		SVGOutputDir:    ".",                                    // SVG ファイルの出力先
		Timezone:        "UTC",                                  // タイムゾーン
		CommitMessage:   "chore: update GitHub profile metrics", // Git コミットメッセージ
		MaxRepositories: 0,                                      // 0 = すべてのリポジトリ
		ExcludeForks:    excludeForks,
		LogLevel:        logLevel, // ログレベル
	}

	// ワークフローを実行
	fmt.Println("\n🚀 メインワークフローを開始します...")
	err = workflow.Run(ctx, cfg.GitHubToken, workflowConfig)
	if err != nil {
		fmt.Printf("エラー: ワークフローの実行に失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ すべての処理が完了しました！")
	os.Exit(0)
}
