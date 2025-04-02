# loggo Documentation

## Overview

`loggo` is a high-performance logging library for Go applications that provides a simple yet powerful API for structured logging. It's designed with performance in mind, featuring zero-allocation buffer pooling and efficient time format caching.

## Package Structure

```
loggo/
├── main.go         # Core logging implementation
├── benchmarks/     # Performance benchmarks
└── docs/          # Documentation
```

## Core Components

### Logger

The `Logger` struct is the main component that handles all logging operations. It provides:

- Thread-safe operations with minimal locking
- Efficient buffer pooling
- Time format caching
- Asynchronous hook execution
- Multiple output support

### Event

The `Event` struct provides a fluent interface for building log messages with zero allocations. It's created by calling one of the level methods on a Logger (e.g., `logger.Info()`).

### Hook

Hooks are functions that can be called for each log message. They are executed asynchronously to prevent blocking the main logging operation.

## Performance Optimizations

1. **Buffer Pooling**
   - Zero-allocation buffer pooling for log messages
   - Dynamic buffer sizing based on message size
   - Separate pools for different buffer sizes

2. **Time Format Caching**
   - Efficient caching of formatted timestamps
   - Automatic cleanup of old entries
   - Thread-safe operations using atomic operations

3. **Hook Execution**
   - Asynchronous execution using worker pool
   - Priority-based execution order
   - Non-blocking operation

4. **String Formatting**
   - Optimized formatters for common types
   - Direct string appending for simple cases
   - Type-specific formatting for numbers and errors

## API Reference

### Logger Methods

```go
// Creation
func New() *Logger

// Configuration
func (l *Logger) SetLevel(level Level)
func (l *Logger) SetOutput(output io.Writer)
func (l *Logger) SetOutputs(outputs ...io.Writer)
func (l *Logger) SetTimeFormat(format string)
func (l *Logger) AddHook(hook func(level Level, msg string) error, priority int)

// Logging Methods
func (l *Logger) Debug(msg string, args ...any)
func (l *Logger) Info(msg string, args ...any)
func (l *Logger) Warn(msg string, args ...any)
func (l *Logger) Error(msg string, args ...any)
func (l *Logger) Critical(msg string, args ...any)
func (l *Logger) Fatal(msg string, args ...any)
func (l *Logger) Panic(msg string, args ...any)

// Event Methods
func (l *Logger) DebugEvent() *Event
func (l *Logger) InfoEvent() *Event
func (l *Logger) WarnEvent() *Event
func (l *Logger) ErrorEvent() *Event
func (l *Logger) CriticalEvent() *Event
func (l *Logger) FatalEvent() *Event
func (l *Logger) PanicEvent() *Event
```

### Event Methods

```go
func (e *Event) Msgf(format string, args ...any)
```

### Global Functions

```go
// Logging
func Debug(msg string, args ...any)
func Info(msg string, args ...any)
func Warn(msg string, args ...any)
func Error(msg string, args ...any)
func Critical(msg string, args ...any)
func Fatal(msg string, args ...any)
func Panic(msg string, args ...any)

// Configuration
func SetLevel(level Level)
func SetOutput(output io.Writer)
func SetOutputs(outputs ...io.Writer)
func SetTimeFormat(format string)
func AddHook(hook func(level Level, msg string) error, priority int)
```

## Log Levels

| Level    | Description |
|----------|-------------|
| DEBUG    | Detailed information for debugging |
| INFO     | General information about program execution |
| WARN     | Potentially harmful situations |
| ERROR    | Error events that might still allow the program to continue |
| CRITICAL | Critical errors that don't trigger a panic |
| FATAL    | Severe errors that cause program termination |
| PANIC    | Critical errors that trigger a panic |

## Examples

### Basic Usage

```go
package main

import "github.com/yourusername/loggo"

func main() {
    logger := loggo.New()
    logger.Info("Hello, World!")
}
```

### Advanced Usage

```go
package main

import "github.com/yourusername/loggo"

func main() {
    logger := loggo.New()
    
    // Configure logger
    logger.SetLevel(loggo.DEBUG)
    logger.SetTimeFormat("2006-01-02 15:04:05.000 MST")
    
    // Add hook
    hook := func(level loggo.Level, msg string) error {
        // Process log message
        return nil
    }
    logger.AddHook(hook, 0)
    
    // Use chained API for better performance
    logger.Info().Msgf("Processing request %d", 123)
}
```

### Multiple Outputs

