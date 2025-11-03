package repository

import (
	"fmt"
	"log"

	"github.com/google/go-github/v76/github"
)

// PaginationResult represents pagination result
type PaginationResult struct {
	HasNextPage bool
	NextPageNum int
}

// CheckPagination determines if there is a next page from the response
//
// Preconditions:
// - resp is a GitHub API response
// - currentCount is the number of items retrieved in the current page
// - perPage is the maximum number of items per page
//
// Postconditions:
// - Returns PaginationResult (contains HasNextPage and NextPageNum)
//
// Invariants:
// - Prioritizes information from response headers
// - Also performs determination based on retrieved count (fallback)
func CheckPagination(resp *github.Response, currentCount, perPage int) PaginationResult {
	// 1. If resp.NextPage is not 0, use information from GitHub API response headers
	if resp.NextPage != 0 {
		log.Printf("Next page (%d) detected from response headers", resp.NextPage)
		return PaginationResult{
			HasNextPage: true,
			NextPageNum: resp.NextPage,
		}
	}

	// 2. Even if NextPage is 0, if retrieved count reaches PerPage, try next page
	if currentCount >= perPage {
		log.Printf("Warning: Could not get next page info from response headers, but retrieved count (%d) reached PerPage (%d), so trying next page", currentCount, perPage)
		return PaginationResult{
			HasNextPage: true,
			NextPageNum: 0, // Manually increment
		}
	}

	// 3. If retrieved count is 30 (GitHub API default), there might be a next page
	if currentCount == DefaultPageSize {
		log.Printf("Warning: Could not get next page info from response headers, but retrieved count is %d (GitHub API default), so trying next page", DefaultPageSize)
		return PaginationResult{
			HasNextPage: true,
			NextPageNum: 0, // Manually increment
		}
	}

	// 4. If retrieved count is 0, determine there is no next page
	if currentCount == 0 {
		log.Printf("Retrieved 0 items, ending pagination")
		return PaginationResult{
			HasNextPage: false,
			NextPageNum: 0,
		}
	}

	// Otherwise, no next page
	return PaginationResult{
		HasNextPage: false,
		NextPageNum: 0,
	}
}

// ValidateOwnerAndRepo validates if owner and repo are valid
//
// Preconditions:
// - owner and repo are strings
//
// Postconditions:
// - Returns error if either is empty
//
// Invariants:
// - Only performs empty string check
func ValidateOwnerAndRepo(owner, repo string) error {
	if owner == "" || repo == "" {
		return fmt.Errorf("owner or repo is empty: owner=%s, repo=%s", owner, repo)
	}
	return nil
}

// ValidateUsername validates if username is valid
//
// Preconditions:
// - username is a string
//
// Postconditions:
// - Returns error if empty
//
// Invariants:
// - Only performs empty string check
func ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username is empty")
	}
	return nil
}
