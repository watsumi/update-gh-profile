# update-gh-profile

GitHub ユーザーのリポジトリメトリクスを自動分析し、README.md を更新する GitHub Actions ワークフロー

## 概要

このツールは、GitHub ユーザーの全リポジトリから以下の情報を収集し、可視化します：

- 使用言語ランキング
- コミット推移グラフ
- コミット時間帯分析
- コミットごとの使用言語 Top5
- サマリーカード（スター数、リポジトリ数、コミット数、PR 数）

## セットアップ

### 必要な環境

- Go 1.21 以上
- GitHub Personal Access Token (環境変数 `GITHUB_TOKEN`)

### ローカル実行

```bash
# 依存関係のインストール
go mod download

# 実行
export GITHUB_TOKEN=your_token_here
go run cmd/update-gh-profile/main.go

# または、ビルドして実行
go build -o update-gh-profile cmd/update-gh-profile/main.go
./update-gh-profile
```

### GitHub Actions での使用

このリポジトリを別のリポジトリの GitHub Actions ワークフローで使用できます。

詳細は以下を参照してください：

- [README_ACTION.md](README_ACTION.md) - 基本的な使用手順
- [USAGE_ACTION.md](USAGE_ACTION.md) - 詳細な使用例
- [PRIVATE_REPO_SETUP.md](PRIVATE_REPO_SETUP.md) - プライベートリポジトリでの設定手順

#### クイックスタート

```yaml
- uses: watsumi/update-gh-profile@main
  with:
    github_token: ${{ secrets.GITHUB_TOKEN }}
    exclude_forks: "true"
```

#### Workflow の設定例

プロフィールリポジトリ（`username/username` 形式のリポジトリ）で使用する場合の完全な workflow ファイル例：

```yaml
name: Update GitHub Profile

on:
  schedule:
    # 毎日 00:00 UTC（日本時間 09:00）に実行
    - cron: "0 0 * * *"
  workflow_dispatch: # 手動実行も可能

permissions:
  contents: write # README.md を更新するために必要

jobs:
  update-profile:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Update GitHub Profile
        uses: watsumi/update-gh-profile@main
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          exclude_forks: "true"
          exclude_languages: "HTML,CSS,JSON" # 除外する言語名（カンマ区切り）
        # 注意: このアクションは内部で自動的にコミットとプッシュを実行します
        # permissions: contents: write により認証情報が自動的に設定されるため、
        # github_token_write は不要です
```

**注意事項:**

- **`permissions: contents: write` について**: この権限設定により `secrets.GITHUB_TOKEN` に書き込み権限が付与され、`actions/checkout@v4` でチェックアウトされたリポジトリに対して `git push` が自動的に認証されます。そのため、`github_token_write` パラメーターは不要です（削除されました）。
- **自動コミット・プッシュ**: このアクションは README.md を更新した後、自動的にコミットとプッシュを実行します。追加のステップは不要です。
- **トークンの設定**: `github_token` には `secrets.GITHUB_TOKEN` を渡してください。プライベートリポジトリを読み取る場合は、より広範囲な権限を持つ Personal Access Token を `github_token` に設定してください。
- **認証ユーザーの自動取得**: このツールは認証ユーザー自身のリポジトリのみを取得します。認証ユーザーは自動的に取得されるため、ユーザー名の指定は不要です。
- **フォークの除外**: `exclude_forks: "true"` を設定すると、フォークされたリポジトリは統計から除外されます。
- **言語の除外**: `exclude_languages` パラメーターでランキングから除外する言語を指定できます。カンマ区切りで複数の言語を指定可能です（例: `"HTML,CSS,JSON"`）。大文字小文字は区別されません。

## プロジェクト構成