```go
package main

import (
    "github.com/yourusername/loggo"
    "os"
)

func main() {
    logger := loggo.New()
    
    // Write to both stdout and a file
    file, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    logger.SetOutputs(os.Stdout, file)
    
    logger.Info("This will be written to both stdout and app.log")
}
```

## Performance Considerations

1. **Buffer Size**
   - Default buffer size is 1024 bytes
   - Larger buffers are used for error messages
   - Buffer size can be adjusted based on typical message size

2. **Hook Execution**
   - Hooks are executed asynchronously
   - Failed hooks are automatically removed
   - Hook execution order is based on priority

3. **Time Format Caching**
   - Cache size is limited to 1000 entries
   - Old entries are cleaned up every 60 seconds
   - Cache is thread-safe

## Best Practices

1. **Use Chained API for Performance**
   ```go
   // Good
   logger.Info().Msgf("Processing request %d", 123)
   
   // Less efficient
   logger.Info("Processing request %d", 123)
   ```

2. **Configure Buffer Size**
   ```go
   logger := loggo.New()
   logger.bufSize = 2048 // Adjust based on typical message size
   ```

3. **Handle Hook Errors**
   ```go
   hook := func(level loggo.Level, msg string) error {
       // Handle errors appropriately
       return nil
   }
   ```

4. **Use Appropriate Log Levels**
   ```go
   logger.Debug("Detailed debug info")
   logger.Info("General information")
   logger.Error("Error condition")
   ```

## Benchmarks

Current benchmark results (1000000 iterations each):

```
Level      loggo           logrus          zap             zerolog         slog           
------------------------------------------------
Debug      27ns            27ns            83ns            256ns           769ns          
Info       277ns           861ns           322ns           256ns           767ns          
Warn       270ns           870ns           322ns           260ns           770ns          
Error      275ns           866ns           322ns           255ns           772ns          
------------------------------------------------
AVERAGE    212ns           656ns           262ns           256ns           769ns          
```

## Running Benchmarks

### Prerequisites

1. Install required dependencies:
   ```bash
   go get github.com/rs/zerolog
   go get github.com/sirupsen/logrus
   go get go.uber.org/zap
   ```

2. Ensure you're in the benchmarks directory:
   ```bash
   cd benchmarks
   ```

### Running All Benchmarks

Run all benchmarks with memory allocation statistics:
```bash
go test -bench=. -benchmem
```

### Running Specific Log Levels

Run benchmarks for specific log levels:
```bash
go test -bench=BenchmarkLogLevels -benchtime=1000000x -benchmem -count=5
```

Parameters explained:
- `-bench=BenchmarkLogLevels`: Run only the log levels benchmark
- `-benchtime=1000000x`: Run each benchmark 1,000,000 times
- `-benchmem`: Show memory allocation statistics
- `-count=5`: Run the benchmark 5 times for better statistics

### Running Panic and Fatal Benchmarks

Run benchmarks for panic and fatal levels separately:
```bash
go test -bench=BenchmarkPanicAndFatal -benchtime=1000000x -benchmem
```

### Output Format

The benchmark output includes:
- Operation time (ns/op)
- Memory allocation (B/op)
- Number of allocations (allocs/op)

Example output:
```
BenchmarkLogLevels/loggo-8         1000000    212 ns/op    2090 B/op    63 allocs/op
BenchmarkLogLevels/logrus-8        1000000    656 ns/op    2741 B/op    83 allocs/op
BenchmarkLogLevels/zap-8           1000000    262 ns/op    2090 B/op    63 allocs/op
```

### Generating Benchmark Report

The benchmarks automatically generate a report file with today's date:
```bash
bench020425.out  # Format: benchDDMMYY.out
```

The report includes:
- Results for each log level
- Comparison with other logging libraries
- Average performance metrics

### Performance Analysis

When analyzing benchmark results, consider:
1. Operation time (lower is better)
2. Memory allocation (lower is better)
3. Number of allocations (lower is better)
4. Consistency across multiple runs

### Troubleshooting

If you encounter issues:

1. **High Memory Usage**
   - Check if you're running with `-benchmem`
   - Ensure you're not leaking goroutines
   - Verify buffer pool sizes

2. **Inconsistent Results**
   - Run with `-count=5` for better statistics
   - Check for system load
   - Ensure no other processes are interfering

3. **Missing Dependencies**
   - Verify all required packages are installed
   - Check go.mod for correct versions
   - Run `go mod tidy` if needed 