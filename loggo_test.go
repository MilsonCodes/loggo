package loggo

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestGlobalLogger(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(DEBUG) // Set level to DEBUG to see debug messages

	// Test all log levels
	Debug("debug message")
	Info("info message")
	Warn("warning message")
	Error("error message")
	Critical("critical message")

	// Verify output contains all messages
	output := buf.String()
	expected := []string{
		"debug message",
		"info message",
		"warning message",
		"error message",
		"critical message",
	}

	for _, msg := range expected {
		if !strings.Contains(output, msg) {
			t.Errorf("Expected output to contain %q", msg)
		}
	}
}

func TestCustomLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := New()
	logger.SetOutput(&buf)
	logger.SetLevel(DEBUG)

	logger.Debug("debug message")
	logger.Info("info message")

	output := buf.String()
	if !strings.Contains(output, "debug message") || !strings.Contains(output, "info message") {
		t.Error("Expected output to contain both debug and info messages")
	}
}

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := New()
	logger.SetOutput(&buf)

	// Test level filtering
	logger.SetLevel(WARN)
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warning message")
	logger.Error("error message")

	output := buf.String()
	if strings.Contains(output, "debug message") || strings.Contains(output, "info message") {
		t.Error("Debug and info messages should not be logged at WARN level")
	}
	if !strings.Contains(output, "warning message") || !strings.Contains(output, "error message") {
		t.Error("Warning and error messages should be logged at WARN level")
	}
}

func TestTimeFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := New()
	logger.SetOutput(&buf)
	logger.SetTimeFormat("2006-01-02")

	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, time.Now().Format("2006-01-02")) {
		t.Error("Expected output to contain formatted date")
	}
}

func TestHook(t *testing.T) {
	logger := New()
	defer logger.Close() // Ensure logger is closed after test

	hookCalled := make(chan bool, 1) // Buffered channel to prevent blocking
	hook := func(level Level, msg string) error {
		select {
		case hookCalled <- true:
			// Successfully sent signal
		default:
			// Channel is full, don't block
		}
		return nil
	}

	err := logger.AddHook(hook, 0)
	if err != nil {
		t.Fatalf("Failed to add hook: %v", err)
	}

	logger.Info("Test message")

	// Wait for hook with timeout
	select {
	case <-hookCalled:
		// Hook was called successfully
	case <-time.After(2 * time.Second):
		t.Fatal("Hook was not called within timeout")
	}
}

func TestHookError(t *testing.T) {
	logger := New()
	defer logger.Close() // Ensure logger is closed after test

	hookCalled := make(chan bool, 1) // Buffered channel to prevent blocking
	hook := func(level Level, msg string) error {
		select {
		case hookCalled <- true:
			// Successfully sent signal
		default:
			// Channel is full, don't block
		}
		return fmt.Errorf("hook error")
	}

	err := logger.AddHook(hook, 0)
	if err != nil {
		t.Fatalf("Failed to add hook: %v", err)
	}

	logger.Info("Test message")

	// Wait for hook with timeout
	select {
	case <-hookCalled:
		// Hook was called successfully
	case <-time.After(2 * time.Second):
		t.Fatal("Hook was not called within timeout")
	}

	// Verify hook was removed
	logger.mu.Lock()
	if len(logger.hooks) != 0 {
		t.Error("Hook was not removed after error")
	}
	logger.mu.Unlock()
}

func TestFailingHook(t *testing.T) {
	var buf bytes.Buffer
	logger := New()
	logger.SetOutput(&buf)

	// Create a hook that returns an error
	hook := func(level Level, msg string) error {
		return os.ErrInvalid
	}

	logger.AddHook(hook, 0)
	logger.Info("test message")

	// Wait for hook error to be logged
	time.Sleep(50 * time.Millisecond)

	// Verify hook error was logged
	output := buf.String()
	if !strings.Contains(output, "Hook error") {
		t.Error("Expected hook error to be logged")
	}

	// Clear the buffer for the next check
	buf.Reset()

	// Wait for hook to be removed
	time.Sleep(50 * time.Millisecond)

	// Verify hook was removed by checking no new hook error is logged
	logger.Info("second message")
	time.Sleep(50 * time.Millisecond)
	output = buf.String()
	if strings.Contains(output, "Hook error") {
		t.Error("Hook should have been removed after error")
	}
}

func TestFatal(t *testing.T) {
	// Skip in normal test run as it would exit the process
	if os.Getenv("TEST_FATAL") == "1" {
		var buf bytes.Buffer
		logger := New()
		logger.SetOutput(&buf)
		logger.Fatal("fatal message")
	}
}

func TestCritical(t *testing.T) {
	// Skip in normal test run as it would panic
	if os.Getenv("TEST_CRITICAL") == "1" {
		var buf bytes.Buffer
		logger := New()
		logger.SetOutput(&buf)
		logger.Critical("critical message")
	}
}

func TestPanic(t *testing.T) {
	// Skip in normal test run as it would panic
	if os.Getenv("TEST_PANIC") == "1" {
		var buf bytes.Buffer
		logger := New()
		logger.SetOutput(&buf)
		logger.Panic("panic message")
	}
}
