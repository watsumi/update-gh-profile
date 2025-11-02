.PHONY: generate build test

# コード生成を実行（最新のスキーマを自動ダウンロード）
generate:
	@echo "📝 最新のGraphQLスキーマを取得中..."
	@cd ./tools/generate-github-types && go run . --download
	@echo "✅ スキーマファイルを更新しました（既存の generated.go は手動メンテナンス）"
	@echo "   注意: internal/graphql/generated.go は手動でメンテナンスされています"

# 既存のスキーマファイルを使用してコード生成
generate-local:
	@echo "📝 GraphQL型定義を生成中（ローカルスキーマを使用）..."
	@cd ./tools/generate-github-types && go run . ../../schema.docs.graphql ../../internal/graphql/
	@echo "✅ コード生成が完了しました"

# ビルド
build:
	@go build ./cmd/update-gh-profile/
	@echo "✅ ビルドが完了しました"

# テスト
test:
	@go test ./...

# フォーマット
fmt:
	@go fmt ./...
	@echo "✅ コードフォーマットが完了しました"

# 全て実行（生成→フォーマット→ビルド→テスト）
all: generate fmt build test

