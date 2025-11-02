package repository

// GitHub API 関連の定数
const (
	// MaxPages 最大ページ数（無限ループを防ぐため）
	// GitHub API の制限: per_page=100 の場合、100ページで最大10,000リポジトリ
	MaxPages = 100

	// DefaultPerPage 1ページあたりのデフォルト件数
	DefaultPerPage = 100

	// DefaultPageSize GitHub APIのデフォルトページサイズ（PerPageパラメータが無視された場合）
	DefaultPageSize = 30

	// MaxCommitsForLanguageDetection コミットごとの言語判定で処理する最大コミット数
	// パフォーマンスを考慮して制限を設ける
	MaxCommitsForLanguageDetection = 100
)
