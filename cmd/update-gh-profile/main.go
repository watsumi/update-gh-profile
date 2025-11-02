package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/watsumi/update-gh-profile/internal/config"
	"github.com/watsumi/update-gh-profile/internal/repository"

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

	// リポジトリ一覧の取得（認証ユーザー自身のリポジトリのみ）
	fmt.Println("\n📦 リポジトリ一覧を取得しています...")
	repos, err := repository.FetchUserRepositories(ctx, client, targetUser, excludeForks, true) // 常に認証ユーザーとして取得
	if err != nil {
		fmt.Printf("エラー: リポジトリ一覧の取得に失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ リポジトリ一覧の取得に成功しました: %d 件\n", len(repos))

	// 取得したリポジトリの一部を表示（最大5件）
	maxDisplay := 5
	if len(repos) < maxDisplay {
		maxDisplay = len(repos)
	}
	fmt.Printf("\n取得したリポジトリ（最初の%d件）:\n", maxDisplay)
	for i := 0; i < maxDisplay; i++ {
		repo := repos[i]
		fmt.Printf("  - %s (⭐ %d, Fork: %v)\n",
			repo.GetFullName(),
			repo.GetStargazersCount(),
			repo.GetFork())
	}
	if len(repos) > maxDisplay {
		fmt.Printf("  ... 他 %d 件\n", len(repos)-maxDisplay)
	}

	// GitHub Actions の出力変数を設定（GITHUB_OUTPUT ファイルに書き込む）
	if outputFile := os.Getenv("GITHUB_OUTPUT"); outputFile != "" {
		file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			fmt.Fprintf(file, "repository_count=%d\n", len(repos))
			file.Close()
		}
	}
}
