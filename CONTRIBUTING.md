# Contributing to TFE

Thank you for your interest in contributing to TFE (Terminal File Explorer)! This document provides guidelines for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How to Contribute](#how-to-contribute)
- [Development Guidelines](#development-guidelines)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)

## Code of Conduct

Be respectful and constructive in all interactions with the community.

## Getting Started

### Prerequisites

- Go 1.24 or higher
- A terminal with Nerd Fonts installed (for proper icon display)
- Git for version control

### Setting Up Development Environment

1. **Fork the repository** on GitHub

2. **Clone your fork:**
```bash
git clone https://github.com/YOUR_USERNAME/tfe.git
cd tfe
```

3. **Install dependencies:**
```bash
go mod download
```

4. **Build and run:**
```bash
go build -o tfe
./tfe
```

5. **Run tests:**
```bash
go test ./...
```

## How to Contribute

### Reporting Bugs

Before creating a bug report, please check existing issues to avoid duplicates.

**When reporting a bug, please include:**
- Your operating system and version
- Go version (`go version`)
- Terminal emulator you're using
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Screenshots (if applicable)

**Create an issue with:**
- A clear, descriptive title
- Detailed description of the problem
- Steps to reproduce
- Any relevant logs or error messages

### Suggesting Enhancements

Enhancement suggestions are welcome! Please:
- Use a clear, descriptive title
- Provide detailed description of the proposed feature
- Explain why this enhancement would be useful
- Include examples of how it would work (if applicable)

### Your First Code Contribution

Good first issues for new contributors are labeled with `good first issue` on GitHub.

## Development Guidelines

### Code Style

TFE follows standard Go conventions:

- Use `go fmt` to format your code
- Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use meaningful variable and function names
- Add comments for exported functions (GoDoc style)

### Architecture

TFE follows a **modular architecture**. Please read `CLAUDE.md` for detailed information about:
- Module responsibilities
- File organization principles
- How to add new features

**Key principles:**
- Keep `main.go` minimal (entry point only)
- One responsibility per file
- Files should be under 800 lines
- Group related functionality together

### Adding New Features

When adding a new feature:

1. **Check if it fits an existing module** (see `CLAUDE.md`)
2. **Create a new module if needed** for substantial features
3. **Update documentation** (`CLAUDE.md`, `README.md`, `HOTKEYS.md` as applicable)
4. **Add tests** for new functionality (minimum 40% coverage for new code)
5. **Update `CHANGELOG.md`** with your changes

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run with race detection
go test -race ./...
```

### Writing Tests

- All new features should include tests
- Aim for **minimum 40% code coverage** for new code
- Test files should be named `*_test.go`
- Follow existing test patterns in the codebase

**Test structure:**
```go
func TestYourFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case 1", "input1", "output1"},
        {"case 2", "input2", "output2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := yourFunction(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Testing Best Practices

- Use table-driven tests when testing multiple cases
- Test edge cases and error conditions
- Use descriptive test names
- Keep tests focused and independent
- Use `t.TempDir()` for temporary files/directories

## Pull Request Process

### Before Submitting

1. **Update your fork:**
```bash
git fetch upstream
git rebase upstream/main
```

2. **Run all checks:**
```bash
go test ./...        # All tests must pass
go vet ./...         # No vet warnings
go fmt ./...         # Code must be formatted
```

3. **Update documentation** if you've:
   - Added a new feature
   - Changed keyboard shortcuts
   - Modified the architecture
   - Changed configuration options

4. **Test manually:**
   - Test your changes in the actual TFE application
   - Verify keyboard shortcuts work
   - Check mouse interactions
   - Test on different terminal sizes if applicable

### Submitting a Pull Request

1. **Create a feature branch:**
```bash
git checkout -b feature/your-feature-name
```

2. **Make your changes** following the guidelines above

3. **Commit with clear messages:**
```bash
git add .
git commit -m "feat: Add feature description

Detailed explanation of what this commit does and why."
```

**Commit message format:**
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `test:` Adding or updating tests
- `refactor:` Code refactoring
- `perf:` Performance improvements
- `style:` Code style changes (formatting, etc.)
- `chore:` Maintenance tasks

4. **Push to your fork:**
```bash
git push origin feature/your-feature-name
```

5. **Create Pull Request** on GitHub

### Pull Request Requirements

Your PR must:
- ✅ Pass all tests (`go test ./...`)
- ✅ Pass `go vet` with no warnings
- ✅ Be formatted with `go fmt`
- ✅ Include tests for new functionality
- ✅ Update relevant documentation
- ✅ Have a clear description of changes
- ✅ Reference any related issues

### PR Description Template

```markdown
## Description
Brief description of what this PR does

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
Describe the tests you added or ran

## Screenshots (if applicable)
Add screenshots for UI changes

## Checklist
- [ ] Tests pass locally
- [ ] Code is formatted
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
```

### Review Process

- A maintainer will review your PR
- Address any feedback or requested changes
- Once approved, your PR will be merged

## Documentation Line Limits

To keep documentation manageable, please adhere to these limits:

- `CLAUDE.md`: 500 lines max
- `README.md`: 400 lines max
- `PLAN.md`: 400 lines max
- `CHANGELOG.md`: 300 lines max

If a file exceeds its limit, archive old content to `docs/archive/`.

## Questions?

- Check existing [GitHub Issues](https://github.com/GGPrompts/TFE/issues)
- Read the [CLAUDE.md](CLAUDE.md) architecture guide
- Open a new issue for questions or discussions

## License

By contributing to TFE, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to TFE!** 🚀
