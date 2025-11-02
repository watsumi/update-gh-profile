# GitHub Action としての使用手順

このリポジトリを別のリポジトリの GitHub Actions ワークフローで使用する手順です。

## 前提条件

1. このリポジトリが GitHub にプッシュされていること
2. 使用するリポジトリで GitHub Actions が有効になっていること

## プライベートリポジトリの場合

このリポジトリが**プライベートリポジトリ**の場合、他のリポジトリから使用するには以下の設定が必要です：

### アクセス権限の設定

1. このリポジトリ（`update-gh-profile`）の GitHub ページに移動
2. **Settings** → **Actions** → **General** に移動
3. ページ下部の **Access** セクションで：
   - **Accessible from repositories owned by 'USERNAME' user** を選択
   - または、特定のリポジトリのみにアクセスを許可する場合は該当オプションを選択
4. **Save** をクリック

詳細は [GitHub 公式ドキュメント](https://docs.github.com/ja/actions/how-tos/reuse-automations/share-across-private-repositories) を参照してください。

> **注意**: プライベートリポジトリのアクションを使用する場合、外部コラボレーターはワークフローの実行ログを確認できます。

## 使用手順

### ステップ 1: このリポジトリを GitHub にプッシュ

```bash
git add .
git commit -m "Add GitHub Action support"
git push origin main
```

### ステップ 2: 別のリポジトリでワークフローを作成

別のリポジトリに `.github/workflows/test-action.yml` を作成：

```yaml
name: Test Repository Fetch Action

on:
  workflow_dispatch: # 手動実行
  # schedule:
  #   - cron: '0 0 * * *' # 毎日実行（オプション）

jobs:
  test:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      # この Action を使用
      - name: Fetch repositories
        id: fetch_repos
        uses: watsumi/update-gh-profile@main
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          username: ${{ github.repository_owner }}
          exclude_forks: "true"

      # 結果を表示
      - name: Display result
        run: |
          echo "✅ リポジトリ取得が完了しました"
          echo "取得数: ${{ steps.fetch_repos.outputs.repository_count }}"
```

### ステップ 3: ワークフローを実行

1. GitHub のリポジトリページで「Actions」タブを開く
2. 「Test Repository Fetch Action」を選択
3. 「Run workflow」ボタンをクリック
4. 実行結果を確認

## 動作確認のポイント

✅ **成功時のログ**:

- `✓ GITHUB_TOKEN が設定されています`
- `✓ 対象ユーザー: [username]`
- `✅ リポジトリ一覧の取得に成功しました: X 件`
- `取得したリポジトリ（最初の5件）:`

✅ **出力変数**:

- `repository_count`: 取得したリポジトリ数が設定されます

## トラブルシューティング

### エラー: `repository not found`

- このリポジトリ (`watsumi/update-gh-profile`) が GitHub にプッシュされているか確認
- ブランチ名が正しいか確認（`@main`）

### エラー: `GITHUB_TOKEN 環境変数が設定されていません`

- `env` セクションに `GITHUB_TOKEN` が設定されているか確認
- GitHub の Secrets が正しく設定されているか確認

### エラー: `go: module github.com/watsumi/update-gh-profile@latest found`

- このリポジトリが GitHub にプッシュされているか確認
- `@main` の部分を正しいブランチ名に変更してください

## 参考

- [GitHub Actions の公式ドキュメント](https://docs.github.com/ja/actions)
- [Composite Actions のドキュメント](https://docs.github.com/ja/actions/creating-actions/creating-a-composite-action)
