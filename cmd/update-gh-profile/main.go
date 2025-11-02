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

	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
)

func main() {
	// コマンドライン引数のパース
	var (
		usernameFlag    = flag.String("username", "", "[非推奨・無視されます] このツールは認証ユーザー自身のリポジトリのみを取得します")
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

	fmt.Println("✓ GITHUB_TOKEN が設定されています")

	// GitHub API クライアントの初期化
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GitHubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// 認証ユーザーを取得（必須）
	authUser, _, err := client.Users.Get(ctx, "")
	if err != nil {
		fmt.Printf("エラー: 認証ユーザー情報の取得に失敗しました: %v\n", err)
		os.Exit(1)
	}
	authenticatedUsername := authUser.GetLogin()

	// 対象ユーザー名の決定（優先順位: コマンドライン引数 > 環境変数 > 認証ユーザー）
	targetUser := *usernameFlag
	if targetUser == "" {
		targetUser = cfg.GetTargetUser()
		if targetUser == "" {
			// 認証ユーザーを使用（デフォルト）
			targetUser = authenticatedUsername
		}
	}

	// 認証ユーザー以外を指定した場合はエラー
	if targetUser != authenticatedUsername {
		fmt.Printf("エラー: 認証ユーザー（%s）以外のリポジトリを取得することはできません\n", authenticatedUsername)
		fmt.Printf("指定されたユーザー: %s\n", targetUser)
		fmt.Println("\nこのツールは認証ユーザー自身のリポジトリのみを取得できます。")
		os.Exit(1)
	}

	// 認証ユーザー自身であることを確認
	fmt.Printf("✓ 認証ユーザー: %s（プライベートリポジトリも取得します）\n", targetUser)

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
		EnableGitPush:   false,                                  // デフォルトではプッシュしない（テストモード）
		MaxRepositories: 0,                                      // 0 = すべてのリポジトリ
		ExcludeForks:    excludeForks,
		LogLevel:        logLevel, // ログレベル
	}

	// ワークフローを実行
	fmt.Println("\n🚀 メインワークフローを開始します...")
	err = workflow.Run(ctx, client, workflowConfig)
	if err != nil {
		fmt.Printf("エラー: ワークフローの実行に失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ すべての処理が完了しました！")
	os.Exit(0)
}
