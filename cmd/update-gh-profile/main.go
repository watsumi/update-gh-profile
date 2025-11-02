package main

import (
	"context"
	"fmt"
	"os"

	"github.com/watsumi/update-gh-profile/internal/config"

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
	fmt.Println("次のステップ: リポジトリ情報の取得機能を実装します")
}
