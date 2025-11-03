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
			name:   "Default format (RFC3339)",
			t:      timestamp,
			format: "",
			want:   "2024-01-15T10:30:00Z",
		},
		{
			name:   "Custom format",
			t:      timestamp,
			format: "2006-01-02 15:04:05",
			want:   "2024-01-15 10:30:00",
		},
		{
			name:   "Date only format",
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
		contains string // Expected string to be contained
	}{
		{
			name:     "UTC timezone",
			t:        timestamp,
			timezone: "UTC",
			wantErr:  false,
			contains: "2024-01-15T10:30:00Z",
		},
		{
			name:     "Asia/Tokyo timezone",
			t:        timestamp,
			timezone: "Asia/Tokyo",
			wantErr:  false,
			contains: "2024-01-15T19:30:00+09:00", // UTC+9 hours
		},
		{
			name:     "Invalid timezone",
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
					t.Errorf("FormatTimestampWithTimezone() should have returned an error, but returned nil")
				}
			} else {
				if err != nil {
					t.Errorf("FormatTimestampWithTimezone() error = %v, did not expect error", err)
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
			name:     "UTC timezone",
			t:        timestamp,
			timezone: "UTC",
			wantErr:  false,
			want:     "*Last updated: 2024-01-15T10:30:00Z*",
		},
		{
			name:     "No timezone",
			t:        timestamp,
			timezone: "",
			wantErr:  false,
			want:     "*Last updated: 2024-01-15T10:30:00Z*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateTimestampMarkdown(tt.t, tt.timezone)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GenerateTimestampMarkdown() should have returned an error, but returned nil")
				}
			} else {
				if err != nil {
					t.Errorf("GenerateTimestampMarkdown() error = %v, did not expect error", err)
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
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create test README file
	testReadme := filepath.Join(testDir, "README.md")
	initialContent := `# Test README

<!-- START_STATS -->
existing content
<!-- END_STATS -->
`

	err = os.WriteFile(testReadme, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Add timestamp
	timestamp := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	err = AddUpdateTimestamp(testReadme, "STATS", timestamp, "UTC")
	if err != nil {
		t.Fatalf("AddUpdateTimestamp() error = %v", err)
	}

	// Read file and verify
	content, err := ReadFile(testReadme)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Verify timestamp is included
	if !strings.Contains(content, "*Last updated:") {
		t.Errorf("AddUpdateTimestamp() timestamp was not added")
	}

	if !strings.Contains(content, "2024-01-15T10:30:00Z") {
		t.Errorf("AddUpdateTimestamp() expected timestamp not included")
	}

	// Verify existing content is preserved
	if !strings.Contains(content, "existing content") {
		t.Errorf("AddUpdateTimestamp() existing content not preserved")
	}
}

func TestGetCurrentTimestamp(t *testing.T) {
	result := GetCurrentTimestamp()

	// Verify it's close to current time (within 5 minutes)
	now := time.Now().UTC()
	diff := now.Sub(result)

	if diff < 0 {
		diff = -diff
	}

	if diff > 5*time.Minute {
		t.Errorf("GetCurrentTimestamp() returned time is too far from current time: %v", diff)
	}
}
