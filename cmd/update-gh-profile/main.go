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
		usernameFlag    = flag.String("username", "", "GitHub ユーザー名（省略時は環境変数 GITHUB_USERNAME または認証ユーザー）")
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

	// 対象ユーザー名の決定（優先順位: コマンドライン引数 > 環境変数 > 認証ユーザー）
	targetUser := *usernameFlag
	if targetUser == "" {
		targetUser = cfg.GetTargetUser()
		if targetUser == "" {
			user, _, err := client.Users.Get(ctx, "")
			if err != nil {
				fmt.Printf("エラー: ユーザー情報の取得に失敗しました: %v\n", err)
				os.Exit(1)
			}
			targetUser = user.GetLogin()
			fmt.Printf("✓ 認証ユーザー: %s\n", targetUser)
		} else {
			fmt.Printf("✓ 対象ユーザー（環境変数）: %s\n", targetUser)
		}
	} else {
		fmt.Printf("✓ 対象ユーザー（コマンドライン）: %s\n", targetUser)
	}

	// フォーク除外の設定
	excludeForks, err := strconv.ParseBool(*excludeForksStr)
	if err != nil {
		fmt.Printf("警告: exclude-forks の値が不正です（%s）。デフォルト値 true を使用します\n", *excludeForksStr)
		excludeForks = true
	}

	fmt.Println("\n✅ GitHub API クライアントの初期化に成功しました！")

	// リポジトリ一覧の取得
	fmt.Println("\n📦 リポジトリ一覧を取得しています...")
	repos, err := repository.FetchUserRepositories(ctx, client, targetUser, excludeForks)
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
