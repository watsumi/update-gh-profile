package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/v76/github"
)

// FetchUserRepositories fetches a list of repositories for the authenticated user using GitHub API
//
// Preconditions:
// - username is a non-empty string
// - client is a valid GitHub client
// - isAuthenticatedUser is true (only supports authenticated user's own repositories)
//
// Postconditions:
// - If forks are excluded, only Fork=false repositories are returned
// - Return value is a slice of repository structs
// - Fetches repositories owned by the authenticated user (including private ones)
//
// Invariants:
// - Waits and retries when API rate limit is reached
// - Only fetches repositories owned by the authenticated user
func FetchUserRepositories(ctx context.Context, client *github.Client, username string, excludeForks bool, isAuthenticatedUser bool) ([]*github.Repository, error) {
	if err := ValidateUsername(username); err != nil {
		return nil, err
	}

	if !isAuthenticatedUser {
		return nil, fmt.Errorf("this tool can only fetch repositories owned by the authenticated user")
	}

	log.Printf("Fetching repository list: authenticated user=%s, exclude forks=%v", username, excludeForks)

	// Options for pagination
	// Type: "all" allows fetching private repositories as well
	// Reference: https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#list-repositories-for-the-authenticated-user
	opt := &github.RepositoryListOptions{
		Type:      "owner", // Choose from all, owner, member (for authenticated user, "all" allows fetching private repos)
		Sort:      "updated",
		Direction: "desc",
		ListOptions: github.ListOptions{
			PerPage: DefaultPerPage,
			Page:    0, // First page
		},
	}

	var allRepos []*github.Repository

	// Pagination loop: repeat until all pages are fetched
	// Passing empty string for username fetches repositories owned by the authenticated user (including private)
	// Reference: https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#list-repositories-for-the-authenticated-user
	for pageNum := 1; pageNum <= MaxPages; pageNum++ {
		// Set page number (GitHub API starts from 1)
		if pageNum > 1 {
			opt.Page = pageNum
		}

		repos, resp, err := client.Repositories.List(ctx, "", opt)

		if err != nil {
			return nil, fmt.Errorf("failed to fetch repository list: %w", err)
		}

		// Check and handle rate limit
		if err := HandleRateLimit(ctx, resp); err != nil {
			return nil, fmt.Errorf("failed to handle rate limit: %w", err)
		}

		// Add fetched repositories
		allRepos = append(allRepos, repos...)

		// Debug: output pagination information to log
		log.Printf("Fetched repositories: %d (total: %d)", len(repos), len(allRepos))
		log.Printf("Pagination info: current page=%d (manual=%d), next page=%d, last page=%d, PerPage=%d",
			opt.Page, pageNum, resp.NextPage, resp.LastPage, opt.PerPage)

		// Check if there is a next page (using common function)
		paginationResult := CheckPagination(resp, len(repos), opt.PerPage)

		if !paginationResult.HasNextPage {
			log.Printf("No next page, ending pagination (fetched: %d, PerPage: %d)", len(repos), opt.PerPage)
			break
		}

		// Check max page count (before advancing to next page)
		if pageNum >= MaxPages {
			log.Printf("Warning: reached max page count (%d). Ending pagination (total: %d)", MaxPages, len(allRepos))
			break
		}

		// Determine next page number
		if paginationResult.NextPageNum != 0 {
			pageNum = paginationResult.NextPageNum - 1 // -1 because pageNum is incremented in the loop
		}
		log.Printf("Fetching next page (page number: %d / max: %d)...", pageNum+1, MaxPages)
	}

	log.Printf("Finished fetching all repositories: %d", len(allRepos))

	// Exclude fork repositories
	if excludeForks {
		var filteredRepos []*github.Repository
		for _, repo := range allRepos {
			// Check if fork using GetFork() method
			if !repo.GetFork() {
				filteredRepos = append(filteredRepos, repo)
			}
		}
		allRepos = filteredRepos
		log.Printf("Repositories after excluding forks: %d", len(allRepos))
	}

	return allRepos, nil
}

// FetchRepositoryLanguages fetches language statistics for the specified repository
//
// Preconditions:
// - owner and repo are valid repository identifiers
// - client is a valid GitHub client
//
// Postconditions:
// - Returned map is in the format map[string]int{language name: bytes}
// - Returns nil and error on error
//
// Invariants:
// - Returns appropriate error on API error
// - Waits and retries when rate limit is reached
func FetchRepositoryLanguages(ctx context.Context, client *github.Client, owner, repo string) (map[string]int, error) {
	if err := ValidateOwnerAndRepo(owner, repo); err != nil {
		return nil, err
	}

	// Call GitHub API /repos/{owner}/{repo}/languages endpoint
	// Reference: https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#list-repository-languages
	languages, resp, err := client.Repositories.ListLanguages(ctx, owner, repo)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch language information for repository %s/%s: %w", owner, repo, err)
	}

	// Check and handle rate limit
	if err := HandleRateLimit(ctx, resp); err != nil {
		return nil, fmt.Errorf("failed to handle rate limit: %w", err)
	}

	// languages is in the format map[string]int{language name: bytes}
	return languages, nil
}

