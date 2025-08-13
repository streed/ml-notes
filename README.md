# ML Notes

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

A powerful command-line note-taking application with semantic vector search capabilities, powered by SQLite and sqlite-vec.

## ✨ Features

- 📝 **Simple Note Management** - Create, list, and retrieve notes with an intuitive CLI
- 🔍 **Semantic Search** - Find notes using AI-powered vector similarity search
- 🚀 **Fast & Lightweight** - Built with Go and SQLite for maximum performance
- 🔌 **Ollama Integration** - Use local LLMs for generating embeddings
- 📊 **Vector Database** - Built-in sqlite-vec for efficient similarity search
- 🛠️ **Configurable** - Customize storage paths, models, and embedding dimensions
- 🐛 **Debug Mode** - Built-in debugging for troubleshooting configuration issues

## 📋 Table of Contents

- [Installation](#installation)
  - [From Source](#from-source)
  - [Pre-built Binaries](#pre-built-binaries)
  - [Using Make](#using-make)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Usage](#usage)
  - [Managing Notes](#managing-notes)
  - [Searching](#searching)
  - [Vector Search](#vector-search)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

## 🚀 Installation

### Prerequisites

- Go 1.22 or higher
- CGO support (for sqlite-vec)
- (Optional) [Ollama](https://ollama.ai) for enhanced embeddings

### From Source

```bash
# Clone the repository
git clone https://github.com/streed/ml-notes.git
cd ml-notes

# Build the application
make build

# Install to your PATH
sudo make install
```

### Pre-built Binaries

Download the latest release from the [Releases](https://github.com/streed/ml-notes/releases) page.

```bash
# Download and extract (Linux/macOS example)
wget https://github.com/streed/ml-notes/releases/latest/download/ml-notes-linux-amd64.tar.gz
tar -xzf ml-notes-linux-amd64.tar.gz
sudo mv ml-notes /usr/local/bin/
```

### Using Make

```bash
# Build only
make build

# Install to /usr/local/bin (requires sudo)
make install

# Build and install in one step
make all

# Clean build artifacts
make clean

# Run tests
make test

# Development build with race detector
make dev
```

## 🎯 Quick Start

1. **Initialize configuration:**
```bash
ml-notes init
```

2. **Add your first note:**
```bash
ml-notes add -t "My First Note" -c "This is the content of my first note"
```

3. **List your notes:**
```bash
ml-notes list
```

4. **Search with vector similarity:**
```bash
ml-notes search --vector "machine learning concepts"
```

## ⚙️ Configuration

ML Notes stores its configuration in `~/.config/ml-notes/config.json`.

### Initial Setup

Run the interactive setup:
```bash
ml-notes init -i
```

Or configure with flags:
```bash
ml-notes init \
  --data-dir ~/.local/share/ml-notes \
  --ollama-endpoint http://localhost:11434
```

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `data_directory` | Where notes database is stored | `~/.local/share/ml-notes` |
| `ollama_endpoint` | Ollama API endpoint | `http://localhost:11434` |
| `embedding_model` | Model for embeddings | `nomic-embed-text:v1.5` |
| `vector_dimensions` | Embedding vector size | Auto-detected |
| `enable_vector_search` | Enable/disable vector search | `true` |
| `debug` | Enable debug logging | `false` |

### Managing Configuration

```bash
# View current configuration
ml-notes config show

# Update settings
ml-notes config set ollama-endpoint http://192.168.1.100:11434
ml-notes config set embedding-model nomic-embed-text:v1.5
ml-notes config set debug true

# Detect model dimensions
ml-notes detect-dimensions
```

## 📚 Usage

### Managing Notes

#### Add a Note
```bash
# Interactive mode
ml-notes add -t "Title"

# With content
ml-notes add -t "Title" -c "Content"

# From stdin
echo "Content" | ml-notes add -t "Title"
```

#### List Notes
```bash
# List recent notes
ml-notes list

# With pagination
ml-notes list --limit 10 --offset 20

# Short format
ml-notes list --short
```

#### Get a Note
```bash
ml-notes get <note-id>
```

### Searching

#### Text Search
```bash
# Search in titles and content
ml-notes search "golang"

# Limit results
ml-notes search --limit 5 "machine learning"
```

#### Vector Search
```bash
# Semantic similarity search
ml-notes search --vector "neural networks and deep learning"

# Finds related notes even without exact matches
ml-notes search --vector "AI concepts"
```

### Advanced Features

#### Reindexing
After changing the embedding model or dimensions:
```bash
ml-notes reindex
```

#### Debug Mode
```bash
# Enable for current command
ml-notes --debug search --vector "test"

# Enable persistently
ml-notes config set debug true
```

## 🔧 Development

### Project Structure
```
ml-notes/
├── cmd/              # CLI commands
│   ├── root.go      # Root command
│   ├── add.go       # Add note command
│   ├── list.go      # List notes command
│   ├── get.go       # Get note command
│   ├── search.go    # Search command
│   ├── init.go      # Init configuration
│   ├── config.go    # Config management
│   ├── reindex.go   # Reindex embeddings
│   └── detect.go    # Detect dimensions
├── internal/         # Internal packages
│   ├── config/      # Configuration management
│   ├── database/    # Database operations
│   ├── embeddings/  # Embedding generation
│   ├── logger/      # Logging utilities
│   ├── models/      # Data models
│   └── search/      # Search implementation
├── main.go          # Entry point
├── go.mod           # Go modules
├── Makefile         # Build automation
└── README.md        # Documentation
```

### Building from Source

```bash
# Get dependencies
go mod download

# Build with CGO
CGO_ENABLED=1 go build -o ml-notes

# Run tests
go test ./...

# Run with race detector
go build -race -o ml-notes-dev
```

### Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package tests
go test ./internal/embeddings/...
```

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 🐛 Troubleshooting

### Common Issues

**Dimension Mismatch:**
```bash
# Auto-detect and fix dimensions
ml-notes detect-dimensions
ml-notes reindex
```

**Ollama Connection:**
```bash
# Check Ollama is running
curl http://localhost:11434/api/tags

# Update endpoint if needed
ml-notes config set ollama-endpoint http://your-ollama:11434
```

**Debug Information:**
```bash
# Enable debug mode
ml-notes --debug <command>
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [sqlite-vec](https://github.com/asg017/sqlite-vec) - Vector search SQLite extension
- [Ollama](https://ollama.ai) - Local LLM inference
- [Nomic AI](https://nomic.ai) - Embedding models
- [Cobra](https://github.com/spf13/cobra) - CLI framework

## 📮 Support

- 🐛 [Report bugs](https://github.com/streed/ml-notes/issues)
- 💡 [Request features](https://github.com/streed/ml-notes/issues)
- 📖 [Documentation](https://github.com/streed/ml-notes/wiki)

---

Made with ❤️ by the ML Notes community