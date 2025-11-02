package repository

import (
	"context"
	"sync"
	"time"

	"github.com/google/go-github/v76/github"
)

// RateLimiter GitHub API レート制限を管理する
type RateLimiter struct {
	mu              sync.Mutex
	client          *github.Client
	remaining       int
	limit           int
	resetTime       time.Time
	requestInterval time.Duration // リクエスト間隔（デフォルト: 100ms）
	lastRequest     time.Time
}

// NewRateLimiter 新しいレートリミッターを作成する
func NewRateLimiter(client *github.Client) *RateLimiter {
	return &RateLimiter{
		client:          client,
		requestInterval: 100 * time.Millisecond, // デフォルト: 100ms間隔
	}
}

// SetRequestInterval リクエスト間隔を設定する
func (r *RateLimiter) SetRequestInterval(interval time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requestInterval = interval
}

// WaitIfNeeded レート制限を考慮して必要に応じて待機する
func (r *RateLimiter) WaitIfNeeded(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// リクエスト間隔をチェック
	now := time.Now()
	elapsed := now.Sub(r.lastRequest)
	if elapsed < r.requestInterval {
		waitTime := r.requestInterval - elapsed
		r.mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// 待機完了
		}
		r.mu.Lock()
	}
	r.lastRequest = time.Now()

	// レート制限の状態を確認（残り数が少ない場合）
	if r.remaining > 0 && r.remaining < 100 {
		// 残り数が少ない場合は間隔を広げる
		waitTime := r.requestInterval * 2
		r.mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// 待機完了
		}
		return nil
	}

	// レート制限がリセットされるまでの時間を確認
	if !r.resetTime.IsZero() && r.remaining == 0 {
		waitTime := time.Until(r.resetTime)
		if waitTime > 0 {
			// リセット時刻まで待機（最大10秒まで）
			maxWait := 10 * time.Second
			if waitTime > maxWait {
				waitTime = maxWait
			}
			r.mu.Unlock()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitTime):
				// 待機完了
			}
			return nil
		}
	}

	return nil
}

// UpdateRateLimitInfo レート制限情報を更新する
func (r *RateLimiter) UpdateRateLimitInfo(remaining, limit int, resetTime time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.remaining = remaining
	r.limit = limit
	r.resetTime = resetTime
}

// GetRemaining 残りのリクエスト数を取得する
func (r *RateLimiter) GetRemaining() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.remaining
}
