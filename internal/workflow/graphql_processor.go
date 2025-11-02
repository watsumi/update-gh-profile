package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/watsumi/update-gh-profile/internal/logger"
	"github.com/watsumi/update-gh-profile/internal/repository"

	"github.com/google/go-github/v76/github"
)

// AggregateGraphQLData GraphQLから取得したデータを集計する
func AggregateGraphQLData(ctx context.Context, token string, username, userID string, excludeForks bool) (
	map[string]int, // languageTotals
	map[string]map[string]int, // commitHistories
	map[string]map[int]int, // timeDistributions
	map[string]map[string]int, // allCommitLanguages
	int, // totalCommits
	int, // totalPRs
	[]*github.Repository, // repos (サマリー統計用)
	error,
) {
	logger.Info("GraphQLを使用してリポジトリ情報を一括取得します")

	// 1. リポジトリ情報をGraphQLで一括取得（生成された型を使用）
	repoGraphQLData, err := repository.FetchRepositoriesWithGraphQLGenerated(ctx, token, username, excludeForks)
	if err != nil {
		logger.LogError(err, "GraphQLによるリポジトリ情報の取得に失敗しました")
		return nil, nil, nil, nil, 0, 0, nil, fmt.Errorf("GraphQLによるリポジトリ情報の取得に失敗しました: %w", err)
	}

	logger.Info("%d 個のリポジトリ情報をGraphQLで取得しました", len(repoGraphQLData))

	// 2. ユーザー詳細情報を取得（コミット数、PR数など）（生成された型を使用）
	userDetails, err := repository.FetchUserDetailsWithGraphQLGenerated(ctx, token, username)
	if err != nil {
		logger.LogError(err, "GraphQLによるユーザー詳細情報の取得に失敗しました")
		// エラーは致命的ではないので、続行
	}

	// 3. コミット時間帯を取得（過去1年間）
	since := time.Now().AddDate(-1, 0, 0)
	until := time.Now()
	timeDistribution, err := repository.FetchProductiveTimeWithGraphQL(ctx, token, username, userID, since, until)
	if err != nil {
		logger.LogError(err, "GraphQLによるコミット時間帯の取得に失敗しました")
		timeDistribution = make(map[int]int) // 空のマップで続行
	}

	// 4. コミットごとの言語を取得
	commitLanguages, err := repository.FetchCommitLanguagesWithGraphQL(ctx, token, username)
	if err != nil {
		logger.LogError(err, "GraphQLによるコミット言語情報の取得に失敗しました")
		commitLanguages = make(map[string]map[string]int) // 空のマップで続行
	}

	// 5. データを集計
	languageTotals := make(map[string]int)
	commitHistories := make(map[string]map[string]int)

	// リポジトリごとの言語データを集計
	for _, repo := range repoGraphQLData {
		repoKey := fmt.Sprintf("%s/%s", repo.Owner.Login, repo.Name)

		// 言語データを集計
		for _, lang := range repo.Languages.Nodes {
			languageTotals[lang.Name] += lang.Size
		}

		// コミット履歴を集計（日付ごと）
		if repo.DefaultBranchRef.Target.History.Nodes != nil {
			history := make(map[string]int)
			for _, commit := range repo.DefaultBranchRef.Target.History.Nodes {
				// 日付を取得（YYYY-MM-DD形式）
				date := ""
				if commit.CommittedDate != "" {
					t, err := time.Parse(time.RFC3339, commit.CommittedDate)
					if err == nil {
						date = t.UTC().Format("2006-01-02")
						history[date]++
					}
				} else if commit.Author.Date != "" {
					t, err := time.Parse(time.RFC3339, commit.Author.Date)
					if err == nil {
						date = t.UTC().Format("2006-01-02")
						history[date]++
					}
				}
			}
			if len(history) > 0 {
				commitHistories[repoKey] = history
			}
		}
	}

	// コミット時間帯をリポジトリごとに集計
	timeDistributions := make(map[string]map[int]int)
	// GraphQLから取得した時間帯分布をそのまま使用
	// （全てのリポジトリをまとめた結果なので、1つのエントリとして扱う）
	if len(timeDistribution) > 0 {
		timeDistributions["all"] = timeDistribution
	}

	// コミットごとの言語データを集計
	allCommitLanguages := commitLanguages

	// 総コミット数と総PR数を計算
	var totalCommits, totalPRs, totalStars int
	if userDetails != nil {
		// ユーザー詳細からPR数を取得
		totalPRs = userDetails.PullRequests.TotalCount
		// リポジトリごとのコミット数とスター数を合計
		for _, repo := range repoGraphQLData {
			if repo.DefaultBranchRef.Target.History.TotalCount > 0 {
				totalCommits += repo.DefaultBranchRef.Target.History.TotalCount
			}
			totalStars += repo.StargazerCount
		}
	} else {
		// フォールバック: リポジトリデータから集計
		for _, repo := range repoGraphQLData {
			if repo.DefaultBranchRef.Target.History.TotalCount > 0 {
				totalCommits += repo.DefaultBranchRef.Target.History.TotalCount
			}
			totalStars += repo.StargazerCount
		}
	}

	logger.Info("GraphQLデータの集計が完了しました: 言語数=%d, コミット履歴数=%d, 時間帯分布=%d",
		len(languageTotals), len(commitHistories), len(timeDistributions))

	// GraphQLデータからgithub.Repository構造体を作成（既存のコードとの互換性のため）
	var repos []*github.Repository
	for _, repoData := range repoGraphQLData {
		repo := &github.Repository{
			Name:            github.String(repoData.Name),
			StargazersCount: github.Int(repoData.StargazerCount),
			Owner: &github.User{
				Login: github.String(repoData.Owner.Login),
			},
		}
		repos = append(repos, repo)
	}

	return languageTotals, commitHistories, timeDistributions, allCommitLanguages, totalCommits, totalPRs, repos, nil
}
