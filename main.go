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
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
)

// Level represents the logging level
type Level int

const (
	DEBUG    Level = iota // Detailed information for debugging purposes
	INFO                  // General information about program execution
	WARN                  // Potentially harmful situations that might need attention
	ERROR                 // Error events that might still allow the program to continue
	CRITICAL              // Critical errors that trigger a panic
	FATAL                 // Severe errors that cause program termination
)

// Hook represents a logging hook function
type Hook func(level Level, msg string) error

// Logger represents the main logger struct
type Logger struct {
	level      Level
	output     io.Writer
	timeFormat string
	hooks      []Hook
}

// String returns the string representation of the log level
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
	case FATAL:
		return "FATAL"
	case CRITICAL:
		return "PANIC"
	default:
		return "UNKNOWN"
	}
}

// New creates a new logger instance
func New() *Logger {
	return &Logger{
		level:      INFO,
		output:     os.Stdout,
		timeFormat: "2006-01-02 15:04:05.999 MST",
	}
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level Level) {
	l.level = level
}

// SetOutput sets the output writer
func (l *Logger) SetOutput(output io.Writer) {
	l.output = output
}

// SetTimeFormat sets the time format for log messages
func (l *Logger) SetTimeFormat(format string) {
	l.timeFormat = format
}

// AddHook adds a new hook to the logger
func (l *Logger) AddHook(hook Hook) {
	l.hooks = append(l.hooks, hook)
}

// log is the internal logging function
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
		colorRed,    // CRITICAL
		colorPurple, // FATAL
	}[level]

	timestamp := time.Now().Format(l.timeFormat)
	message := fmt.Sprintf(msg, args...)

	fmt.Fprintf(l.output, "[%s] %s%s%s: %s\n",
		timestamp,
		levelColor,
		levelStr,
		colorReset,
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
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...any) {
	l.log(DEBUG, msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...any) {
	l.log(INFO, msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...any) {
	l.log(WARN, msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...any) {
	l.log(ERROR, msg, args...)
}

// Fatal logs a fatal error message and exits the program
func (l *Logger) Fatal(msg string, args ...any) {
	l.log(FATAL, msg, args...)
}

// Panic logs a panic message and triggers a panic
func (l *Logger) Critical(msg string, args ...any) {
	l.log(CRITICAL, msg, args...)
}

// global logger instance
var globalLogger = New()

// Global logging functions
func Debug(msg string, args ...any) {
	globalLogger.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	globalLogger.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	globalLogger.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	globalLogger.Error(msg, args...)
}

func Fatal(msg string, args ...any) {
	globalLogger.Fatal(msg, args...)
}

func Critical(msg string, args ...any) {
	globalLogger.Critical(msg, args...)
}

// Global configuration functions
func SetLevel(level Level) {
	globalLogger.SetLevel(level)
}

func SetOutput(output io.Writer) {
	globalLogger.SetOutput(output)
}

func SetTimeFormat(format string) {
	globalLogger.SetTimeFormat(format)
}

// Global hook function
func AddHook(hook Hook) {
	globalLogger.AddHook(hook)
}
