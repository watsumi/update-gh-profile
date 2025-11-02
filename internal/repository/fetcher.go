package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/go-github/v56/github"
)

const (
	// maxPages 最大ページ数（無限ループを防ぐため）
	// GitHub API の制限: per_page=100 の場合、100ページで最大10,000リポジトリ
	maxPages = 100
)

// FetchUserRepositories GitHub API を使用して認証ユーザーのリポジトリ一覧を取得する
//
// Preconditions:
// - username が非空文字列であること
// - client が有効な GitHub クライアントであること
// - isAuthenticatedUser が true であること（認証ユーザー自身のみ対応）
//
// Postconditions:
// - フォークを除外する場合は、Fork=false のリポジトリのみが返される
// - スライスはリポジトリ構造体のスライスである
// - 認証ユーザー自身のリポジトリ（プライベート含む）を取得する
//
// Invariants:
// - API レート制限に達した場合は待機して再試行する
// - 認証ユーザー自身のリポジトリのみを取得する
func FetchUserRepositories(ctx context.Context, client *github.Client, username string, excludeForks bool, isAuthenticatedUser bool) ([]*github.Repository, error) {
	if username == "" {
		return nil, fmt.Errorf("username が空です")
	}

	if !isAuthenticatedUser {
		return nil, fmt.Errorf("このツールは認証ユーザー自身のリポジトリのみを取得できます")
	}

	log.Printf("リポジトリ一覧を取得しています: 認証ユーザー=%s, フォーク除外=%v", username, excludeForks)

	// ページネーション用のオプション
	// Type: "all" を指定することで、プライベートリポジトリも含めて取得
	// 参考: https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#list-repositories-for-the-authenticated-user
	opt := &github.RepositoryListOptions{
		Type:      "all", // all, owner, member から選択（認証ユーザーの場合は all でプライベートも取得可能）
		Sort:      "updated",
		Direction: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100, // 1ページあたりの最大件数（GitHub APIの最大値）
			Page:    0,   // 最初のページ（0から開始、または1から開始）
		},
	}

	var allRepos []*github.Repository

	// ページネーションループ: 全ページを取得するまで繰り返す
	// username を空文字列にすると、認証ユーザー自身のリポジトリ（プライベート含む）を取得
	// 参考: https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#list-repositories-for-the-authenticated-user
	for pageNum := 1; pageNum <= maxPages; pageNum++ {
		// ページ番号を設定（GitHub APIは1から開始）
		if pageNum > 1 {
			opt.Page = pageNum
		}

		repos, resp, err := client.Repositories.List(ctx, "", opt)

		if err != nil {
			return nil, fmt.Errorf("リポジトリ一覧の取得に失敗しました: %w", err)
		}

		// レート制限のチェックと処理
		if err := handleRateLimit(ctx, resp); err != nil {
			return nil, fmt.Errorf("レート制限の処理に失敗しました: %w", err)
		}

		// 取得したリポジトリを追加
		allRepos = append(allRepos, repos...)

		// デバッグ: ページネーション情報をログに出力
		log.Printf("取得したリポジトリ数: %d (累計: %d)", len(repos), len(allRepos))
		log.Printf("ページネーション情報: 現在ページ=%d (手動=%d), 次ページ=%d, 最終ページ=%d, PerPage=%d",
			opt.Page, pageNum, resp.NextPage, resp.LastPage, opt.PerPage)

		// 次のページがあるか確認
		// 1. resp.NextPage が 0 でない場合は、GitHub API のレスポンスヘッダーから取得した情報を使用
		// 2. resp.NextPage が 0 の場合でも、取得した件数が PerPage に達している場合は次のページがある可能性がある
		// 3. 取得した件数が30件（GitHub APIのデフォルト）の場合は、次のページがある可能性がある
		// 4. 取得した件数が 0 の場合は、次のページがないと判断
		hasNextPage := false
		if resp.NextPage != 0 {
			hasNextPage = true
			log.Printf("レスポンスヘッダーから次ページ (%d) を検出", resp.NextPage)
		} else if len(repos) >= opt.PerPage {
			// NextPage が 0 でも、取得した件数が PerPage に達している場合は次のページを試みる
			hasNextPage = true
			log.Printf("警告: レスポンスヘッダーから次ページ情報が取得できませんでしたが、取得件数 (%d) が PerPage (%d) に達しているため、次のページを試みます", len(repos), opt.PerPage)
		} else if len(repos) == 30 {
			// GitHub APIのデフォルトのページサイズは30件
			// 30件取得できている場合は、PerPageパラメータが無視されている可能性がある
			// 次のページを試みる
			hasNextPage = true
			log.Printf("警告: レスポンスヘッダーから次ページ情報が取得できませんでしたが、取得件数が30件（GitHub APIのデフォルト）のため、次のページを試みます")
		} else if len(repos) == 0 {
			// 0件取得した場合は終了
			log.Printf("0件取得したため、ページネーションを終了します")
			break
		}

		if !hasNextPage {
			log.Printf("次のページがないため、ページネーションを終了します（取得件数: %d, PerPage: %d）", len(repos), opt.PerPage)
			break // 次のページがない場合は終了
		}

		// 最大ページ数チェック（次のページに進む前に確認）
		if pageNum >= maxPages {
			log.Printf("警告: 最大ページ数 (%d) に達しました。ページネーションを終了します（累計: %d 件）", maxPages, len(allRepos))
			break
		}

		// 次のページ番号を決定
		if resp.NextPage != 0 {
			pageNum = resp.NextPage - 1 // pageNum はループでインクリメントされるため -1
		}
		log.Printf("次のページを取得します（ページ番号: %d / 最大: %d）...", pageNum+1, maxPages)
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

// FetchRepositoryLanguages 指定されたリポジトリの言語統計を取得する
//
// Preconditions:
// - owner と repo が有効なリポジトリ識別子であること
// - client が有効な GitHub クライアントであること
//
// Postconditions:
// - 返される map は map[string]int{言語名: バイト数} の形式である
// - エラー時は nil と error を返す
//
// Invariants:
// - API エラー時は適切なエラーを返す
// - レート制限に達した場合は待機して再試行する
func FetchRepositoryLanguages(ctx context.Context, client *github.Client, owner, repo string) (map[string]int, error) {
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner または repo が空です: owner=%s, repo=%s", owner, repo)
	}

	// GitHub API の /repos/{owner}/{repo}/languages エンドポイントを呼び出す
	// 参考: https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#list-repository-languages
	languages, resp, err := client.Repositories.ListLanguages(ctx, owner, repo)

	if err != nil {
		return nil, fmt.Errorf("リポジトリ %s/%s の言語情報取得に失敗しました: %w", owner, repo, err)
	}

	// レート制限のチェックと処理
	if err := handleRateLimit(ctx, resp); err != nil {
		return nil, fmt.Errorf("レート制限の処理に失敗しました: %w", err)
	}

	// languages は map[string]int{言語名: バイト数} の形式
	return languages, nil
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
