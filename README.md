# ML Notes

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

A powerful command-line note-taking application with semantic vector search capabilities, powered by SQLite and sqlite-vec.

## ✨ Features

- 📝 **Simple Note Management** - Create, list, and retrieve notes with an intuitive CLI
- 🔍 **Semantic Search** - Find notes using AI-powered vector similarity search
- 🚀 **Fast & Lightweight** - Built with Go and SQLite for maximum performance
- 🔌 **Ollama Integration** - Use local LLMs for generating embeddings and summaries
- 📊 **Vector Database** - Built-in sqlite-vec for efficient similarity search
- 🛠️ **Configurable** - Customize storage paths, models, and embedding dimensions
- 🐛 **Debug Mode** - Built-in debugging for troubleshooting configuration issues
- 🤖 **MCP Server** - Model Context Protocol server for LLM integration
- ✨ **AI Summarization** - Generate intelligent summaries of notes and search results

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
- [MCP Server](#mcp-server)
  - [Claude Desktop Integration](#claude-desktop-integration)
  - [Available Tools](#available-tools)
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
# Interactive setup (recommended for first-time users)
ml-notes init --interactive

# Or quick setup with defaults
ml-notes init

# Or custom setup with flags
ml-notes init \
  --data-dir ~/my-notes \
  --ollama-endpoint http://localhost:11434 \
  --summarization-model llama3.2:latest \
  --enable-summarization
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
  --ollama-endpoint http://localhost:11434 \
  --summarization-model llama3.2:latest \
  --enable-summarization
```

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `data_directory` | Where notes database is stored | `~/.local/share/ml-notes` |
| `ollama_endpoint` | Ollama API endpoint | `http://localhost:11434` |
| `embedding_model` | Model for embeddings | `nomic-embed-text:v1.5` |
| `vector_dimensions` | Embedding vector size | Auto-detected |
| `enable_vector_search` | Enable/disable vector search | `true` |
| `summarization_model` | Model for AI summarization | `llama3.2:latest` |
| `enable_summarization` | Enable/disable summarization | `true` |
| `debug` | Enable debug logging | `false` |

### Managing Configuration

```bash
# View current configuration
ml-notes config show

# Update settings
ml-notes config set ollama-endpoint http://localhost:11434
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

### AI-Powered Summarization

#### Summarize Individual Notes
```bash
# Get a note with its summary
ml-notes get 123 --summarize

# Summarize a specific note
ml-notes summarize 123
```

#### Summarize Search Results
```bash
# Search with automatic summarization
ml-notes search "machine learning" --summarize

# Vector search with summary
ml-notes search --vector "deep learning concepts" --summarize
```

#### Bulk Summarization
```bash
# Summarize multiple notes
ml-notes summarize 1 2 3 4 5

# Summarize recent notes
ml-notes summarize --recent 10

# Summarize all notes
ml-notes summarize --all

# Use a specific model for summarization
ml-notes summarize --model llama3.2:latest --recent 5
```

#### Configuration
```bash
# Enable/disable summarization
ml-notes config set enable-summarization true

# Set the summarization model
ml-notes config set summarization-model llama3.2:latest

# View current settings
ml-notes config show
```

## 🤖 MCP Server

ML Notes includes a Model Context Protocol (MCP) server that allows LLMs like Claude to interact with your notes programmatically.

### Claude Desktop Integration

Add ML Notes to your Claude Desktop configuration:

1. Open your Claude Desktop config file:
   - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - Windows: `%APPDATA%\Claude\claude_desktop_config.json`
   - Linux: `~/.config/claude/claude_desktop_config.json`

2. Add the ML Notes MCP server:
```json
{
  "mcpServers": {
    "ml-notes": {
      "command": "ml-notes",
      "args": ["mcp"]
    }
  }
}
```

3. Restart Claude Desktop

### Available Tools

The MCP server provides the following tools to LLMs:

#### Note Management
- **add_note** - Create a new note with title and content
- **get_note** - Retrieve a specific note by ID
- **update_note** - Modify existing note title or content
- **delete_note** - Remove a note from the database
- **list_notes** - List notes with pagination support

#### Search Capabilities
- **search_notes** - Search using vector similarity or text matching
  - Supports both semantic vector search and keyword search
  - Configurable result limits
  - Automatically uses best search method

### Resources

The MCP server exposes these resources:
- `notes://recent` - Get the most recently created notes
- `notes://stats` - Database statistics and configuration
- `notes://config` - Current ML Notes configuration
- `notes://note/{id}` - Access specific note by ID

### Prompts

Pre-configured prompts for common tasks:
- **search_notes** - Structured search prompt with query parameters
- **summarize_notes** - Generate summaries of your note collection

### Starting the MCP Server

```bash
# Start MCP server (for use with LLM clients)
ml-notes mcp

# The server communicates via stdio for Claude Desktop integration
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
│   ├── detect.go    # Detect dimensions
│   └── mcp.go       # MCP server command
├── internal/         # Internal packages
│   ├── config/      # Configuration management
│   ├── database/    # Database operations
│   ├── embeddings/  # Embedding generation
│   ├── logger/      # Logging utilities
│   ├── mcp/         # MCP server implementation
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