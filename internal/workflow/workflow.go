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

	"github.com/google/go-github/v76/github"
)

// Config ワークフロー設定
type Config struct {
	RepoPath        string          // リポジトリパス（README.md がある場所）
	SVGOutputDir    string          // SVG ファイルの出力ディレクトリ
	Timezone        string          // タイムゾーン（例: "Asia/Tokyo", "UTC"）
	CommitMessage   string          // Git コミットメッセージ
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
func Run(ctx context.Context, tokenRead string, tokenWrite string, config Config) error {
	// ロガーの設定
	if config.LogLevel != 0 {
		logger.DefaultLogger.SetLevel(config.LogLevel)
	}

	logger.Info("ワークフローを開始します")

	// トークンの検証（既に渡されているが念のため確認）
	if tokenRead == "" {
		logger.Error("GITHUB_TOKEN_READ が設定されていません")
		return fmt.Errorf("GITHUB_TOKEN_READ が設定されていません")
	}

	// 認証ユーザー情報をGraphQLで取得（生成された型を使用）
	username, userID, err := repository.FetchViewerGenerated(ctx, tokenRead)
	if err != nil {
		logger.LogError(err, "認証ユーザー情報の取得に失敗しました")
		return fmt.Errorf("認証ユーザー情報の取得に失敗しました: %w", err)
	}
	logger.Info("認証ユーザー: %s", username)

	// 1-2. GraphQLを使用してデータを一括取得・集計
	fmt.Println("\n📊 GraphQLを使用してリポジトリデータを一括取得・集計しています...")
	logger.Info("GraphQLを使用してデータを取得します")

	languageTotals, commitHistories, timeDistributions, allCommitLanguages, totalCommits, totalPRs, repos, err := AggregateGraphQLData(
		ctx, tokenRead, username, userID, config.ExcludeForks)
	if err != nil {
		logger.LogError(err, "GraphQLデータの取得・集計に失敗しました")
		return fmt.Errorf("GraphQLデータの取得・集計に失敗しました: %w", err)
	}

	if len(languageTotals) == 0 {
		logger.Warning("リポジトリデータが見つかりませんでした")
		return fmt.Errorf("リポジトリデータが見つかりませんでした")
	}

	logger.Info("GraphQLデータの取得が完了しました: 言語数=%d, コミット履歴数=%d, 総コミット数=%d, 総PR数=%d",
		len(languageTotals), len(commitHistories), totalCommits, totalPRs)
	fmt.Printf("✅ GraphQLでデータを取得しました（言語: %d種類, コミット履歴: %dリポジトリ）\n",
		len(languageTotals), len(commitHistories))

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
	var reposForSummary []*github.Repository
	if len(repos) > 0 {
		reposForSummary = repos
	}
	summaryStats := aggregator.AggregateSummaryStats(reposForSummary, totalCommits, totalPRs)

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
	err = git.CommitAndPush(repoPath, commitMsg, nil, "origin", "", tokenWrite)
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
