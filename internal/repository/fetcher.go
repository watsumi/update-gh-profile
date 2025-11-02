package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/v56/github"
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
	if err := ValidateUsername(username); err != nil {
		return nil, err
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
			PerPage: DefaultPerPage,
			Page:    0, // 最初のページ
		},
	}

	var allRepos []*github.Repository

	// ページネーションループ: 全ページを取得するまで繰り返す
	// username を空文字列にすると、認証ユーザー自身のリポジトリ（プライベート含む）を取得
	// 参考: https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#list-repositories-for-the-authenticated-user
	for pageNum := 1; pageNum <= MaxPages; pageNum++ {
		// ページ番号を設定（GitHub APIは1から開始）
		if pageNum > 1 {
			opt.Page = pageNum
		}

		repos, resp, err := client.Repositories.List(ctx, "", opt)

		if err != nil {
			return nil, fmt.Errorf("リポジトリ一覧の取得に失敗しました: %w", err)
		}

		// レート制限のチェックと処理
		if err := HandleRateLimit(ctx, resp); err != nil {
			return nil, fmt.Errorf("レート制限の処理に失敗しました: %w", err)
		}

		// 取得したリポジトリを追加
		allRepos = append(allRepos, repos...)

		// デバッグ: ページネーション情報をログに出力
		log.Printf("取得したリポジトリ数: %d (累計: %d)", len(repos), len(allRepos))
		log.Printf("ページネーション情報: 現在ページ=%d (手動=%d), 次ページ=%d, 最終ページ=%d, PerPage=%d",
			opt.Page, pageNum, resp.NextPage, resp.LastPage, opt.PerPage)

		// 次のページがあるか確認（共通関数を使用）
		paginationResult := CheckPagination(resp, len(repos), opt.PerPage)

		if !paginationResult.HasNextPage {
			log.Printf("次のページがないため、ページネーションを終了します（取得件数: %d, PerPage: %d）", len(repos), opt.PerPage)
			break
		}

		// 最大ページ数チェック（次のページに進む前に確認）
		if pageNum >= MaxPages {
			log.Printf("警告: 最大ページ数 (%d) に達しました。ページネーションを終了します（累計: %d 件）", MaxPages, len(allRepos))
			break
		}

		// 次のページ番号を決定
		if paginationResult.NextPageNum != 0 {
			pageNum = paginationResult.NextPageNum - 1 // pageNum はループでインクリメントされるため -1
		}
		log.Printf("次のページを取得します（ページ番号: %d / 最大: %d）...", pageNum+1, MaxPages)
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
	if err := ValidateOwnerAndRepo(owner, repo); err != nil {
		return nil, err
	}

	// GitHub API の /repos/{owner}/{repo}/languages エンドポイントを呼び出す
	// 参考: https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#list-repository-languages
	languages, resp, err := client.Repositories.ListLanguages(ctx, owner, repo)

	if err != nil {
		return nil, fmt.Errorf("リポジトリ %s/%s の言語情報取得に失敗しました: %w", owner, repo, err)
	}

	// レート制限のチェックと処理
	if err := HandleRateLimit(ctx, resp); err != nil {
		return nil, fmt.Errorf("レート制限の処理に失敗しました: %w", err)
	}

	// languages は map[string]int{言語名: バイト数} の形式
	return languages, nil
}

// FetchCommits 指定されたリポジトリのコミット履歴を取得する
//
// Preconditions:
// - owner と repo が有効なリポジトリ識別子であること
// - client が有効な GitHub クライアントであること
//
// Postconditions:
// - 返されるスライスはコミット情報のリストである
// - ページネーションを使用して全コミットを取得する（最大100ページまで）
//
// Invariants:
// - API レート制限に達した場合は待機して再試行する
func FetchCommits(ctx context.Context, client *github.Client, owner, repo string) ([]*github.RepositoryCommit, error) {
	if err := ValidateOwnerAndRepo(owner, repo); err != nil {
		return nil, err
	}

	log.Printf("リポジトリ %s/%s のコミット履歴を取得しています...", owner, repo)

	// ページネーション用のオプション
	opt := &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			PerPage: DefaultPerPage,
			Page:    0, // 最初のページ
		},
	}

	var allCommits []*github.RepositoryCommit

	// ページネーションループ: 全ページを取得するまで繰り返す
	// 参考: https://docs.github.com/en/rest/commits/commits?apiVersion=2022-11-28#list-commits
	for pageNum := 1; pageNum <= MaxPages; pageNum++ {
		// ページ番号を設定
		if pageNum > 1 {
			opt.Page = pageNum
		}

		commits, resp, err := client.Repositories.ListCommits(ctx, owner, repo, opt)

		if err != nil {
			return nil, fmt.Errorf("リポジトリ %s/%s のコミット履歴取得に失敗しました: %w", owner, repo, err)
		}

		// レート制限のチェックと処理
		if err := HandleRateLimit(ctx, resp); err != nil {
			return nil, fmt.Errorf("レート制限の処理に失敗しました: %w", err)
		}

		// 取得したコミットを追加
		allCommits = append(allCommits, commits...)

		log.Printf("取得したコミット数: %d (累計: %d)", len(commits), len(allCommits))

		// 次のページがあるか確認（共通関数を使用）
		paginationResult := CheckPagination(resp, len(commits), opt.PerPage)

		if !paginationResult.HasNextPage {
			log.Printf("次のページがないため、ページネーションを終了します（取得件数: %d）", len(commits))
			break
		}

		// 最大ページ数チェック
		if pageNum >= MaxPages {
			log.Printf("警告: 最大ページ数 (%d) に達しました。ページネーションを終了します（累計: %d 件）", MaxPages, len(allCommits))
			break
		}

		// 次のページ番号を決定
		if paginationResult.NextPageNum != 0 {
			pageNum = paginationResult.NextPageNum - 1 // pageNum はループでインクリメントされるため -1
		}
	}

	log.Printf("リポジトリ %s/%s のコミット履歴取得完了: %d 件", owner, repo, len(allCommits))
	return allCommits, nil
}

