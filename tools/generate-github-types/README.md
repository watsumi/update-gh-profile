# GitHub GraphQL 型生成ツール

このツールは、GitHubのGraphQLスキーマから必要な型定義を抽出してGoコードを生成します。

## 使用方法

```bash
# スキーマファイルから型を生成
go run ./tools/generate-github-types/main.go schema.docs.graphql

# 出力ディレクトリを指定
go run ./tools/generate-github-types/main.go schema.docs.graphql internal/graphql/generated
```

## 生成されるファイル

- `internal/graphql/generated/types.go` - 型定義
- `internal/graphql/generated/queries.go` - クエリ構造体

## 追加の型を抽出する場合

`tools/generate-github-types/main.go`の`typesToExtract`スライスに型名を追加してください。

```go
var typesToExtract = []string{
    "User",
    "Repository",
    // 追加したい型名をここに追加
}
```

