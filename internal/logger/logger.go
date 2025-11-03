package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

// LogLevel log level
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
)

// Logger logger struct
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

// NewLogger creates a new logger
func NewLogger(level LogLevel, output io.Writer) *Logger {
	if output == nil {
		output = os.Stderr
	}

	return &Logger{
		level:  level,
		logger: log.New(output, "", log.LstdFlags),
	}
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// Debug outputs debug level log
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= LogLevelDebug {
		l.logger.Printf("[DEBUG] "+format, args...)
	}
}

// Info outputs info level log
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= LogLevelInfo {
		l.logger.Printf("[INFO] "+format, args...)
	}
}

// Warning outputs warning level log
func (l *Logger) Warning(format string, args ...interface{}) {
	if l.level <= LogLevelWarning {
		l.logger.Printf("[WARNING] "+format, args...)
	}
}

// Error outputs error level log
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= LogLevelError {
		l.logger.Printf("[ERROR] "+format, args...)
	}
}

// Fatal outputs fatal error level log and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.logger.Printf("[FATAL] "+format, args...)
	os.Exit(1)
}

// ParseLogLevel parses LogLevel from string
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
		return LogLevelInfo // Default is Info
	}
}

// String converts LogLevel to string
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

// DefaultLogger default logger instance
var DefaultLogger = NewLogger(LogLevelInfo, os.Stderr)

// Convenience functions (using DefaultLogger)
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

// LogError logs an error (does nothing if error is nil)
func LogError(err error, format string, args ...interface{}) error {
	if err != nil {
		msg := fmt.Sprintf(format, args...)
		DefaultLogger.Error("%s: %v", msg, err)
	}
	return err
}

// LogErrorWithContext logs an error with context information
func LogErrorWithContext(err error, context string, format string, args ...interface{}) error {
	if err != nil {
		msg := fmt.Sprintf(format, args...)
		DefaultLogger.Error("[%s] %s: %v", context, msg, err)
	}
	return err
}
