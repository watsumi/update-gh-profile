package aggregator

// LanguageStat 言語統計情報
type LanguageStat struct {
	Language       string  // 言語名
	Bytes          int     // 総バイト数
	Percentage     float64 // 全体に占める割合（パーセンテージ）
	RepositoryCount int     // 使用されているリポジトリ数
}

// SummaryStats サマリー統計情報
type SummaryStats struct {
	TotalStars        int // 合計スター数
	RepositoryCount   int // リポジトリ数
	TotalCommits      int // 総コミット数
	TotalPullRequests int // 総プルリクエスト数
}

// AggregatedMetrics 集計されたメトリクス
type AggregatedMetrics struct {
	Languages              []LanguageStat      // ランキング済み言語スライス
	TotalBytes             int                 // 全言語の総バイト数
	RepositoryCount        int                 // 対象リポジトリ数
	CommitHistory          map[string]int      // 日付ごとのコミット数
	CommitTimeDistribution map[int]int         // 時間帯ごとのコミット数
	CommitLanguages        map[string]int      // コミットごとの使用言語Top5
	SummaryStats           SummaryStats        // サマリー統計
}
