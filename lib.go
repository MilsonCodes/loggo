package loggo

import (
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strconv"
	"sync"
	"time"
)

// Color codes for terminal output.
// These are pre-calculated constants to avoid string allocations.
const (
	colorReset  = "\033[0m"  // Reset color
	colorRed    = "\033[31m" // Red color
	colorGreen  = "\033[32m" // Green color
	colorYellow = "\033[33m" // Yellow color
	colorBlue   = "\033[34m" // Blue color
	colorPurple = "\033[35m" // Purple color
	colorCyan   = "\033[36m" // Cyan color
)

// levelColors maps log levels to their corresponding color codes
var levelColors = map[Level]string{
	DEBUG:    colorCyan,
	INFO:     colorGreen,
	WARN:     colorYellow,
	ERROR:    colorRed,
	CRITICAL: colorRed,
	FATAL:    colorRed,
	PANIC:    colorRed,
}

// paddedLevelStrings maps log levels to their padded string representations
var paddedLevelStrings = map[Level]string{
	DEBUG:    "[DEBUG]",
	INFO:     "[INFO] ",
	WARN:     "[WARN] ",
	ERROR:    "[ERROR]",
	CRITICAL: "[CRIT] ",
	FATAL:    "[FATAL]",
	PANIC:    "[PANIC]",
}

// Package level variables for testing
var (
	// exitFunc allows overriding os.Exit for testing
	exitFunc = os.Exit
	// panicFunc allows overriding panic for testing
	panicFunc = func(v string) { panic(v) }
)

// multiWriter is a custom writer that writes to multiple outputs
type multiWriter struct {
	writers []io.Writer
	mu      sync.Mutex
}

// newMultiWriter creates a new multiWriter with the given writers
func newMultiWriter(writers ...io.Writer) *multiWriter {
	return &multiWriter{
		writers: writers,
	}
}

// write writes the given data to all registered writers
func (w *multiWriter) write(data []byte) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, writer := range w.writers {
		writer.Write(data)
	}
}

// workerPool manages a pool of workers for executing jobs
type workerPool struct {
	jobs    chan func()
	wg      sync.WaitGroup
	stopper chan struct{}
	workers int
}

// newWorkerPool creates a new worker pool with the specified number of workers
func newWorkerPool(workers int) *workerPool {
	pool := &workerPool{
		jobs:    make(chan func(), workers*2),
		stopper: make(chan struct{}),
		workers: workers,
	}

	for i := 0; i < workers; i++ {
		pool.wg.Add(1)
		go pool.worker()
	}

	return pool
}

// worker processes jobs from the queue
func (p *workerPool) worker() {
	defer p.wg.Done()

	for {
		select {
		case job := <-p.jobs:
			job()
		case <-p.stopper:
			return
		}
	}
}

// event represents a log event that can be built using a chained API.
// The event type provides a fluent interface for building log messages
// with zero allocations. It is created by calling one of the level methods
// on a Logger (e.g., logger.Debug(), logger.Info(), etc.).
//
// Example:
//
//	logger := loggo.New()
//	logger.Info().Msgf("Processing request %d", 123)
//
// Performance Note: Events are designed for zero-allocation logging by
// writing directly to a pooled buffer. The Msgf method formats and writes
// the message in a single operation, minimizing memory allocations.
type event struct {
	logger *Logger
	level  Level
	buf    *[]byte
}

// msgf formats and writes the message to the event buffer.
// The format string and arguments follow the same rules as fmt.Sprintf.
// After writing the message, the buffer is returned to the pool.
//
// Example:
//
//	logger.Info().msgf("Processing request %d from %s", 123, "user")
//
// Performance Note: This method writes directly to the buffer without
// intermediate string allocations. The buffer is automatically returned
// to the pool after use.
func (e *event) msgf(format string, args ...any) {
	if e == nil {
		return
	}
	defer e.logger.putBuffer(e.buf)

	// Format timestamp
	timestamp := e.logger.getFormattedTime()

	// Pre-allocate buffer with estimated size
	// Format: color + level + reset + timestamp + ": " + message + "\n"
	estimatedSize := len(levelColors[e.level]) + len(e.level.PaddedString()) +
		len(colorReset) + len(timestamp) + 2 + len(format) + 1

	// Resize buffer if needed
	if cap(*e.buf) < estimatedSize {
		newBuf := e.logger.getBuffer(estimatedSize)
		*newBuf = append(*newBuf, *e.buf...)
		e.buf = newBuf
	}

	// Write the formatted message directly to the buffer
	*e.buf = fmt.Appendf(*e.buf, "%s%s%s %s: ",
		levelColors[e.level],
		e.level.PaddedString(),
		colorReset,
		timestamp,
	)

	// Optimize common formatting patterns
	if len(args) == 0 {
		*e.buf = append(*e.buf, format...)
	} else if len(args) == 1 {
		switch v := args[0].(type) {
		case string:
			*e.buf = append(*e.buf, v...)
		case int:
			*e.buf = strconv.AppendInt(*e.buf, int64(v), 10)
		case int64:
			*e.buf = strconv.AppendInt(*e.buf, v, 10)
		case float64:
			*e.buf = strconv.AppendFloat(*e.buf, v, 'f', -1, 64)
		case error:
			*e.buf = append(*e.buf, v.Error()...)
		default:
			*e.buf = fmt.Appendf(*e.buf, format, args...)
		}
	} else {
		*e.buf = fmt.Appendf(*e.buf, format, args...)
	}

	*e.buf = append(*e.buf, '\n')

	// Write to output
	e.logger.output.write(*e.buf)

	// Execute hooks if any exist, but only format message if hooks are present
	if len(e.logger.hooks) > 0 {
		// Only format message if hooks are present
		message := fmt.Sprintf(format, args...)
		e.logger.executeHooks(e.level, message)
	}

	if e.level == FATAL {
		e.logger.wg.Wait()
		e.logger.workerPool.stop()
		exitFunc(1)
	}
	if e.level == PANIC {
		message := fmt.Sprintf(format, args...)
		e.logger.wg.Wait()
		e.logger.workerPool.stop()
		panicFunc(message)
	}
}

