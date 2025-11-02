package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"
)

// TypeDefinition 型定義の情報
type TypeDefinition struct {
	Name        string
	Description string
	Fields      []FieldDefinition
}

// FieldDefinition フィールド定義の情報
type FieldDefinition struct {
	Name        string
	Type        string
	Description string
	Required    bool
}

// 抽出する型のリスト（必要に応じて追加）
var typesToExtract = []string{
	"User",
	"Repository",
	"RepositoryOwner",
	"Language",
	"Commit",
	"CommitHistory",
	"PageInfo",
	"ContributionsCollection",
	"CommitContributionsByRepository",
	"RepositoryConnection",
}

func main() {
	var schemaPath string
	outputDir := "internal/graphql/generated"
	downloadLatest := false

	// コマンドライン引数のパース
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch arg {
		case "--download", "-d":
			downloadLatest = true
		case "--output", "-o":
			if i+1 < len(os.Args) {
				outputDir = os.Args[i+1]
				i++
			}
		default:
			if schemaPath == "" {
				schemaPath = arg
			} else if outputDir == "internal/graphql/generated" && !strings.HasPrefix(arg, "-") {
				outputDir = arg
			}
		}
	}

	// スキーマファイルが指定されていない、または最新版をダウンロードする場合
	if schemaPath == "" || downloadLatest {
		fmt.Println("📥 最新のGraphQLスキーマを取得中...")

		// デフォルトのスキーマファイルパス
		defaultSchemaPath := "schema.docs.graphql"
		if schemaPath == "" {
			schemaPath = defaultSchemaPath
		}

		ctx := context.Background()

		// GitHub公式からダウンロードを試みる
		err := downloadSchemaFromGitHubDocs(ctx, schemaPath)
		if err != nil {
			// ローカルのファイルが存在する場合はそれを使用
			if _, statErr := os.Stat(schemaPath); statErr == nil {
				fmt.Printf("⚠️  最新版の取得に失敗したため、既存のファイルを使用します: %s\n", schemaPath)
			} else if schemaPath == defaultSchemaPath {
				fmt.Fprintf(os.Stderr, "エラー: スキーマファイルのダウンロードに失敗し、ローカルファイルも見つかりません\n")
				fmt.Fprintf(os.Stderr, "Usage: %s [schema.docs.graphql] [--download] [--output <dir>]\n", os.Args[0])
				fmt.Fprintf(os.Stderr, "  または、ローカルの schema.docs.graphql ファイルを指定してください\n")
				os.Exit(1)
			} else {
				fmt.Fprintf(os.Stderr, "エラー: スキーマファイルが見つかりません: %s\n", schemaPath)
				os.Exit(1)
			}
		} else {
			fmt.Printf("✅ 最新のスキーマファイルを取得しました: %s\n", schemaPath)
		}
	}

	// スキーマファイルが存在するか確認
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "エラー: スキーマファイルが見つかりません: %s\n", schemaPath)
		fmt.Fprintf(os.Stderr, "Usage: %s [schema.docs.graphql] [--download] [--output <dir>]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  --download, -d: 最新のスキーマを自動ダウンロード\n")
		fmt.Fprintf(os.Stderr, "  --output, -o: 出力ディレクトリを指定\n")
		os.Exit(1)
	}

	fmt.Printf("📖 スキーマファイルを読み込み中: %s\n", schemaPath)
	types, err := parseSchema(schemaPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: スキーマの解析に失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ %d個の型定義を抽出しました\n", len(types))

	// 出力ディレクトリを作成
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: 出力ディレクトリの作成に失敗しました: %v\n", err)
		os.Exit(1)
	}

	// 現在は internal/graphql/generated.go は手動でメンテナンスされています
	// 将来的に自動生成機能を追加する予定です
	fmt.Println("ℹ️  internal/graphql/generated.go は手動でメンテナンスされています")
	fmt.Printf("✅ スキーマファイルの処理が完了しました（%d個の型を抽出）\n", len(types))

	fmt.Println("✅ コード生成が完了しました！")
}

