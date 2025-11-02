package readme

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFormatTimestamp(t *testing.T) {
	timestamp := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name   string
		t      time.Time
		format string
		want   string
	}{
		{
			name:   "デフォルトフォーマット（RFC3339）",
			t:      timestamp,
			format: "",
			want:   "2024-01-15T10:30:00Z",
		},
		{
			name:   "カスタムフォーマット",
			t:      timestamp,
			format: "2006-01-02 15:04:05",
			want:   "2024-01-15 10:30:00",
		},
		{
			name:   "日付のみフォーマット",
			t:      timestamp,
			format: "2006-01-02",
			want:   "2024-01-15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTimestamp(tt.t, tt.format)
			if result != tt.want {
				t.Errorf("FormatTimestamp() = %q, want %q", result, tt.want)
			}
		})
	}
}

func TestFormatTimestampWithTimezone(t *testing.T) {
	timestamp := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		t        time.Time
		timezone string
		wantErr  bool
		contains string // 期待される文字列が含まれるか
	}{
		{
			name:     "UTC タイムゾーン",
			t:        timestamp,
			timezone: "UTC",
			wantErr:  false,
			contains: "2024-01-15T10:30:00Z",
		},
		{
			name:     "Asia/Tokyo タイムゾーン",
			t:        timestamp,
			timezone: "Asia/Tokyo",
			wantErr:  false,
			contains: "2024-01-15T19:30:00+09:00", // UTC+9時間
		},
		{
			name:     "無効なタイムゾーン",
			t:        timestamp,
			timezone: "Invalid/Timezone",
			wantErr:  true,
			contains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatTimestampWithTimezone(tt.t, tt.timezone)

			if tt.wantErr {
				if err == nil {
					t.Errorf("FormatTimestampWithTimezone() はエラーを返すべきでしたが、nil を返しました")
				}
			} else {
				if err != nil {
					t.Errorf("FormatTimestampWithTimezone() エラー = %v, エラーを期待していませんでした", err)
					return
				}

				if tt.contains != "" && !strings.Contains(result, tt.contains) {
					t.Errorf("FormatTimestampWithTimezone() = %q, want contains %q", result, tt.contains)
				}
			}
		})
	}
}

func TestGenerateTimestampMarkdown(t *testing.T) {
	timestamp := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		t        time.Time
		timezone string
		wantErr  bool
		want     string
	}{
		{
			name:     "UTC タイムゾーン",
			t:        timestamp,
			timezone: "UTC",
			wantErr:  false,
			want:     "*最終更新: 2024-01-15T10:30:00Z*",
		},
		{
			name:     "タイムゾーンなし",
			t:        timestamp,
			timezone: "",
			wantErr:  false,
			want:     "*最終更新: 2024-01-15T10:30:00Z*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateTimestampMarkdown(tt.t, tt.timezone)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GenerateTimestampMarkdown() はエラーを返すべきでしたが、nil を返しました")
				}
			} else {
				if err != nil {
					t.Errorf("GenerateTimestampMarkdown() エラー = %v, エラーを期待していませんでした", err)
					return
				}

				if !strings.Contains(result, tt.want) {
					t.Errorf("GenerateTimestampMarkdown() = %q, want contains %q", result, tt.want)
				}
			}
		})
	}
}

func TestAddUpdateTimestamp(t *testing.T) {
	testDir := "test_timestamp"
	defer func() {
		os.RemoveAll(testDir)
	}()

	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("テストディレクトリの作成に失敗しました: %v", err)
	}

	// テスト用のREADMEファイルを作成
	testReadme := filepath.Join(testDir, "README.md")
	initialContent := `# Test README

<!-- START_STATS -->
existing content
<!-- END_STATS -->
`

	err = os.WriteFile(testReadme, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("テストファイルの作成に失敗しました: %v", err)
	}

	// タイムスタンプを追加
	timestamp := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	err = AddUpdateTimestamp(testReadme, "STATS", timestamp, "UTC")
	if err != nil {
		t.Fatalf("AddUpdateTimestamp() エラー = %v", err)
	}

	// ファイルを読み込んで確認
	content, err := ReadFile(testReadme)
	if err != nil {
		t.Fatalf("ファイルの読み込みに失敗しました: %v", err)
	}

	// タイムスタンプが含まれているか確認
	if !strings.Contains(content, "*最終更新:") {
		t.Errorf("AddUpdateTimestamp() タイムスタンプが追加されていません")
	}

	if !strings.Contains(content, "2024-01-15T10:30:00Z") {
		t.Errorf("AddUpdateTimestamp() 期待されるタイムスタンプが含まれていません")
	}

	// 既存のコンテンツが保持されているか確認
	if !strings.Contains(content, "existing content") {
		t.Errorf("AddUpdateTimestamp() 既存のコンテンツが保持されていません")
	}
}

func TestGetCurrentTimestamp(t *testing.T) {
	result := GetCurrentTimestamp()

	// 現在時刻の近くであることを確認（5分以内）
	now := time.Now().UTC()
	diff := now.Sub(result)

	if diff < 0 {
		diff = -diff
	}

	if diff > 5*time.Minute {
		t.Errorf("GetCurrentTimestamp() 返された時刻が現在時刻から遠すぎます: %v", diff)
	}
}
