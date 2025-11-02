# GitHub GraphQL 型生成ツール

このツールは、GitHub の GraphQL スキーマから必要な型定義を抽出して Go コードを生成します。

## 使用方法

### 最新のスキーマを自動ダウンロードして生成（推奨）

```bash
# Makefileを使用
make generate

# 直接実行
go run ./tools/generate-github-types/main.go --download
```

### 既存のスキーマファイルを使用

```bash
# Makefileを使用
make generate-local

# 直接実行
go run ./tools/generate-github-types/main.go ./schema.docs.graphql
```

### オプション

- `--download`, `-d`: 最新のスキーマを GitHub 公式リポジトリから自動ダウンロード
- `--output`, `-o <dir>`: 出力ディレクトリを指定（デフォルト: `internal/graphql/generated`）

## スキーマファイルの取得方法

### 方法 1: 自動ダウンロード（推奨）

`--download`オプションを使用すると、GitHub 公式ドキュメントリポジトリから最新のスキーマを自動取得します。

```bash
go run ./tools/generate-github-types/main.go --download
```

### 方法 2: 手動ダウンロード

1. GitHub 公式ドキュメントからスキーマを取得:

   - https://docs.github.com/en/graphql/overview/public-schema
   - または、GitHub 公式ドキュメントリポジトリから:
     https://raw.githubusercontent.com/github/docs/main/content/graphql/reference/schema.docs.graphql

2. `schema.docs.graphql`として保存

3. コード生成を実行:
   ```bash
   make generate-local
   ```

### 方法 3: イントロスペクションクエリを使用

GitHub GraphQL API のイントロスペクション機能を使って、実行時にスキーマを取得することも可能です（実装は今後の拡張予定）。

## 生成されるファイル

- `internal/graphql/generated.go` - 型定義とクエリ構造体

## 追加の型を抽出する場合

`tools/generate-github-types/main.go`の`typesToExtract`スライスに型名を追加してください。

```go
var typesToExtract = []string{
    "User",
    "Repository",
    // 追加したい型名をここに追加
}
```

## CI/CD での使用

GitHub Actions などの CI/CD では、以下のように最新スキーマを自動取得できます：

```yaml
- name: Generate GraphQL types
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  run: make generate
```

## トラブルシューティング

### スキーマのダウンロードに失敗する場合

1. インターネット接続を確認
2. GitHub 公式リポジトリの URL が変更されていないか確認
3. ローカルの`schema.docs.graphql`ファイルを使用: `make generate-local`

### スキーマが古い場合

`--download`オプションを使用して最新版を取得してください。