このプロジェクトは、[Go 公式ドキュメントの推奨レイアウト](https://go.dev/doc/modules/layout)に基づいて構成されています。

### ディレクトリ構成

```
update-gh-profile/
├── cmd/
│   └── update-gh-profile/        # エントリーポイント（実行可能コマンド）
│       └── main.go
├── internal/                     # 内部パッケージ（外部からインポート不可）
│   ├── fetcher/                  # Repository Fetcher: GitHub API からのデータ取得
│   ├── aggregator/               # Metrics Aggregator: メトリクスデータの集計と分析
│   ├── generator/                # SVG Generator: SVG グラフの生成
│   ├── updater/                  # README Updater: README.md の更新と Git 操作
│   └── config/                   # 設定管理: 環境変数と設定値の処理
├── pkg/
│   └── models/                   # データモデル（外部公開可能なパッケージ）
├── .github/
│   └── workflows/                # GitHub Actions ワークフロー定義
├── go.mod                        # Go モジュール定義
├── go.sum                        # 依存関係のチェックサム
└── README.md                     # このファイル
```

### 各ディレクトリの説明

#### `cmd/update-gh-profile/`

- **目的**: 実行可能なコマンドラインプログラムのエントリーポイント
- **特徴**: `package main` を宣言し、`func main()` を含む
- **理由**: 複数のコマンドを作成する場合、`cmd/` ディレクトリに各コマンドを配置することで、構造が明確になります（[Go 公式ドキュメント](https://go.dev/doc/modules/layout#multiple-commands)を参照）
- **インストール**: `go install github.com/watsumi/update-gh-profile/cmd/update-gh-profile@latest` でインストール可能

#### `internal/`

- **目的**: プロジェクト内部でのみ使用されるパッケージ
- **特徴**: Go の仕様により、`internal/` ディレクトリ内のパッケージは、このモジュールの外部からインポートできません（[Go 公式ドキュメント](https://go.dev/doc/modules/layout#package-or-command-with-supporting-packages)を参照）
- **メリット**:
  - 外部への影響を気にせずリファクタリング可能
  - 内部実装の詳細を外部に公開しない
- **構成**: 設計ドキュメントの 4 つの主要コンポーネントに対応
  - `fetcher/`: GitHub API を使用したリポジトリ情報取得
  - `aggregator/`: 収集したデータの集計と分析
  - `generator/`: メトリクスデータから SVG グラフ生成
  - `updater/`: README.md の更新と Git 操作
  - `config/`: 環境変数や設定値の管理

#### `pkg/models/`

- **目的**: 外部にも公開可能なパッケージ（将来的にライブラリとして公開する場合）
- **特徴**: 他のプロジェクトからもインポート可能
- **使用例**: テストコードや、将来的にこのプロジェクトをライブラリとして使用する場合

#### `.github/workflows/`

- **目的**: GitHub Actions ワークフローの定義ファイル
- **内容**: スケジュール実行や手動実行の設定

### Go 公式ドキュメントに基づく設計

このプロジェクトは、[Go 公式ドキュメントの「Packages and commands in the same repository」](https://go.dev/doc/modules/layout#packages-and-commands-in-the-same-repository)パターンに従っています：

- **実行可能コマンド**: `cmd/update-gh-profile/` に配置
- **内部パッケージ**: `internal/` に配置（外部からインポート不可）
- **公開可能パッケージ**: `pkg/` に配置（将来の拡張のため）

また、サーバープロジェクトと同様に、すべての内部ロジックを `internal/` ディレクトリに配置し、コマンドは `cmd/` ディレクトリに配置しています（[Go 公式ドキュメントの「Server project」](https://go.dev/doc/modules/layout#server-project)を参照）。

### テストファイルの配置

- **単体テスト**: 各パッケージと同じディレクトリに `*_test.go` ファイルを配置（例: `internal/fetcher/repository_test.go`）
- **統合テスト**: `test/` ディレクトリに配置（今後追加予定）

## 開発状況

現在、実装を進めています。詳細は `.kiro/specs/github/` を参照してください。

## 参考資料

- [Go 公式ドキュメント: Organizing a Go module](https://go.dev/doc/modules/layout) - Go プロジェクトの構成に関する公式ガイドライン
