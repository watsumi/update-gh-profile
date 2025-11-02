package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	ghgraphql "github.com/watsumi/update-gh-profile/internal/graphql"
	"github.com/watsumi/update-gh-profile/internal/logger"
)

// FetchViewerGenerated 生成された型を使用して認証ユーザー情報を取得
func FetchViewerGenerated(ctx context.Context, token string) (string, string, error) {
	graphqlClient, err := newGraphQLClient(ctx, token)
	if err != nil {
		return "", "", fmt.Errorf("GraphQLクライアントの作成に失敗しました: %w", err)
	}

	var query ghgraphql.ViewerQuery
	err = graphqlClient.Query(ctx, &query, nil)
	if err != nil {
		return "", "", fmt.Errorf("GraphQLクエリの実行に失敗しました: %w", err)
	}

	return query.Viewer.Login, query.Viewer.ID, nil
}

// FetchRepositoriesWithGraphQLGenerated 生成された型を使用してリポジトリ情報を一括取得
func FetchRepositoriesWithGraphQLGenerated(ctx context.Context, token string, username string, excludeForks bool) ([]*RepositoryGraphQLData, error) {
	graphqlClient, err := newGraphQLClient(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("GraphQLクライアントの作成に失敗しました: %w", err)
	}

	var allRepos []*RepositoryGraphQLData
	var after *string

	// 文字列ベースのクエリを使用（オプショナル変数を適切に処理するため）
	// 既存のQueryReposPerLanguageと同じ形式で統一
	queryStr := `query ReposQuery($login: String!, $isFork: Boolean, $first: Int, $after: String) {
  user(login: $login) {
    repositories(isFork: $isFork, first: $first, after: $after, ownerAffiliations: OWNER) {
      nodes {
        name
        owner {
          login
        }
        primaryLanguage {
          name
        }
        languages(first: 100) {
          edges {
            node {
              name
            }
            size
          }
          totalSize
        }
        stargazerCount
        defaultBranchRef {
          target {
            ... on Commit {
              history(first: 10) {
                totalCount
                nodes {
                  committedDate
                  author {
                    date
                  }
                }
              }
            }
          }
        }
      }
      pageInfo {
        endCursor
        hasNextPage
      }
    }
  }
}`

	for {
		// 変数をmapに変換
		variables := map[string]interface{}{
			"login":  username,
			"isFork": !excludeForks,
			"first":  50, // ページサイズを50に縮小してタイムアウトを防ぐ
		}
		if after != nil {
			variables["after"] = *after
		}

		// Execメソッドで文字列クエリと生成された型を組み合わせて使用
		// Execはdataフィールドを自動unwrapするため、直接型を渡す
		var query ghgraphql.ReposQuery

		// リトライロジック: 502 Bad Gatewayなどの一時的なエラーに対応
		const maxRetries = 5
		const baseRetryDelay = 5 * time.Second
		var lastErr error

		for attempt := 0; attempt < maxRetries; attempt++ {
			if attempt > 0 {
				// 指数バックオフでリトライ前に待機
				// 1回目: 5秒、2回目: 10秒、3回目: 20秒、4回目: 40秒
				retryDelay := baseRetryDelay * time.Duration(1<<uint(attempt-1))
				logger.Info("GraphQLクエリのリトライ: %d/%d回目（%v待機後）", attempt+1, maxRetries, retryDelay)
				select {
				case <-ctx.Done():
					return nil, fmt.Errorf("コンテキストがキャンセルされました: %w", ctx.Err())
				case <-time.After(retryDelay):
					// 待機完了
				}
			}

			logger.Debug("GraphQLクエリを実行中（試行 %d/%d）...", attempt+1, maxRetries)

			// リクエスト前に少し待機してGitHub APIへの負荷を軽減
			if attempt == 0 {
				select {
				case <-ctx.Done():
					return nil, fmt.Errorf("コンテキストがキャンセルされました: %w", ctx.Err())
				case <-time.After(500 * time.Millisecond):
					// 500ms待機
				}
			}

			err = graphqlClient.Exec(ctx, queryStr, &query, variables)
			if err == nil {
				if attempt > 0 {
					logger.Info("GraphQLクエリが成功しました（%d回目の試行で成功）", attempt+1)
				}
				break // 成功
			}

			lastErr = err
			errStr := err.Error()
			logger.Warning("GraphQLクエリの実行に失敗しました（試行 %d/%d）: %v", attempt+1, maxRetries, err)

			// 502 Bad Gateway や 503 Service Unavailable などの一時的なエラーの場合はリトライ
			lowerErrStr := strings.ToLower(errStr)
			isRetryableError := strings.Contains(lowerErrStr, "502") ||
				strings.Contains(lowerErrStr, "503") ||
				strings.Contains(lowerErrStr, "504") ||
				strings.Contains(lowerErrStr, "timeout") ||
				strings.Contains(lowerErrStr, "bad gateway") ||
				strings.Contains(lowerErrStr, "service unavailable") ||
				strings.Contains(lowerErrStr, "gateway timeout") ||
				strings.Contains(lowerErrStr, "request_error") ||
				strings.Contains(lowerErrStr, "network") ||
				strings.Contains(lowerErrStr, "connection")

			if !isRetryableError {
				// 一時的なエラーでない場合は即座に返す
				return nil, fmt.Errorf("GraphQLクエリの実行に失敗しました（リトライ不可なエラー）: %w", err)
			}

			// 最後の試行でエラーが残っている場合はログに記録
			if attempt == maxRetries-1 {
				logger.Error("GraphQLクエリの実行に失敗しました（%d回リトライ後）: %v", maxRetries, lastErr)
				return nil, fmt.Errorf("GraphQLクエリの実行に失敗しました（%d回リトライ後）: %w", maxRetries, lastErr)
			}
		}

		if lastErr != nil {
			return nil, fmt.Errorf("GraphQLクエリの実行に失敗しました: %w", lastErr)
		}

		// 生成された型からRepositoryGraphQLDataに変換
		for _, repo := range query.User.Repositories.Nodes {
			repoData := &RepositoryGraphQLData{
				Name:           repo.Name,
				StargazerCount: repo.StargazerCount,
			}

			// Owner
			repoData.Owner.Login = repo.Owner.Login

			// PrimaryLanguage
			if repo.PrimaryLanguage != nil {
				repoData.PrimaryLanguage.Name = repo.PrimaryLanguage.Name
			}

			// Languages
			if repo.Languages != nil {
				repoData.Languages.TotalSize = repo.Languages.TotalSize
				// LanguageConnectionのnodesはLanguage型（sizeフィールドなし）
				// edgesにsizeがあるので、edgesを使用するか、nodesだけを保存する
				// ここではedgesを使用してsizeも取得
				if len(repo.Languages.Edges) > 0 {
					for _, edge := range repo.Languages.Edges {
						repoData.Languages.Nodes = append(repoData.Languages.Nodes, struct {
							Name string `json:"name"`
							Size int    `json:"size"`
						}{
							Name: edge.Node.Name,
							Size: edge.Size,
						})
					}
				} else {
					// edgesが空の場合はnodesを使用
					for _, lang := range repo.Languages.Nodes {
						repoData.Languages.Nodes = append(repoData.Languages.Nodes, struct {
							Name string `json:"name"`
							Size int    `json:"size"`
						}{
							Name: lang.Name,
							Size: 0, // size情報がない場合は0
						})
					}
				}
			}

			// DefaultBranchRef
			if repo.DefaultBranchRef != nil {
				history := repo.DefaultBranchRef.Target.Commit.History
				repoData.DefaultBranchRef.Target.History.TotalCount = history.TotalCount
				for _, commit := range history.Nodes {
					if commit != nil {
						repoData.DefaultBranchRef.Target.History.Nodes = append(repoData.DefaultBranchRef.Target.History.Nodes, struct {
							CommittedDate string `json:"committedDate"`
							Author        struct {
								Date string `json:"date"`
							} `json:"author"`
						}{
							CommittedDate: commit.CommittedDate.Format(time.RFC3339),
							Author: struct {
								Date string `json:"date"`
							}{
								Date: commit.Author.Date.Format(time.RFC3339),
							},
						})
					}
				}
			}

			allRepos = append(allRepos, repoData)
		}

		if !query.User.Repositories.PageInfo.HasNextPage {
			break
		}
		after = stringPtr(query.User.Repositories.PageInfo.EndCursor)

		// 次のページ取得前に待機してGitHub APIへの負荷を軽減
		logger.Debug("次のページを取得する前に1秒待機します...")
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("コンテキストがキャンセルされました: %w", ctx.Err())
		case <-time.After(1 * time.Second):
			// 1秒待機
		}
	}

	return allRepos, nil
}

