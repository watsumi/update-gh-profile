package aggregator

import (
	"log"
	"sort"
)

// AggregateCommitHistory 日付ごとのコミット数を集計する
//
// Preconditions:
// - commitHistories が map[string]map[string]int{リポジトリ名: {日付: コミット数}} の形式であること
//
// Postconditions:
// - 返される map は map[string]int{日付: 合計コミット数} の形式である
// - 日付は YYYY-MM-DD 形式で記録される
//
// Invariants:
// - 全リポジトリの日付ごとのコミット数が合算される
func AggregateCommitHistory(commitHistories map[string]map[string]int) map[string]int {
	log.Printf("コミット履歴の集計を開始: %d リポジトリ", len(commitHistories))

	// 日付ごとの合計コミット数を格納する map
	aggregated := make(map[string]int)

	// 各リポジトリのコミット履歴を集計
	for repoName, history := range commitHistories {
		log.Printf("  %s: %d 日分のコミット履歴を集計中", repoName, len(history))
		for date, count := range history {
			aggregated[date] += count
		}
	}

	log.Printf("コミット履歴の集計完了: %d 日分", len(aggregated))
	return aggregated
}

// SortCommitHistoryByDate コミット履歴を日付順でソートする
//
// Preconditions:
// - commitHistory が map[string]int{日付: コミット数} の形式であること
//
// Postconditions:
// - 返されるスライスは日付順（昇順）にソートされた日付とコミット数のペアのスライスである
//
// Invariants:
// - 日付は YYYY-MM-DD 形式で記録される
// - ソート順序は日付の文字列比較で行われる（YYYY-MM-DD形式なので正しくソートされる）
func SortCommitHistoryByDate(commitHistory map[string]int) []DateCommitPair {
	if len(commitHistory) == 0 {
		return []DateCommitPair{}
	}

	// DateCommitPair スライスを作成
	var pairs []DateCommitPair
	for date, count := range commitHistory {
		pairs = append(pairs, DateCommitPair{
			Date:  date,
			Count: count,
		})
	}

	// 日付順（昇順）でソート
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Date < pairs[j].Date
	})

	return pairs
}

// DateCommitPair 日付とコミット数のペア
type DateCommitPair struct {
	Date  string // 日付（YYYY-MM-DD形式）
	Count int    // コミット数
}

// AggregateCommitTimeDistribution 時間帯ごとのコミット数を集計する
//
// Preconditions:
// - timeDistributions が map[string]map[int]int{リポジトリ名: {時間帯: コミット数}} の形式であること
//
// Postconditions:
// - 返される map は map[int]int{時間帯: 合計コミット数} の形式である
// - 時間帯は 0-23 の範囲で記録される
//
// Invariants:
// - 全リポジトリの時間帯ごとのコミット数が合算される
func AggregateCommitTimeDistribution(timeDistributions map[string]map[int]int) map[int]int {
	log.Printf("コミット時間帯分布の集計を開始: %d リポジトリ", len(timeDistributions))

	// 時間帯ごとの合計コミット数を格納する map（0-23時）
	aggregated := make(map[int]int)

	// 各リポジトリの時間帯分布を集計
	for repoName, distribution := range timeDistributions {
		log.Printf("  %s: %d 時間帯分のデータを集計中", repoName, len(distribution))
		for hour, count := range distribution {
			// 時間帯が0-23の範囲内であることを確認
			if hour < 0 || hour > 23 {
				log.Printf("警告: リポジトリ %s の時間帯 %d が範囲外です。スキップします", repoName, hour)
				continue
			}
			aggregated[hour] += count
		}
	}

	log.Printf("コミット時間帯分布の集計完了: %d 時間帯", len(aggregated))
	return aggregated
}

// SortCommitTimeDistributionByHour コミット時間帯分布を時間帯順でソートする
//
// Preconditions:
// - timeDistribution が map[int]int{時間帯: コミット数} の形式であること
//
// Postconditions:
// - 返されるスライスは時間帯順（昇順、0-23時）にソートされた時間帯とコミット数のペアのスライスである
//
// Invariants:
// - 時間帯は 0-23 の範囲で記録される
func SortCommitTimeDistributionByHour(timeDistribution map[int]int) []HourCommitPair {
	if len(timeDistribution) == 0 {
		return []HourCommitPair{}
	}

	// HourCommitPair スライスを作成
	var pairs []HourCommitPair
	for hour, count := range timeDistribution {
		// 範囲チェック
		if hour < 0 || hour > 23 {
			continue
		}
		pairs = append(pairs, HourCommitPair{
			Hour:  hour,
			Count: count,
		})
	}

	// 時間帯順（昇順）でソート
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Hour < pairs[j].Hour
	})

	return pairs
}

// HourCommitPair 時間帯とコミット数のペア
type HourCommitPair struct {
	Hour  int // 時間帯（0-23時）
	Count int // コミット数
}
