package aggregator

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/google/go-github/v76/github"
)

// AggregateLanguages 全リポジトリの言語データを集計する
//
// Preconditions:
// - repositories がリポジトリ構造体のスライスであること
// - languageData が map[string]map[string]int{リポジトリ名: {言語: バイト数}} の形式であること
//
// Postconditions:
// - 返される map は map[string]int{言語名: 総バイト数} の形式である
// - フォークされたリポジトリのデータは除外されている（repositoriesで既に除外されている前提）
//
// Invariants:
// - 同じ言語のデータは合算される
// - フォークリポジトリは除外される
func AggregateLanguages(repositories []*github.Repository, languageData map[string]map[string]int) map[string]int {
	log.Printf("言語データの集計を開始: %d リポジトリ", len(repositories))

	// 言語ごとの総バイト数を集計する map
	languageTotals := make(map[string]int)
	// 言語ごとの使用リポジトリ数をカウントする map
	languageRepoCounts := make(map[string]map[string]bool)

	for _, repo := range repositories {
		// フォークリポジトリはスキップ（repositoriesで既に除外されている前提だが、念のため）
		if repo.GetFork() {
			continue
		}

		// リポジトリ名を生成（owner/repo形式）
		repoKey := fmt.Sprintf("%s/%s", repo.GetOwner().GetLogin(), repo.GetName())

		// このリポジトリの言語データを取得
		langs, ok := languageData[repoKey]
		if !ok {
			continue // 言語データがない場合はスキップ
		}

		// 各言語のバイト数を合算
		for lang, bytes := range langs {
			languageTotals[lang] += bytes

			// 言語ごとの使用リポジトリ数をカウント
			if languageRepoCounts[lang] == nil {
				languageRepoCounts[lang] = make(map[string]bool)
			}
			languageRepoCounts[lang][repoKey] = true
		}
	}

	log.Printf("言語データの集計完了: %d 言語", len(languageTotals))
	return languageTotals
}

// RankLanguages 言語データをランキング化する
//
// Preconditions:
// - languageTotals が map[string]int{言語名: 総バイト数} の形式であること
//
// Postconditions:
// - 返されるスライスは LanguageStat 構造体のスライスで、バイト数降順にソートされている
// - 各 LanguageStat にはパーセンテージが含まれる
//
// Invariants:
// - パーセンテージの合計は 100% になる（丸め誤差を除く）
func RankLanguages(languageTotals map[string]int) []LanguageStat {
	if len(languageTotals) == 0 {
		return []LanguageStat{}
	}

	// 総バイト数を計算
	totalBytes := 0
	for _, bytes := range languageTotals {
		totalBytes += bytes
	}

	if totalBytes == 0 {
		log.Printf("警告: 総バイト数が0です")
		return []LanguageStat{}
	}

	// LanguageStat スライスを作成
	var ranked []LanguageStat
	for lang, bytes := range languageTotals {
		percentage := float64(bytes) / float64(totalBytes) * 100.0
		ranked = append(ranked, LanguageStat{
			Language:   lang,
			Bytes:      bytes,
			Percentage: percentage,
		})
	}

	// バイト数降順でソート
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Bytes > ranked[j].Bytes
	})

	log.Printf("言語ランキング生成完了: %d 言語（総バイト数: %d）", len(ranked), totalBytes)
	return ranked
}

// FilterMinorLanguages 閾値以下の言語を除外する
//
// Preconditions:
// - rankedLanguages がランキング済み言語スライスであること
// - threshold が 0 以上 100 以下であること
//
// Postconditions:
// - 返されるスライスは閾値以上のパーセンテージを持つ言語のみを含む
//
// Invariants:
// - 元のスライスの順序が保持される
func FilterMinorLanguages(rankedLanguages []LanguageStat, threshold float64) []LanguageStat {
	if threshold < 0 || threshold > 100 {
		log.Printf("警告: 閾値が範囲外です（%f）。すべての言語を含みます", threshold)
		return rankedLanguages
	}

	var filtered []LanguageStat
	for _, lang := range rankedLanguages {
		if lang.Percentage >= threshold {
			filtered = append(filtered, lang)
		}
	}

	log.Printf("閾値（%.2f%%）によるフィルタリング完了: %d 言語 → %d 言語", threshold, len(rankedLanguages), len(filtered))
	return filtered
}

// FilterExcludedLanguages 指定された言語を除外する
//
// Preconditions:
// - rankedLanguages がランキング済み言語スライスであること
// - excludedLanguages が除外する言語名のスライスであること（空でも可）
//
// Postconditions:
// - 返されるスライスは除外リストに含まれていない言語のみを含む
//
// Invariants:
// - 元のスライスの順序が保持される
// - 大文字小文字を区別せずに比較する
func FilterExcludedLanguages(rankedLanguages []LanguageStat, excludedLanguages []string) []LanguageStat {
	if len(excludedLanguages) == 0 {
		return rankedLanguages
	}

	// 除外リストを大文字小文字を区別しない比較のためにmapに変換
	excludedMap := make(map[string]bool)
	for _, lang := range excludedLanguages {
		// 空白を削除し、大文字小文字を区別しない比較のために小文字に変換
		normalized := strings.TrimSpace(strings.ToLower(lang))
		if normalized != "" {
			excludedMap[normalized] = true
		}
	}

	var filtered []LanguageStat
	for _, lang := range rankedLanguages {
		// 言語名を正規化して比較
		normalized := strings.ToLower(lang.Language)
		if !excludedMap[normalized] {
			filtered = append(filtered, lang)
		}
	}

	log.Printf("除外言語によるフィルタリング完了: %d 言語 → %d 言語（除外: %v）", len(rankedLanguages), len(filtered), excludedLanguages)
	return filtered
}