// FetchUserDetailsWithGraphQLGenerated 生成された型を使用してユーザー詳細情報を取得
func FetchUserDetailsWithGraphQLGenerated(ctx context.Context, token string, username string) (*UserDetailsGraphQLData, error) {
	graphqlClient, err := newGraphQLClient(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("GraphQLクライアントの作成に失敗しました: %w", err)
	}

	// 変数をmapに変換
	variables := map[string]interface{}{
		"login": username,
	}

	var query ghgraphql.UserDetailsQuery
	err = graphqlClient.Query(ctx, &query, variables)
	if err != nil {
		return nil, fmt.Errorf("GraphQLクエリの実行に失敗しました: %w", err)
	}

	// 生成された型からUserDetailsGraphQLDataに変換
	userDetails := &UserDetailsGraphQLData{
		ID:        query.User.ID,
		Name:      query.User.Name,
		Email:     query.User.Email,
		CreatedAt: query.User.CreatedAt,
	}

	if query.User.Repositories != nil {
		userDetails.Repositories.TotalCount = query.User.Repositories.TotalCount
		for _, repo := range query.User.Repositories.Nodes {
			node := struct {
				StargazerCount   int `json:"stargazerCount"`
				DefaultBranchRef struct {
					Target struct {
						History struct {
							TotalCount int `json:"totalCount"`
						} `json:"history"`
					} `json:"target"`
				} `json:"defaultBranchRef"`
			}{
				StargazerCount: repo.StargazerCount,
			}

			if repo.DefaultBranchRef != nil {
				node.DefaultBranchRef.Target.History.TotalCount = repo.DefaultBranchRef.Target.Commit.History.TotalCount
			}

			userDetails.Repositories.Nodes = append(userDetails.Repositories.Nodes, node)
		}
	}

	if query.User.ContributionsCollection != nil && query.User.ContributionsCollection.ContributionCalendar != nil {
		for _, week := range query.User.ContributionsCollection.ContributionCalendar.Weeks {
			var days []struct {
				ContributionCount int    `json:"contributionCount"`
				Date              string `json:"date"`
			}
			for _, day := range week.ContributionDays {
				days = append(days, struct {
					ContributionCount int    `json:"contributionCount"`
					Date              string `json:"date"`
				}{
					ContributionCount: day.ContributionCount,
					Date:              day.Date,
				})
			}
			userDetails.ContributionsCollection.ContributionCalendar.Weeks = append(
				userDetails.ContributionsCollection.ContributionCalendar.Weeks,
				struct {
					ContributionDays []struct {
						ContributionCount int    `json:"contributionCount"`
						Date              string `json:"date"`
					} `json:"contributionDays"`
				}{
					ContributionDays: days,
				},
			)
		}
	}

	userDetails.PullRequests.TotalCount = query.User.PullRequests.TotalCount
	userDetails.Issues.TotalCount = query.User.Issues.TotalCount

	return userDetails, nil
}

// ヘルパー関数
func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
