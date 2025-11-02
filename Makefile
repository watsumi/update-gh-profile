.PHONY: generate build test

# コード生成を実行
generate:
	@echo "📝 GraphQL型定義を生成中..."
	@go run ./tools/generate-github-types/main.go ./schema.docs.graphql ./internal/graphql/generated
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

