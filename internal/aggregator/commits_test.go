package aggregator

import (
	"testing"
)

func TestAggregateCommitHistory(t *testing.T) {
	tests := []struct {
		name           string
		commitHistories map[string]map[string]int
		wantDates      int
		wantTotal      int
	}{
		{
			name: "正常系: 複数リポジトリ",
			commitHistories: map[string]map[string]int{
				"repo1": {
					"2024-01-01": 5,
					"2024-01-02": 3,
				},
				"repo2": {
					"2024-01-01": 2,
					"2024-01-03": 4,
				},
			},
			wantDates: 3, // 2024-01-01, 2024-01-02, 2024-01-03
			wantTotal: 14, // 5 + 3 + 2 + 4
		},
		{
			name: "空のmap",
			commitHistories: map[string]map[string]int{},
			wantDates: 0,
			wantTotal: 0,
		},
		{
			name: "単一リポジトリ",
			commitHistories: map[string]map[string]int{
				"repo1": {
					"2024-01-01": 10,
				},
			},
			wantDates: 1,
			wantTotal: 10,
		},
		{
			name: "同じ日付のコミットを複数リポジトリから集計",
			commitHistories: map[string]map[string]int{
				"repo1": {
					"2024-01-01": 5,
				},
				"repo2": {
					"2024-01-01": 3,
				},
				"repo3": {
					"2024-01-01": 2,
				},
			},
			wantDates: 1,
			wantTotal: 10, // 5 + 3 + 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aggregated := AggregateCommitHistory(tt.commitHistories)

			if len(aggregated) != tt.wantDates {
				t.Errorf("AggregateCommitHistory() dates = %d, want %d", len(aggregated), tt.wantDates)
			}

			total := 0
			for _, count := range aggregated {
				total += count
			}

			if total != tt.wantTotal {
				t.Errorf("AggregateCommitHistory() total = %d, want %d", total, tt.wantTotal)
			}
		})
	}
}

func TestSortCommitHistoryByDate(t *testing.T) {
	history := map[string]int{
		"2024-01-03": 5,
		"2024-01-01": 10,
		"2024-01-02": 3,
	}

	sorted := SortCommitHistoryByDate(history)

	if len(sorted) != 3 {
		t.Errorf("SortCommitHistoryByDate() length = %d, want 3", len(sorted))
	}

	// 日付順（昇順）でソートされていることを確認
	expectedDates := []string{"2024-01-01", "2024-01-02", "2024-01-03"}
	for i, pair := range sorted {
		if pair.Date != expectedDates[i] {
			t.Errorf("SortCommitHistoryByDate() [%d] = %s, want %s", i, pair.Date, expectedDates[i])
		}
	}

	// 値が保持されていることを確認
	if sorted[0].Count != 10 || sorted[1].Count != 3 || sorted[2].Count != 5 {
		t.Errorf("SortCommitHistoryByDate() values not preserved")
	}
}

func TestAggregateCommitTimeDistribution(t *testing.T) {
	tests := []struct {
		name             string
		timeDistributions map[string]map[int]int
		wantHours        int
		wantTotal        int
	}{
		{
			name: "正常系: 複数リポジトリ",
			timeDistributions: map[string]map[int]int{
				"repo1": {
					9:  5, // 9時
					10: 3, // 10時
				},
				"repo2": {
					9:  2, // 9時
					14: 4, // 14時
				},
			},
			wantHours: 3, // 9時, 10時, 14時
			wantTotal: 14, // 5 + 3 + 2 + 4
		},
		{
			name: "空のmap",
			timeDistributions: map[string]map[int]int{},
			wantHours: 0,
			wantTotal: 0,
		},
		{
			name: "同じ時間帯のコミットを複数リポジトリから集計",
			timeDistributions: map[string]map[int]int{
				"repo1": {
					9: 5,
				},
				"repo2": {
					9: 3,
				},
				"repo3": {
					9: 2,
				},
			},
			wantHours: 1,
			wantTotal: 10, // 5 + 3 + 2
		},
		{
			name: "範囲外の時間帯をスキップ",
			timeDistributions: map[string]map[int]int{
				"repo1": {
					9:  5,
					25: 10, // 範囲外（スキップされる）
					-1: 3,  // 範囲外（スキップされる）
				},
			},
			wantHours: 1, // 9時のみ
			wantTotal: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aggregated := AggregateCommitTimeDistribution(tt.timeDistributions)

			if len(aggregated) != tt.wantHours {
				t.Errorf("AggregateCommitTimeDistribution() hours = %d, want %d", len(aggregated), tt.wantHours)
			}

			total := 0
			for _, count := range aggregated {
				total += count
			}

			if total != tt.wantTotal {
				t.Errorf("AggregateCommitTimeDistribution() total = %d, want %d", total, tt.wantTotal)
			}

			// 時間帯が0-23の範囲内であることを確認
			for hour := range aggregated {
				if hour < 0 || hour > 23 {
					t.Errorf("AggregateCommitTimeDistribution() hour %d out of range [0-23]", hour)
				}
			}
		})
	}
}

func TestSortCommitTimeDistributionByHour(t *testing.T) {
	distribution := map[int]int{
		14: 5,
		9:  10,
		22: 3,
	}

	sorted := SortCommitTimeDistributionByHour(distribution)

	if len(sorted) != 3 {
		t.Errorf("SortCommitTimeDistributionByHour() length = %d, want 3", len(sorted))
	}

	// 時間帯順（昇順）でソートされていることを確認
	expectedHours := []int{9, 14, 22}
	for i, pair := range sorted {
		if pair.Hour != expectedHours[i] {
			t.Errorf("SortCommitTimeDistributionByHour() [%d] = %d, want %d", i, pair.Hour, expectedHours[i])
		}
	}

	// 値が保持されていることを確認
	if sorted[0].Count != 10 || sorted[1].Count != 5 || sorted[2].Count != 3 {
		t.Errorf("SortCommitTimeDistributionByHour() values not preserved")
	}
}