// stop stops the worker pool and waits for all workers to finish
func (p *workerPool) stop() {
	close(p.stopper)
	p.wg.Wait()
}

// submit submits a job to the worker pool
func (p *workerPool) submit(job func()) {
	select {
	case p.jobs <- job:
	case <-p.stopper:
	}
}

// sortedKeys returns a sorted slice of map keys
func sortedKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return fmt.Sprintf("%v", keys[i]) < fmt.Sprintf("%v", keys[j])
	})
	return keys
}

// getBuffer gets a buffer from the pool
func (l *Logger) getBuffer(size int) *[]byte {
	if size > l.bufSize*4 {
		buf := make([]byte, 0, size)
		return &buf
	}
	buf := l.pool.Get().(*[]byte)
	*buf = (*buf)[:0]
	return buf
}

// putBuffer returns a buffer to the pool
func (l *Logger) putBuffer(buf *[]byte) {
	if buf == nil {
		return
	}
	if cap(*buf) > l.bufSize*4 {
		return
	}
	l.pool.Put(buf)
}

// newEvent creates a new event with the given level
func (l *Logger) newEvent(level Level) *event {
	if level < l.level {
		return nil
	}
	buf := l.getBuffer(l.bufSize)
	return &event{
		logger: l,
		level:  level,
		buf:    buf,
	}
}

// getFormattedTime returns a formatted timestamp, using caching for efficiency
func (l *Logger) getFormattedTime() string {
	now := time.Now()
	key := now.Unix()

	// Check if we have a cached value for this second
	if key == l.timeKey {
		return l.timeValue
	}

	// Format the time
	formatted := now.Format(l.timeFormat)

	// Update cache
	l.timeKey = key
	l.timeValue = formatted

	// Clean up old cache entries if needed
	l.cleanupTimeCache()

	return formatted
}

// cleanupTimeCache removes old entries from the time format cache
func (l *Logger) cleanupTimeCache() {
	if l.cleanupInProgress {
		return
	}

	now := time.Now().Unix()
	if now-l.lastCleanup < 60 { // Clean up at most once per minute
		return
	}

	l.cleanupInProgress = true
	defer func() {
		l.cleanupInProgress = false
		l.lastCleanup = now
	}()

	l.timeCache.Range(func(key, value any) bool {
		if key.(int64) < now-3600 { // Remove entries older than 1 hour
			l.timeCache.Delete(key)
		}
		return true
	})
}

// executeHooks executes all registered hooks asynchronously
func (l *Logger) executeHooks(level Level, msg string) {
	l.wg.Add(1)
	l.workerPool.submit(func() {
		defer l.wg.Done()

		// Sort hooks by priority (higher priority first)
		hooks := make([]Hook, len(l.hooks))
		copy(hooks, l.hooks)
		slices.SortFunc(hooks, func(a, b Hook) int {
			return b.priority - a.priority
		})

		// Execute hooks
		for _, hook := range hooks {
			if err := hook.fn(level, msg); err != nil {
				// Log the error and remove the hook
				fmt.Fprintf(os.Stderr, "Hook error: %v\n", err)
				l.removeHook(hook.id)
			}
		}
	})
}

// removeHook removes a hook by its ID
func (l *Logger) removeHook(id string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i, hook := range l.hooks {
		if hook.id == id {
			l.hooks = append(l.hooks[:i], l.hooks[i+1:]...)
			return
		}
	}
}

// Internal event creation methods
func (l *Logger) debugEvent() *event    { return l.newEvent(DEBUG) }
func (l *Logger) infoEvent() *event     { return l.newEvent(INFO) }
func (l *Logger) warnEvent() *event     { return l.newEvent(WARN) }
func (l *Logger) errorEvent() *event    { return l.newEvent(ERROR) }
func (l *Logger) criticalEvent() *event { return l.newEvent(CRITICAL) }
func (l *Logger) fatalEvent() *event    { return l.newEvent(FATAL) }
func (l *Logger) panicEvent() *event    { return l.newEvent(PANIC) }
