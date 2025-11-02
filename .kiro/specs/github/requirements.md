# Requirements Document

## Introduction

この機能は、GitHub ユーザーの全リポジトリからコードメトリクスを収集し、使用言語ランキング、コミット推移、コミット時間帯、サマリーカードなどの多様な SVG グラフを生成して README.md を自動更新する GitHub Actions ワークフローを提供します。これにより、ユーザーの技術スタック、開発活動の推移、作業時間帯などの包括的な情報が一目でわかるプロフィールを自動的に維持できます。

## Requirements

### Requirement 1: GitHub リポジトリメトリクスの取得

**Objective:** GitHub Actions ワークフローとして、対象ユーザーの全リポジトリからコードメトリクスを収集する機能を提供する

#### Acceptance Criteria

1. WHEN ワークフローがトリガーされる THEN システム SHALL GitHub API を使用して対象ユーザーの全リポジトリ一覧を取得する
2. WHEN リポジトリ一覧が取得される THEN システム SHALL 各リポジトリの言語情報を収集する
3. WHEN リポジトリ一覧が取得される THEN システム SHALL 各リポジトリのコード行数を収集する
4. WHEN リポジトリ一覧が取得される THEN システム SHALL 各リポジトリのコミット数を収集する
5. WHEN リポジトリ一覧が取得される THEN システム SHALL 各リポジトリのスター数を収集する
6. WHEN リポジトリ一覧が取得される THEN システム SHALL 各リポジトリのプルリクエスト数を収集する
7. WHEN リポジトリ一覧が取得される THEN システム SHALL 各リポジトリのコミット履歴データ（日付ごとのコミット数）を収集する
8. WHEN リポジトリ一覧が取得される THEN システム SHALL 各リポジトリのコミット時間帯データ（時間帯ごとのコミット数）を収集する
9. WHEN リポジトリ一覧が取得される THEN システム SHALL 各リポジトリのコミットごとの言語使用状況を収集する
10. IF リポジトリがフォークである THEN システム SHALL そのリポジトリを集計対象から除外する
11. WHEN API レート制限に達する THEN システム SHALL 適切な待機時間を設けて処理を継続する

### Requirement 2: メトリクスデータの集計と分析

**Objective:** 収集したメトリクスデータを集計し、ユーザーの技術スタック傾向を分析する

#### Acceptance Criteria

1. WHEN 言語データが収集される THEN システム SHALL 言語ごとの総コード行数を集計する
2. WHEN 言語データが収集される THEN システム SHALL 言語ごとの使用リポジトリ数を集計する
3. WHEN 集計が完了する THEN システム SHALL 使用量が多い順に言語をランキング化する
4. IF 言語の合計が閾値（例: 全コードの 1%未満）以下の場合 THEN システム SHALL その言語を集計結果から除外する
5. WHEN コミット履歴データが収集される THEN システム SHALL 日付ごとの合計コミット数を集計する
6. WHEN コミット時間帯データが収集される THEN システム SHALL 時間帯ごとの合計コミット数を集計する
7. WHEN コミットごとの言語使用状況が収集される THEN システム SHALL コミットごとに使用言語の Top5 を集計する
8. WHEN メトリクスデータが収集される THEN システム SHALL 全リポジトリの合計スター数、リポジトリ数、総コミット数、総プルリクエスト数を集計する
9. WHEN 集計が完了する THEN システム SHALL メトリクスデータを構造化された形式（JSON など）で保存する

### Requirement 3: SVG グラフの生成

**Objective:** 集計したメトリクスデータから視覚的にわかりやすい SVG グラフを生成する

#### Acceptance Criteria

1. WHEN 言語ランキングデータが準備される THEN システム SHALL 使用言語のランキングを表示する SVG を生成する
2. WHEN コミット推移データが準備される THEN システム SHALL コミット合計の推移を表示する SVG グラフを生成する
3. WHEN コミットごとの言語使用データが準備される THEN システム SHALL コミットごとの使用言語 Top5 を表示する SVG を生成する
4. WHEN コミット時間帯データが準備される THEN システム SHALL コミットが多い時間帯を表示する SVG グラフを生成する
5. WHEN サマリーデータが準備される THEN システム SHALL スター数、リポジトリ数、コミット数、プルリクエスト数を表示するサマリーカードの SVG を生成する
6. WHEN SVG が生成される THEN システム SHALL グラフに必要な情報（言語名、使用量、パーセンテージ、日付、時間帯など）を含める
7. WHEN SVG が生成される THEN システム SHALL グラフをモダンで見やすいデザインで作成する
8. WHEN SVG が生成される THEN システム SHALL 各 SVG グラフを GitHub Actions のワークスペースに保存する

### Requirement 4: README.md の自動更新

**Objective:** 生成した SVG グラフとメトリクス情報を README.md に自動的に反映する

#### Acceptance Criteria

1. WHEN グラフ生成が完了する THEN システム SHALL README.md の既存のメトリクスセクションを更新する
2. IF README.md にメトリクスセクションが存在しない場合 THEN システム SHALL 適切な位置にメトリクスセクションを追加する
3. WHEN README.md が更新される THEN システム SHALL 生成した SVG グラフを埋め込む
4. WHEN README.md が更新される THEN システム SHALL 最後に更新された日時を記録する
5. WHEN README.md が更新される THEN システム SHALL 変更内容をコミットする
6. WHEN 変更がコミットされる THEN システム SHALL 変更を main ブランチにプッシュする

### Requirement 5: GitHub Actions ワークフローの設定

**Objective:** 定期的または手動で実行可能な GitHub Actions ワークフローを提供する

#### Acceptance Criteria

1. WHEN ワークフローが設定される THEN システム SHALL スケジュール実行（例: 毎日または毎週）をサポートする
2. WHEN ワークフローが設定される THEN システム SHALL 手動実行（workflow_dispatch）をサポートする
3. WHEN ワークフローが実行される THEN システム SHALL 必要な GitHub トークンの権限を要求する
4. IF 対象ユーザーが指定されていない場合 THEN システム SHALL リポジトリ所有者を対象ユーザーとして使用する
5. WHEN ワークフローが実行される THEN システム SHALL 実行ログとエラー情報を適切に記録する
6. WHEN エラーが発生する THEN システム SHALL ワークフロー実行を適切に失敗としてマークする
