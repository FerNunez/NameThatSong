package logger

import (
	"context"
	"fmt"
	"os"
	"time"
)

// LogLevel represents the severity of the log message
type LogLevel string

const (
	LevelError LogLevel = "ERROR"
	LevelWarn  LogLevel = "WARN"
	LevelInfo  LogLevel = "INFO"
	LevelDebug LogLevel = "DEBUG"
)

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

// F is a helper function to create Fields more easily
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Logger interface defines the logging methods
type Logger interface {
	Error(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Debug(ctx context.Context, msg string, fields ...Field)
	SetLevel(level LogLevel)
}

// SimpleLogger implements the Logger interface with human-readable output
type SimpleLogger struct {
	level LogLevel
}

// Global logger instance
var defaultLogger Logger

// Init initializes the global logger
func Init(level LogLevel) {
	defaultLogger = &SimpleLogger{level: level}
}

// GetLogLevelFromEnv returns the log level from environment variable
func GetLogLevelFromEnv() LogLevel {
	switch os.Getenv("LOG_LEVEL") {
	case "ERROR":
		return LevelError
	case "WARN":
		return LevelWarn
	case "INFO":
		return LevelInfo
	case "DEBUG":
		return LevelDebug
	default:
		return LevelInfo // Default to INFO
	}
}

// shouldLog determines if a message should be logged based on the current level
func (s *SimpleLogger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LevelError: 1,
		LevelWarn:  2,
		LevelInfo:  3,
		LevelDebug: 4,
	}
	return levels[level] <= levels[s.level]
}

// log is the internal method that handles the actual logging
func (s *SimpleLogger) log(ctx context.Context, level LogLevel, msg string, fields []Field) {
	if !s.shouldLog(level) {
		return
	}

	timestamp := time.Now().Format("15:04:05")
	
	// Build the log message
	logMsg := fmt.Sprintf("[%s] %s: %s", timestamp, level, msg)
	
	// Add context information if available
	if ctx != nil {
		if userID := ctx.Value("user_id"); userID != nil {
			logMsg += fmt.Sprintf(" | user_id=%v", userID)
		}
		if requestID := ctx.Value("request_id"); requestID != nil {
			logMsg += fmt.Sprintf(" | request_id=%v", requestID)
		}
	}
	
	// Add fields
	if len(fields) > 0 {
		logMsg += " |"
		for _, field := range fields {
			logMsg += fmt.Sprintf(" %s=%v", field.Key, field.Value)
		}
	}
	
	fmt.Println(logMsg)
}

// SetLevel sets the minimum log level
func (s *SimpleLogger) SetLevel(level LogLevel) {
	s.level = level
}

// Error logs an error message
func (s *SimpleLogger) Error(ctx context.Context, msg string, fields ...Field) {
	s.log(ctx, LevelError, msg, fields)
}

// Warn logs a warning message
func (s *SimpleLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	s.log(ctx, LevelWarn, msg, fields)
}

// Info logs an info message
func (s *SimpleLogger) Info(ctx context.Context, msg string, fields ...Field) {
	s.log(ctx, LevelInfo, msg, fields)
}

// Debug logs a debug message
func (s *SimpleLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	s.log(ctx, LevelDebug, msg, fields)
}

// Package-level functions that use the global logger

// Error logs an error message using the global logger
func Error(ctx context.Context, msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Error(ctx, msg, fields...)
	}
}

// Warn logs a warning message using the global logger
func Warn(ctx context.Context, msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Warn(ctx, msg, fields...)
	}
}

// Info logs an info message using the global logger
func Info(ctx context.Context, msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Info(ctx, msg, fields...)
	}
}

// Debug logs a debug message using the global logger
func Debug(ctx context.Context, msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Debug(ctx, msg, fields...)
	}
}

// SetLevel sets the log level on the global logger
func SetLevel(level LogLevel) {
	if defaultLogger != nil {
		defaultLogger.SetLevel(level)
	}
}
