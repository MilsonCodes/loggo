// Package loggo provides a simple, efficient, and feature-rich logging library for Go applications.
// It supports multiple log levels, colored output, customizable time formats, and hooks for external integrations.
package loggo

import (
	"fmt"
	"io"
	"os"
	"slices"
	"time"
)

// Color codes for terminal output
const (
	colorReset  = "\033[0m"  // Reset color
	colorRed    = "\033[31m" // Red color
	colorGreen  = "\033[32m" // Green color
	colorYellow = "\033[33m" // Yellow color
	colorBlue   = "\033[34m" // Blue color
	colorPurple = "\033[35m" // Purple color
	colorCyan   = "\033[36m" // Cyan color
)

// Level represents the logging level.
// Higher levels indicate more severe conditions.
type Level int

const (
	DEBUG    Level = iota // Detailed information for debugging purposes
	INFO                  // General information about program execution
	WARN                  // Potentially harmful situations that might need attention
	ERROR                 // Error events that might still allow the program to continue
	CRITICAL              // Critical errors that do NOT trigger a panic, fatal severity without exit code 1
	FATAL                 // Severe errors that cause program termination
	PANIC                 // Critical errors that trigger a panic
)

// Hook represents a logging hook function that can be called for each log message.
// The function receives the log level and message, and returns an error if the hook fails.
type Hook func(level Level, msg string) error

// Logger represents the main logger struct that handles all logging operations.
type Logger struct {
	level      Level     // Current logging level
	output     io.Writer // Output destination for log messages
	timeFormat string    // Format string for timestamps
	hooks      []Hook    // List of registered hooks
}

// String returns the string representation of the log level.
// It returns "UNKNOWN" for undefined levels.
func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case CRITICAL:
		return "CRITICAL"
	case FATAL:
		return "FATAL"
	case PANIC:
		return "PANIC"
	default:
		return "UNKNOWN"
	}
}

// New creates and returns a new logger instance with default settings:
// - INFO level
// - stdout output
// - RFC3339 time format with millisecond precision
func New() *Logger {
	return &Logger{
		level:      INFO,
		output:     os.Stdout,
		timeFormat: "2006-01-02 15:04:05.999 MST",
	}
}

// SetLevel sets the minimum logging level for the logger.
// Messages with levels below this will be ignored.
func (l *Logger) SetLevel(level Level) {
	l.level = level
}

// SetOutput sets the output destination for log messages.
// It accepts any type that implements the io.Writer interface.
func (l *Logger) SetOutput(output io.Writer) {
	l.output = output
}

// SetTimeFormat sets the format string for timestamps in log messages.
// The format string should follow Go's time format layout.
func (l *Logger) SetTimeFormat(format string) {
	l.timeFormat = format
}

// AddHook adds a new hook function to the logger.
// Hooks are called for each log message and can be used for external integrations.
// If a hook returns an error, it will be logged and the hook will be removed.
func (l *Logger) AddHook(hook Hook) {
	l.hooks = append(l.hooks, hook)
}

// log is the internal logging function that handles all log messages.
// It formats the message, applies colors, and executes any registered hooks.
func (l *Logger) log(level Level, msg string, args ...any) {
	if level < l.level {
		return
	}

	levelStr := level.String()
	levelColor := []string{
		colorCyan,   // DEBUG
		colorGreen,  // INFO
		colorYellow, // WARN
		colorRed,    // ERROR
		colorPurple, // CRITICAL
		colorRed,    // FATAL
		colorPurple, // PANIC
	}[level]

	timestamp := time.Now().Format(l.timeFormat)
	message := fmt.Sprintf(msg, args...)

	fmt.Fprintf(l.output, "[%s%s%s] %s: %s\n",
		levelColor,
		levelStr,
		colorReset,
		timestamp,
		message,
	)

	// Execute hooks
	for i := 0; i < len(l.hooks); i++ {
		if err := l.hooks[i](level, message); err != nil {
			// Log the hook error and remove the failing hook
			fmt.Fprintf(l.output, "[%s] %sERROR%s: Hook error: %v\n",
				timestamp,
				colorRed,
				colorReset,
				err,
			)
			l.hooks = slices.Delete(l.hooks, i, i+1)
			i-- // Adjust index since we removed an element
		}
	}

	if level == FATAL {
		os.Exit(1)
	}
	if level == PANIC {
		panic(message)
	}
}

// Debug logs a debug message with the given format string and arguments.
// Debug messages are only logged if the current level is DEBUG or lower.
func (l *Logger) Debug(msg string, args ...any) {
	l.log(DEBUG, msg, args...)
}

// Info logs an info message with the given format string and arguments.
// Info messages are only logged if the current level is INFO or lower.
func (l *Logger) Info(msg string, args ...any) {
	l.log(INFO, msg, args...)
}

// Warn logs a warning message with the given format string and arguments.
// Warning messages are only logged if the current level is WARN or lower.
func (l *Logger) Warn(msg string, args ...any) {
	l.log(WARN, msg, args...)
}

// Error logs an error message with the given format string and arguments.
// Error messages are only logged if the current level is ERROR or lower.
func (l *Logger) Error(msg string, args ...any) {
	l.log(ERROR, msg, args...)
}

// Critical logs a critical message with the given format string and arguments.
// Critical messages are only logged if the current level is CRITICAL or lower.
// Unlike Fatal, this does not terminate the program.
func (l *Logger) Critical(msg string, args ...any) {
	l.log(CRITICAL, msg, args...)
}

// Fatal logs a fatal error message with the given format string and arguments,
// then exits the program with status code 1.
// Fatal messages are only logged if the current level is FATAL or lower.
func (l *Logger) Fatal(msg string, args ...any) {
	l.log(FATAL, msg, args...)
}

// Panic logs a panic message with the given format string and arguments,
// then triggers a panic with the formatted message.
// Panic messages are only logged if the current level is PANIC or lower.
func (l *Logger) Panic(msg string, args ...any) {
	l.log(PANIC, msg, args...)
}

// Global logging functions that use the default logger instance.

// globalLogger is the default logger instance used by the global logging functions.
var globalLogger = New()

// Debug logs a debug message using the global logger.
func Debug(msg string, args ...any) {
	globalLogger.Debug(msg, args...)
}

// Info logs an info message using the global logger.
func Info(msg string, args ...any) {
	globalLogger.Info(msg, args...)
}

// Warn logs a warning message using the global logger.
func Warn(msg string, args ...any) {
	globalLogger.Warn(msg, args...)
}

// Error logs an error message using the global logger.
func Error(msg string, args ...any) {
	globalLogger.Error(msg, args...)
}

// Critical logs a critical message using the global logger.
func Critical(msg string, args ...any) {
	globalLogger.Critical(msg, args...)
}

// Fatal logs a fatal error message using the global logger and exits the program.
func Fatal(msg string, args ...any) {
	globalLogger.Fatal(msg, args...)
}

// Panic logs a panic message using the global logger and triggers a panic.
func Panic(msg string, args ...any) {
	globalLogger.Panic(msg, args...)
}

// Global configuration functions that modify the default logger instance.

// SetLevel sets the logging level for the global logger.
func SetLevel(level Level) {
	globalLogger.SetLevel(level)
}

// SetOutput sets the output destination for the global logger.
func SetOutput(output io.Writer) {
	globalLogger.SetOutput(output)
}

// SetTimeFormat sets the time format for the global logger.
func SetTimeFormat(format string) {
	globalLogger.SetTimeFormat(format)
}

// AddHook adds a new hook to the global logger.
func AddHook(hook Hook) {
	globalLogger.AddHook(hook)
}
