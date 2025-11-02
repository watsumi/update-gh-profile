# プロジェクト構成

このプロジェクトは、Go言語の標準的なプロジェクトレイアウトに基づいて構成されています。

## 推奨ディレクトリ構成

```
update-gh-profile/
├── cmd/
│   └── update-gh-profile/
│       └── main.go              # エントリーポイント
├── internal/
│   ├── fetcher/                 # Repository Fetcher（設計ドキュメント参照）
│   │   ├── repository.go       # リポジトリ情報取得
│   │   ├── languages.go         # 言語情報取得
│   │   ├── commits.go           # コミット情報取得
│   │   └── pullrequests.go      # PR情報取得
│   ├── aggregator/              # Metrics Aggregator（設計ドキュメント参照）
│   │   ├── languages.go         # 言語データ集計
│   │   ├── commits.go           # コミット推移集計
│   │   └── summary.go           # サマリー統計集計
│   ├── generator/               # SVG Generator（設計ドキュメント参照）
│   │   ├── language_chart.go    # 言語ランキンググラフ
│   │   ├── commit_history.go    # コミット推移グラフ
│   │   ├── commit_time.go       # コミット時間帯グラフ
│   │   └── summary_card.go      # サマリーカード
│   ├── updater/                 # README Updater（設計ドキュメント参照）
│   │   ├── readme.go            # README更新処理
│   │   └── git.go               # Git操作
│   └── config/                  # 設定管理
│       └── config.go            # 環境変数と設定
├── pkg/
│   └── models/                  # データモデル（外部公開可能）
│       ├── repository.go        # Repository構造体
│       ├── metrics.go           # Metrics構造体
│       └── stats.go             # Stats構造体
├── .github/
│   └── workflows/
│       └── update-readme.yml    # GitHub Actionsワークフロー
├── test/                        # テストファイル
│   ├── fetcher_test.go
│   ├── aggregator_test.go
│   └── integration_test.go
├── go.mod
├── go.sum
├── README.md
└── .gitignore
```

## 各ディレクトリの説明

### `cmd/`
- **目的**: アプリケーションのエントリーポイント
- **使用**: `main.go` のみ配置
- **理由**: 複数のバイナリを作る可能性を考慮（将来的に拡張可能）

### `internal/`
- **目的**: プロジェクト内部でのみ使用するパッケージ
- **特徴**: 外部パッケージからインポート不可（Goの仕様）
- **構成**: 設計ドキュメントのコンポーネントに基づいて分割
  - `fetcher/`: Repository Fetcher
  - `aggregator/`: Metrics Aggregator
  - `generator/`: SVG Generator
  - `updater/`: README Updater
  - `config/`: 設定管理

### `pkg/`
- **目的**: 外部公開可能なパッケージ（現在は主にデータモデル）
- **使用**: テストコードなどから使用可能

### `.github/workflows/`
- **目的**: GitHub Actionsワークフロー定義

### `test/`
- **目的**: 統合テストやエンドツーエンドテスト
- **注意**: 単体テスト（`*_test.go`）は各パッケージと同ディレクトリに配置

## 参考リポジトリ

### 1. Standard Go Project Layout
**リポジトリ**: https://github.com/golang-standards/project-layout
- Goコミュニティで広く使われる標準的なプロジェクトレイアウト
- `cmd/`, `internal/`, `pkg/` などの推奨構成を説明

### 2. GitHub CLI (gh)
**リポジトリ**: https://github.com/cli/cli
- GitHub APIを使用するCLIツール
- 類似した機能を持つ実装例
- ディレクトリ構成: `cmd/`, `internal/`, `pkg/` を使用

### 3. Terraform
**リポジトリ**: https://github.com/hashicorp/terraform
- 大規模なGoプロジェクトの構成例
- `internal/` パッケージの使い方の参考

### 4. GitHub Actions ツール例
**リポジトリ**: 
- https://github.com/github/super-linter
- https://github.com/actions/toolkit
- Goで実装されたActionsツールの例

## このプロジェクトでの採用方針

設計ドキュメントに基づき、以下の構成を採用：

1. **シンプルな開始**: 最初は `cmd/` と `internal/` のみ
2. **段階的な拡張**: 必要に応じて `pkg/` を追加
3. **コンポーネント分離**: 設計ドキュメントの4つのコンポーネントを `internal/` 内で分離
4. **テスト**: 各パッケージに `*_test.go` を配置