// FetchCommits fetches commit history for the specified repository
//
// Preconditions:
// - owner and repo are valid repository identifiers
// - client is a valid GitHub client
//
// Postconditions:
// - Returned slice is a list of commit information
// - Fetches all commits using pagination (up to 100 pages)
//
// Invariants:
// - Waits and retries when API rate limit is reached
func FetchCommits(ctx context.Context, client *github.Client, owner, repo string) ([]*github.RepositoryCommit, error) {
	if err := ValidateOwnerAndRepo(owner, repo); err != nil {
		return nil, err
	}

	log.Printf("Fetching commit history for repository %s/%s...", owner, repo)

	// Options for pagination
	opt := &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			PerPage: DefaultPerPage,
			Page:    0, // First page
		},
	}

	var allCommits []*github.RepositoryCommit

	// Pagination loop: repeat until all pages are fetched
	// Reference: https://docs.github.com/en/rest/commits/commits?apiVersion=2022-11-28#list-commits
	for pageNum := 1; pageNum <= MaxPages; pageNum++ {
		// Set page number
		if pageNum > 1 {
			opt.Page = pageNum
		}

		commits, resp, err := client.Repositories.ListCommits(ctx, owner, repo, opt)

		if err != nil {
			return nil, fmt.Errorf("failed to fetch commit history for repository %s/%s: %w", owner, repo, err)
		}

		// Check and handle rate limit
		if err := HandleRateLimit(ctx, resp); err != nil {
			return nil, fmt.Errorf("failed to handle rate limit: %w", err)
		}

		// Add fetched commits
		allCommits = append(allCommits, commits...)

		log.Printf("Fetched commits: %d (total: %d)", len(commits), len(allCommits))

		// Check if there is a next page (using common function)
		paginationResult := CheckPagination(resp, len(commits), opt.PerPage)

		if !paginationResult.HasNextPage {
			log.Printf("No next page, ending pagination (fetched: %d)", len(commits))
			break
		}

		// Check max page count
		if pageNum >= MaxPages {
			log.Printf("Warning: reached max page count (%d). Ending pagination (total: %d)", MaxPages, len(allCommits))
			break
		}

		// Determine next page number
		if paginationResult.NextPageNum != 0 {
			pageNum = paginationResult.NextPageNum - 1 // -1 because pageNum is incremented in the loop
		}
	}

	log.Printf("Finished fetching commit history for repository %s/%s: %d", owner, repo, len(allCommits))
	return allCommits, nil
}

// FetchCommitHistory fetches commit count per date for the specified repository
//
// Preconditions:
// - owner and repo are valid repository identifiers
// - client is a valid GitHub client
//
// Postconditions:
// - Returned map is in the format map[string]int{date(YYYY-MM-DD): commit count}
//
// Invariants:
// - Dates are recorded in YYYY-MM-DD format
// - Dates are recorded in UTC
func FetchCommitHistory(ctx context.Context, client *github.Client, owner, repo string) (map[string]int, error) {
	commits, err := FetchCommits(ctx, client, owner, repo)
	if err != nil {
		return nil, err
	}

	// Aggregate commit count per date
	history := make(map[string]int)
	for _, commit := range commits {
		// Get commit timestamp (using committer's timestamp)
		if commit.Commit == nil || commit.Commit.Committer == nil || commit.Commit.Committer.Date == nil {
			continue
		}

		// Get date in UTC (YYYY-MM-DD format)
		date := commit.Commit.Committer.Date.Time.UTC()
		dateStr := date.Format("2006-01-02")
		history[dateStr]++
	}

	log.Printf("Finished aggregating commit history for repository %s/%s: %d days", owner, repo, len(history))
	return history, nil
}

// FetchCommitTimeDistribution fetches commit count per time slot for the specified repository
//
// Preconditions:
// - owner and repo are valid repository identifiers
// - client is a valid GitHub client
//
// Postconditions:
// - Returned map is in the format map[int]int{time slot(0-23): commit count}
// - Time slots are aggregated in UTC
//
// Invariants:
// - Time slots are recorded in the range 0-23
func FetchCommitTimeDistribution(ctx context.Context, client *github.Client, owner, repo string) (map[int]int, error) {
	commits, err := FetchCommits(ctx, client, owner, repo)
	if err != nil {
		return nil, err
	}

	// Aggregate commit count per time slot (0-23 hours)
	distribution := make(map[int]int)
	for _, commit := range commits {
		// Get commit timestamp (using committer's timestamp)
		if commit.Commit == nil || commit.Commit.Committer == nil || commit.Commit.Committer.Date == nil {
			continue
		}

		// Get time slot in UTC (0-23 hours)
		date := commit.Commit.Committer.Date.Time.UTC()
		hour := date.Hour()
		distribution[hour]++
	}

	log.Printf("Finished aggregating commit time distribution for repository %s/%s: %d time slots", owner, repo, len(distribution))
	return distribution, nil
}

