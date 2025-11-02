package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// downloadSchemaFromGitHub GitHub GraphQL APIから最新のスキーマをダウンロード
func downloadSchemaFromGitHub(ctx context.Context, token, outputPath string) error {
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN が設定されていません（イントロスペクションには認証が必要です）")
	}

	fmt.Println("📥 GitHub GraphQL APIから最新のスキーマをダウンロード中...")

	// イントロスペクションクエリ
	introspectionQuery := `
		query IntrospectionQuery {
			__schema {
				queryType {
					name
				}
				mutationType {
					name
				}
				subscriptionType {
					name
				}
				types {
					...FullType
				}
				directives {
					name
					description
					locations
					args {
						...InputValue
					}
				}
			}
		}

		fragment FullType on __Type {
			kind
			name
			description
			fields(includeDeprecated: true) {
				name
				description
				args {
					...InputValue
				}
				type {
					...TypeRef
				}
				isDeprecated
				deprecationReason
			}
			inputFields {
				...InputValue
			}
			interfaces {
				...TypeRef
			}
			enumValues(includeDeprecated: true) {
				name
				description
				isDeprecated
				deprecationReason
			}
			possibleTypes {
				...TypeRef
			}
		}

		fragment InputValue on __InputValue {
			name
			description
			type {
				...TypeRef
			}
			defaultValue
		}

		fragment TypeRef on __Type {
			kind
			name
			ofType {
				kind
				name
				ofType {
					kind
					name
					ofType {
						kind
						name
						ofType {
							kind
							name
							ofType {
								kind
								name
								ofType {
									kind
									name
									ofType {
										kind
										name
									}
								}
							}
						}
					}
				}
			}
		}
	`

	// GraphQLリクエストを作成
	reqBody := map[string]interface{}{
		"query": introspectionQuery,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("リクエストボディの作成に失敗しました: %w", err)
	}

	// HTTPリクエストを作成
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.github.com/graphql", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("HTTPリクエストの作成に失敗しました: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "update-gh-profile/1.0")

	// HTTPリクエストを実行
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("GraphQLリクエストの実行に失敗しました: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GraphQL APIエラー (ステータス: %d): %s", resp.StatusCode, string(body))
	}

	// レスポンスを読み取る
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("レスポンスの読み取りに失敗しました: %w", err)
	}

	// レスポンスをパース
	var graphQLResp struct {
		Data struct {
			Schema json.RawMessage `json:"__schema"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(body, &graphQLResp); err != nil {
		return fmt.Errorf("レスポンスのパースに失敗しました: %w", err)
	}

	if len(graphQLResp.Errors) > 0 {
		return fmt.Errorf("GraphQLエラー: %v", graphQLResp.Errors)
	}

	// スキーマをGraphQL SDL形式に変換（簡易版: JSONをそのまま保存）
	// 実際には、JSONスキーマをGraphQL SDLに変換する必要がありますが、
	// ここでは簡易的にJSONを保存し、後で変換ツールを使用することもできます
	// ただし、既存のパーサーはGraphQL SDLを期待しているので、変換が必要です

	// 簡易対応: スキーマが存在する場合のみ保存
	if len(graphQLResp.Data.Schema) == 0 {
		return fmt.Errorf("スキーマデータが空です")
	}

	// 注: イントロスペクション結果はJSON形式なので、GraphQL SDL形式に変換する必要があります
	// ここでは、既存のschema.docs.graphqlを使用するか、変換ツールが必要です
	// より実用的なアプローチとして、GitHub公式のスキーマファイルをダウンロードします

	fmt.Println("⚠️  イントロスペクション結果はJSON形式のため、GraphQL SDLへの変換が必要です")
	fmt.Println("   代わりに、GitHub公式のスキーマファイルをダウンロードします...")

	return downloadSchemaFromGitHubDocs(ctx, outputPath)
}

// downloadSchemaFromGitHubDocs GitHub公式ドキュメントからスキーマをダウンロード
func downloadSchemaFromGitHubDocs(ctx context.Context, outputPath string) error {
	// GitHub公式のスキーマエンドポイント
	// GitHub公式ドキュメントリポジトリからスキーマファイルを取得
	fmt.Println("📥 GitHub公式のスキーマファイルをダウンロード中...")

	// GitHub公式ドキュメントリポジトリのスキーマファイルURL
	// 複数の候補URLを試行
	schemaURLs := []string{
		// GitHub公式ドキュメントサイトから直接取得（推奨）
		"https://docs.github.com/public/fpt/schema.docs.graphql",
		// フォールバックURL
		"https://docs.github.com/public/schema.docs.graphql",
		// GitHub公式ドキュメントリポジトリのスキーマ
		"https://raw.githubusercontent.com/github/docs/main/data/graphql/schema.docs.graphql",
		"https://raw.githubusercontent.com/github/docs/main/content/graphql/reference/schema.docs.graphql",
		"https://raw.githubusercontent.com/github/docs/main/content/graphql/schema.docs.graphql",
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	var resp *http.Response
	var lastErr error

	// 複数のURLを試行
	for _, schemaURL := range schemaURLs {
		fmt.Printf("   試行中: %s\n", schemaURL)
		req, err := http.NewRequestWithContext(ctx, "GET", schemaURL, nil)
		if err != nil {
			lastErr = err
			continue
		}

		req.Header.Set("User-Agent", "update-gh-profile/1.0")

		resp, err = client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusOK {
			break // 成功したURLを使用
		}

		resp.Body.Close()
		lastErr = fmt.Errorf("HTTPエラー (ステータス: %d)", resp.StatusCode)
		resp = nil
	}

	if resp == nil {
		// 全てのURLで失敗した場合
		fmt.Printf("⚠️  スキーマファイルのダウンロードに失敗しました\n")
		fmt.Println("   ローカルの schema.docs.graphql を使用するか、手動でダウンロードしてください")
		fmt.Println("   参考: https://docs.github.com/en/graphql/overview/public-schema")
		fmt.Println("   手動ダウンロード: curl -o schema.docs.graphql https://docs.github.com/public/schema.docs.graphql")
		return fmt.Errorf("スキーマファイルをダウンロードできませんでした: %w", lastErr)
	}
	defer resp.Body.Close()

	// ファイルに保存
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("ファイルの作成に失敗しました: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("ファイルの書き込みに失敗しました: %w", err)
	}

	fmt.Printf("✅ スキーマファイルをダウンロードしました: %s\n", outputPath)
	return nil
}
