package loggo

import (
	"io"
)

// Global logging functions that use the default logger instance.

// globalLogger is the default logger instance used by the global logging functions.
var globalLogger = New()

// Debug logs a debug message using the global logger.
func Debug(msg string) {
	globalLogger.Debug(msg)
}

// Debugf logs a formatted debug message using the global logger.
func Debugf(msg string, args ...any) {
	globalLogger.Debugf(msg, args...)
}

// Info logs an info message using the global logger.
func Info(msg string) {
	globalLogger.Info(msg)
}

// Infof logs a formatted info message using the global logger.
func Infof(msg string, args ...any) {
	globalLogger.Infof(msg, args...)
}

// Warn logs a warning message using the global logger.
func Warn(msg string) {
	globalLogger.Warn(msg)
}

// Warnf logs a formatted warning message using the global logger.
func Warnf(msg string, args ...any) {
	globalLogger.Warnf(msg, args...)
}

// Error logs an error message using the global logger.
func Error(msg string) {
	globalLogger.Error(msg)
}

// Errorf logs a formatted error message using the global logger.
func Errorf(msg string, args ...any) {
	globalLogger.Errorf(msg, args...)
}

// Critical logs a critical message using the global logger.
func Critical(msg string) {
	globalLogger.Critical(msg)
}

// Criticalf logs a formatted critical message using the global logger.
func Criticalf(msg string, args ...any) {
	globalLogger.Criticalf(msg, args...)
}

// Fatal logs a fatal error message using the global logger and exits the program.
func Fatal(msg string) {
	globalLogger.Fatal(msg)
}

// Fatalf logs a formatted fatal error message using the global logger and exits the program.
func Fatalf(msg string, args ...any) {
	globalLogger.Fatalf(msg, args...)
}

// Panic logs a panic message using the global logger and triggers a panic.
func Panic(msg string) {
	globalLogger.Panic(msg)
}

// Panicf logs a formatted panic message using the global logger and triggers a panic.
func Panicf(msg string, args ...any) {
	globalLogger.Panicf(msg, args...)
}

// Global configuration functions that modify the default logger instance.

// SetLevel sets the logging level for the global logger.
func SetLevel(level Level) {
	globalLogger.SetLevel(level)
}

// SetOutputs sets multiple output destinations for the global logger.
// All log messages will be written to all specified outputs.
func SetOutputs(outputs ...io.Writer) {
	globalLogger.SetOutputs(outputs...)
}

// SetOutput sets a single output destination for the global logger.
// This is a convenience method for when only one output is needed.
func SetOutput(output io.Writer) {
	globalLogger.SetOutput(output)
}

// SetTimeFormat sets the time format for the global logger.
func SetTimeFormat(format string) {
	globalLogger.SetTimeFormat(format)
}

// AddHook adds a new hook to the global logger.
func AddHook(hook func(level Level, msg string) error, priority int) {
	globalLogger.AddHook(hook, priority)
}

// SetExitFunc allows overriding the exit function for testing.
// This should only be used in test code.
// The original function will be restored when the test completes.
func SetExitFunc(fn func(int)) {
	exitFunc = fn
}

// SetPanicFunc allows overriding the panic function for testing.
// This should only be used in test code.
// The original function will be restored when the test completes.
func SetPanicFunc(fn func(string)) {
	panicFunc = fn
}

// Close stops the logger and cleans up resources.
// This should be called when the logger is no longer needed.
func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Stop the worker pool
	if l.workerPool != nil {
		l.workerPool.stop()
	}

	// Wait for any pending hooks to complete
	l.wg.Wait()

	// Clear hooks
	l.hooks = nil
}
