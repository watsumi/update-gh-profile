package aggregator

import (
	"testing"

	"github.com/google/go-github/v56/github"
)

func TestAggregateSummaryStats(t *testing.T) {
	// テスト用のリポジトリデータを作成
	repos := []*github.Repository{
		{
			StargazersCount: github.Int(10),
			Fork:            github.Bool(false),
		},
		{
			StargazersCount: github.Int(20),
			Fork:            github.Bool(false),
		},
		{
			StargazersCount: github.Int(5),
			Fork:            github.Bool(false),
		},
	}

	tests := []struct {
		name         string
		repositories []*github.Repository
		totalCommits int
		totalPRs     int
		wantStars    int
		wantRepos    int
		wantCommits  int
		wantPRs      int
	}{
		{
			name:         "正常系: 複数リポジトリ",
			repositories: repos,
			totalCommits: 100,
			totalPRs:     50,
			wantStars:    35, // 10 + 20 + 5
			wantRepos:    3,
			wantCommits:  100,
			wantPRs:      50,
		},
		{
			name:         "空のリポジトリリスト",
			repositories: []*github.Repository{},
			totalCommits: 0,
			totalPRs:     0,
			wantStars:    0,
			wantRepos:    0,
			wantCommits:  0,
			wantPRs:      0,
		},
		{
			name: "単一リポジトリ",
			repositories: []*github.Repository{
				{
					StargazersCount: github.Int(100),
					Fork:            github.Bool(false),
				},
			},
			totalCommits: 500,
			totalPRs:     200,
			wantStars:    100,
			wantRepos:    1,
			wantCommits:  500,
			wantPRs:      200,
		},
		{
			name: "フォークリポジトリが含まれている場合（除外される）",
			repositories: []*github.Repository{
				{
					StargazersCount: github.Int(10),
					Fork:            github.Bool(false),
				},
				{
					StargazersCount: github.Int(20),
					Fork:            github.Bool(true), // フォーク
				},
				{
					StargazersCount: github.Int(5),
					Fork:            github.Bool(false),
				},
			},
			totalCommits: 50,
			totalPRs:     25,
			wantStars:    15, // 10 + 5 (フォークは除外)
			wantRepos:    2,  // フォークは除外
			wantCommits:  50,
			wantPRs:      25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := AggregateSummaryStats(tt.repositories, tt.totalCommits, tt.totalPRs)

			if stats.TotalStars != tt.wantStars {
				t.Errorf("AggregateSummaryStats() TotalStars = %d, want %d", stats.TotalStars, tt.wantStars)
			}

			if stats.RepositoryCount != tt.wantRepos {
				t.Errorf("AggregateSummaryStats() RepositoryCount = %d, want %d", stats.RepositoryCount, tt.wantRepos)
			}

			if stats.TotalCommits != tt.wantCommits {
				t.Errorf("AggregateSummaryStats() TotalCommits = %d, want %d", stats.TotalCommits, tt.wantCommits)
			}

			if stats.TotalPullRequests != tt.wantPRs {
				t.Errorf("AggregateSummaryStats() TotalPullRequests = %d, want %d", stats.TotalPullRequests, tt.wantPRs)
			}
		})
	}
}
