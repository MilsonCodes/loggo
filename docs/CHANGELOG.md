# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of loggo
- High-performance logging library with zero-allocation buffer pooling
- Support for multiple log levels (Debug, Info, Warn, Error, Critical, Fatal, Panic)
- Colored output support
- Customizable time format
- Asynchronous hook execution
- Multiple output support
- Thread-safe operations
- Simple and chained API styles
- Comprehensive test suite
- Performance benchmarks comparing against other logging libraries

### Performance
- Average operation time: 212ns
- Memory allocation: 2,090 B/op
- Number of allocations: 63 allocs/op
- Competitive performance against zerolog and zap
- Significantly faster than logrus and slog

### Documentation
- Comprehensive API documentation
- Performance benchmarks and comparisons
- Usage examples and best practices
- Contributing guidelines
- Changelog 