package generator

import (
	"strings"
	"testing"
)

func TestGenerateCommitTimeChart(t *testing.T) {
	tests := []struct {
		name             string
		timeDistribution map[int]int
		wantContains     []string
		wantNotContains  []string
	}{
		{
			name: "正常系: 複数の時間帯データ",
			timeDistribution: map[int]int{
				9:  10,
				10: 15,
				14: 8,
				22: 5,
			},
			wantContains: []string{
				"Commit Time Distribution",
				"UTC",
				":00",
				"<svg",
				"<rect",
			},
			wantNotContains: []string{},
		},
		{
			name:             "空のデータ",
			timeDistribution: map[int]int{},
			wantContains: []string{
				"Commit Time Distribution",
				"No data available",
			},
			wantNotContains: []string{},
		},
		{
			name: "全時間帯にデータがある場合",
			timeDistribution: func() map[int]int {
				data := make(map[int]int)
				for i := 0; i < 24; i++ {
					data[i] = i * 2 // 時間帯ごとに異なる値を設定
				}
				return data
			}(),
			wantContains: []string{
				"Commit Time Distribution",
				":00",
				"High",
				"Low",
			},
			wantNotContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svg, err := GenerateCommitTimeChart(tt.timeDistribution)
			if err != nil {
				t.Errorf("GenerateCommitTimeChart() error = %v", err)
				return
			}

			// SVG形式の基本的な検証
			if !strings.HasPrefix(svg, "<?xml") {
				t.Errorf("GenerateCommitTimeChart() SVG should start with <?xml")
			}

			if !strings.Contains(svg, "<svg") {
				t.Errorf("GenerateCommitTimeChart() SVG should contain <svg> tag")
			}

			// 期待される文字列が含まれているか確認
			for _, want := range tt.wantContains {
				if !strings.Contains(svg, want) {
					t.Errorf("GenerateCommitTimeChart() should contain %q", want)
				}
			}

			// 含まれていないことを期待する文字列の確認
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(svg, notWant) {
					t.Errorf("GenerateCommitTimeChart() should not contain %q", notWant)
				}
			}
		})
	}
}
