# Update GH Profile

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
```

**注意事項:**

- **トークンの設定**: `github_token` には `secrets.GITHUB_TOKEN` を渡してください。プライベートリポジトリを読み取る場合は、より広範囲な権限を持つ Personal Access Token を `github_token` に設定してください。
- **フォークの除外**: `exclude_forks: "true"` を設定すると、フォークされたリポジトリは統計から除外されます。
- **言語の除外**: `exclude_languages` パラメーターでランキングから除外する言語を指定できます。カンマ区切りで複数の言語を指定可能です（例: `"HTML,CSS,JSON"`）。大文字小文字は区別されません。
