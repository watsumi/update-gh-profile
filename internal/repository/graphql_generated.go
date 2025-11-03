package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	ghgraphql "github.com/watsumi/update-gh-profile/internal/graphql"
	"github.com/watsumi/update-gh-profile/internal/logger"
)

// FetchViewerGenerated fetches authenticated user information using generated types
func FetchViewerGenerated(ctx context.Context, token string) (string, string, error) {
	graphqlClient, err := newGraphQLClient(ctx, token)
	if err != nil {
		return "", "", fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	var query ghgraphql.ViewerQuery
	err = graphqlClient.Query(ctx, &query, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to execute GraphQL query: %w", err)
	}

	return query.Viewer.Login, query.Viewer.ID, nil
}

// FetchRepositoriesWithGraphQLGenerated fetches repository information in bulk using generated types
func FetchRepositoriesWithGraphQLGenerated(ctx context.Context, token string, username string, excludeForks bool) ([]*RepositoryGraphQLData, error) {
	graphqlClient, err := newGraphQLClient(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	var allRepos []*RepositoryGraphQLData
	var after *string

	// Use string-based query (to properly handle optional variables)
	// Unified with the same format as existing QueryReposPerLanguage
	queryStr := `query ReposQuery($login: String!, $isFork: Boolean, $first: Int, $after: String) {
  user(login: $login) {
    repositories(isFork: $isFork, first: $first, after: $after, ownerAffiliations: OWNER) {
      nodes {
        name
        owner {
          login
        }
        primaryLanguage {
          name
        }
        languages(first: 100) {
          edges {
            node {
              name
            }
            size
          }
          totalSize
        }
        stargazerCount
        defaultBranchRef {
          target {
            ... on Commit {
              history(first: 5) {
                totalCount
                nodes {
                  committedDate
                  author {
                    date
                  }
                }
              }
            }
          }
        }
      }
      pageInfo {
        endCursor
        hasNextPage
      }
    }
  }
}`

	for {
		// Convert variables to map
		variables := map[string]interface{}{
			"login":  username,
			"isFork": !excludeForks,
			"first":  30, // Reduce page size to 30 to prevent timeout
		}
		if after != nil {
			variables["after"] = *after
		}

		// Use Exec method to combine string query with generated types
		// Exec automatically unwraps data field, so pass type directly
		var query ghgraphql.ReposQuery

		// Retry logic: handle transient errors like 502 Bad Gateway
		const maxRetries = 5
		const baseRetryDelay = 5 * time.Second
		var lastErr error

		for attempt := 0; attempt < maxRetries; attempt++ {
			if attempt > 0 {
				// Wait before retry with exponential backoff
				// 1st: 5s, 2nd: 10s, 3rd: 20s, 4th: 40s
				retryDelay := baseRetryDelay * time.Duration(1<<uint(attempt-1))
				logger.Info("Retrying GraphQL query: %d/%d (waiting %v)", attempt+1, maxRetries, retryDelay)
				select {
				case <-ctx.Done():
					return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
				case <-time.After(retryDelay):
					// Wait completed
				}
			}

			logger.Debug("Executing GraphQL query (attempt %d/%d)...", attempt+1, maxRetries)

			// Wait a bit before request to reduce load on GitHub API
			// Wait on first request and after pagination
			if attempt == 0 {
				var waitTime time.Duration
				if after == nil {
					// First page request
					waitTime = 1 * time.Second
				} else {
					// Request after pagination (longer wait)
					waitTime = 2 * time.Second
				}
				logger.Debug("Waiting %v before request...", waitTime)
				select {
				case <-ctx.Done():
					return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
				case <-time.After(waitTime):
					// Wait completed
				}
			}

			err = graphqlClient.Exec(ctx, queryStr, &query, variables)
			if err == nil {
				if attempt > 0 {
					logger.Info("GraphQL query succeeded (succeeded on attempt %d)", attempt+1)
				}
				break // Success
			}

			lastErr = err
			errStr := err.Error()
			logger.Warning("Failed to execute GraphQL query (attempt %d/%d): %v", attempt+1, maxRetries, err)

			// Retry on transient errors like 502 Bad Gateway or 503 Service Unavailable
			lowerErrStr := strings.ToLower(errStr)
			logger.Debug("Error string (lowercase): %s", lowerErrStr)
			isRetryableError := strings.Contains(lowerErrStr, "502") ||
				strings.Contains(lowerErrStr, "503") ||
				strings.Contains(lowerErrStr, "504") ||
				strings.Contains(lowerErrStr, "timeout") ||
				strings.Contains(lowerErrStr, "bad gateway") ||
				strings.Contains(lowerErrStr, "service unavailable") ||
				strings.Contains(lowerErrStr, "gateway timeout") ||
				strings.Contains(lowerErrStr, "request_error") ||
				strings.Contains(lowerErrStr, "network") ||
				strings.Contains(lowerErrStr, "connection") ||
				strings.Contains(lowerErrStr, "stream error") ||
				strings.Contains(lowerErrStr, "stream id") ||
				strings.Contains(lowerErrStr, "cancel") ||
				strings.Contains(lowerErrStr, "json_decode_error") ||
				strings.Contains(lowerErrStr, "json decode") ||
				strings.Contains(lowerErrStr, "decode error") ||
				strings.Contains(lowerErrStr, "unmarshal") ||
				strings.Contains(lowerErrStr, "parse error")

			if !isRetryableError {
				// Return immediately if not a transient error
				return nil, fmt.Errorf("failed to execute GraphQL query (non-retryable error): %w", err)
			}

			// Log error if remaining after last attempt
			if attempt == maxRetries-1 {
				logger.Error("Failed to execute GraphQL query (after %d retries): %v", maxRetries, lastErr)
				return nil, fmt.Errorf("failed to execute GraphQL query (after %d retries): %w", maxRetries, lastErr)
			}
		}

		// If lastErr is not nil (retry failed), error has already been returned
		// Check here is unnecessary (because we break when err == nil)

		// Convert from generated type to RepositoryGraphQLData
		for _, repo := range query.User.Repositories.Nodes {
			repoData := &RepositoryGraphQLData{
				Name:           repo.Name,
				StargazerCount: repo.StargazerCount,
			}

			// Owner
			repoData.Owner.Login = repo.Owner.Login

			// PrimaryLanguage
			if repo.PrimaryLanguage != nil {
				repoData.PrimaryLanguage.Name = repo.PrimaryLanguage.Name
			}

			// Languages
			if repo.Languages != nil {
				repoData.Languages.TotalSize = repo.Languages.TotalSize
				// LanguageConnection's nodes are Language type (no size field)
				// edges have size, so use edges or save only nodes
				// Here we use edges to also get size
				if len(repo.Languages.Edges) > 0 {
					for _, edge := range repo.Languages.Edges {
						repoData.Languages.Nodes = append(repoData.Languages.Nodes, struct {
							Name string `json:"name"`
							Size int    `json:"size"`
						}{
							Name: edge.Node.Name,
							Size: edge.Size,
						})
					}
				} else {
					// Use nodes if edges are empty
					for _, lang := range repo.Languages.Nodes {
						repoData.Languages.Nodes = append(repoData.Languages.Nodes, struct {
							Name string `json:"name"`
							Size int    `json:"size"`
						}{
							Name: lang.Name,
							Size: 0, // Use 0 if size information is not available
						})
					}
				}
			}

			// DefaultBranchRef
			if repo.DefaultBranchRef != nil {
				history := repo.DefaultBranchRef.Target.Commit.History
				repoData.DefaultBranchRef.Target.History.TotalCount = history.TotalCount
				for _, commit := range history.Nodes {
					if commit != nil {
						repoData.DefaultBranchRef.Target.History.Nodes = append(repoData.DefaultBranchRef.Target.History.Nodes, struct {
							CommittedDate string `json:"committedDate"`
							Author        struct {
								Date string `json:"date"`
							} `json:"author"`
						}{
							CommittedDate: commit.CommittedDate.Format(time.RFC3339),
							Author: struct {
								Date string `json:"date"`
							}{
								Date: commit.Author.Date.Format(time.RFC3339),
							},
						})
					}
				}
			}

			allRepos = append(allRepos, repoData)
		}

		if !query.User.Repositories.PageInfo.HasNextPage {
			break
		}
		after = stringPtr(query.User.Repositories.PageInfo.EndCursor)

		// Wait before fetching next page to reduce load on GitHub API
		// Set longer wait time for pagination
		waitTime := 2 * time.Second
		logger.Info("Waiting %v before fetching next page... (fetched: %d repositories)", waitTime, len(allRepos))
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		case <-time.After(waitTime):
			// Wait completed
		}
	}

	return allRepos, nil
}

// FetchUserDetailsWithGraphQLGenerated fetches user details using generated types
func FetchUserDetailsWithGraphQLGenerated(ctx context.Context, token string, username string) (*UserDetailsGraphQLData, error) {
	graphqlClient, err := newGraphQLClient(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	// Convert variables to map
	variables := map[string]interface{}{
		"login": username,
	}

	var query ghgraphql.UserDetailsQuery
	err = graphqlClient.Query(ctx, &query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to execute GraphQL query: %w", err)
	}

	// Convert from generated type to UserDetailsGraphQLData
	userDetails := &UserDetailsGraphQLData{
		ID:        query.User.ID,
		Name:      query.User.Name,
		Email:     query.User.Email,
		CreatedAt: query.User.CreatedAt,
	}

	if query.User.Repositories != nil {
		userDetails.Repositories.TotalCount = query.User.Repositories.TotalCount
		for _, repo := range query.User.Repositories.Nodes {
			node := struct {
				StargazerCount   int `json:"stargazerCount"`
				DefaultBranchRef struct {
					Target struct {
						History struct {
							TotalCount int `json:"totalCount"`
						} `json:"history"`
					} `json:"target"`
				} `json:"defaultBranchRef"`
			}{
				StargazerCount: repo.StargazerCount,
			}

			if repo.DefaultBranchRef != nil {
				node.DefaultBranchRef.Target.History.TotalCount = repo.DefaultBranchRef.Target.Commit.History.TotalCount
			}

			userDetails.Repositories.Nodes = append(userDetails.Repositories.Nodes, node)
		}
	}

	if query.User.ContributionsCollection != nil && query.User.ContributionsCollection.ContributionCalendar != nil {
		for _, week := range query.User.ContributionsCollection.ContributionCalendar.Weeks {
			var days []struct {
				ContributionCount int    `json:"contributionCount"`
				Date              string `json:"date"`
			}
			for _, day := range week.ContributionDays {
				days = append(days, struct {
					ContributionCount int    `json:"contributionCount"`
					Date              string `json:"date"`
				}{
					ContributionCount: day.ContributionCount,
					Date:              day.Date,
				})
			}
			userDetails.ContributionsCollection.ContributionCalendar.Weeks = append(
				userDetails.ContributionsCollection.ContributionCalendar.Weeks,
				struct {
					ContributionDays []struct {
						ContributionCount int    `json:"contributionCount"`
						Date              string `json:"date"`
					} `json:"contributionDays"`
				}{
					ContributionDays: days,
				},
			)
		}
	}

	userDetails.PullRequests.TotalCount = query.User.PullRequests.TotalCount
	userDetails.Issues.TotalCount = query.User.Issues.TotalCount

	return userDetails, nil
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
