package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/watsumi/update-gh-profile/internal/logger"
	"github.com/watsumi/update-gh-profile/internal/repository"

	"github.com/google/go-github/v76/github"
)

// RepoData data retrieved from repository
type RepoData struct {
	Owner            string
	RepoName         string
	Languages        map[string]int
	CommitHistory    map[string]int
	TimeDistribution map[int]int
	CommitCount      int
	CommitLanguages  map[string]map[string]int
	PRCount          int
	Error            error
}

// ProcessRepository processes repository and retrieves data
func ProcessRepository(ctx context.Context, client *github.Client, rateLimiter *repository.RateLimiter, repo *github.Repository) *RepoData {
	owner := repo.GetOwner().GetLogin()
	repoName := repo.GetName()

	data := &RepoData{
		Owner:            owner,
		RepoName:         repoName,
		Languages:        make(map[string]int),
		CommitHistory:    make(map[string]int),
		TimeDistribution: make(map[int]int),
		CommitLanguages:  make(map[string]map[string]int),
	}

	// Check rate limit and wait
	if err := rateLimiter.WaitIfNeeded(ctx); err != nil {
		data.Error = fmt.Errorf("failed to wait for rate limit: %w", err)
		return data
	}

	// Fetch language data
	langs, err := repository.FetchRepositoryLanguages(ctx, client, owner, repoName)
	if err != nil {
		logger.LogErrorWithContext(err, fmt.Sprintf("%s/%s", owner, repoName), "failed to fetch language data")
		// Language data fetch failure is not fatal, so try to fetch other data
	} else {
		data.Languages = langs
		// Update rate limit info (update here if it can be obtained from response in FetchRepositoryLanguages)
	}

	// Fetch commit history
	if err := rateLimiter.WaitIfNeeded(ctx); err != nil {
		data.Error = fmt.Errorf("failed to wait for rate limit: %w", err)
		return data
	}

	commitHistory, err := repository.FetchCommitHistory(ctx, client, owner, repoName)
	if err != nil {
		logger.LogErrorWithContext(err, fmt.Sprintf("%s/%s", owner, repoName), "failed to fetch commit history")
	} else {
		data.CommitHistory = commitHistory
	}

	// Fetch commit time distribution
	if err := rateLimiter.WaitIfNeeded(ctx); err != nil {
		data.Error = fmt.Errorf("failed to wait for rate limit: %w", err)
		return data
	}

	timeDist, err := repository.FetchCommitTimeDistribution(ctx, client, owner, repoName)
	if err != nil {
		logger.LogErrorWithContext(err, fmt.Sprintf("%s/%s", owner, repoName), "failed to fetch commit time distribution")
	} else {
		data.TimeDistribution = timeDist
	}

	// Fetch commit count
	if err := rateLimiter.WaitIfNeeded(ctx); err != nil {
		data.Error = fmt.Errorf("failed to wait for rate limit: %w", err)
		return data
	}

	commits, err := repository.FetchCommits(ctx, client, owner, repoName)
	if err != nil {
		logger.LogErrorWithContext(err, fmt.Sprintf("%s/%s", owner, repoName), "failed to fetch commit data")
	} else {
		data.CommitCount = len(commits)
	}

	// Fetch languages per commit (skip if commit count is too large)
	if data.CommitCount > 0 && data.CommitCount <= 100 {
		if err := rateLimiter.WaitIfNeeded(ctx); err != nil {
			// Ignore error (optional data)
		} else {
			commitLangs, err := repository.FetchCommitLanguages(ctx, client, owner, repoName)
			if err != nil {
				logger.LogErrorWithContext(err, fmt.Sprintf("%s/%s", owner, repoName), "failed to fetch commit language data")
			} else if len(commitLangs) > 0 {
				data.CommitLanguages = commitLangs
			}
		}
	}

	// Fetch pull request count
	if err := rateLimiter.WaitIfNeeded(ctx); err != nil {
		data.Error = fmt.Errorf("failed to wait for rate limit: %w", err)
		return data
	}

	prCount, err := repository.FetchPullRequests(ctx, client, owner, repoName)
	if err != nil {
		logger.LogErrorWithContext(err, fmt.Sprintf("%s/%s", owner, repoName), "failed to fetch pull request data")
	} else {
		data.PRCount = prCount
	}

	return data
}

// ProcessRepositoriesInParallel processes repositories in parallel
func ProcessRepositoriesInParallel(ctx context.Context, client *github.Client, repos []*github.Repository, maxConcurrency int) ([]*RepoData, error) {
	if maxConcurrency <= 0 {
		maxConcurrency = 5 // Default: 5 parallel processes
	}

	rateLimiter := repository.NewRateLimiter(client)
	rateLimiter.SetRequestInterval(150 * time.Millisecond) // Request every 150ms

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrency) // Limit concurrency with semaphore
	results := make([]*RepoData, len(repos))
	var mu sync.Mutex

	// Get rate limit info from first repository
	if len(repos) > 0 {
		// Fetch authenticated user info and update rate limit info
		// (Actual API calls are made in ProcessRepository)
	}

	logger.Info("Processing repositories in parallel: total=%d, max concurrency=%d", len(repos), maxConcurrency)

	for i, repo := range repos {
		wg.Add(1)
		go func(idx int, r *github.Repository) {
			defer wg.Done()

			// Limit concurrency with semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Process repository
			data := ProcessRepository(ctx, client, rateLimiter, r)

			// Save result
			mu.Lock()
			results[idx] = data
			mu.Unlock()

			if data.Error != nil {
				logger.Warning("[%d/%d] Error occurred while processing %s/%s: %v",
					idx+1, len(repos), data.Owner, data.RepoName, data.Error)
			} else {
				logger.Debug("[%d/%d] Completed processing %s/%s", idx+1, len(repos), data.Owner, data.RepoName)
			}
			fmt.Printf("  [%d/%d] Completed processing %s/%s\n", idx+1, len(repos), data.Owner, data.RepoName)
		}(i, repo)
	}

	wg.Wait()
	logger.Info("Completed processing all repositories")

	return results, nil
}
