package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/watsumi/update-gh-profile/internal/aggregator"
	"github.com/watsumi/update-gh-profile/internal/config"
	"github.com/watsumi/update-gh-profile/internal/generator"
	"github.com/watsumi/update-gh-profile/internal/repository"

	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
)

// formatBytes バイト数を人間が読みやすい形式に変換する
func formatBytes(bytes int) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func main() {
	// コマンドライン引数のパース
	var (
		usernameFlag    = flag.String("username", "", "[非推奨・無視されます] このツールは認証ユーザー自身のリポジトリのみを取得します")
		excludeForksStr = flag.String("exclude-forks", "true", "フォークリポジトリを除外するか（true/false）")
	)
	flag.Parse()

	fmt.Println("update-gh-profile: GitHub プロフィール自動更新ツール")
	fmt.Println("初期化完了")

	// 設定を読み込む
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("エラー: 設定の読み込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Printf("エラー: 設定の検証に失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ GITHUB_TOKEN が設定されています")

	// GitHub API クライアントの初期化
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GitHubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// 認証ユーザーを取得（必須）
	authUser, _, err := client.Users.Get(ctx, "")
	if err != nil {
		fmt.Printf("エラー: 認証ユーザー情報の取得に失敗しました: %v\n", err)
		os.Exit(1)
	}
	authenticatedUsername := authUser.GetLogin()

	// 対象ユーザー名の決定（優先順位: コマンドライン引数 > 環境変数 > 認証ユーザー）
	targetUser := *usernameFlag
	if targetUser == "" {
		targetUser = cfg.GetTargetUser()
		if targetUser == "" {
			// 認証ユーザーを使用（デフォルト）
			targetUser = authenticatedUsername
		}
	}

	// 認証ユーザー以外を指定した場合はエラー
	if targetUser != authenticatedUsername {
		fmt.Printf("エラー: 認証ユーザー（%s）以外のリポジトリを取得することはできません\n", authenticatedUsername)
		fmt.Printf("指定されたユーザー: %s\n", targetUser)
		fmt.Println("\nこのツールは認証ユーザー自身のリポジトリのみを取得できます。")
		os.Exit(1)
	}

	// 認証ユーザー自身であることを確認
	fmt.Printf("✓ 認証ユーザー: %s（プライベートリポジトリも取得します）\n", targetUser)

	// フォーク除外の設定
	excludeForks, err := strconv.ParseBool(*excludeForksStr)
	if err != nil {
		fmt.Printf("警告: exclude-forks の値が不正です（%s）。デフォルト値 true を使用します\n", *excludeForksStr)
		excludeForks = true
	}

	fmt.Println("\n✅ GitHub API クライアントの初期化に成功しました！")

	// リポジトリ一覧の取得（認証ユーザー自身のリポジトリのみ）
	fmt.Println("\n📦 リポジトリ一覧を取得しています...")
	repos, err := repository.FetchUserRepositories(ctx, client, targetUser, excludeForks, true) // 常に認証ユーザーとして取得
	if err != nil {
		fmt.Printf("エラー: リポジトリ一覧の取得に失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ リポジトリ一覧の取得に成功しました: %d 件\n", len(repos))

	// 取得したリポジトリの一部を表示（最大5件）
	maxDisplay := 5
	if len(repos) < maxDisplay {
		maxDisplay = len(repos)
	}
	fmt.Printf("\n取得したリポジトリ（最初の%d件）:\n", maxDisplay)
	for i := 0; i < maxDisplay; i++ {
		repo := repos[i]
		fmt.Printf("  - %s (⭐ %d, Fork: %v)\n",
			repo.GetFullName(),
			repo.GetStargazersCount(),
			repo.GetFork())
	}
	if len(repos) > maxDisplay {
		fmt.Printf("  ... 他 %d 件\n", len(repos)-maxDisplay)
	}

	// 言語情報の取得テスト（最初の3件のリポジトリに対して）
	if len(repos) > 0 {
		fmt.Println("\n📊 リポジトリの言語情報を取得しています...")
		testCount := 3
		if len(repos) < testCount {
			testCount = len(repos)
		}

		for i := 0; i < testCount; i++ {
			repo := repos[i]
			owner := repo.GetOwner().GetLogin()
			repoName := repo.GetName()

			fmt.Printf("\n  [%d/%d] %s/%s の言語情報を取得中...\n", i+1, testCount, owner, repoName)

			languages, err := repository.FetchRepositoryLanguages(ctx, client, owner, repoName)
			if err != nil {
				fmt.Printf("    ⚠️  エラー: %v\n", err)
				continue
			}

			if len(languages) == 0 {
				fmt.Printf("    ℹ️  言語情報が見つかりませんでした\n")
				continue
			}

			fmt.Printf("    ✅ 言語数: %d\n", len(languages))

			// 言語情報を表示（上位5言語まで）
			type langStat struct {
				name  string
				bytes int
			}
			var langList []langStat
			totalBytes := 0
			for lang, bytes := range languages {
				langList = append(langList, langStat{name: lang, bytes: bytes})
				totalBytes += bytes
			}

			// バイト数でソート（降順）
			for i := 0; i < len(langList)-1; i++ {
				for j := i + 1; j < len(langList); j++ {
					if langList[i].bytes < langList[j].bytes {
						langList[i], langList[j] = langList[j], langList[i]
					}
				}
			}

			maxLangDisplay := 5
			if len(langList) < maxLangDisplay {
				maxLangDisplay = len(langList)
			}

			fmt.Printf("    📈 主要な言語（上位%d言語）:\n", maxLangDisplay)
			for j := 0; j < maxLangDisplay; j++ {
				lang := langList[j]
				percentage := float64(lang.bytes) / float64(totalBytes) * 100
				fmt.Printf("      - %s: %.1f%% (%s)\n",
					lang.name,
					percentage,
					formatBytes(lang.bytes))
			}
			if len(langList) > maxLangDisplay {
				fmt.Printf("      ... 他 %d 言語\n", len(langList)-maxLangDisplay)
			}
		}
		fmt.Println("\n✅ 言語情報の取得テストが完了しました")
	}

	// 言語データ集計のテスト（全リポジトリの言語情報を集計）
	if len(repos) > 0 {
		fmt.Println("\n📊 全リポジトリの言語データを集計しています...")

		// リポジトリごとの言語データを格納する map
		// map[リポジトリ名]map[言語名]バイト数
		languageData := make(map[string]map[string]int)

		// 各リポジトリの言語情報を取得（最初の5件のリポジトリに対して）
		testCount := 5
		if len(repos) < testCount {
			testCount = len(repos)
		}

		for i := 0; i < testCount; i++ {
			repo := repos[i]
			owner := repo.GetOwner().GetLogin()
			repoName := repo.GetName()
			repoKey := fmt.Sprintf("%s/%s", owner, repoName)

			fmt.Printf("  [%d/%d] %s の言語情報を取得中...\n", i+1, testCount, repoKey)

			languages, err := repository.FetchRepositoryLanguages(ctx, client, owner, repoName)
			if err != nil {
				fmt.Printf("    ⚠️  エラー: %v\n", err)
				continue
			}

			if len(languages) == 0 {
				fmt.Printf("    ℹ️  言語情報が見つかりませんでした\n")
				continue
			}

			// 言語データを保存
			languageData[repoKey] = languages
			fmt.Printf("    ✅ %d 言語を取得しました\n", len(languages))
		}

		// 言語データを集計
		fmt.Printf("\n📈 言語データを集計中...\n")
		languageTotals := aggregator.AggregateLanguages(repos[:testCount], languageData)

		if len(languageTotals) > 0 {
			fmt.Printf("✅ 集計完了: %d 言語\n", len(languageTotals))

			// ランキングを生成
			rankedLanguages := aggregator.RankLanguages(languageTotals)

			// 上位5言語を表示
			maxDisplay := 5
			if len(rankedLanguages) < maxDisplay {
				maxDisplay = len(rankedLanguages)
			}

			fmt.Printf("\n🏆 言語ランキング（上位%d言語）:\n", maxDisplay)
			for i := 0; i < maxDisplay; i++ {
				lang := rankedLanguages[i]
				fmt.Printf("  %d. %s: %.1f%% (%s)\n",
					i+1,
					lang.Language,
					lang.Percentage,
					formatBytes(lang.Bytes))
			}

			// 閾値（1%）でフィルタリングして表示
			filtered := aggregator.FilterMinorLanguages(rankedLanguages, 1.0)
			fmt.Printf("\n📌 閾値1%%以上: %d 言語\n", len(filtered))
			if len(filtered) < len(rankedLanguages) {
				fmt.Printf("  （%d 言語が除外されました）\n", len(rankedLanguages)-len(filtered))
			}

			// SVG グラフを生成
			fmt.Printf("\n🎨 言語ランキングの SVG グラフを生成中...\n")
			svg, err := generator.GenerateLanguageChart(rankedLanguages, 10)
			if err != nil {
				fmt.Printf("  ⚠️  SVG生成エラー: %v\n", err)
			} else {
				// SVG をファイルに保存（テスト用）
				outputPath := "language_chart.svg"
				err = generator.SaveSVG(svg, outputPath)
				if err != nil {
					fmt.Printf("  ⚠️  ファイル保存エラー: %v\n", err)
				} else {
					fmt.Printf("  ✅ SVG グラフを生成しました: %s\n", outputPath)
					fmt.Printf("    （SVGサイズ: %d バイト）\n", len(svg))
				}
			}
		} else {
			fmt.Println("⚠️  集計できる言語データがありませんでした")
		}

		fmt.Println("\n✅ 言語データ集計のテストが完了しました")
	}

	// コミット情報の取得テスト（最初の1件のリポジトリに対して）
	if len(repos) > 0 {
		fmt.Println("\n📝 リポジトリのコミット情報を取得しています...")
		repo := repos[0]
		owner := repo.GetOwner().GetLogin()
		repoName := repo.GetName()

		fmt.Printf("\n  [1/1] %s/%s のコミット情報を取得中...\n", owner, repoName)

		// コミット履歴の取得
		commits, err := repository.FetchCommits(ctx, client, owner, repoName)
		if err != nil {
			fmt.Printf("    ⚠️  エラー: %v\n", err)
		} else {
			fmt.Printf("    ✅ コミット数: %d\n", len(commits))

			// 最新の5件のコミットを表示
			maxCommitDisplay := 5
			if len(commits) < maxCommitDisplay {
				maxCommitDisplay = len(commits)
			}
			if maxCommitDisplay > 0 {
				fmt.Printf("    📋 最新のコミット（%d件）:\n", maxCommitDisplay)
				for i := 0; i < maxCommitDisplay; i++ {
					commit := commits[i]
					if commit.Commit != nil {
						message := commit.Commit.GetMessage()
						// メッセージの最初の行のみ表示（改行を除去）
						if len(message) > 50 {
							message = message[:50] + "..."
						}
						date := "N/A"
						if commit.Commit.Committer != nil && commit.Commit.Committer.Date != nil {
							date = commit.Commit.Committer.Date.Time.Format("2006-01-02 15:04")
						}
						fmt.Printf("      - %s (%s)\n", message, date)
					}
				}
			}
		}

		// 日付ごとのコミット数の取得
		commitHistory, err := repository.FetchCommitHistory(ctx, client, owner, repoName)
		if err != nil {
			fmt.Printf("    ⚠️  日付ごとのコミット数取得エラー: %v\n", err)
		} else {
			fmt.Printf("    ✅ コミット履歴: %d 日分\n", len(commitHistory))

			// 最新の5日分を表示
			type dateCount struct {
				date  string
				count int
			}
			var historyList []dateCount
			for date, count := range commitHistory {
				historyList = append(historyList, dateCount{date: date, count: count})
			}

			// 日付でソート（降順）
			for i := 0; i < len(historyList)-1; i++ {
				for j := i + 1; j < len(historyList); j++ {
					if historyList[i].date < historyList[j].date {
						historyList[i], historyList[j] = historyList[j], historyList[i]
					}
				}
			}

			maxHistoryDisplay := 5
			if len(historyList) < maxHistoryDisplay {
				maxHistoryDisplay = len(historyList)
			}
			if maxHistoryDisplay > 0 {
				fmt.Printf("    📅 最近のコミット履歴（%d日分）:\n", maxHistoryDisplay)
				for i := 0; i < maxHistoryDisplay; i++ {
					item := historyList[i]
					fmt.Printf("      - %s: %d コミット\n", item.date, item.count)
				}
			}
		}

		// 時間帯ごとのコミット数の取得
		timeDistribution, err := repository.FetchCommitTimeDistribution(ctx, client, owner, repoName)
		if err != nil {
			fmt.Printf("    ⚠️  時間帯ごとのコミット数取得エラー: %v\n", err)
		} else {
			fmt.Printf("    ✅ コミット時間帯分布: %d 時間帯\n", len(timeDistribution))

			// コミット数が多い時間帯トップ5を表示
			type hourCount struct {
				hour  int
				count int
			}
			var hourList []hourCount
			for hour, count := range timeDistribution {
				hourList = append(hourList, hourCount{hour: hour, count: count})
			}

			// コミット数でソート（降順）
			for i := 0; i < len(hourList)-1; i++ {
				for j := i + 1; j < len(hourList); j++ {
					if hourList[i].count < hourList[j].count {
						hourList[i], hourList[j] = hourList[j], hourList[i]
					}
				}
			}

			maxHourDisplay := 5
			if len(hourList) < maxHourDisplay {
				maxHourDisplay = len(hourList)
			}
			if maxHourDisplay > 0 {
				fmt.Printf("    🕐 コミットが多い時間帯（UTC、上位%d）:\n", maxHourDisplay)
				for i := 0; i < maxHourDisplay; i++ {
					item := hourList[i]
					fmt.Printf("      - %02d時: %d コミット\n", item.hour, item.count)
				}
			}
		}

		fmt.Println("\n✅ コミット情報の取得テストが完了しました")
	}

	// コミット履歴と時間帯分布の集計テスト（最初の3件のリポジトリに対して）
	if len(repos) > 0 {
		fmt.Println("\n📊 全リポジトリのコミット履歴と時間帯分布を集計しています...")

		// リポジトリごとのコミット履歴を格納する map
		commitHistories := make(map[string]map[string]int)
		// リポジトリごとの時間帯分布を格納する map
		timeDistributions := make(map[string]map[int]int)

		testCount := 3
		if len(repos) < testCount {
			testCount = len(repos)
		}

		// 各リポジトリのコミット履歴と時間帯分布を取得
		for i := 0; i < testCount; i++ {
			repo := repos[i]
			owner := repo.GetOwner().GetLogin()
			repoName := repo.GetName()
			repoKey := fmt.Sprintf("%s/%s", owner, repoName)

			fmt.Printf("  [%d/%d] %s のコミット情報を取得中...\n", i+1, testCount, repoKey)

			// コミット履歴の取得
			history, err := repository.FetchCommitHistory(ctx, client, owner, repoName)
			if err != nil {
				fmt.Printf("    ⚠️  コミット履歴取得エラー: %v\n", err)
			} else {
				commitHistories[repoKey] = history
				fmt.Printf("    ✅ コミット履歴: %d 日分\n", len(history))
			}

			// 時間帯分布の取得
			timeDist, err := repository.FetchCommitTimeDistribution(ctx, client, owner, repoName)
			if err != nil {
				fmt.Printf("    ⚠️  時間帯分布取得エラー: %v\n", err)
			} else {
				timeDistributions[repoKey] = timeDist
				fmt.Printf("    ✅ 時間帯分布: %d 時間帯\n", len(timeDist))
			}
		}

		// コミット履歴を集計
		if len(commitHistories) > 0 {
			fmt.Printf("\n📅 コミット履歴を集計中...\n")
			aggregatedHistory := aggregator.AggregateCommitHistory(commitHistories)

			if len(aggregatedHistory) > 0 {
				fmt.Printf("✅ 集計完了: %d 日分\n", len(aggregatedHistory))

				// 日付順でソート
				sortedHistory := aggregator.SortCommitHistoryByDate(aggregatedHistory)

				// 最新の5日分を表示
				maxDisplay := 5
				if len(sortedHistory) < maxDisplay {
					maxDisplay = len(sortedHistory)
				}

				if maxDisplay > 0 {
					startIdx := len(sortedHistory) - maxDisplay
					if startIdx < 0 {
						startIdx = 0
					}
					fmt.Printf("\n📈 最近のコミット履歴（%d日分）:\n", maxDisplay)
					for i := startIdx; i < len(sortedHistory); i++ {
						pair := sortedHistory[i]
						fmt.Printf("  - %s: %d コミット\n", pair.Date, pair.Count)
					}
				}

				// SVG グラフを生成
				fmt.Printf("\n🎨 コミット推移の SVG グラフを生成中...\n")
				svg, err := generator.GenerateCommitHistoryChart(aggregatedHistory)
				if err != nil {
					fmt.Printf("  ⚠️  SVG生成エラー: %v\n", err)
				} else {
					// SVG をファイルに保存（テスト用）
					outputPath := "commit_history_chart.svg"
					err = os.WriteFile(outputPath, []byte(svg), 0644)
					if err != nil {
						fmt.Printf("  ⚠️  ファイル保存エラー: %v\n", err)
					} else {
						fmt.Printf("  ✅ SVG グラフを生成しました: %s\n", outputPath)
						fmt.Printf("    （SVGサイズ: %d バイト）\n", len(svg))
					}
				}
			}
		}

		// 時間帯分布を集計
		if len(timeDistributions) > 0 {
			fmt.Printf("\n🕐 コミット時間帯分布を集計中...\n")
			aggregatedTimeDist := aggregator.AggregateCommitTimeDistribution(timeDistributions)

			if len(aggregatedTimeDist) > 0 {
				fmt.Printf("✅ 集計完了: %d 時間帯\n", len(aggregatedTimeDist))

				// 時間帯順でソート
				sortedTimeDist := aggregator.SortCommitTimeDistributionByHour(aggregatedTimeDist)

				// コミット数が多い時間帯トップ5を表示
				type hourCount struct {
					hour  int
					count int
				}
				var hourList []hourCount
				for _, pair := range sortedTimeDist {
					hourList = append(hourList, hourCount{hour: pair.Hour, count: pair.Count})
				}

				// コミット数でソート（降順）
				for i := 0; i < len(hourList)-1; i++ {
					for j := i + 1; j < len(hourList); j++ {
						if hourList[i].count < hourList[j].count {
							hourList[i], hourList[j] = hourList[j], hourList[i]
						}
					}
				}

				maxDisplay := 5
				if len(hourList) < maxDisplay {
					maxDisplay = len(hourList)
				}
				if maxDisplay > 0 {
					fmt.Printf("\n🏆 コミットが多い時間帯（UTC、上位%d）:\n", maxDisplay)
					for i := 0; i < maxDisplay; i++ {
						item := hourList[i]
						fmt.Printf("  %d. %02d時: %d コミット\n", i+1, item.hour, item.count)
					}
				}
			}

			// SVG グラフを生成
			fmt.Printf("\n🎨 コミット時間帯分布の SVG グラフを生成中...\n")
			svg, err := generator.GenerateCommitTimeChart(aggregatedTimeDist)
			if err != nil {
				fmt.Printf("  ⚠️  SVG生成エラー: %v\n", err)
			} else {
				// SVG をファイルに保存（テスト用）
				outputPath := "commit_time_chart.svg"
				err = os.WriteFile(outputPath, []byte(svg), 0644)
				if err != nil {
					fmt.Printf("  ⚠️  ファイル保存エラー: %v\n", err)
				} else {
					fmt.Printf("  ✅ SVG グラフを生成しました: %s\n", outputPath)
					fmt.Printf("    （SVGサイズ: %d バイト）\n", len(svg))
				}
			}
		}

		fmt.Println("\n✅ コミット履歴・時間帯分布の集計テストが完了しました")
	}

	// コミットごとの言語Top5集計のテスト（最初の2件のリポジトリに対して）
	if len(repos) > 0 {
		fmt.Println("\n🔍 コミットごとの言語Top5を集計しています...")

		// リポジトリごとのコミット言語データを格納する map
		// map[コミットSHA]map[言語名]出現回数 の形式で統合
		allCommitLanguages := make(map[string]map[string]int)

		testCount := 2
		if len(repos) < testCount {
			testCount = len(repos)
		}

		// 各リポジトリのコミット言語データを取得
		for i := 0; i < testCount; i++ {
			repo := repos[i]
			owner := repo.GetOwner().GetLogin()
			repoName := repo.GetName()
			repoKey := fmt.Sprintf("%s/%s", owner, repoName)

			fmt.Printf("  [%d/%d] %s のコミット言語データを取得中...\n", i+1, testCount, repoKey)

			commitLanguages, err := repository.FetchCommitLanguages(ctx, client, owner, repoName)
			if err != nil {
				fmt.Printf("    ⚠️  エラー: %v\n", err)
				continue
			}

			if len(commitLanguages) == 0 {
				fmt.Printf("    ℹ️  コミット言語データが見つかりませんでした\n")
				continue
			}

			// コミットごとの言語データを統合（SHAをキーとして統合）
			for sha, langs := range commitLanguages {
				// SHAにリポジトリ名をプレフィックスとして付与（同じSHAが複数リポジトリにある場合を考慮）
				uniqueSHA := fmt.Sprintf("%s:%s", repoKey, sha)
				allCommitLanguages[uniqueSHA] = langs
			}

			fmt.Printf("    ✅ %d コミット分の言語データを取得しました\n", len(commitLanguages))
		}

		// コミットごとの言語Top5を集計
		if len(allCommitLanguages) > 0 {
			fmt.Printf("\n📊 コミットごとの言語Top5を集計中...\n")
			top5Languages := aggregator.AggregateCommitLanguages(allCommitLanguages)

			if len(top5Languages) > 0 {
				fmt.Printf("✅ 集計完了: %d 言語（Top5）\n", len(top5Languages))

				// 使用回数でソートして表示
				type langCount struct {
					lang  string
					count int
				}
				var langList []langCount
				for lang, count := range top5Languages {
					langList = append(langList, langCount{lang: lang, count: count})
				}

				// 使用回数でソート（降順）
				for i := 0; i < len(langList)-1; i++ {
					for j := i + 1; j < len(langList); j++ {
						if langList[i].count < langList[j].count {
							langList[i], langList[j] = langList[j], langList[i]
						}
					}
				}

				fmt.Printf("\n🏆 コミットごとの使用言語 Top5:\n")
				for i, item := range langList {
					fmt.Printf("  %d. %s: %d ファイル\n", i+1, item.lang, item.count)
				}

				// SVG グラフを生成
				fmt.Printf("\n🎨 コミットごとの使用言語Top5の SVG グラフを生成中...\n")
				svg, err := generator.GenerateCommitLanguagesChart(top5Languages)
				if err != nil {
					fmt.Printf("  ⚠️  SVG生成エラー: %v\n", err)
				} else {
					// SVG をファイルに保存（テスト用）
					outputPath := "commit_languages_chart.svg"
					err = os.WriteFile(outputPath, []byte(svg), 0644)
					if err != nil {
						fmt.Printf("  ⚠️  ファイル保存エラー: %v\n", err)
					} else {
						fmt.Printf("  ✅ SVG グラフを生成しました: %s\n", outputPath)
						fmt.Printf("    （SVGサイズ: %d バイト）\n", len(svg))
					}
				}
			} else {
				fmt.Println("⚠️  集計できる言語データがありませんでした")
			}
		} else {
			fmt.Println("⚠️  コミット言語データがありませんでした")
		}

		fmt.Println("\n✅ コミットごとの言語Top5集計のテストが完了しました")
	}

	// コミットごとの言語使用状況の取得テスト（最初の1件のリポジトリに対して、最初の10コミットのみ）
	if len(repos) > 0 {
		fmt.Println("\n🔍 コミットごとの言語使用状況を取得しています...")
		repo := repos[0]
		owner := repo.GetOwner().GetLogin()
		repoName := repo.GetName()

		fmt.Printf("\n  [1/1] %s/%s のコミットごとの言語使用状況を取得中（最初の10コミットのみ）...\n", owner, repoName)

		commitLanguages, err := repository.FetchCommitLanguages(ctx, client, owner, repoName)
		if err != nil {
			fmt.Printf("    ⚠️  エラー: %v\n", err)
		} else {
			fmt.Printf("    ✅ 処理完了: %d コミット分の言語情報を取得しました\n", len(commitLanguages))

			// 最初の5コミット分の言語使用状況を表示
			maxCommitDisplay := 5
			count := 0
			for sha, langs := range commitLanguages {
				if count >= maxCommitDisplay {
					break
				}
				fmt.Printf("\n    📝 コミット %s で使用された言語:\n", sha[:7])
				if len(langs) == 0 {
					fmt.Printf("      ℹ️  言語情報なし\n")
				} else {
					// 言語を出現回数でソート
					type langCount struct {
						lang  string
						count int
					}
					var langList []langCount
					for lang, cnt := range langs {
						langList = append(langList, langCount{lang: lang, count: cnt})
					}

					// 出現回数でソート（降順）
					for i := 0; i < len(langList)-1; i++ {
						for j := i + 1; j < len(langList); j++ {
							if langList[i].count < langList[j].count {
								langList[i], langList[j] = langList[j], langList[i]
							}
						}
					}

					for _, item := range langList {
						fmt.Printf("      - %s: %d ファイル\n", item.lang, item.count)
					}
				}
				count++
			}
			if len(commitLanguages) > maxCommitDisplay {
				fmt.Printf("\n    ... 他 %d コミット\n", len(commitLanguages)-maxCommitDisplay)
			}

			// 全コミットを通しての言語使用回数（Top5）を集計
			allLangCounts := make(map[string]int)
			for _, langs := range commitLanguages {
				for lang, count := range langs {
					allLangCounts[lang] += count
				}
			}

			if len(allLangCounts) > 0 {
				type langCount struct {
					lang  string
					count int
				}
				var langList []langCount
				for lang, cnt := range allLangCounts {
					langList = append(langList, langCount{lang: lang, count: cnt})
				}

				// 出現回数でソート（降順）
				for i := 0; i < len(langList)-1; i++ {
					for j := i + 1; j < len(langList); j++ {
						if langList[i].count < langList[j].count {
							langList[i], langList[j] = langList[j], langList[i]
						}
					}
				}

				maxLangDisplay := 5
				if len(langList) < maxLangDisplay {
					maxLangDisplay = len(langList)
				}
				if maxLangDisplay > 0 {
					fmt.Printf("\n    📊 全コミットを通しての使用言語 Top%d:\n", maxLangDisplay)
					for i := 0; i < maxLangDisplay; i++ {
						item := langList[i]
						fmt.Printf("      - %s: %d ファイル\n", item.lang, item.count)
					}
				}
			}
		}

		fmt.Println("\n✅ コミットごとの言語使用状況の取得テストが完了しました")
	}

	// プルリクエスト情報の取得テスト（最初の3件のリポジトリに対して）
	if len(repos) > 0 {
		fmt.Println("\n🔀 リポジトリのプルリクエスト情報を取得しています...")
		testCount := 3
		if len(repos) < testCount {
			testCount = len(repos)
		}

		for i := 0; i < testCount; i++ {
			repo := repos[i]
			owner := repo.GetOwner().GetLogin()
			repoName := repo.GetName()

			fmt.Printf("\n  [%d/%d] %s/%s のプルリクエスト数を取得中...\n", i+1, testCount, owner, repoName)

			prCount, err := repository.FetchPullRequests(ctx, client, owner, repoName)
			if err != nil {
				fmt.Printf("    ⚠️  エラー: %v\n", err)
				continue
			}

			fmt.Printf("    ✅ プルリクエスト数: %d 件\n", prCount)
		}

		fmt.Println("\n✅ プルリクエスト情報の取得テストが完了しました")
	}

	// サマリー統計集計のテスト（全リポジトリの統計を集計）
	if len(repos) > 0 {
		fmt.Println("\n📊 サマリー統計を集計しています...")

		// 全リポジトリのコミット数とプルリクエスト数を取得
		// 注: 実際の運用では全リポジトリを取得するが、テストでは最初の3件のみ
		testCount := 3
		if len(repos) < testCount {
			testCount = len(repos)
		}

		totalCommits := 0
		totalPRs := 0

		// 各リポジトリのコミット数とプルリクエスト数を取得
		for i := 0; i < testCount; i++ {
			repo := repos[i]
			owner := repo.GetOwner().GetLogin()
			repoName := repo.GetName()

			fmt.Printf("  [%d/%d] %s/%s の統計を取得中...\n", i+1, testCount, owner, repoName)

			// コミット数を取得（すでに取得済みの場合は再利用できるが、今回は簡単のため再取得）
			commits, err := repository.FetchCommits(ctx, client, owner, repoName)
			if err != nil {
				fmt.Printf("    ⚠️  コミット数取得エラー: %v\n", err)
			} else {
				totalCommits += len(commits)
				fmt.Printf("    ✅ コミット数: %d\n", len(commits))
			}

			// プルリクエスト数を取得
			prCount, err := repository.FetchPullRequests(ctx, client, owner, repoName)
			if err != nil {
				fmt.Printf("    ⚠️  プルリクエスト数取得エラー: %v\n", err)
			} else {
				totalPRs += prCount
				fmt.Printf("    ✅ プルリクエスト数: %d\n", prCount)
			}
		}

		// サマリー統計を集計
		fmt.Printf("\n📈 サマリー統計を集計中...\n")
		summaryStats := aggregator.AggregateSummaryStats(repos[:testCount], totalCommits, totalPRs)

		fmt.Printf("\n📊 サマリー統計:\n")
		fmt.Printf("  ⭐ 合計スター数: %d\n", summaryStats.TotalStars)
		fmt.Printf("  📦 リポジトリ数: %d\n", summaryStats.RepositoryCount)
		fmt.Printf("  📝 総コミット数: %d\n", summaryStats.TotalCommits)
		fmt.Printf("  🔀 総プルリクエスト数: %d\n", summaryStats.TotalPullRequests)

		// SVG カードを生成
		fmt.Printf("\n🎨 サマリーカードの SVG を生成中...\n")
		svg, err := generator.GenerateSummaryCard(summaryStats)
		if err != nil {
			fmt.Printf("  ⚠️  SVG生成エラー: %v\n", err)
		} else {
			// SVG をファイルに保存（テスト用）
			outputPath := "summary_card.svg"
			err = os.WriteFile(outputPath, []byte(svg), 0644)
			if err != nil {
				fmt.Printf("  ⚠️  ファイル保存エラー: %v\n", err)
			} else {
				fmt.Printf("  ✅ SVG カードを生成しました: %s\n", outputPath)
				fmt.Printf("    （SVGサイズ: %d バイト）\n", len(svg))
			}
		}

		fmt.Println("\n✅ サマリー統計集計のテストが完了しました")
	}

	// GitHub Actions の出力変数を設定（GITHUB_OUTPUT ファイルに書き込む）
	if outputFile := os.Getenv("GITHUB_OUTPUT"); outputFile != "" {
		file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			fmt.Fprintf(file, "repository_count=%d\n", len(repos))
			file.Close()
		}
	}
}