// FetchCommitHistory 指定されたリポジトリの日付ごとのコミット数を取得する
//
// Preconditions:
// - owner と repo が有効なリポジトリ識別子であること
// - client が有効な GitHub クライアントであること
//
// Postconditions:
// - 返される map は map[string]int{日付(YYYY-MM-DD): コミット数} の形式である
//
// Invariants:
// - 日付は YYYY-MM-DD 形式で記録される
// - 日付は UTC で記録される
func FetchCommitHistory(ctx context.Context, client *github.Client, owner, repo string) (map[string]int, error) {
	commits, err := FetchCommits(ctx, client, owner, repo)
	if err != nil {
		return nil, err
	}

	// 日付ごとのコミット数を集計
	history := make(map[string]int)
	for _, commit := range commits {
		// コミットの日時を取得（コミッターの日時を使用）
		if commit.Commit == nil || commit.Commit.Committer == nil || commit.Commit.Committer.Date == nil {
			continue
		}

		// UTC で日付を取得（YYYY-MM-DD形式）
		date := commit.Commit.Committer.Date.Time.UTC()
		dateStr := date.Format("2006-01-02")
		history[dateStr]++
	}

	log.Printf("リポジトリ %s/%s のコミット履歴集計完了: %d 日分", owner, repo, len(history))
	return history, nil
}

// FetchCommitTimeDistribution 指定されたリポジトリの時間帯ごとのコミット数を取得する
//
// Preconditions:
// - owner と repo が有効なリポジトリ識別子であること
// - client が有効な GitHub クライアントであること
//
// Postconditions:
// - 返される map は map[int]int{時間帯(0-23): コミット数} の形式である
// - 時間帯は UTC で集計される
//
// Invariants:
// - 時間帯は 0-23 の範囲で記録される
func FetchCommitTimeDistribution(ctx context.Context, client *github.Client, owner, repo string) (map[int]int, error) {
	commits, err := FetchCommits(ctx, client, owner, repo)
	if err != nil {
		return nil, err
	}

	// 時間帯ごとのコミット数を集計（0-23時）
	distribution := make(map[int]int)
	for _, commit := range commits {
		// コミットの日時を取得（コミッターの日時を使用）
		if commit.Commit == nil || commit.Commit.Committer == nil || commit.Commit.Committer.Date == nil {
			continue
		}

		// UTC で時間帯を取得（0-23時）
		date := commit.Commit.Committer.Date.Time.UTC()
		hour := date.Hour()
		distribution[hour]++
	}

	log.Printf("リポジトリ %s/%s のコミット時間帯分布集計完了: %d 時間帯", owner, repo, len(distribution))
	return distribution, nil
}