// FetchCommitLanguages fetches languages used per commit for the specified repository
//
// Preconditions:
// - owner and repo are valid repository identifiers
// - client is a valid GitHub client
//
// Postconditions:
// - Returned map is in the format map[string]map[string]int{commit SHA: {language name: occurrence count}}
// - Extracts languages from changed files for each commit
//
// Invariants:
// - Extracts languages from changed files for each commit
// - Language names are case-sensitive (Go, Python, etc.)
func FetchCommitLanguages(ctx context.Context, client *github.Client, owner, repo string) (map[string]map[string]int, error) {
	if err := ValidateOwnerAndRepo(owner, repo); err != nil {
		return nil, err
	}

	log.Printf("Fetching language usage per commit for repository %s/%s...", owner, repo)

	// First, fetch commit list
	commits, err := FetchCommits(ctx, client, owner, repo)
	if err != nil {
		return nil, err
	}

	// Map to store language usage per commit
	// map[commit SHA]map[language name]occurrence count
	commitLanguages := make(map[string]map[string]int)

	// Fetch detailed information for each commit (including changed file information)
	maxCommits := MaxCommitsForLanguageDetection
	if len(commits) < maxCommits {
		maxCommits = len(commits)
	}

	log.Printf("Fetching language information per commit: processing %d commits", maxCommits)

	for i := 0; i < maxCommits; i++ {
		commit := commits[i]
		sha := commit.GetSHA()

		// Fetch detailed information for commit (including changed file information)
		commitDetail, resp, err := client.Repositories.GetCommit(ctx, owner, repo, sha, &github.ListOptions{})
		if err != nil {
			log.Printf("Warning: failed to fetch details for commit %s: %v", sha[:7], err)
			continue
		}

		// Check and handle rate limit
		if err := HandleRateLimit(ctx, resp); err != nil {
			return nil, fmt.Errorf("failed to handle rate limit: %w", err)
		}

		// Aggregate languages used in this commit
		langs := make(map[string]int)

		if commitDetail.Files != nil {
			for _, file := range commitDetail.Files {
				// Detect language from filename (using common function)
				lang := DetectLanguageFromFilename(file.GetFilename())
				if lang != "" {
					langs[lang]++
				}
			}
		}

		if len(langs) > 0 {
			commitLanguages[sha] = langs
		}

		// Output progress to log (every 10 commits)
		if (i+1)%10 == 0 {
			log.Printf("Progress: processed %d/%d commits", i+1, maxCommits)
		}
	}

	log.Printf("Finished fetching language usage per commit for repository %s/%s: %d commits", owner, repo, len(commitLanguages))
	return commitLanguages, nil
}

// FetchPullRequests fetches pull request count for the specified repository
//
// Preconditions:
// - owner and repo are valid repository identifiers
// - client is a valid GitHub client
//
// Postconditions:
// - Returned value is the total number of pull requests
// - Aggregates pull requests in all states (open, closed, all)
//
// Invariants:
// - Fetches all PRs using pagination (up to 100 pages)
// - Waits and retries when API rate limit is reached
func FetchPullRequests(ctx context.Context, client *github.Client, owner, repo string) (int, error) {
	if err := ValidateOwnerAndRepo(owner, repo); err != nil {
		return 0, err
	}

	log.Printf("Fetching pull request count for repository %s/%s...", owner, repo)

	// Options for pagination
	// State: "all" fetches PRs in all states (open, closed)
	// Reference: https://docs.github.com/en/rest/pulls/pulls?apiVersion=2022-11-28#list-pull-requests
	opt := &github.PullRequestListOptions{
		State: "all", // Choose from "open", "closed", "all"
		ListOptions: github.ListOptions{
			PerPage: DefaultPerPage,
			Page:    0, // First page
		},
	}

	var totalCount int

	// Pagination loop: repeat until all pages are fetched
	for pageNum := 1; pageNum <= MaxPages; pageNum++ {
		// Set page number
		if pageNum > 1 {
			opt.Page = pageNum
		}

		pullRequests, resp, err := client.PullRequests.List(ctx, owner, repo, opt)

		if err != nil {
			return 0, fmt.Errorf("failed to fetch pull requests for repository %s/%s: %w", owner, repo, err)
		}

		// Check and handle rate limit
		if err := HandleRateLimit(ctx, resp); err != nil {
			return 0, fmt.Errorf("failed to handle rate limit: %w", err)
		}

		// Add fetched PR count
		totalCount += len(pullRequests)

		log.Printf("Fetched pull requests: %d (total: %d)", len(pullRequests), totalCount)

		// Check if there is a next page (using common function)
		paginationResult := CheckPagination(resp, len(pullRequests), opt.PerPage)

		if !paginationResult.HasNextPage {
			log.Printf("No next page, ending pagination (fetched: %d)", len(pullRequests))
			break
		}

		// Check max page count
		if pageNum >= MaxPages {
			log.Printf("Warning: reached max page count (%d). Ending pagination (total: %d)", MaxPages, totalCount)
			break
		}

		// Determine next page number
		if paginationResult.NextPageNum != 0 {
			pageNum = paginationResult.NextPageNum - 1 // -1 because pageNum is incremented in the loop
		}
	}

	log.Printf("Finished fetching pull request count for repository %s/%s: %d", owner, repo, totalCount)
	return totalCount, nil
}
