package logger

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LogLevelDebug, &buf)

	logger.Debug("test message %d", 123)

	output := buf.String()
	if !strings.Contains(output, "[DEBUG]") {
		t.Errorf("Debug() ログに [DEBUG] が含まれていません: %q", output)
	}
	if !strings.Contains(output, "test message 123") {
		t.Errorf("Debug() ログにメッセージが含まれていません: %q", output)
	}
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LogLevelInfo, &buf)

	logger.Info("test message %d", 456)

	output := buf.String()
	if !strings.Contains(output, "[INFO]") {
		t.Errorf("Info() ログに [INFO] が含まれていません: %q", output)
	}
}

func TestLogger_Warning(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LogLevelWarning, &buf)

	logger.Warning("test message")

	output := buf.String()
	if !strings.Contains(output, "[WARNING]") {
		t.Errorf("Warning() ログに [WARNING] が含まれていません: %q", output)
	}
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LogLevelError, &buf)

	logger.Error("test message")

	output := buf.String()
	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("Error() ログに [ERROR] が含まれていません: %q", output)
	}
}

func TestLogger_LevelFiltering(t *testing.T) {
	tests := []struct {
		name        string
		level       LogLevel
		wantDebug   bool
		wantInfo    bool
		wantWarning bool
		wantError   bool
	}{
		{
			name:        "Debug レベル",
			level:       LogLevelDebug,
			wantDebug:   true,
			wantInfo:    true,
			wantWarning: true,
			wantError:   true,
		},
		{
			name:        "Info レベル",
			level:       LogLevelInfo,
			wantDebug:   false,
			wantInfo:    true,
			wantWarning: true,
			wantError:   true,
		},
		{
			name:        "Warning レベル",
			level:       LogLevelWarning,
			wantDebug:   false,
			wantInfo:    false,
			wantWarning: true,
			wantError:   true,
		},
		{
			name:        "Error レベル",
			level:       LogLevelError,
			wantDebug:   false,
			wantInfo:    false,
			wantWarning: false,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(tt.level, &buf)

			logger.Debug("debug")
			logger.Info("info")
			logger.Warning("warning")
			logger.Error("error")

			output := buf.String()

			if tt.wantDebug && !strings.Contains(output, "[DEBUG]") {
				t.Errorf("Debug ログが出力されるべきでした")
			}
			if !tt.wantDebug && strings.Contains(output, "[DEBUG]") {
				t.Errorf("Debug ログが出力されるべきではありませんでした")
			}

			if tt.wantInfo && !strings.Contains(output, "[INFO]") {
				t.Errorf("Info ログが出力されるべきでした")
			}
			if !tt.wantInfo && strings.Contains(output, "[INFO]") {
				t.Errorf("Info ログが出力されるべきではありませんでした")
			}

			if tt.wantWarning && !strings.Contains(output, "[WARNING]") {
				t.Errorf("Warning ログが出力されるべきでした")
			}
			if !tt.wantWarning && strings.Contains(output, "[WARNING]") {
				t.Errorf("Warning ログが出力されるべきではありませんでした")
			}

			if tt.wantError && !strings.Contains(output, "[ERROR]") {
				t.Errorf("Error ログが出力されるべきでした")
			}
			if !tt.wantError && strings.Contains(output, "[ERROR]") {
				t.Errorf("Error ログが出力されるべきではありませんでした")
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", LogLevelDebug},
		{"DEBUG", LogLevelDebug},
		{"info", LogLevelInfo},
		{"INFO", LogLevelInfo},
		{"warning", LogLevelWarning},
		{"WARNING", LogLevelWarning},
		{"warn", LogLevelWarning},
		{"WARN", LogLevelWarning},
		{"error", LogLevelError},
		{"ERROR", LogLevelError},
		{"unknown", LogLevelInfo}, // デフォルト
		{"", LogLevelInfo},        // デフォルト
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("ParseLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LogLevelDebug, "DEBUG"},
		{LogLevelInfo, "INFO"},
		{LogLevelWarning, "WARNING"},
		{LogLevelError, "ERROR"},
		{LogLevel(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("LogLevel.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestLogError(t *testing.T) {
	var buf bytes.Buffer
	DefaultLogger = NewLogger(LogLevelError, &buf)
	defer func() {
		DefaultLogger = NewLogger(LogLevelInfo, nil)
	}()

	err := fmt.Errorf("test error")
	LogError(err, "処理に失敗しました")

	output := buf.String()
	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("LogError() ログに [ERROR] が含まれていません")
	}
	if !strings.Contains(output, "処理に失敗しました") {
		t.Errorf("LogError() ログにメッセージが含まれていません")
	}
	if !strings.Contains(output, "test error") {
		t.Errorf("LogError() ログにエラー内容が含まれていません")
	}
}

func TestLogError_Nil(t *testing.T) {
	var buf bytes.Buffer
	DefaultLogger = NewLogger(LogLevelError, &buf)
	defer func() {
		DefaultLogger = NewLogger(LogLevelInfo, nil)
	}()

	err := LogError(nil, "処理に失敗しました")
	if err != nil {
		t.Errorf("LogError() nil エラーの場合は nil を返すべきでした")
	}

	output := buf.String()
	if output != "" {
		t.Errorf("LogError() nil エラーの場合はログを出力すべきではありませんでした")
	}
}