// FetchCommitLanguages 指定されたリポジトリのコミットごとの使用言語を取得する
//
// Preconditions:
// - owner と repo が有効なリポジトリ識別子であること
// - client が有効な GitHub クライアントであること
//
// Postconditions:
// - 返される map は map[string]map[string]int{コミットSHA: {言語名: 出現回数}} の形式である
// - 各コミットに対して、変更されたファイルから言語を抽出する
//
// Invariants:
// - 各コミットの変更ファイルから言語を抽出する
// - 言語名は大文字小文字を区別する（Go, Python など）
func FetchCommitLanguages(ctx context.Context, client *github.Client, owner, repo string) (map[string]map[string]int, error) {
	if err := ValidateOwnerAndRepo(owner, repo); err != nil {
		return nil, err
	}

	log.Printf("リポジトリ %s/%s のコミットごとの言語使用状況を取得しています...", owner, repo)

	// まずコミット一覧を取得
	commits, err := FetchCommits(ctx, client, owner, repo)
	if err != nil {
		return nil, err
	}

	// コミットごとの言語使用状況を格納する map
	// map[コミットSHA]map[言語名]出現回数
	commitLanguages := make(map[string]map[string]int)

	// 各コミットの詳細情報を取得（変更ファイル情報を含む）
	maxCommits := MaxCommitsForLanguageDetection
	if len(commits) < maxCommits {
		maxCommits = len(commits)
	}

	log.Printf("コミットごとの言語情報を取得中: %d 件のコミットを処理します", maxCommits)

	for i := 0; i < maxCommits; i++ {
		commit := commits[i]
		sha := commit.GetSHA()

		// コミットの詳細情報を取得（変更ファイル情報を含む）
		commitDetail, resp, err := client.Repositories.GetCommit(ctx, owner, repo, sha, &github.ListOptions{})
		if err != nil {
			log.Printf("警告: コミット %s の詳細取得に失敗しました: %v", sha[:7], err)
			continue
		}

		// レート制限のチェックと処理
		if err := HandleRateLimit(ctx, resp); err != nil {
			return nil, fmt.Errorf("レート制限の処理に失敗しました: %w", err)
		}

		// このコミットで使用された言語を集計
		langs := make(map[string]int)

		if commitDetail.Files != nil {
			for _, file := range commitDetail.Files {
				// ファイル名から言語を判定（共通関数を使用）
				lang := DetectLanguageFromFilename(file.GetFilename())
				if lang != "" {
					langs[lang]++
				}
			}
		}

		if len(langs) > 0 {
			commitLanguages[sha] = langs
		}

		// 進捗をログに出力（10コミットごと）
		if (i+1)%10 == 0 {
			log.Printf("進捗: %d/%d コミットを処理しました", i+1, maxCommits)
		}
	}

	log.Printf("リポジトリ %s/%s のコミットごとの言語使用状況取得完了: %d コミット", owner, repo, len(commitLanguages))
	return commitLanguages, nil
}

// FetchPullRequests 指定されたリポジトリのプルリクエスト数を取得する
//
// Preconditions:
// - owner と repo が有効なリポジトリ識別子であること
// - client が有効な GitHub クライアントであること
//
// Postconditions:
// - 返される値はプルリクエストの総数である
// - すべての状態（open, closed, all）のプルリクエストを集計する
//
// Invariants:
// - ページネーションを使用して全PRを取得する（最大100ページまで）
// - API レート制限に達した場合は待機して再試行する
func FetchPullRequests(ctx context.Context, client *github.Client, owner, repo string) (int, error) {
	if err := ValidateOwnerAndRepo(owner, repo); err != nil {
		return 0, err
	}

	log.Printf("リポジトリ %s/%s のプルリクエスト数を取得しています...", owner, repo)

	// ページネーション用のオプション
	// State: "all" を指定することで、すべての状態（open, closed）のPRを取得
	// 参考: https://docs.github.com/en/rest/pulls/pulls?apiVersion=2022-11-28#list-pull-requests
	opt := &github.PullRequestListOptions{
		State: "all", // "open", "closed", "all" から選択
		ListOptions: github.ListOptions{
			PerPage: DefaultPerPage,
			Page:    0, // 最初のページ
		},
	}

	var totalCount int

	// ページネーションループ: 全ページを取得するまで繰り返す
	for pageNum := 1; pageNum <= MaxPages; pageNum++ {
		// ページ番号を設定
		if pageNum > 1 {
			opt.Page = pageNum
		}

		pullRequests, resp, err := client.PullRequests.List(ctx, owner, repo, opt)

		if err != nil {
			return 0, fmt.Errorf("リポジトリ %s/%s のプルリクエスト取得に失敗しました: %w", owner, repo, err)
		}

		// レート制限のチェックと処理
		if err := HandleRateLimit(ctx, resp); err != nil {
			return 0, fmt.Errorf("レート制限の処理に失敗しました: %w", err)
		}

		// 取得したPR数を追加
		totalCount += len(pullRequests)

		log.Printf("取得したプルリクエスト数: %d (累計: %d)", len(pullRequests), totalCount)

		// 次のページがあるか確認（共通関数を使用）
		paginationResult := CheckPagination(resp, len(pullRequests), opt.PerPage)

		if !paginationResult.HasNextPage {
			log.Printf("次のページがないため、ページネーションを終了します（取得件数: %d）", len(pullRequests))
			break
		}

		// 最大ページ数チェック
		if pageNum >= MaxPages {
			log.Printf("警告: 最大ページ数 (%d) に達しました。ページネーションを終了します（累計: %d 件）", MaxPages, totalCount)
			break
		}

		// 次のページ番号を決定
		if paginationResult.NextPageNum != 0 {
			pageNum = paginationResult.NextPageNum - 1 // pageNum はループでインクリメントされるため -1
		}
	}

	log.Printf("リポジトリ %s/%s のプルリクエスト数取得完了: %d 件", owner, repo, totalCount)
	return totalCount, nil
}
