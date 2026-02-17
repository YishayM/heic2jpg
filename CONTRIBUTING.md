# Contributing to heic2jpg

Thank you for your interest in contributing to heic2jpg! This document provides guidelines and instructions for contributing.

## Development Setup

### Prerequisites

- Go 1.21 or later
- C compiler (gcc on Linux, Xcode Command Line Tools on macOS, MinGW on Windows)
- Git

### Clone and Build

```bash
# Clone the repository
git clone https://github.com/yourusername/heic2jpg.git
cd heic2jpg

# Build the project
make build

# The binary will be in the dist/ directory
./dist/heic2jpg --version
```

### Development Dependencies

On Linux, you may need additional development libraries:
```bash
sudo apt-get install libde265-dev
```

On macOS, install Xcode Command Line Tools:
```bash
xcode-select --install
```

## Running Tests

Run all tests:
```bash
go test ./...
```

Run tests with verbose output:
```bash
go test -v ./...
```

Run tests with coverage:
```bash
go test -cover ./...
```

## Code Style

### Formatting

All Go code must be formatted with `gofmt`:
```bash
# Format all files
gofmt -w .

# Check formatting
gofmt -l .
```

### Linting

We recommend using `golangci-lint` for comprehensive linting:
```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run
```

### Code Guidelines

- Write clear, descriptive variable and function names
- Add comments for exported functions and types
- Keep functions focused and concise
- Handle errors explicitly - don't ignore them
- Write tests for new functionality
- Follow Go best practices and idioms

## Submitting Changes

### Before You Submit

1. **Test your changes**: Run `go test ./...` and ensure all tests pass
2. **Format your code**: Run `gofmt -w .`
3. **Lint your code**: Run `golangci-lint run` if available
4. **Update documentation**: Update README.md if you've added features or changed behavior
5. **Add tests**: Include tests for new functionality

### Pull Request Process

1. **Fork the repository** and create a new branch from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following the code style guidelines

3. **Commit your changes** with clear, descriptive commit messages:
   ```bash
   git commit -m "Add feature: description of what you added"
   ```

4. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

5. **Open a Pull Request** with:
   - Clear title describing the change
   - Description of what changed and why
   - Reference any related issues
   - Screenshots/examples if applicable

6. **Respond to feedback**: Maintainers may request changes. Please respond promptly and update your PR as needed.

### Commit Message Guidelines

- Use present tense ("Add feature" not "Added feature")
- Use imperative mood ("Move cursor to..." not "Moves cursor to...")
- Keep first line under 72 characters
- Reference issues and PRs when relevant

Examples:
```
Add support for custom JPEG quality setting
Fix EXIF preservation for rotated images
Update README with new installation instructions
```

## Reporting Issues

### Bug Reports

When reporting bugs, please include:
- Operating system and version
- Go version (`go version`)
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Error messages or logs
- Sample HEIC file if possible (if not sensitive)

### Feature Requests

When requesting features, please include:
- Clear description of the feature
- Use case - why is this feature needed?
- Examples of how it would work
- Any alternative solutions you've considered

## Questions?

If you have questions about contributing, feel free to:
- Open an issue with the "question" label
- Start a discussion in GitHub Discussions (if enabled)

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers and help them learn
- Focus on constructive feedback
- Assume good intentions

Thank you for contributing to heic2jpg!

