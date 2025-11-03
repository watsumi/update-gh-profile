package repository

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/hasura/go-graphql-client"
	"golang.org/x/oauth2"
)

// GraphQLQueries GraphQLクエリ定義
var (
	// QueryReposPerLanguage リポジトリごとの言語情報を取得するクエリ
	QueryReposPerLanguage = `
query ReposPerLanguage($login: String!, $endCursor: String) {
  user(login: $login) {
    repositories(isFork: false, first: 100, after: $endCursor, ownerAffiliations: OWNER) {
      nodes {
        name
        owner {
          login
        }
        primaryLanguage {
          name
        }
        languages(first: 100) {
          nodes {
            name
            size
          }
          totalSize
        }
        stargazerCount
        defaultBranchRef {
          target {
            ... on Commit {
              history(first: 100) {
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

	// QueryUserDetails ユーザーの詳細情報を取得するクエリ
	QueryUserDetails = `
query UserDetails($login: String!) {
  user(login: $login) {
    id
    name
    email
    createdAt
    repositories(first: 100, privacy: PUBLIC, isFork: false, ownerAffiliations: OWNER, orderBy: {direction: DESC, field: STARGAZERS}) {
      totalCount
      nodes {
        stargazerCount
        defaultBranchRef {
          target {
            ... on Commit {
              history(first: 1) {
                totalCount
              }
            }
          }
        }
      }
    }
    contributionsCollection {
      contributionCalendar {
        weeks {
          contributionDays {
            contributionCount
            date
          }
        }
      }
    }
    pullRequests(first: 1) {
      totalCount
    }
    issues(first: 1) {
      totalCount
    }
  }
}`

	// QueryCommitLanguages コミットごとの言語使用状況を取得するクエリ
	QueryCommitLanguages = `
query CommitLanguages($login: String!) {
  user(login: $login) {
    contributionsCollection {
      commitContributionsByRepository(maxRepositories: 100) {
        repository {
          name
          owner {
            login
          }
          primaryLanguage {
            name
          }
          languages(first: 20) {
            edges {
              node {
                name
              }
              size
            }
          }
          defaultBranchRef {
            target {
              ... on Commit {
                history(first: 50) {
                  edges {
                    node {
                      oid
                      message
                      committedDate
                      additions
                      deletions
                    }
                  }
                }
              }
            }
          }
        }
        contributions {
          totalCount
        }
      }
    }
  }
}`

	// QueryProductiveTime コミット時間帯を取得するクエリ
	QueryProductiveTime = `
query ProductiveTime($login: String!, $userId: ID!, $since: GitTimestamp!, $until: GitTimestamp!) {
  user(login: $login) {
    contributionsCollection {
      commitContributionsByRepository(maxRepositories: 50) {
        repository {
          name
          owner {
            login
          }
          defaultBranchRef {
            target {
              ... on Commit {
                history(first: 50, since: $since, until: $until, author: {id: $userId}) {
                  edges {
                    node {
                      committedDate
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`

	// QueryViewer 認証ユーザー情報を取得するクエリ
	QueryViewer = `
query Viewer {
  viewer {
    login
    id
  }
}`
)

// newGraphQLClient GraphQLクライアントを作成する
func newGraphQLClient(ctx context.Context, token string) (*graphql.Client, error) {
	if token == "" {
		return nil, fmt.Errorf("認証トークンが設定されていません")
	}

	// OAuth2トークンを使用してHTTPクライアントを作成
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))

	// HTTPクライアントにタイムアウトを設定
	// 読み取りタイムアウト: 60秒
	httpClient.Timeout = 60 * time.Second

	// Transportに接続タイムアウトを設定
	if oauth2Transport, ok := httpClient.Transport.(*oauth2.Transport); ok {
		if baseTransport, ok := oauth2Transport.Base.(*http.Transport); ok {
			if baseTransport.DialContext == nil {
				baseTransport.DialContext = (&net.Dialer{
					Timeout: 30 * time.Second,
				}).DialContext
			}
		}
	}

	// GraphQLクライアントを作成
	graphqlClient := graphql.NewClient("https://api.github.com/graphql", httpClient)

	return graphqlClient, nil
}

// FetchViewer GraphQLを使用して認証ユーザー情報を取得する
func FetchViewer(ctx context.Context, token string) (string, string, error) {
	graphqlClient, err := newGraphQLClient(ctx, token)
	if err != nil {
		return "", "", fmt.Errorf("GraphQLクライアントの作成に失敗しました: %w", err)
	}

	variables := map[string]interface{}{}

	var response struct {
		Viewer struct {
			Login string `json:"login"`
			ID    string `json:"id"`
		} `json:"viewer"`
	}

	err = graphqlClient.Exec(ctx, QueryViewer, &response, variables)
	if err != nil {
		return "", "", fmt.Errorf("GraphQLクエリの実行に失敗しました: %w", err)
	}

	return response.Viewer.Login, response.Viewer.ID, nil
}

// FetchRepositoriesWithGraphQL GraphQLを使用してリポジトリ情報を一括取得する
func FetchRepositoriesWithGraphQL(ctx context.Context, token string, username string, excludeForks bool) ([]*RepositoryGraphQLData, error) {
	graphqlClient, err := newGraphQLClient(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("GraphQLクライアントの作成に失敗しました: %w", err)
	}

	var allRepos []*RepositoryGraphQLData
	var endCursor *string

	for {
		// GraphQLクエリの変数を準備
		variables := map[string]interface{}{
			"login": username,
		}
		if endCursor != nil {
			variables["endCursor"] = *endCursor
		}

		// GraphQLクエリを実行（文字列クエリを使用）
		var response struct {
			User struct {
				Repositories struct {
					Nodes    []*RepositoryGraphQLData `json:"nodes"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"repositories"`
			} `json:"user"`
		}

		// Execを使用してクエリを実行
		err := graphqlClient.Exec(ctx, QueryReposPerLanguage, &response, variables)
		if err != nil {
			return nil, fmt.Errorf("GraphQLクエリの実行に失敗しました: %w", err)
		}

		allRepos = append(allRepos, response.User.Repositories.Nodes...)

		if !response.User.Repositories.PageInfo.HasNextPage {
			break
		}
		endCursor = &response.User.Repositories.PageInfo.EndCursor
	}

	return allRepos, nil
}

// RepositoryGraphQLData GraphQLから取得したリポジトリデータ
type RepositoryGraphQLData struct {
	Name            string    `json:"name"`
	Owner           OwnerData `json:"owner"`
	PrimaryLanguage struct {
		Name string `json:"name"`
	} `json:"primaryLanguage"`
	Languages struct {
		Nodes []struct {
			Name string `json:"name"`
			Size int    `json:"size"`
		} `json:"nodes"`
		TotalSize int `json:"totalSize"`
	} `json:"languages"`
	StargazerCount   int `json:"stargazerCount"`
	DefaultBranchRef struct {
		Target struct {
			History struct {
				TotalCount int `json:"totalCount"`
				Nodes      []struct {
					CommittedDate string `json:"committedDate"`
					Author        struct {
						Date string `json:"date"`
					} `json:"author"`
				} `json:"nodes"`
			} `json:"history"`
		} `json:"target"`
	} `json:"defaultBranchRef"`
}

// OwnerData オーナー情報
type OwnerData struct {
	Login string `json:"login"`
}

// FetchUserDetailsWithGraphQL GraphQLを使用してユーザー詳細情報を取得する
func FetchUserDetailsWithGraphQL(ctx context.Context, token string, username string) (*UserDetailsGraphQLData, error) {
	graphqlClient, err := newGraphQLClient(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("GraphQLクライアントの作成に失敗しました: %w", err)
	}

	variables := map[string]interface{}{
		"login": username,
	}

	var response struct {
		User UserDetailsGraphQLData `json:"user"`
	}

	err = graphqlClient.Exec(ctx, QueryUserDetails, &response, variables)
	if err != nil {
		return nil, fmt.Errorf("GraphQLクエリの実行に失敗しました: %w", err)
	}

	return &response.User, nil
}

// UserDetailsGraphQLData GraphQLから取得したユーザー詳細データ
type UserDetailsGraphQLData struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	CreatedAt    string `json:"createdAt"`
	Repositories struct {
		TotalCount int `json:"totalCount"`
		Nodes      []struct {
			StargazerCount   int `json:"stargazerCount"`
			DefaultBranchRef struct {
				Target struct {
					History struct {
						TotalCount int `json:"totalCount"`
					} `json:"history"`
				} `json:"target"`
			} `json:"defaultBranchRef"`
		} `json:"nodes"`
	} `json:"repositories"`
	ContributionsCollection struct {
		ContributionCalendar struct {
			Weeks []struct {
				ContributionDays []struct {
					ContributionCount int    `json:"contributionCount"`
					Date              string `json:"date"`
				} `json:"contributionDays"`
			} `json:"weeks"`
		} `json:"contributionCalendar"`
	} `json:"contributionsCollection"`
	PullRequests struct {
		TotalCount int `json:"totalCount"`
	} `json:"pullRequests"`
	Issues struct {
		TotalCount int `json:"totalCount"`
	} `json:"issues"`
}

// FetchCommitLanguagesWithGraphQL GraphQLを使用してコミットごとの言語使用状況を取得する
// リポジトリの複数言語情報を使用して、より多くの言語を取得する
func FetchCommitLanguagesWithGraphQL(ctx context.Context, token string, username string) (map[string]map[string]int, error) {
	graphqlClient, err := newGraphQLClient(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("GraphQLクライアントの作成に失敗しました: %w", err)
	}

	variables := map[string]interface{}{
		"login": username,
	}

	var response struct {
		User struct {
			ContributionsCollection struct {
				CommitContributionsByRepository []struct {
					Repository struct {
						Name            string    `json:"name"`
						Owner           OwnerData `json:"owner"`
						PrimaryLanguage struct {
							Name string `json:"name"`
						} `json:"primaryLanguage"`
						Languages struct {
							Edges []struct {
								Node struct {
									Name string `json:"name"`
								} `json:"node"`
								Size int `json:"size"`
							} `json:"edges"`
						} `json:"languages"`
						DefaultBranchRef struct {
							Target struct {
								History struct {
									Edges []struct {
										Node struct {
											Oid           string `json:"oid"`
											Message       string `json:"message"`
											CommittedDate string `json:"committedDate"`
											Additions     int    `json:"additions"`
											Deletions     int    `json:"deletions"`
										} `json:"node"`
									} `json:"edges"`
								} `json:"history"`
							} `json:"target"`
						} `json:"defaultBranchRef"`
					} `json:"repository"`
					Contributions struct {
						TotalCount int `json:"totalCount"`
					} `json:"contributions"`
				} `json:"commitContributionsByRepository"`
			} `json:"contributionsCollection"`
		} `json:"user"`
	}

	err = graphqlClient.Exec(ctx, QueryCommitLanguages, &response, variables)
	if err != nil {
		return nil, fmt.Errorf("GraphQLクエリの実行に失敗しました: %w", err)
	}

	// データを変換（リポジトリの複数言語情報を使用）
	commitLanguages := make(map[string]map[string]int)

	for _, repoContrib := range response.User.ContributionsCollection.CommitContributionsByRepository {
		// リポジトリで使用されている言語リストを取得
		repoLanguages := make(map[string]int)
		
		// languagesエッジから言語とサイズを取得（サイズが大きい順にソート済み）
		for _, langEdge := range repoContrib.Repository.Languages.Edges {
			if langEdge.Node.Name != "" {
				// サイズを重みとして使用（大きなファイルほど重要）
				repoLanguages[langEdge.Node.Name] = langEdge.Size
			}
		}
		
		// プライマリ言語も追加（languagesに含まれていない場合）
		if repoContrib.Repository.PrimaryLanguage.Name != "" {
			if _, exists := repoLanguages[repoContrib.Repository.PrimaryLanguage.Name]; !exists {
				repoLanguages[repoContrib.Repository.PrimaryLanguage.Name] = 1
			}
		}

		// 各コミットにリポジトリの言語を適用
		for _, edge := range repoContrib.Repository.DefaultBranchRef.Target.History.Edges {
			commitKey := edge.Node.CommittedDate
			if commitKey == "" {
				continue
			}

			if commitLanguages[commitKey] == nil {
				commitLanguages[commitKey] = make(map[string]int)
			}

			// リポジトリの各言語をコミットに追加（サイズに比例した重み）
			for lang, size := range repoLanguages {
				// サイズの平方根を使用して重み付け（大きな差を緩和）
				weight := 1
				if size > 1000 {
					// 大きなファイルの場合は重みを増やす（ただし上限あり）
					weight = 2
				}
				commitLanguages[commitKey][lang] += weight
			}
		}
	}

	return commitLanguages, nil
}

// FetchProductiveTimeWithGraphQL GraphQLを使用してコミット時間帯を取得する
func FetchProductiveTimeWithGraphQL(ctx context.Context, token string, username, userID string, since, until time.Time) (map[int]int, error) {
	graphqlClient, err := newGraphQLClient(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("GraphQLクライアントの作成に失敗しました: %w", err)
	}

	variables := map[string]interface{}{
		"login":  username,
		"userId": userID,
		"since":  since.Format(time.RFC3339),
		"until":  until.Format(time.RFC3339),
	}

	var response struct {
		User struct {
			ContributionsCollection struct {
				CommitContributionsByRepository []struct {
					Repository struct {
						Name             string    `json:"name"`
						Owner            OwnerData `json:"owner"`
						DefaultBranchRef struct {
							Target struct {
								History struct {
									Edges []struct {
										Node struct {
											CommittedDate string `json:"committedDate"`
										} `json:"node"`
									} `json:"edges"`
								} `json:"history"`
							} `json:"target"`
						} `json:"defaultBranchRef"`
					} `json:"repository"`
				} `json:"commitContributionsByRepository"`
			} `json:"contributionsCollection"`
		} `json:"user"`
	}

	err = graphqlClient.Exec(ctx, QueryProductiveTime, &response, variables)
	if err != nil {
		return nil, fmt.Errorf("GraphQLクエリの実行に失敗しました: %w", err)
	}

	// 時間帯ごとのコミット数を集計
	timeDistribution := make(map[int]int)
	for _, repoContrib := range response.User.ContributionsCollection.CommitContributionsByRepository {
		for _, edge := range repoContrib.Repository.DefaultBranchRef.Target.History.Edges {
			committedDate, err := time.Parse(time.RFC3339, edge.Node.CommittedDate)
			if err != nil {
				continue
			}
			hour := committedDate.UTC().Hour()
			timeDistribution[hour]++
		}
	}

	return timeDistribution, nil
}
