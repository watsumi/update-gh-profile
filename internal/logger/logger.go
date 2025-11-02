package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

// LogLevel ログレベル
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
)

// Logger ロガー構造体
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

// NewLogger 新しいロガーを作成する
func NewLogger(level LogLevel, output io.Writer) *Logger {
	if output == nil {
		output = os.Stderr
	}

	return &Logger{
		level:  level,
		logger: log.New(output, "", log.LstdFlags),
	}
}

// SetLevel ログレベルを設定する
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// Debug デバッグレベルのログを出力する
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= LogLevelDebug {
		l.logger.Printf("[DEBUG] "+format, args...)
	}
}

// Info 情報レベルのログを出力する
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= LogLevelInfo {
		l.logger.Printf("[INFO] "+format, args...)
	}
}

// Warning 警告レベルのログを出力する
func (l *Logger) Warning(format string, args ...interface{}) {
	if l.level <= LogLevelWarning {
		l.logger.Printf("[WARNING] "+format, args...)
	}
}

// Error エラーレベルのログを出力する
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= LogLevelError {
		l.logger.Printf("[ERROR] "+format, args...)
	}
}

// Fatal 致命的エラーレベルのログを出力して終了する
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.logger.Printf("[FATAL] "+format, args...)
	os.Exit(1)
}

// ParseLogLevel 文字列から LogLevel を解析する
func ParseLogLevel(levelStr string) LogLevel {
	switch levelStr {
	case "debug", "DEBUG":
		return LogLevelDebug
	case "info", "INFO":
		return LogLevelInfo
	case "warning", "WARNING", "warn", "WARN":
		return LogLevelWarning
	case "error", "ERROR":
		return LogLevelError
	default:
		return LogLevelInfo // デフォルトは Info
	}
}

// String LogLevel を文字列に変換する
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarning:
		return "WARNING"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// DefaultLogger デフォルトのロガーインスタンス
var DefaultLogger = NewLogger(LogLevelInfo, os.Stderr)

// 便利な関数（DefaultLogger を使用）
func Debug(format string, args ...interface{}) {
	DefaultLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	DefaultLogger.Info(format, args...)
}

func Warning(format string, args ...interface{}) {
	DefaultLogger.Warning(format, args...)
}

func Error(format string, args ...interface{}) {
	DefaultLogger.Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	DefaultLogger.Fatal(format, args...)
}

// LogError エラーをログに記録する（エラーが nil の場合は何もしない）
func LogError(err error, format string, args ...interface{}) error {
	if err != nil {
		msg := fmt.Sprintf(format, args...)
		DefaultLogger.Error("%s: %v", msg, err)
	}
	return err
}

// LogErrorWithContext エラーをコンテキスト情報とともにログに記録する
func LogErrorWithContext(err error, context string, format string, args ...interface{}) error {
	if err != nil {
		msg := fmt.Sprintf(format, args...)
		DefaultLogger.Error("[%s] %s: %v", context, msg, err)
	}
	return err
}
