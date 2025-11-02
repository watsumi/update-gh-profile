package aggregator

import (
	"log"
	"sort"
)

// AggregateCommitLanguages コミットごとの使用言語Top5を集計する
//
// Preconditions:
// - commitLanguages が map[string]map[string]map[string]int{リポジトリ名: {コミットSHA: {言語名: 出現回数}}} の形式であること
//   または map[string]map[string]int{コミットSHA: {言語名: 出現回数}} の形式であること
//
// Postconditions:
// - 返される map は map[string]int{言語名: 使用回数} の形式で、Top5のみを含む
// - 使用回数が多い順にソートされている
//
// Invariants:
// - 使用回数が多い順にソートされ、上位5つが返される
func AggregateCommitLanguages(commitLanguages map[string]map[string]int) map[string]int {
	log.Printf("コミットごとの言語使用状況の集計を開始: %d コミット", len(commitLanguages))

	// 言語ごとの使用回数を集計する map
	languageCounts := make(map[string]int)

	// 各コミットの言語使用状況を集計
	for commitSHA, langs := range commitLanguages {
		// SHAが短い場合の処理
		shaDisplay := commitSHA
		if len(commitSHA) > 7 {
			shaDisplay = commitSHA[:7]
		}
		log.Printf("  コミット %s: %d 言語を使用", shaDisplay, len(langs))
		for lang, count := range langs {
			languageCounts[lang] += count
		}
	}

	log.Printf("言語ごとの使用回数集計完了: %d 言語", len(languageCounts))

	// 使用回数でソートしてTop5を抽出
	top5 := extractTop5Languages(languageCounts)

	log.Printf("コミットごとの言語Top5集計完了: %d 言語", len(top5))
	return top5
}

// extractTop5Languages 言語ごとの使用回数からTop5を抽出する
//
// Preconditions:
// - languageCounts が map[string]int{言語名: 使用回数} の形式であること
//
// Postconditions:
// - 返される map は使用回数が多い順にソートされたTop5のみを含む
//
// Invariants:
// - 使用回数が同じ場合は、言語名の辞書順でソートされる
func extractTop5Languages(languageCounts map[string]int) map[string]int {
	if len(languageCounts) == 0 {
		return make(map[string]int)
	}

	// 言語と使用回数のペアのスライスを作成
	type langCount struct {
		lang  string
		count int
	}
	var langList []langCount
	for lang, count := range languageCounts {
		langList = append(langList, langCount{lang: lang, count: count})
	}

	// 使用回数降順でソート（使用回数が同じ場合は言語名昇順）
	sort.Slice(langList, func(i, j int) bool {
		if langList[i].count != langList[j].count {
			return langList[i].count > langList[j].count
		}
		return langList[i].lang < langList[j].lang
	})

	// Top5を抽出
	maxCount := 5
	if len(langList) < maxCount {
		maxCount = len(langList)
	}

	result := make(map[string]int)
	for i := 0; i < maxCount; i++ {
		result[langList[i].lang] = langList[i].count
	}

	return result
}
