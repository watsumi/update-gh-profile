package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/watsumi/update-gh-profile/internal/aggregator"
	"github.com/watsumi/update-gh-profile/internal/generator"
	"github.com/watsumi/update-gh-profile/internal/git"
	"github.com/watsumi/update-gh-profile/internal/logger"
	"github.com/watsumi/update-gh-profile/internal/readme"
	"github.com/watsumi/update-gh-profile/internal/repository"

	"github.com/google/go-github/v56/github"
)

// Config ワークフロー設定
type Config struct {
	RepoPath        string          // リポジトリパス（README.md がある場所）
	SVGOutputDir    string          // SVG ファイルの出力ディレクトリ
	Timezone        string          // タイムゾーン（例: "Asia/Tokyo", "UTC"）
	CommitMessage   string          // Git コミットメッセージ
	EnableGitPush   bool            // Git プッシュを有効にするか
	MaxRepositories int             // 処理する最大リポジトリ数（0 = すべて）
	ExcludeForks    bool            // フォークリポジトリを除外するか
	LogLevel        logger.LogLevel // ログレベル
}

// Run メイン処理フローを実行する
//
// Preconditions:
// - ctx が有効な context.Context であること
// - client が初期化された GitHub API クライアントであること
// - config が有効な Config 構造体であること
//
// Postconditions:
// - README.md が更新される
// - SVG ファイルが生成・保存される
// - 変更があれば Git コミット・プッシュされる
//
// Invariants:
// - エラーが発生した場合は適切に処理される
func Run(ctx context.Context, client *github.Client, config Config) error {
	// ロガーの設定
	if config.LogLevel != 0 {
		logger.DefaultLogger.SetLevel(config.LogLevel)
	}

	logger.Info("ワークフローを開始します")

	// 認証ユーザーを取得
	authUser, _, err := client.Users.Get(ctx, "")
	if err != nil {
		logger.LogError(err, "認証ユーザー情報の取得に失敗しました")
		return fmt.Errorf("認証ユーザー情報の取得に失敗しました: %w", err)
	}
	username := authUser.GetLogin()
	logger.Info("認証ユーザー: %s", username)

	// 1. リポジトリ一覧の取得
	logger.Info("リポジトリ一覧を取得しています...")
	fmt.Println("📦 リポジトリ一覧を取得しています...")
	repos, err := repository.FetchUserRepositories(ctx, client, username, config.ExcludeForks, true)
	if err != nil {
		logger.LogError(err, "リポジトリ一覧の取得に失敗しました")
		return fmt.Errorf("リポジトリ一覧の取得に失敗しました: %w", err)
	}

	if len(repos) == 0 {
		logger.Warning("リポジトリが見つかりませんでした")
		return fmt.Errorf("リポジトリが見つかりませんでした")
	}

	logger.Info("%d 個のリポジトリを取得しました", len(repos))
	fmt.Printf("✅ %d 個のリポジトリを取得しました\n", len(repos))

	// 最大リポジトリ数の制限
	if config.MaxRepositories > 0 && len(repos) > config.MaxRepositories {
		repos = repos[:config.MaxRepositories]
		fmt.Printf("📊 最初の %d 個のリポジトリのみを処理します\n", config.MaxRepositories)
	}

	// 2. データの取得と集計（並列処理）
	fmt.Println("\n📊 リポジトリデータを取得・集計しています...")
	logger.Info("リポジトリを並列処理します: 総数=%d", len(repos))

	// 並列処理でリポジトリデータを取得
	maxConcurrency := 5 // 最大並列数（環境変数から設定可能にする場合の拡張ポイント）
	repoDataList, err := ProcessRepositoriesInParallel(ctx, client, repos, maxConcurrency)
	if err != nil {
		logger.LogError(err, "リポジトリの並列処理に失敗しました")
		return fmt.Errorf("リポジトリの並列処理に失敗しました: %w", err)
	}

	// 取得したデータを集計
	languageTotals := make(map[string]int)
	commitHistories := make(map[string]map[string]int)    // repoKey -> date -> count
	timeDistributions := make(map[string]map[int]int)     // repoKey -> hour -> count
	allCommitLanguages := make(map[string]map[string]int) // repoKey -> commitSHA -> languages
	var totalCommits, totalPRs int

	for _, data := range repoDataList {
		if data == nil {
			continue
		}

		repoKey := fmt.Sprintf("%s/%s", data.Owner, data.RepoName)

		// 言語データを集計
		for lang, bytes := range data.Languages {
			languageTotals[lang] += bytes
		}

		// コミット履歴を集計
		if len(data.CommitHistory) > 0 {
			commitHistories[repoKey] = data.CommitHistory
			logger.Debug("%s: %d 日分のコミット履歴を取得しました", repoKey, len(data.CommitHistory))
		}

		// コミット時間帯を集計
		if len(data.TimeDistribution) > 0 {
			timeDistributions[repoKey] = data.TimeDistribution
		}

		// コミット数を集計
		totalCommits += data.CommitCount
		if data.CommitCount > 0 {
			logger.Debug("%s: %d コミットを取得しました", repoKey, data.CommitCount)
		}

		// コミットごとの言語を集計
		if len(data.CommitLanguages) > 0 {
			for sha, langs := range data.CommitLanguages {
				uniqueSHA := fmt.Sprintf("%s:%s", repoKey, sha)
				allCommitLanguages[uniqueSHA] = langs
			}
		}

		// プルリクエスト数を集計
		totalPRs += data.PRCount
		if data.PRCount > 0 {
			logger.Debug("%s: %d プルリクエストを取得しました", repoKey, data.PRCount)
		}
	}

	// 3. データの集計とランキング生成
	fmt.Println("\n📈 データを集計・ランキング生成中...")

	// 言語ランキング
	var rankedLanguages []aggregator.LanguageStat
	if len(languageTotals) > 0 {
		rankedLanguages = aggregator.RankLanguages(languageTotals)
		rankedLanguages = aggregator.FilterMinorLanguages(rankedLanguages, 1.0) // 1%以上の言語のみ
	}

	// コミット履歴の集計
	logger.Info("コミット履歴を集計しています...")
	aggregatedHistoryMap := aggregator.AggregateCommitHistory(commitHistories)
	aggregatedHistory := aggregator.SortCommitHistoryByDate(aggregatedHistoryMap)
	logger.Info("コミット履歴の集計が完了しました: %d 日分", len(aggregatedHistory))

	// コミット時間帯の集計
	logger.Info("コミット時間帯を集計しています...")
	aggregatedTimeDistMap := aggregator.AggregateCommitTimeDistribution(timeDistributions)
	aggregatedTimeDist := aggregator.SortCommitTimeDistributionByHour(aggregatedTimeDistMap)
	logger.Info("コミット時間帯の集計が完了しました: %d 時間帯", len(aggregatedTimeDist))

	// コミットごとの言語Top5
	top5Languages := aggregator.AggregateCommitLanguages(allCommitLanguages)

	// サマリー統計
	summaryStats := aggregator.AggregateSummaryStats(repos, totalCommits, totalPRs)

	// 4. SVG グラフの生成
	fmt.Println("\n🎨 SVG グラフを生成しています...")

	svgOutputDir := config.SVGOutputDir
	if svgOutputDir == "" {
		svgOutputDir = "."
	}

	// 出力ディレクトリの作成
	err = os.MkdirAll(svgOutputDir, 0755)
	if err != nil {
		return fmt.Errorf("出力ディレクトリの作成に失敗しました: %w", err)
	}

	svgs := make(map[string]string)

	// 言語ランキング SVG
	if len(rankedLanguages) > 0 {
		langSVG, err := generator.GenerateLanguageChart(rankedLanguages, 10)
		if err == nil {
			langPath := filepath.Join(svgOutputDir, "language_chart.svg")
			err = generator.SaveSVG(langSVG, langPath)
			if err != nil {
				logger.LogError(err, "言語ランキング SVG の保存に失敗しました")
			} else {
				svgs["language_chart.svg"] = langPath
				logger.Info("言語ランキング SVG を生成しました: %s", langPath)
				fmt.Printf("  ✅ 言語ランキング SVG を生成: %s\n", langPath)
			}
		}
	}

	// コミット推移 SVG
	if len(aggregatedHistory) > 0 {
		// DateCommitPair スライスを map[string]int に変換
		historyMap := make(map[string]int)
		for _, pair := range aggregatedHistory {
			historyMap[pair.Date] = pair.Count
		}
		historySVG, err := generator.GenerateCommitHistoryChart(historyMap)
		if err == nil {
			historyPath := filepath.Join(svgOutputDir, "commit_history_chart.svg")
			err = generator.SaveSVG(historySVG, historyPath)
			if err == nil {
				svgs["commit_history_chart.svg"] = historyPath
				fmt.Printf("  ✅ コミット推移 SVG を生成: %s\n", historyPath)
			}
		}
	}

	// コミット時間帯 SVG
	if len(aggregatedTimeDist) > 0 {
		// HourCommitPair スライスを map[int]int に変換
		timeDistMap := make(map[int]int)
		for _, pair := range aggregatedTimeDist {
			timeDistMap[pair.Hour] = pair.Count
		}
		timeSVG, err := generator.GenerateCommitTimeChart(timeDistMap)
		if err == nil {
			timePath := filepath.Join(svgOutputDir, "commit_time_chart.svg")
			err = generator.SaveSVG(timeSVG, timePath)
			if err == nil {
				svgs["commit_time_chart.svg"] = timePath
				fmt.Printf("  ✅ コミット時間帯 SVG を生成: %s\n", timePath)
			}
		}
	}

	// コミットごとの言語Top5 SVG
	if len(top5Languages) > 0 {
		commitLangSVG, err := generator.GenerateCommitLanguagesChart(top5Languages)
		if err == nil {
			commitLangPath := filepath.Join(svgOutputDir, "commit_languages_chart.svg")
			err = generator.SaveSVG(commitLangSVG, commitLangPath)
			if err == nil {
				svgs["commit_languages_chart.svg"] = commitLangPath
				fmt.Printf("  ✅ コミット言語Top5 SVG を生成: %s\n", commitLangPath)
			}
		}
	}

	// サマリーカード SVG
	if summaryStats.RepositoryCount > 0 {
		summarySVG, err := generator.GenerateSummaryCard(summaryStats)
		if err == nil {
			summaryPath := filepath.Join(svgOutputDir, "summary_card.svg")
			err = generator.SaveSVG(summarySVG, summaryPath)
			if err == nil {
				svgs["summary_card.svg"] = summaryPath
				fmt.Printf("  ✅ サマリーカード SVG を生成: %s\n", summaryPath)
			}
		}
	}

	// 5. README.md の更新
	fmt.Println("\n📝 README.md を更新しています...")

	readmePath := filepath.Join(config.RepoPath, "README.md")
	if config.RepoPath == "" {
		readmePath = "README.md"
	}

	// README が存在しない場合は作成
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		err = os.WriteFile(readmePath, []byte("# GitHub Profile\n\n"), 0644)
		if err != nil {
			return fmt.Errorf("README.md の作成に失敗しました: %w", err)
		}
		fmt.Printf("  ℹ️  README.md を作成しました\n")
	}

	// SVG グラフを埋め込み
	svgSections := map[string]string{
		"LANGUAGE_STATS":   "language_chart.svg",
		"COMMIT_HISTORY":   "commit_history_chart.svg",
		"COMMIT_TIME":      "commit_time_chart.svg",
		"COMMIT_LANGUAGES": "commit_languages_chart.svg",
		"SUMMARY_STATS":    "summary_card.svg",
	}

	for sectionTag, svgFile := range svgSections {
		if svgPath, ok := svgs[svgFile]; ok {
			// 相対パスに変換
			relPath, err := filepath.Rel(config.RepoPath, svgPath)
			if err != nil {
				relPath = svgFile
			}

			err = readme.EmbedSVGWithCustomPath(readmePath, relPath, sectionTag, "")
			if err != nil {
				logger.LogErrorWithContext(err, sectionTag, "セクションの更新に失敗しました")
				fmt.Printf("  ⚠️  セクション %s の更新に失敗: %v\n", sectionTag, err)
			} else {
				logger.Info("セクション %s を更新しました", sectionTag)
				fmt.Printf("  ✅ セクション %s を更新\n", sectionTag)
			}
		}
	}

	// 更新日時の追加
	if config.Timezone == "" {
		config.Timezone = "UTC"
	}
	timestamp := time.Now().UTC()
	err = readme.AddUpdateTimestamp(readmePath, "UPDATE_TIMESTAMP", timestamp, config.Timezone)
	if err != nil {
		logger.LogError(err, "更新日時の追加に失敗しました")
		fmt.Printf("  ⚠️  更新日時の追加に失敗: %v\n", err)
	} else {
		logger.Info("更新日時を追加しました")
		fmt.Printf("  ✅ 更新日時を追加\n")
	}

	// 6. Git コミット・プッシュ
	if !config.EnableGitPush {
		fmt.Println("\n✅ 処理が完了しました（Git プッシュはスキップされました）")
		return nil
	}

	fmt.Println("\n🔀 Git 操作を実行しています...")

	repoPath := config.RepoPath
	if repoPath == "" {
		repoPath = "."
	}

	// Git リポジトリか確認
	if !git.IsGitRepository(repoPath) {
		logger.Warning("Git リポジトリではないため、コミット・プッシュをスキップします")
		fmt.Println("  ℹ️  Git リポジトリではないため、コミット・プッシュをスキップします")
		return nil
	}

	// 変更があるか確認
	hasChanges, err := git.HasChanges(repoPath)
	if err != nil {
		logger.LogError(err, "変更の確認に失敗しました")
		return fmt.Errorf("変更の確認に失敗しました: %w", err)
	}

	if !hasChanges {
		logger.Info("変更がないため、コミット・プッシュをスキップします")
		fmt.Println("  ℹ️  変更がないため、コミット・プッシュをスキップします")
		return nil
	}

	// コミットメッセージ
	commitMsg := config.CommitMessage
	if commitMsg == "" {
		commitMsg = "chore: update GitHub profile metrics"
	}

	// コミット・プッシュ
	logger.Info("Git コミット・プッシュを実行しています...")
	err = git.CommitAndPush(repoPath, commitMsg, nil, "origin", "")
	if err != nil {
		logger.LogError(err, "Git コミット・プッシュに失敗しました")
		return fmt.Errorf("Git コミット・プッシュに失敗しました: %w", err)
	}

	logger.Info("Git コミット・プッシュが完了しました")
	fmt.Println("  ✅ Git コミット・プッシュが完了しました")

	logger.Info("すべての処理が完了しました")
	fmt.Println("\n✅ すべての処理が完了しました！")

	return nil
}
