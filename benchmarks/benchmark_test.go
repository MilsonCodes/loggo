// go test -bench=BenchmarkLogLevels -benchtime=1000000x -benchmem
// Package benchmarks provides performance comparison tests for the loggo library
// against other popular logging libraries in the Go ecosystem.
package benchmarks

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"loggo"

	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// BenchmarkLogLevels compares the performance of loggo against other logging libraries.
// It runs each benchmark 10000 times and calculates the average.
//
// The benchmark compares:
// - Operation time (ns/op)
// - Memory allocation (B/op)
// - Number of allocations (allocs/op)
//
// Libraries compared:
// - loggo: Our high-performance logging library
// - logrus: Popular structured logging library
// - zap: High-performance structured logging library
// - zerolog: Zero-allocation JSON logger
// - slog: Go's built-in structured logging
func BenchmarkLogLevels(b *testing.B) {
	// Test message and arguments
	msg := "Benchmark test message %d"
	args := []any{123}

	// Initialize loggers with consistent configuration
	loggoLogger := loggo.New()
	defer loggoLogger.Close() // Ensure logger is closed after benchmark
	var loggoBuf bytes.Buffer
	loggoLogger.SetOutput(&loggoBuf)

	// Override the default exit behavior for testing
	loggo.SetExitFunc(func(int) {})

	// Initialize logrus with minimal configuration
	logrusLogger := logrus.New()
	var logrusBuf bytes.Buffer
	logrusLogger.SetOutput(&logrusBuf)
	logrusLogger.SetFormatter(&logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: true,
	})

	// Initialize zap with production configuration
	zapCore := zap.NewProductionEncoderConfig()
	zapCore.TimeKey = "" // Disable timestamps for benchmarking
	encoder := zapcore.NewJSONEncoder(zapCore)
	var zapBuf bytes.Buffer
	zapLogger := zap.New(zapcore.NewCore(
		encoder,
		zapcore.AddSync(&zapBuf),
		zap.InfoLevel,
	))
	defer zapLogger.Sync()

	// Initialize zerolog with timestamp
	var zerologBuf bytes.Buffer
	zerologLogger := zerolog.New(&zerologBuf).With().Timestamp().Logger()

	// Initialize slog with debug level
	var slogBuf bytes.Buffer
	slogLogger := slog.New(slog.NewTextHandler(&slogBuf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Define benchmark levels and their corresponding logging functions
	levels := []struct {
		name    string
		loggo   func(string, ...any)
		logrus  func(string, ...any)
		zap     func(string, ...any)
		zerolog func(string, ...any)
		slog    func(string, ...any)
	}{
		{
			name:    "Debug",
			loggo:   func(msg string, args ...any) { loggoLogger.Debugf(msg, args...) },
			logrus:  func(msg string, args ...any) { logrusLogger.Debugf(msg, args...) },
			zap:     func(msg string, args ...any) { zapLogger.Debug(fmt.Sprintf(msg, args...)) },
			zerolog: func(msg string, args ...any) { zerologLogger.Debug().Msgf(msg, args...) },
			slog:    func(msg string, args ...any) { slogLogger.Debug(fmt.Sprintf(msg, args...)) },
		},
		{
			name:    "Info",
			loggo:   func(msg string, args ...any) { loggoLogger.Infof(msg, args...) },
			logrus:  func(msg string, args ...any) { logrusLogger.Infof(msg, args...) },
			zap:     func(msg string, args ...any) { zapLogger.Info(fmt.Sprintf(msg, args...)) },
			zerolog: func(msg string, args ...any) { zerologLogger.Info().Msgf(msg, args...) },
			slog:    func(msg string, args ...any) { slogLogger.Info(fmt.Sprintf(msg, args...)) },
		},
		{
			name:    "Warn",
			loggo:   func(msg string, args ...any) { loggoLogger.Warnf(msg, args...) },
			logrus:  func(msg string, args ...any) { logrusLogger.Warnf(msg, args...) },
			zap:     func(msg string, args ...any) { zapLogger.Warn(fmt.Sprintf(msg, args...)) },
			zerolog: func(msg string, args ...any) { zerologLogger.Warn().Msgf(msg, args...) },
			slog:    func(msg string, args ...any) { slogLogger.Warn(fmt.Sprintf(msg, args...)) },
		},
		{
			name:    "Error",
			loggo:   func(msg string, args ...any) { loggoLogger.Errorf(msg, args...) },
			logrus:  func(msg string, args ...any) { logrusLogger.Errorf(msg, args...) },
			zap:     func(msg string, args ...any) { zapLogger.Error(fmt.Sprintf(msg, args...)) },
			zerolog: func(msg string, args ...any) { zerologLogger.Error().Msgf(msg, args...) },
			slog:    func(msg string, args ...any) { slogLogger.Error(fmt.Sprintf(msg, args...)) },
		},
	}

	// Run benchmarks for each level
	const iterations = 1000000
	results := make(map[string][]time.Duration)

	for _, level := range levels {
		results[level.name] = make([]time.Duration, 0, iterations*5) // 5 loggers

		// Test each logger
		for i := 0; i < iterations; i++ {
			start := time.Now()
			level.loggo(msg, args...)
			results[level.name] = append(results[level.name], time.Since(start))
			loggoBuf.Reset()
		}

		// Test logrus
		for i := 0; i < iterations; i++ {
			start := time.Now()
			level.logrus(msg, args...)
			results[level.name] = append(results[level.name], time.Since(start))
			logrusBuf.Reset()
		}

		// Test zap
		for i := 0; i < iterations; i++ {
			start := time.Now()
			level.zap(msg, args...)
			results[level.name] = append(results[level.name], time.Since(start))
			zapBuf.Reset()
		}

		// Test zerolog
		for i := 0; i < iterations; i++ {
			start := time.Now()
			level.zerolog(msg, args...)
			results[level.name] = append(results[level.name], time.Since(start))
			zerologBuf.Reset()
		}

		// Test slog
		for i := 0; i < iterations; i++ {
			start := time.Now()
			level.slog(msg, args...)
			results[level.name] = append(results[level.name], time.Since(start))
			slogBuf.Reset()
		}
	}

	// Generate benchmark report
	filename := fmt.Sprintf("bench%s.out", time.Now().Format("020106"))
	f, err := os.Create(filename)
	if err != nil {
		b.Fatalf("Failed to create output file: %v", err)
	}
	defer f.Close()

	// Write results to file
	fmt.Fprintf(f, "Logging Performance Results (%d iterations each)\n", iterations)
	fmt.Fprintf(f, "================================================\n")
	fmt.Fprintf(f, "%-10s %-15s %-15s %-15s %-15s %-15s\n",
		"Function", "loggo", "zerolog", "zap", "logrus", "slog")
	fmt.Fprintf(f, "------------------------------------------------\n")

	// Calculate averages for each logger
	var loggoTotal, zerologTotal, zapTotal, logrusTotal, slogTotal time.Duration

	for _, fn := range levels {
		fnResults := results[fn.name]

		loggoAvg := average(fnResults[0:iterations])
		zerologAvg := average(fnResults[iterations*3 : iterations*4])
		zapAvg := average(fnResults[iterations*2 : iterations*3])
		logrusAvg := average(fnResults[iterations : iterations*2])
		slogAvg := average(fnResults[iterations*4 : iterations*5])

		// Add to totals for overall average
		loggoTotal += loggoAvg
		zerologTotal += zerologAvg
		zapTotal += zapAvg
		logrusTotal += logrusAvg
		slogTotal += slogAvg

		fmt.Fprintf(f, "%-10s %-15v %-15v %-15v %-15v %-15v\n",
			fn.name,
			loggoAvg,
			zerologAvg,
			zapAvg,
			logrusAvg,
			slogAvg,
		)
	}

	// Calculate and print overall averages
	fmt.Fprintf(f, "------------------------------------------------\n")
	fmt.Fprintf(f, "%-10s %-15v %-15v %-15v %-15v %-15v\n",
		"AVERAGE",
		loggoTotal/time.Duration(len(levels)),
		zerologTotal/time.Duration(len(levels)),
		zapTotal/time.Duration(len(levels)),
		logrusTotal/time.Duration(len(levels)),
		slogTotal/time.Duration(len(levels)),
	)
}

// average calculates the average duration from a slice of durations.
// It returns 0 if the slice is empty to avoid division by zero.
func average(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	var sum time.Duration
	for _, d := range durations {
		sum += d
	}
	return sum / time.Duration(len(durations))
}
