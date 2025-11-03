package repository

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/hasura/go-graphql-client"
	"golang.org/x/oauth2"
)

// GraphQLQueries GraphQL query definitions
var (
	// QueryReposPerLanguage Query to fetch language information per repository
	QueryReposPerLanguage = `
query ReposPerLanguage($login: String!, $endCursor: String) {
  user(login: $login) {
    repositories(isFork: false, first: 100, after: $endCursor, ownerAffiliations: OWNER) {
      nodes {
        name
        owner {
          login
        }
        primaryLanguage {
          name
        }
        languages(first: 100) {
          nodes {
            name
            size
          }
          totalSize
        }
        stargazerCount
        defaultBranchRef {
          target {
            ... on Commit {
              history(first: 100) {
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

	// QueryUserDetails Query to fetch user details
	QueryUserDetails = `
query UserDetails($login: String!) {
  user(login: $login) {
    id
    name
    email
    createdAt
    repositories(first: 100, privacy: PUBLIC, isFork: false, ownerAffiliations: OWNER, orderBy: {direction: DESC, field: STARGAZERS}) {
      totalCount
      nodes {
        stargazerCount
        defaultBranchRef {
          target {
            ... on Commit {
              history(first: 1) {
                totalCount
              }
            }
          }
        }
      }
    }
    contributionsCollection {
      contributionCalendar {
        weeks {
          contributionDays {
            contributionCount
            date
          }
        }
      }
    }
    pullRequests(first: 1) {
      totalCount
    }
    issues(first: 1) {
      totalCount
    }
  }
}`

	// QueryCommitLanguages Query to fetch language usage per commit
	QueryCommitLanguages = `
query CommitLanguages($login: String!) {
  user(login: $login) {
    contributionsCollection {
      commitContributionsByRepository(maxRepositories: 100) {
        repository {
          name
          owner {
            login
          }
          primaryLanguage {
            name
          }
          languages(first: 20) {
            edges {
              node {
                name
              }
              size
            }
          }
          defaultBranchRef {
            target {
              ... on Commit {
                history(first: 50) {
                  edges {
                    node {
                      oid
                      message
                      committedDate
                      additions
                      deletions
                    }
                  }
                }
              }
            }
          }
        }
        contributions {
          totalCount
        }
      }
    }
  }
}`

	// QueryProductiveTime Query to fetch commit time distribution
	QueryProductiveTime = `
query ProductiveTime($login: String!, $userId: ID!, $since: GitTimestamp!, $until: GitTimestamp!) {
  user(login: $login) {
    contributionsCollection {
      commitContributionsByRepository(maxRepositories: 50) {
        repository {
          name
          owner {
            login
          }
          defaultBranchRef {
            target {
              ... on Commit {
                history(first: 50, since: $since, until: $until, author: {id: $userId}) {
                  edges {
                    node {
                      committedDate
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`

	// QueryViewer Query to fetch authenticated user information
	QueryViewer = `
query Viewer {
  viewer {
    login
    id
  }
}`
)

// newGraphQLClient creates a GraphQL client
func newGraphQLClient(ctx context.Context, token string) (*graphql.Client, error) {
	if token == "" {
		return nil, fmt.Errorf("authentication token is not set")
	}

	// Create HTTP client with OAuth2 token
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))

	// Set timeout for HTTP client
	// Read timeout: 60 seconds
	httpClient.Timeout = 60 * time.Second

	// Set connection timeout for Transport
	if oauth2Transport, ok := httpClient.Transport.(*oauth2.Transport); ok {
		if baseTransport, ok := oauth2Transport.Base.(*http.Transport); ok {
			if baseTransport.DialContext == nil {
				baseTransport.DialContext = (&net.Dialer{
					Timeout: 30 * time.Second,
				}).DialContext
			}
		}
	}

	// Create GraphQL client
	graphqlClient := graphql.NewClient("https://api.github.com/graphql", httpClient)

	return graphqlClient, nil
}

// FetchViewer fetches authenticated user information using GraphQL
func FetchViewer(ctx context.Context, token string) (string, string, error) {
	graphqlClient, err := newGraphQLClient(ctx, token)
	if err != nil {
		return "", "", fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	variables := map[string]interface{}{}

	var response struct {
		Viewer struct {
			Login string `json:"login"`
			ID    string `json:"id"`
		} `json:"viewer"`
	}

	err = graphqlClient.Exec(ctx, QueryViewer, &response, variables)
	if err != nil {
		return "", "", fmt.Errorf("failed to execute GraphQL query: %w", err)
	}

	return response.Viewer.Login, response.Viewer.ID, nil
}

// FetchRepositoriesWithGraphQL fetches repository information in bulk using GraphQL
func FetchRepositoriesWithGraphQL(ctx context.Context, token string, username string, excludeForks bool) ([]*RepositoryGraphQLData, error) {
	graphqlClient, err := newGraphQLClient(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	var allRepos []*RepositoryGraphQLData
	var endCursor *string

	for {
		// Prepare variables for GraphQL query
		variables := map[string]interface{}{
			"login": username,
		}
		if endCursor != nil {
			variables["endCursor"] = *endCursor
		}

		// Execute GraphQL query (using string query)
		var response struct {
			User struct {
				Repositories struct {
					Nodes    []*RepositoryGraphQLData `json:"nodes"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"repositories"`
			} `json:"user"`
		}

		// Execute query using Exec
		err := graphqlClient.Exec(ctx, QueryReposPerLanguage, &response, variables)
		if err != nil {
			return nil, fmt.Errorf("failed to execute GraphQL query: %w", err)
		}

		allRepos = append(allRepos, response.User.Repositories.Nodes...)

		if !response.User.Repositories.PageInfo.HasNextPage {
			break
		}
		endCursor = &response.User.Repositories.PageInfo.EndCursor
	}

	return allRepos, nil
}

// RepositoryGraphQLData Repository data fetched from GraphQL
type RepositoryGraphQLData struct {
	Name            string    `json:"name"`
	Owner           OwnerData `json:"owner"`
	PrimaryLanguage struct {
		Name string `json:"name"`
	} `json:"primaryLanguage"`
	Languages struct {
		Nodes []struct {
			Name string `json:"name"`
			Size int    `json:"size"`
		} `json:"nodes"`
		TotalSize int `json:"totalSize"`
	} `json:"languages"`
	StargazerCount   int `json:"stargazerCount"`
	DefaultBranchRef struct {
		Target struct {
			History struct {
				TotalCount int `json:"totalCount"`
				Nodes      []struct {
					CommittedDate string `json:"committedDate"`
					Author        struct {
						Date string `json:"date"`
					} `json:"author"`
				} `json:"nodes"`
			} `json:"history"`
		} `json:"target"`
	} `json:"defaultBranchRef"`
}

// OwnerData Owner information
type OwnerData struct {
	Login string `json:"login"`
}

// FetchUserDetailsWithGraphQL fetches user details using GraphQL
func FetchUserDetailsWithGraphQL(ctx context.Context, token string, username string) (*UserDetailsGraphQLData, error) {
	graphqlClient, err := newGraphQLClient(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	variables := map[string]interface{}{
		"login": username,
	}

	var response struct {
		User UserDetailsGraphQLData `json:"user"`
	}

	err = graphqlClient.Exec(ctx, QueryUserDetails, &response, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to execute GraphQL query: %w", err)
	}

	return &response.User, nil
}

// UserDetailsGraphQLData User details data fetched from GraphQL
type UserDetailsGraphQLData struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	CreatedAt    string `json:"createdAt"`
	Repositories struct {
		TotalCount int `json:"totalCount"`
		Nodes      []struct {
			StargazerCount   int `json:"stargazerCount"`
			DefaultBranchRef struct {
				Target struct {
					History struct {
						TotalCount int `json:"totalCount"`
					} `json:"history"`
				} `json:"target"`
			} `json:"defaultBranchRef"`
		} `json:"nodes"`
	} `json:"repositories"`
	ContributionsCollection struct {
		ContributionCalendar struct {
			Weeks []struct {
				ContributionDays []struct {
					ContributionCount int    `json:"contributionCount"`
					Date              string `json:"date"`
				} `json:"contributionDays"`
			} `json:"weeks"`
		} `json:"contributionCalendar"`
	} `json:"contributionsCollection"`
	PullRequests struct {
		TotalCount int `json:"totalCount"`
	} `json:"pullRequests"`
	Issues struct {
		TotalCount int `json:"totalCount"`
	} `json:"issues"`
}

// FetchCommitLanguagesWithGraphQL fetches language usage per commit using GraphQL
// Uses multiple language information per repository to fetch more languages
func FetchCommitLanguagesWithGraphQL(ctx context.Context, token string, username string) (map[string]map[string]int, error) {
	graphqlClient, err := newGraphQLClient(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	variables := map[string]interface{}{
		"login": username,
	}

	var response struct {
		User struct {
			ContributionsCollection struct {
				CommitContributionsByRepository []struct {
					Repository struct {
						Name            string    `json:"name"`
						Owner           OwnerData `json:"owner"`
						PrimaryLanguage struct {
							Name string `json:"name"`
						} `json:"primaryLanguage"`
						Languages struct {
							Edges []struct {
								Node struct {
									Name string `json:"name"`
								} `json:"node"`
								Size int `json:"size"`
							} `json:"edges"`
						} `json:"languages"`
						DefaultBranchRef struct {
							Target struct {
								History struct {
									Edges []struct {
										Node struct {
											Oid           string `json:"oid"`
											Message       string `json:"message"`
											CommittedDate string `json:"committedDate"`
											Additions     int    `json:"additions"`
											Deletions     int    `json:"deletions"`
										} `json:"node"`
									} `json:"edges"`
								} `json:"history"`
							} `json:"target"`
						} `json:"defaultBranchRef"`
					} `json:"repository"`
					Contributions struct {
						TotalCount int `json:"totalCount"`
					} `json:"contributions"`
				} `json:"commitContributionsByRepository"`
			} `json:"contributionsCollection"`
		} `json:"user"`
	}

	err = graphqlClient.Exec(ctx, QueryCommitLanguages, &response, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to execute GraphQL query: %w", err)
	}

	// Convert data (using multiple language information per repository)
	commitLanguages := make(map[string]map[string]int)

	for _, repoContrib := range response.User.ContributionsCollection.CommitContributionsByRepository {
		// Get list of languages used in the repository
		repoLanguages := make(map[string]int)

		// Get languages and sizes from languages edges (sorted by size descending)
		for _, langEdge := range repoContrib.Repository.Languages.Edges {
			if langEdge.Node.Name != "" {
				// Use size as weight (larger files are more important)
				repoLanguages[langEdge.Node.Name] = langEdge.Size
			}
		}

		// Also add primary language (if not already in languages)
		if repoContrib.Repository.PrimaryLanguage.Name != "" {
			if _, exists := repoLanguages[repoContrib.Repository.PrimaryLanguage.Name]; !exists {
				repoLanguages[repoContrib.Repository.PrimaryLanguage.Name] = 1
			}
		}

		// Apply repository languages to each commit
		for _, edge := range repoContrib.Repository.DefaultBranchRef.Target.History.Edges {
			commitKey := edge.Node.CommittedDate
			if commitKey == "" {
				continue
			}

			if commitLanguages[commitKey] == nil {
				commitLanguages[commitKey] = make(map[string]int)
			}

			// Add each repository language to commit (weighted by size)
			for lang, size := range repoLanguages {
				// Use square root of size for weighting (to mitigate large differences)
				weight := 1
				if size > 1000 {
					// Increase weight for large files (with upper limit)
					weight = 2
				}
				commitLanguages[commitKey][lang] += weight
			}
		}
	}

	return commitLanguages, nil
}

// FetchProductiveTimeWithGraphQL fetches commit time distribution using GraphQL
func FetchProductiveTimeWithGraphQL(ctx context.Context, token string, username, userID string, since, until time.Time) (map[int]int, error) {
	graphqlClient, err := newGraphQLClient(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	variables := map[string]interface{}{
		"login":  username,
		"userId": userID,
		"since":  since.Format(time.RFC3339),
		"until":  until.Format(time.RFC3339),
	}

	var response struct {
		User struct {
			ContributionsCollection struct {
				CommitContributionsByRepository []struct {
					Repository struct {
						Name             string    `json:"name"`
						Owner            OwnerData `json:"owner"`
						DefaultBranchRef struct {
							Target struct {
								History struct {
									Edges []struct {
										Node struct {
											CommittedDate string `json:"committedDate"`
										} `json:"node"`
									} `json:"edges"`
								} `json:"history"`
							} `json:"target"`
						} `json:"defaultBranchRef"`
					} `json:"repository"`
				} `json:"commitContributionsByRepository"`
			} `json:"contributionsCollection"`
		} `json:"user"`
	}

	err = graphqlClient.Exec(ctx, QueryProductiveTime, &response, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to execute GraphQL query: %w", err)
	}

	// Aggregate commit count by time slot
	timeDistribution := make(map[int]int)
	for _, repoContrib := range response.User.ContributionsCollection.CommitContributionsByRepository {
		for _, edge := range repoContrib.Repository.DefaultBranchRef.Target.History.Edges {
			committedDate, err := time.Parse(time.RFC3339, edge.Node.CommittedDate)
			if err != nil {
				continue
			}
			hour := committedDate.UTC().Hour()
			timeDistribution[hour]++
		}
	}

	return timeDistribution, nil
}
