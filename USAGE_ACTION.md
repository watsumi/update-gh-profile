# GitHub Action としての使用方法

このリポジトリを GitHub Action として使用する方法を説明します。

## 基本的な使用方法

別のリポジトリのワークフローで、以下のように使用できます：

```yaml
name: Test Repository Fetch

on:
  workflow_dispatch:
  schedule:
    - cron: '0 0 * * *' # 毎日実行

jobs:
  test:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - uses: actions/checkout@v4
      
      - name: Fetch repositories
        uses: watsumi/update-gh-profile@main
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          username: ${{ github.repository_owner }}
          exclude_forks: "true"
```

## 入力パラメータ

| パラメータ | 説明 | 必須 | デフォルト値 |
|----------|------|------|------------|
| `github_token` | GitHub Personal Access Token | 任意 | `secrets.GITHUB_TOKEN` |
| `username` | GitHub ユーザー名 | 任意 | リポジトリのオーナー |
| `exclude_forks` | フォークリポジトリを除外するか（`true`/`false`） | 任意 | `true` |

## 出力値

| 出力 | 説明 |
|------|------|
| `repository_count` | 取得したリポジトリ数 |

## 使用例

### 例1: 基本的な使用（デフォルト設定）

```yaml
- uses: watsumi/update-gh-profile@main
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### 例2: 特定のユーザーのリポジトリを取得

```yaml
- uses: watsumi/update-gh-profile@main
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  with:
    username: "octocat"
```

### 例3: フォークも含めて取得

```yaml
- uses: watsumi/update-gh-profile@main
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  with:
    username: "octocat"
    exclude_forks: "false"
```

## 注意事項

1. **トークンの権限**: `GITHUB_TOKEN` には `repo` スコープが必要です。
2. **ブランチ指定**: `@main` の部分は、使用したいブランチ名またはタグに変更してください。
3. **初回実行**: このリポジトリが GitHub にプッシュされている必要があります。
4. **プライベートリポジトリの場合**: 
   - このリポジトリがプライベートの場合、Settings > Actions > General でアクセス権限を設定する必要があります。
   - 詳細は [プライベートリポジトリ間で共有する](https://docs.github.com/ja/actions/how-tos/reuse-automations/share-across-private-repositories) を参照してください。

## トラブルシューティング

### エラー: `repository not found`

- このリポジトリが GitHub にプッシュされているか確認してください
- ブランチ名が正しいか確認してください（`@main` など）

### エラー: `GITHUB_TOKEN 環境変数が設定されていません`

- `env` セクションに `GITHUB_TOKEN` が設定されているか確認してください
- または `with` セクションで `github_token` を指定してください

