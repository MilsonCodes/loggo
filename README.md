# loggo

A high-performance, feature-rich logging library for Go applications. Built with performance and simplicity in mind.

![Loggo](https://github.com/user-attachments/assets/7e372f4d-2692-4315-8fa1-b7aed4ca9fea)
## Features

- 🚀 **High Performance**: Optimized for speed with zero-allocation buffer pooling
- 🎨 **Colored Output**: Built-in support for terminal colors
- ⏰ **Customizable Time Format**: Flexible timestamp formatting
- 🔄 **Asynchronous Hooks**: Non-blocking hook execution
- 📊 **Multiple Log Levels**: Debug, Info, Warn, Error, Critical, Fatal, and Panic
- 🔄 **Multiple Outputs**: Support for writing to multiple destinations
- 🛡️ **Thread-Safe**: Designed for concurrent use
- 🎯 **Simple API**: Both simple and chained API styles

## Installation

```bash
go get github.com/milsoncodes/loggo
```

## Quick Start

```go
package main

import "github.com/milsoncodes/loggo"

func main() {
    // Simple API
    loggo.Info("Hello, %s!", "World")

    // Or create a custom logger
    logger := loggo.New()
    logger.SetLevel(loggo.DEBUG)
    logger.Info("Custom logger message")
}
```

## API Reference

### Global Functions

```go
// Simple API
loggo.Debug(msg string, args ...any)
loggo.Info(msg string, args ...any)
loggo.Warn(msg string, args ...any)
loggo.Error(msg string, args ...any)
loggo.Critical(msg string, args ...any)
loggo.Fatal(msg string, args ...any)
loggo.Panic(msg string, args ...any)

// Configuration
loggo.SetLevel(level Level)
loggo.SetOutput(output io.Writer)
loggo.SetOutputs(outputs ...io.Writer)
loggo.SetTimeFormat(format string)
loggo.AddHook(hook func(level Level, msg string) error, priority int)
```

### Logger Instance

```go
// Create a new logger
logger := loggo.New()

// Configure logger
logger.SetLevel(level Level)
logger.SetOutput(output io.Writer)
logger.SetOutputs(outputs ...io.Writer)
logger.SetTimeFormat(format string)
logger.AddHook(hook func(level Level, msg string) error, priority int)

// Logging methods
logger.Debug(msg string, args ...any)
logger.Info(msg string, args ...any)
logger.Warn(msg string, args ...any)
logger.Error(msg string, args ...any)
logger.Critical(msg string, args ...any)
logger.Fatal(msg string, args ...any)
logger.Panic(msg string, args ...any)
```

## Performance

Benchmark results (1000000 iterations each):

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

## Advanced Usage

### Custom Time Format

```go
logger := loggo.New()
logger.SetTimeFormat("2006-01-02 15:04:05.000 MST")
```

### Multiple Outputs

```go
logger := loggo.New()
logger.SetOutputs(os.Stdout, logFile)
```

### Custom Hooks

```go
logger := loggo.New()
hook := func(level loggo.Level, msg string) error {
    // Process log message
    return nil
}
logger.AddHook(hook, 0) // Priority 0 (highest)
```

## Log Levels

- `DEBUG`: Detailed information for debugging
- `INFO`: General information about program execution
- `WARN`: Potentially harmful situations
- `ERROR`: Error events that might still allow the program to continue
- `CRITICAL`: Critical errors that don't trigger a panic
- `FATAL`: Severe errors that cause program termination
- `PANIC`: Critical errors that trigger a panic

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 
