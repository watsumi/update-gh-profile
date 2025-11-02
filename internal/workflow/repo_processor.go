package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/watsumi/update-gh-profile/internal/logger"
	"github.com/watsumi/update-gh-profile/internal/repository"

	"github.com/google/go-github/v76/github"
)

// RepoData リポジトリから取得したデータ
type RepoData struct {
	Owner            string
	RepoName         string
	Languages        map[string]int
	CommitHistory    map[string]int
	TimeDistribution map[int]int
	CommitCount      int
	CommitLanguages  map[string]map[string]int
	PRCount          int
	Error            error
}

// ProcessRepository リポジトリを処理してデータを取得する
func ProcessRepository(ctx context.Context, client *github.Client, rateLimiter *repository.RateLimiter, repo *github.Repository) *RepoData {
	owner := repo.GetOwner().GetLogin()
	repoName := repo.GetName()

	data := &RepoData{
		Owner:            owner,
		RepoName:         repoName,
		Languages:        make(map[string]int),
		CommitHistory:    make(map[string]int),
		TimeDistribution: make(map[int]int),
		CommitLanguages:  make(map[string]map[string]int),
	}

	// レート制限をチェックして待機
	if err := rateLimiter.WaitIfNeeded(ctx); err != nil {
		data.Error = fmt.Errorf("レート制限の待機に失敗しました: %w", err)
		return data
	}

	// 言語データの取得
	langs, err := repository.FetchRepositoryLanguages(ctx, client, owner, repoName)
	if err != nil {
		logger.LogErrorWithContext(err, fmt.Sprintf("%s/%s", owner, repoName), "言語データの取得に失敗しました")
		// 言語データの取得失敗は致命的ではないので、他のデータは取得を試みる
	} else {
		data.Languages = langs
		// レート制限情報を更新（FetchRepositoryLanguages内でレスポンスから取得できる場合はここで更新）
	}

	// コミット履歴の取得
	if err := rateLimiter.WaitIfNeeded(ctx); err != nil {
		data.Error = fmt.Errorf("レート制限の待機に失敗しました: %w", err)
		return data
	}

	commitHistory, err := repository.FetchCommitHistory(ctx, client, owner, repoName)
	if err != nil {
		logger.LogErrorWithContext(err, fmt.Sprintf("%s/%s", owner, repoName), "コミット履歴の取得に失敗しました")
	} else {
		data.CommitHistory = commitHistory
	}

	// コミット時間帯の取得
	if err := rateLimiter.WaitIfNeeded(ctx); err != nil {
		data.Error = fmt.Errorf("レート制限の待機に失敗しました: %w", err)
		return data
	}

	timeDist, err := repository.FetchCommitTimeDistribution(ctx, client, owner, repoName)
	if err != nil {
		logger.LogErrorWithContext(err, fmt.Sprintf("%s/%s", owner, repoName), "コミット時間帯の取得に失敗しました")
	} else {
		data.TimeDistribution = timeDist
	}

	// コミット数の取得
	if err := rateLimiter.WaitIfNeeded(ctx); err != nil {
		data.Error = fmt.Errorf("レート制限の待機に失敗しました: %w", err)
		return data
	}

	commits, err := repository.FetchCommits(ctx, client, owner, repoName)
	if err != nil {
		logger.LogErrorWithContext(err, fmt.Sprintf("%s/%s", owner, repoName), "コミットデータの取得に失敗しました")
	} else {
		data.CommitCount = len(commits)
	}

	// コミットごとの言語取得（コミット数が多い場合はスキップ）
	if data.CommitCount > 0 && data.CommitCount <= 100 {
		if err := rateLimiter.WaitIfNeeded(ctx); err != nil {
			// エラーは無視（オプショナルなデータ）
		} else {
			commitLangs, err := repository.FetchCommitLanguages(ctx, client, owner, repoName)
			if err != nil {
				logger.LogErrorWithContext(err, fmt.Sprintf("%s/%s", owner, repoName), "コミット言語データの取得に失敗しました")
			} else if len(commitLangs) > 0 {
				data.CommitLanguages = commitLangs
			}
		}
	}

	// プルリクエスト数の取得
	if err := rateLimiter.WaitIfNeeded(ctx); err != nil {
		data.Error = fmt.Errorf("レート制限の待機に失敗しました: %w", err)
		return data
	}

	prCount, err := repository.FetchPullRequests(ctx, client, owner, repoName)
	if err != nil {
		logger.LogErrorWithContext(err, fmt.Sprintf("%s/%s", owner, repoName), "プルリクエストデータの取得に失敗しました")
	} else {
		data.PRCount = prCount
	}

	return data
}

// ProcessRepositoriesInParallel リポジトリを並列処理する
func ProcessRepositoriesInParallel(ctx context.Context, client *github.Client, repos []*github.Repository, maxConcurrency int) ([]*RepoData, error) {
	if maxConcurrency <= 0 {
		maxConcurrency = 5 // デフォルト: 5つの並列処理
	}

	rateLimiter := repository.NewRateLimiter(client)
	rateLimiter.SetRequestInterval(150 * time.Millisecond) // 150ms間隔でリクエスト

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrency) // セマフォで並列数を制限
	results := make([]*RepoData, len(repos))
	var mu sync.Mutex

	// 最初のリポジトリでレート制限情報を取得
	if len(repos) > 0 {
		// 認証ユーザー情報を取得してレート制限情報を更新
		// （実際のAPI呼び出しはProcessRepository内で行われる）
	}

	logger.Info("リポジトリを並列処理します: 総数=%d, 最大並列数=%d", len(repos), maxConcurrency)

	for i, repo := range repos {
		wg.Add(1)
		go func(idx int, r *github.Repository) {
			defer wg.Done()

			// セマフォで並列数を制限
			sem <- struct{}{}
			defer func() { <-sem }()

			// リポジトリを処理
			data := ProcessRepository(ctx, client, rateLimiter, r)

			// 結果を保存
			mu.Lock()
			results[idx] = data
			mu.Unlock()

			if data.Error != nil {
				logger.Warning("[%d/%d] %s/%s の処理中にエラーが発生しました: %v",
					idx+1, len(repos), data.Owner, data.RepoName, data.Error)
			} else {
				logger.Debug("[%d/%d] %s/%s の処理が完了しました", idx+1, len(repos), data.Owner, data.RepoName)
			}
			fmt.Printf("  [%d/%d] %s/%s を処理完了\n", idx+1, len(repos), data.Owner, data.RepoName)
		}(i, repo)
	}

	wg.Wait()
	logger.Info("すべてのリポジトリの処理が完了しました")

	return results, nil
}
