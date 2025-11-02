package repository

import (
	"fmt"
	"log"

	"github.com/google/go-github/v56/github"
)

// PaginationResult ページネーションの結果を表す構造体
type PaginationResult struct {
	HasNextPage bool
	NextPageNum int
}

// CheckPagination レスポンスから次のページがあるかを判定する
//
// Preconditions:
// - resp が GitHub API レスポンスであること
// - currentCount が現在のページで取得した件数であること
// - perPage が1ページあたりの最大件数であること
//
// Postconditions:
// - PaginationResult を返す（HasNextPage と NextPageNum を含む）
//
// Invariants:
// - レスポンスヘッダーから取得した情報を優先する
// - 取得件数からも判定を行う（フォールバック）
func CheckPagination(resp *github.Response, currentCount, perPage int) PaginationResult {
	// 1. resp.NextPage が 0 でない場合は、GitHub API のレスポンスヘッダーから取得した情報を使用
	if resp.NextPage != 0 {
		log.Printf("レスポンスヘッダーから次ページ (%d) を検出", resp.NextPage)
		return PaginationResult{
			HasNextPage: true,
			NextPageNum: resp.NextPage,
		}
	}

	// 2. NextPage が 0 でも、取得した件数が PerPage に達している場合は次のページを試みる
	if currentCount >= perPage {
		log.Printf("警告: レスポンスヘッダーから次ページ情報が取得できませんでしたが、取得件数 (%d) が PerPage (%d) に達しているため、次のページを試みます", currentCount, perPage)
		return PaginationResult{
			HasNextPage: true,
			NextPageNum: 0, // 手動でインクリメントする
		}
	}

	// 3. 取得した件数が30件（GitHub APIのデフォルト）の場合は、次のページがある可能性がある
	if currentCount == DefaultPageSize {
		log.Printf("警告: レスポンスヘッダーから次ページ情報が取得できませんでしたが、取得件数が%d件（GitHub APIのデフォルト）のため、次のページを試みます", DefaultPageSize)
		return PaginationResult{
			HasNextPage: true,
			NextPageNum: 0, // 手動でインクリメントする
		}
	}

	// 4. 取得した件数が 0 の場合は、次のページがないと判断
	if currentCount == 0 {
		log.Printf("0件取得したため、ページネーションを終了します")
		return PaginationResult{
			HasNextPage: false,
			NextPageNum: 0,
		}
	}

	// それ以外の場合は次のページなし
	return PaginationResult{
		HasNextPage: false,
		NextPageNum: 0,
	}
}

// ValidateOwnerAndRepo owner と repo が有効かどうかを検証する
//
// Preconditions:
// - owner と repo が文字列であること
//
// Postconditions:
// - どちらかが空の場合はエラーを返す
//
// Invariants:
// - 空文字列チェックのみを行う
func ValidateOwnerAndRepo(owner, repo string) error {
	if owner == "" || repo == "" {
		return fmt.Errorf("owner または repo が空です: owner=%s, repo=%s", owner, repo)
	}
	return nil
}

// ValidateUsername username が有効かどうかを検証する
//
// Preconditions:
// - username が文字列であること
//
// Postconditions:
// - 空の場合はエラーを返す
//
// Invariants:
// - 空文字列チェックのみを行う
func ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username が空です")
	}
	return nil
}
