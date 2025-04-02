# Contributing to loggo

Thank you for your interest in contributing to loggo! This document provides guidelines and instructions for contributing to the project.

## Development Setup

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/yourusername/loggo.git
   cd loggo
   ```
3. Create a new branch for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Code Style

- Follow the standard Go code style
- Use `gofmt` to format your code
- Add comments for exported functions and types
- Keep functions focused and small
- Write tests for new functionality

## Testing

1. Run the test suite:
   ```bash
   go test ./...
   ```

2. Run benchmarks:
   ```bash
   go test -bench=. ./benchmarks/...
   ```

3. Run with race detector:
   ```bash
   go test -race ./...
   ```

## Performance Considerations

When contributing code, please consider:

1. **Memory Allocations**
   - Minimize allocations in hot paths
   - Use buffer pooling where appropriate
   - Avoid unnecessary string conversions

2. **Concurrency**
   - Ensure thread safety
   - Use appropriate synchronization primitives
   - Consider lock contention

3. **Benchmarks**
   - Add benchmarks for new features
   - Ensure changes don't degrade performance
   - Document performance implications

## Pull Request Process

1. Update documentation if needed
2. Add tests for new functionality
3. Run the test suite
4. Run benchmarks
5. Create a pull request with a clear description

## Commit Messages

- Limit the first line to 72 characters or less
- Reference issues and pull requests liberally after the first line

## Release Process

1. Update version in `go.mod`
2. Update documentation
3. Create a release tag
4. Update CHANGELOG.md

## Questions?

Feel free to open an issue for any questions or concerns about contributing. 