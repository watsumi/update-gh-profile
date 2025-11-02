package repository

import (
	"context"
	"log"
	"time"

	"github.com/google/go-github/v56/github"
)

// HandleRateLimit API レート制限を検出し、適切に待機する
//
// Preconditions:
// - resp が GitHub API レスポンスであること
//
// Postconditions:
// - レート制限に達している場合は、制限解除まで待機する
//
// Invariants:
// - 待機時間はレスポンスヘッダーから計算される
func HandleRateLimit(ctx context.Context, resp *github.Response) error {
	// レート制限の状態を確認
	if resp.Rate.Remaining == 0 {
		// レート制限に達している場合
		// Reset はレート制限がリセットされる時刻（time.Time型）
		resetTime := resp.Rate.Reset.Time
		waitDuration := time.Until(resetTime)

		// 待機時間が負の場合は0にする（既にリセット済み）
		if waitDuration < 0 {
			waitDuration = 0
		}

		// 待機時間に少し余裕を追加（1秒）
		waitDuration += time.Second

		if waitDuration > 0 {
			log.Printf("レート制限に達しました。%v 待機します...", waitDuration)
			select {
			case <-ctx.Done():
				return ctx.Err() // コンテキストがキャンセルされた場合
			case <-time.After(waitDuration):
				// 待機完了
			}
		}
	} else {
		// レート制限に余裕がある場合は、残りリクエスト数をログに記録
		log.Printf("レート制限の残り: %d/%d (リセット時刻: %v)",
			resp.Rate.Remaining,
			resp.Rate.Limit,
			resp.Rate.Reset.Time.Format("2006-01-02 15:04:05"))
	}

	return nil
}
