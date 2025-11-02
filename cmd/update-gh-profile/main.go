package main

import (
	"context"
	"fmt"
	"os"

	"github.com/watsumi/update-gh-profile/internal/config"
	"github.com/watsumi/update-gh-profile/internal/repository"

	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
)

func main() {
	fmt.Println("update-gh-profile: GitHub プロフィール自動更新ツール")
	fmt.Println("初期化完了")

	// 設定を読み込む
	// config.Load() は *Config と error を返します
	cfg, err := config.Load()
	if err != nil {
		// エラーハンドリング: エラーが発生した場合は処理を中断
		fmt.Printf("エラー: 設定の読み込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	// 設定値の検証
	// Validate() メソッドを呼び出して、設定が正しいか確認
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

	// 対象ユーザー名の取得
	targetUser := cfg.GetTargetUser()
	if targetUser == "" {
		// デフォルトでは認証ユーザーを取得
		user, _, err := client.Users.Get(ctx, "")
		if err != nil {
			fmt.Printf("エラー: ユーザー情報の取得に失敗しました: %v\n", err)
			os.Exit(1)
		}
		targetUser = user.GetLogin()
		fmt.Printf("✓ 認証ユーザー: %s\n", targetUser)
	} else {
		fmt.Printf("✓ 対象ユーザー: %s\n", targetUser)
	}

	fmt.Println("\n✅ GitHub API クライアントの初期化に成功しました！")

	// リポジトリ一覧の取得
	fmt.Println("\n📦 リポジトリ一覧を取得しています...")
	repos, err := repository.FetchUserRepositories(ctx, client, targetUser, true) // excludeForks=true
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
}
