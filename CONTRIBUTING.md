# Contributing to ML Notes

First off, thank you for considering contributing to ML Notes! It's people like you that make ML Notes such a great tool.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How Can I Contribute?](#how-can-i-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Enhancements](#suggesting-enhancements)
  - [Pull Requests](#pull-requests)
- [Development Setup](#development-setup)
- [Style Guidelines](#style-guidelines)
- [Commit Messages](#commit-messages)
- [Testing](#testing)

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code. Please be respectful and considerate in your interactions with other contributors.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
3. Create a new branch for your contribution
4. Make your changes
5. Push to your fork and submit a pull request

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When you create a bug report, include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples**
- **Describe the behavior you observed and what you expected**
- **Include system information** (OS, Go version, etc.)
- **Include debug output** if relevant (`ml-notes --debug`)

### Suggesting Enhancements

Enhancement suggestions are welcome! Please provide:

- **A clear and descriptive title**
- **A detailed description of the proposed enhancement**
- **Explain why this enhancement would be useful**
- **List any alternative solutions you've considered**

### Pull Requests

1. **Follow the style guidelines**
2. **Write tests for new functionality**
3. **Update documentation as needed**
4. **Ensure all tests pass**
5. **Write a clear commit message**

## Development Setup

### Prerequisites

- Go 1.22 or higher
- CGO support
- Make
- Git

### Setting Up Your Development Environment

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/ml-notes.git
cd ml-notes

# Add upstream remote
git remote add upstream https://github.com/streed/ml-notes.git

# Install dependencies
make deps

# Build the project
make build

# Run tests
make test

# Run with race detector during development
make dev
```

### Recommended Development Workflow

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes and test:**
   ```bash
   make test
   make lint
   ```

3. **Commit your changes:**
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

4. **Push to your fork:**
   ```bash
   git push origin feature/your-feature-name
   ```

5. **Create a Pull Request** on GitHub

## Style Guidelines

### Go Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Use meaningful variable and function names
- Add comments for exported functions
- Keep functions small and focused
- Handle errors appropriately

### Code Organization

```go
// Good: Clear function with error handling
func ProcessNote(note *Note) error {
    if note == nil {
        return fmt.Errorf("note cannot be nil")
    }
    
    // Process the note
    if err := validateNote(note); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    return nil
}
```

### Running Code Quality Checks

```bash
# Format code
make fmt

# Run linters
make lint

# Run tests with coverage
make test-coverage
```

## Commit Messages

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring
- `test:` - Test additions or changes
- `chore:` - Maintenance tasks

### Examples

```
feat: add support for markdown export
fix: resolve panic when database is locked
docs: update installation instructions
refactor: simplify embedding generation logic
test: add tests for vector search
```

## Testing

### Writing Tests

- Write unit tests for new functions
- Include both positive and negative test cases
- Use table-driven tests where appropriate
- Mock external dependencies

### Example Test

```go
func TestProcessNote(t *testing.T) {
    tests := []struct {
        name    string
        note    *Note
        wantErr bool
    }{
        {
            name:    "valid note",
            note:    &Note{Title: "Test", Content: "Content"},
            wantErr: false,
        },
        {
            name:    "nil note",
            note:    nil,
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ProcessNote(tt.note)
            if (err != nil) != tt.wantErr {
                t.Errorf("ProcessNote() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Running Tests

```bash
# Run all tests
make test

# Run tests for specific package
go test ./internal/embeddings/...

# Run with verbose output
go test -v ./...

# Run with race detector
go test -race ./...
```

## Project Structure

```
ml-notes/
├── cmd/              # CLI commands
├── internal/         # Internal packages
│   ├── config/      # Configuration management
│   ├── database/    # Database operations
│   ├── embeddings/  # Embedding generation
│   ├── logger/      # Logging utilities
│   ├── models/      # Data models
│   └── search/      # Search implementation
├── main.go          # Entry point
├── go.mod           # Dependencies
├── Makefile         # Build automation
└── README.md        # Documentation
```

## Getting Help

- Check the [documentation](README.md)
- Look through [existing issues](https://github.com/streed/ml-notes/issues)
- Ask questions in issues with the "question" label
- Join our discussions

## Recognition

Contributors will be recognized in:
- The project README
- Release notes
- GitHub contributors page

Thank you for contributing to ML Notes! 🎉