// parseSchema GraphQLスキーマファイルを解析して型定義を抽出
func parseSchema(path string) (map[string]TypeDefinition, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("ファイルを開けませんでした: %w", err)
	}
	defer file.Close()

	types := make(map[string]TypeDefinition)
	scanner := bufio.NewScanner(file)

	var currentType *TypeDefinition
	var inTypeDefinition bool

	typePattern := regexp.MustCompile(`^type\s+(\w+)\s*(implements\s+[\w\s&]+)?\s*\{`)
	fieldPattern := regexp.MustCompile(`^\s+(\w+)\s*:\s*([!\w\[\]\(\),&\s]+)`)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// 型定義の開始を検出
		if matches := typePattern.FindStringSubmatch(line); matches != nil {
			typeName := matches[1]
			if contains(typesToExtract, typeName) {
				currentType = &TypeDefinition{
					Name:   typeName,
					Fields: []FieldDefinition{},
				}
				types[typeName] = *currentType
				inTypeDefinition = true
				continue
			}
		}

		// 型定義の中にあるフィールドを抽出
		if inTypeDefinition && currentType != nil {
			if trimmed == "}" || trimmed == "}" {
				inTypeDefinition = false
				currentType = nil
				continue
			}

			if matches := fieldPattern.FindStringSubmatch(line); matches != nil {
				fieldName := matches[1]
				fieldType := strings.TrimSpace(matches[2])
				required := strings.HasSuffix(fieldType, "!")

				field := FieldDefinition{
					Name:     fieldName,
					Type:     fieldType,
					Required: required,
				}
				// currentTypeを更新する必要があるため、mapから取得して更新
				if t, ok := types[currentType.Name]; ok {
					t.Fields = append(t.Fields, field)
					types[currentType.Name] = t
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("スキャンエラー: %w", err)
	}

	return types, nil
}

// contains スライスに要素が含まれているかチェック
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// generateTypes 型定義のGoコードを生成
func generateTypes(outputPath string, types map[string]TypeDefinition) error {
	tmpl := `// Code generated by tools/generate-github-types/main.go. DO NOT EDIT.

package generated

import (
	"time"
)

{{range .Types}}
// {{.Name}} {{.Description}}
type {{.Name}} struct {
{{range .Fields}}
	{{.Name | title}} {{.Type | goType}} ` + "`" + `graphql:"{{.Name}}"` + "`" + `{{end}}
}
{{end}}
`

	funcMap := template.FuncMap{
		"title": strings.Title,
		"goType": func(gqlType string) string {
			return convertGraphQLTypeToGo(gqlType)
		},
	}

	t := template.Must(template.New("types").Funcs(funcMap).Parse(tmpl))

	data := struct {
		Types []TypeDefinition
	}{
		Types: extractTypesInOrder(types),
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("ファイルを作成できませんでした: %w", err)
	}
	defer file.Close()

	return t.Execute(file, data)
}

// generateQueries クエリ構造体のGoコードを生成（将来の拡張用）
func generateQueries(outputPath string) error {
	// 現在は internal/graphql/generated.go に手動で定義されています
	return nil
}

// convertGraphQLTypeToGo GraphQL型をGo型に変換
func convertGraphQLTypeToGo(gqlType string) string {
	gqlType = strings.TrimSpace(gqlType)
	required := strings.HasSuffix(gqlType, "!")
	if required {
		gqlType = strings.TrimSuffix(gqlType, "!")
	}

	// 配列型の処理
	if strings.HasPrefix(gqlType, "[") && strings.HasSuffix(gqlType, "]") {
		inner := strings.TrimPrefix(strings.TrimSuffix(gqlType, "]"), "[")
		goType := convertGraphQLTypeToGo(inner)
		result := "[]" + goType
		if !required {
			result = "*" + result
		}
		return result
	}

	// 基本的な型のマッピング
	switch gqlType {
	case "String":
		if required {
			return "string"
		}
		return "*string"
	case "Int":
		if required {
			return "int"
		}
		return "*int"
	case "Float":
		if required {
			return "float64"
		}
		return "*float64"
	case "Boolean":
		if required {
			return "bool"
		}
		return "*bool"
	case "ID":
		if required {
			return "string"
		}
		return "*string"
	case "DateTime", "GitTimestamp":
		if required {
			return "time.Time"
		}
		return "*time.Time"
	case "URI":
		if required {
			return "string"
		}
		return "*string"
	default:
		// カスタム型（User, Repositoryなど）
		if required {
			return gqlType
		}
		return "*" + gqlType
	}
}

// extractTypesInOrder 型を順序よく抽出
func extractTypesInOrder(types map[string]TypeDefinition) []TypeDefinition {
	var result []TypeDefinition
	for _, typeName := range typesToExtract {
		if t, ok := types[typeName]; ok {
			result = append(result, t)
		}
	}
	return result
}
