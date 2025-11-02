package aggregator

import (
	"log"

	"github.com/google/go-github/v76/github"
)

// AggregateSummaryStats サマリー統計を集計する
//
// Preconditions:
// - repositories がリポジトリ構造体のスライスであること
// - totalCommits が全リポジトリの総コミット数であること
// - totalPRs が全リポジトリの総プルリクエスト数であること
//
// Postconditions:
// - 返される構造体には合計スター数、リポジトリ数、総コミット数、総PR数が含まれる
//
// Invariants:
// - 全リポジトリの値を合算する
// - フォークリポジトリは除外される（repositoriesで既に除外されている前提）
func AggregateSummaryStats(repositories []*github.Repository, totalCommits, totalPRs int) SummaryStats {
	log.Printf("サマリー統計の集計を開始: %d リポジトリ", len(repositories))

	var stats SummaryStats

	// リポジトリ数をカウント（フォークは既に除外されている前提だが、念のため）
	stats.RepositoryCount = 0
	totalStars := 0

	for _, repo := range repositories {
		// フォークリポジトリはスキップ（repositoriesで既に除外されている前提だが、念のため）
		if repo.GetFork() {
			continue
		}

		stats.RepositoryCount++
		totalStars += repo.GetStargazersCount()
	}

	stats.TotalStars = totalStars
	stats.TotalCommits = totalCommits
	stats.TotalPullRequests = totalPRs

	log.Printf("サマリー統計の集計完了:")
	log.Printf("  - 合計スター数: %d", stats.TotalStars)
	log.Printf("  - リポジトリ数: %d", stats.RepositoryCount)
	log.Printf("  - 総コミット数: %d", stats.TotalCommits)
	log.Printf("  - 総プルリクエスト数: %d", stats.TotalPullRequests)

	return stats
}
