// Package loggo provides a high-performance logging library for Go applications.
// It features zero-allocation buffer pooling, efficient time format caching,
// and asynchronous hook execution. The library supports multiple log levels,
// colored output, and customizable time formats.
//
// Key Features:
// - Zero-allocation buffer pooling for log messages
// - Pre-calculated color codes and level strings
// - Efficient time format caching with automatic cleanup
// - Thread-safe operations with minimal locking
// - Asynchronous hook execution
// - Single write operation per log message
// - Memory-efficient string formatting
// - Internal chained API for zero-allocation logging
//
// Example Usage:
//
//	// Simple API (recommended)
//	logger := loggo.New()
//	defer logger.Close() // Ensure proper cleanup
//	logger.Info("Hello, %s!", "World")
//
//	// Using the global logger
//	loggo.Info("Hello, %s!", "World")
//
//	// Advanced usage with chained API (for performance-critical code)
//	logger.Info().Msgf("Hello, %s!", "World")
//
// Important Notes:
// - Always call Close() when you're done with a logger instance to clean up resources
// - The global logger is managed by the package and doesn't need to be closed
// - Hooks are executed asynchronously and may continue running after Close() is called
// - Panic and Fatal levels will still trigger their respective behaviors even after Close()
package loggo

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// Level represents the logging level.
// Higher levels indicate more severe conditions.
type Level int

// Predefined log levels in order of severity.
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
// Hooks are executed asynchronously to prevent blocking the main logging operation.
type Hook struct {
	fn       func(level Level, msg string) error
	priority int    // Higher priority hooks are executed first
	id       string // Unique identifier for the hook
}

// Logger represents the main logger struct that handles all logging operations.
// It includes various performance optimizations:
// - Thread-safe operations with minimal locking
// - Efficient buffer pooling
// - Time format caching
// - Asynchronous hook execution
type Logger struct {
	level             Level          // Current logging level
	output            *multiWriter   // Output destination(s) for log messages
	timeFormat        string         // Format string for timestamps
	hooks             []Hook         // List of registered hooks
	mu                sync.Mutex     // Mutex for thread-safe operations
	wg                sync.WaitGroup // WaitGroup for hook goroutines
	maxHooks          int            // Maximum number of hooks allowed
	bufSize           int            // Buffer size for log messages
	pool              sync.Pool      // Buffer pool for log messages
	timeCache         sync.Map       // Cache for formatted timestamps
	workerPool        *workerPool    // Worker pool for hook execution
	maxCacheSize      int            // Maximum size of time format cache
	cleanupInProgress bool
	lastCleanup       int64     // Last cleanup timestamp
	bufPool           sync.Pool // Additional pool for larger buffers
	timeKey           int64     // Current time key for caching
	timeValue         string    // Current time value
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
		return "CRIT"
	case FATAL:
		return "FATAL"
	case PANIC:
		return "PANIC"
	default:
		return "UNKNOWN"
	}
}

// PaddedString returns the pre-calculated padded string representation of the log level.
// If the level is unknown, it returns "[UNKNOWN]".
func (l Level) PaddedString() string {
	if padded, ok := paddedLevelStrings[l]; ok {
		return padded
	}
	return "[UNKNOWN]"
}

// New creates and returns a new logger instance with default settings.
// Performance Notes:
// - Initializes buffer pools with dynamic sizing
// - Uses sync.Map for efficient concurrent time format caching
// - Sets reasonable defaults for hooks and buffer size
func New() *Logger {
	l := &Logger{
		level:        INFO,
		output:       newMultiWriter(os.Stdout),
		timeFormat:   "2006-01-02 15:04:05.000 MST",
		maxHooks:     100,  // Reasonable limit for hooks
		bufSize:      1024, // Initial buffer size
		maxCacheSize: 1000, // Maximum number of cached time formats
	}

	// Initialize main buffer pool with dynamic sizing
	l.pool = sync.Pool{
		New: func() any {
			buf := make([]byte, 0, l.bufSize)
			return &buf
		},
	}

	// Initialize large buffer pool for bigger messages
	l.bufPool = sync.Pool{
		New: func() any {
			buf := make([]byte, 0, l.bufSize*4)
			return &buf
		},
	}

	// Initialize worker pool for hook execution
	l.workerPool = newWorkerPool(10) // 10 workers by default

	return l
}

// SetLevel sets the minimum logging level for the logger.
// Messages with levels below this will be ignored.
func (l *Logger) SetLevel(level Level) {
	l.level = level
}

// SetOutputs sets multiple output destinations for log messages.
// It accepts any number of writers that implement the io.Writer interface.
// All log messages will be written to all specified outputs.
func (l *Logger) SetOutputs(outputs ...io.Writer) {
	if len(outputs) == 0 {
		l.output = newMultiWriter(os.Stdout)
		return
	}
	l.output = newMultiWriter(outputs...)
}

// SetOutput sets a single output destination for log messages.
// It accepts any type that implements the io.Writer interface.
// This is a convenience method for when only one output is needed.
func (l *Logger) SetOutput(output io.Writer) {
	l.output = newMultiWriter(output)
}

// SetTimeFormat sets the format string for timestamps in log messages.
// The format string should follow Go's time format layout.
func (l *Logger) SetTimeFormat(format string) {
	l.timeFormat = format
}

// AddHook adds a new hook function to the logger.
// Hooks are called asynchronously for each log message and can be used for external integrations.
// If a hook returns an error, it will be logged and the hook will be removed.
// Note: Hook execution order is not guaranteed due to asynchronous execution.
// Returns an error if the maximum number of hooks is reached.
func (l *Logger) AddHook(hook func(level Level, msg string) error, priority int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.hooks) >= l.maxHooks {
		return fmt.Errorf("maximum number of hooks (%d) reached", l.maxHooks)
	}
	l.hooks = append(l.hooks, Hook{
		fn:       hook,
		priority: priority,
		id:       fmt.Sprintf("%p", hook), // Use function pointer as unique identifier
	})
	return nil
}

// Debug logs a debug message using the simple API.
// This is a convenience method that internally uses the chained API.
func (l *Logger) Debug(msg string, args ...any) {
	l.debugEvent().msgf(msg, args...)
}

// Info logs an info message using the simple API.
// This is a convenience method that internally uses the chained API.
func (l *Logger) Info(msg string, args ...any) {
	l.infoEvent().msgf(msg, args...)
}

// Warn logs a warning message using the simple API.
// This is a convenience method that internally uses the chained API.
func (l *Logger) Warn(msg string, args ...any) {
	l.warnEvent().msgf(msg, args...)
}

// Error logs an error message using the simple API.
// This is a convenience method that internally uses the chained API.
func (l *Logger) Error(msg string, args ...any) {
	l.errorEvent().msgf(msg, args...)
}

// Critical logs a critical message using the simple API.
// This is a convenience method that internally uses the chained API.
func (l *Logger) Critical(msg string, args ...any) {
	l.criticalEvent().msgf(msg, args...)
}

// Fatal logs a fatal error message using the simple API.
// This is a convenience method that internally uses the chained API.
func (l *Logger) Fatal(msg string, args ...any) {
	l.fatalEvent().msgf(msg, args...)
}

// Panic logs a panic message using the simple API.
// This is a convenience method that internally uses the chained API.
func (l *Logger) Panic(msg string, args ...any) {
	l.panicEvent().msgf(msg, args...)
}
