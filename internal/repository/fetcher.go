package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/go-github/v56/github"
)

// FetchUserRepositories GitHub API を使用してユーザーのリポジトリ一覧を取得する
//
// Preconditions:
// - username が非空文字列であること
// - client が有効な GitHub クライアントであること
//
// Postconditions:
// - フォークを除外する場合は、Fork=false のリポジトリのみが返される
// - スライスはリポジトリ構造体のスライスである
//
// Invariants:
// - API レート制限に達した場合は待機して再試行する
func FetchUserRepositories(ctx context.Context, client *github.Client, username string, excludeForks bool) ([]*github.Repository, error) {
	if username == "" {
		return nil, fmt.Errorf("username が空です")
	}

	log.Printf("リポジトリ一覧を取得しています: ユーザー=%s, フォーク除外=%v", username, excludeForks)

	// ページネーション用のオプション
	opt := &github.RepositoryListOptions{
		Type:        "all", // all, owner, member から選択
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100}, // 1ページあたりの最大件数
	}

	var allRepos []*github.Repository

	// ページネーションループ: 全ページを取得するまで繰り返す
	for {
		// GitHub API を呼び出してリポジトリ一覧を取得
		repos, resp, err := client.Repositories.List(ctx, username, opt)
		if err != nil {
			return nil, fmt.Errorf("リポジトリ一覧の取得に失敗しました: %w", err)
		}

		// レート制限のチェックと処理
		if err := handleRateLimit(ctx, resp); err != nil {
			return nil, fmt.Errorf("レート制限の処理に失敗しました: %w", err)
		}

		// 取得したリポジトリを追加
		allRepos = append(allRepos, repos...)

		log.Printf("取得したリポジトリ数: %d (累計: %d)", len(repos), len(allRepos))

		// 次のページがあるか確認
		if resp.NextPage == 0 {
			break // 次のページがない場合は終了
		}

		// 次のページを取得するためにページ番号を更新
		opt.Page = resp.NextPage
	}

	log.Printf("全リポジトリ取得完了: %d 件", len(allRepos))

	// フォークリポジトリの除外処理
	if excludeForks {
		var filteredRepos []*github.Repository
		for _, repo := range allRepos {
			// GetFork() メソッドでフォークかどうかを確認
			if !repo.GetFork() {
				filteredRepos = append(filteredRepos, repo)
			}
		}
		allRepos = filteredRepos
		log.Printf("フォーク除外後のリポジトリ数: %d 件", len(allRepos))
	}

	return allRepos, nil
}

// handleRateLimit API レート制限を検出し、適切に待機する
//
// Preconditions:
// - resp が GitHub API レスポンスであること
//
// Postconditions:
// - レート制限に達している場合は、制限解除まで待機する
//
// Invariants:
// - 待機時間はレスポンスヘッダーから計算される
func handleRateLimit(ctx context.Context, resp *github.Response) error {
	// レート制限の状態を確認
	if resp.Rate.Remaining == 0 {
		// レート制限に達している場合
		// Reset はレート制限がリセットされる時刻（time.Time型）
		resetTime := resp.Rate.Reset.Time
		waitDuration := time.Until(resetTime)

		// 待機時間が負の場合は0にする（既にリセット済み）
		if waitDuration < 0 {
			waitDuration = 0
		}

		// 待機時間に少し余裕を追加（1秒）
		waitDuration += time.Second

		if waitDuration > 0 {
			log.Printf("レート制限に達しました。%v 待機します...", waitDuration)
			select {
			case <-ctx.Done():
				return ctx.Err() // コンテキストがキャンセルされた場合
			case <-time.After(waitDuration):
				// 待機完了
			}
		}
	} else {
		// レート制限に余裕がある場合は、残りリクエスト数をログに記録
		log.Printf("レート制限の残り: %d/%d (リセット時刻: %v)",
			resp.Rate.Remaining,
			resp.Rate.Limit,
			resp.Rate.Reset.Time.Format("2006-01-02 15:04:05"))
	}

	return nil
}
