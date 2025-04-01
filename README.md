# Loggo

A simple and efficient logging library for Go applications.

## Features

- Simple and intuitive API
- Multiple log levels
- Customizable output formats
- Thread-safe logging
- Global logger instance for easy use
- Hook system for external integrations

## Installation

```bash
go get github.com/milsoncodes/loggo
```

## Usage

### Using the global logger (recommended for simple use)

```go
import log "github.com/milsoncodes/loggo"

func main() {
    log.Info("Hello, Loggo!")
    log.Debug("Debug message")
    log.Warn("Warning message")
    log.Error("Error message")
    log.Critical("Critical error")
    log.Fatal("Fatal error")
}
```

### Using a custom logger instance

```go
package main

import "github.com/milsoncodes/loggo"

func main() {
    log := loggo.New()
    log.SetLevel(loggo.DEBUG)
    log.Info("Hello, Loggo!")
}
```

### Using Hooks

Hooks allow you to send logs to external systems. Here's an example of sending logs to Grafana Loki:

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
    "time"
    log "github.com/milsoncodes/loggo"
)

type LokiLog struct {
    Streams []struct {
        Stream struct {
            Level string `json:"level"`
        } `json:"stream"`
        Values [][]string `json:"values"`
    } `json:"streams"`
}

func main() {
    // Create a hook that sends logs to Loki
    lokiHook := func(level loggo.Level, msg string) error {
        lokiLog := LokiLog{
            Streams: []struct {
                Stream struct {
                    Level string `json:"level"`
                } `json:"stream"`
                Values [][]string `json:"values"`
            }{
                {
                    Stream: struct {
                        Level string `json:"level"`
                    }{
                        Level: level.String(),
                    },
                    Values: [][]string{
                        {
                            time.Now().Format(time.RFC3339Nano),
                            msg,
                        },
                    },
                },
            },
        }

        jsonData, err := json.Marshal(lokiLog)
        if err != nil {
            return err
        }

        resp, err := http.Post(
            "http://localhost:3100/loki/api/v1/push",
            "application/json",
            bytes.NewBuffer(jsonData),
        )
        if err != nil {
            return err
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
        }

        return nil
    }

    // Add the hook to the logger
    log.AddHook(lokiHook)

    // Now all logs will be sent to both stdout and Loki
    log.Info("Hello, Loggo!")
}
```

## Configuration

You can configure the global logger using the following functions:

```go
log.SetLevel(loggo.DEBUG)  // Set logging level
log.SetOutput(os.Stderr)   // Set output destination
log.SetTimeFormat("2006-01-02 15:04:05.999 MST")  // Set time format
log.AddHook(hook)          // Add a hook for external integrations
```

## License

MIT License 