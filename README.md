# ML Notes

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

A powerful command-line note-taking application with semantic vector search capabilities, powered by SQLite and sqlite-vec.

## ✨ Features

- 📝 **Complete Note Management** - Create, edit, delete, and organize notes with powerful CLI tools
- 🌐 **Modern Web Interface** - Beautiful, responsive web UI with real-time markdown preview and graph visualization
- 🏷️ **Smart Tagging System** - Organize notes with tags, search by tags, and manage tag collections
- 🔍 **Triple Search Methods** - Semantic vector search, traditional text search, and tag-based search
- 📊 **Interactive Graph Visualization** - Explore relationships between notes with D3.js-powered graph views
- 🚀 **Fast & Lightweight** - Built with Go and SQLite for maximum performance
- 🔌 **Ollama Integration** - Use local LLMs for generating embeddings and analysis
- 📊 **Vector Database** - Built-in sqlite-vec for efficient similarity search
- 🧠 **AI-Powered Analysis** - Deep analysis with custom prompts and reasoning visibility
- ✏️ **Advanced Editor Features** - Split-pane markdown editor with synchronized scrolling and focus-based behavior
- 🛠️ **Highly Configurable** - Customize everything from storage paths to AI models
- 🐛 **Debug Support** - Built-in debugging for troubleshooting and development
- 🤖 **MCP Server** - Model Context Protocol server for LLM integration (Claude Desktop)
- 🔄 **Smart Reindexing** - Automatic vector index management and optimization

## 📋 Table of Contents

- [Installation](#installation)
  - [From Source](#from-source)
  - [Pre-built Binaries](#pre-built-binaries)
  - [Using Make](#using-make)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Web Interface](#web-interface)
  - [Starting the Web Server](#starting-the-web-server)
  - [Web UI Features](#web-ui-features)
  - [Graph Visualization](#graph-visualization)
- [CLI Usage](#cli-usage)
  - [Managing Notes](#managing-notes)
  - [Tag Management](#tag-management)
  - [Searching](#searching)
  - [Updating](#updating)
  - [AI-Powered Analysis](#ai-powered-analysis)
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

**Updating**: Once installed, you can update to the latest version with:
```bash
ml-notes update
```

### Using Make

```bash
# Build for current platform
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

### Cross-Platform Builds

ML Notes supports building for multiple platforms:

```bash
# Build for all platforms (Linux, macOS, Windows)
make build-all

# Build for specific platforms
make build-linux    # Linux AMD64
make build-darwin   # macOS (Intel & Apple Silicon)
make build-windows  # Windows AMD64

# Create release packages
make release
```

**Platform Support:**
- ✅ **Linux AMD64** - Full native support
- ✅ **macOS Intel (AMD64)** - Native and cross-compilation support
- ✅ **macOS Apple Silicon (ARM64)** - Native and cross-compilation support  
- ✅ **Windows AMD64** - Native and cross-compilation support

**Cross-Compilation Notes:**
- Cross-compilation requires appropriate toolchains (clang for macOS, mingw for Windows)
- For best results, build natively on target platforms
- CGO is required for sqlite-vec support
- Use `make build-native` for automatic platform detection

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

# Or add a note with tags
ml-notes add -t "Project Ideas" -c "Some great ideas for the next project" --tags "ideas,projects,todo"
```

3. **List your notes:**
```bash
ml-notes list
```

4. **Search with vector similarity:**
```bash
ml-notes search --vector "machine learning concepts"
```

5. **Search by tags:**
```bash
ml-notes search --tags "projects,ideas"
```

6. **Start the web interface:**
```bash
ml-notes serve
# Open http://localhost:8080 in your browser
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
| `data-dir` | Where notes database is stored | `~/.local/share/ml-notes` |
| `ollama-endpoint` | Ollama API endpoint | `http://localhost:11434` |
| `embedding-model` | Model for embeddings | `nomic-embed-text:v1.5` |
| `vector-dimensions` | Embedding vector size | Auto-detected |
| `enable-vector` | Enable/disable vector search | `true` |
| `summarization-model` | Model for AI analysis | `llama3.2:latest` |
| `enable-summarization` | Enable/disable analysis features | `true` |
| `editor` | Default editor for note editing | Auto-detect |
| `debug` | Enable debug logging | `false` |

### Managing Configuration

```bash
# View current configuration
ml-notes config show

# Update settings
ml-notes config set ollama-endpoint http://localhost:11434
ml-notes config set embedding-model nomic-embed-text:v1.5
ml-notes config set editor "code --wait"
ml-notes config set debug true

# Detect model dimensions
ml-notes detect-dimensions
```

## 🌐 Web Interface

ML Notes includes a modern, responsive web interface that provides an intuitive way to manage your notes, visualize relationships, and edit content with a powerful markdown editor.

### Starting the Web Server

```bash
# Start the web server on default port (8080)
ml-notes serve

# Start on specific host and port
ml-notes serve --host 0.0.0.0 --port 3000

# Access the web interface
open http://localhost:8080
```

### Web UI Features

#### 📝 **Smart Markdown Editor**
- **Focus-Based Split Pane**: Editor automatically appears when you focus on editing areas
- **Real-Time Preview**: Live markdown rendering with synchronized scrolling
- **Auto-Scroll**: Preview automatically follows your cursor when typing extends the editor
- **Smooth Transitions**: Elegant animations for pane resizing and focus changes

#### 🏷️ **Tag Management**
- Visual tag display with removal capabilities
- Tag input with comma-separated support
- Filter notes by tags using the sidebar dropdown
- Auto-tagging integration with AI-powered suggestions

#### 🔍 **Integrated Search**
- Real-time search as you type
- Vector search integration for semantic similarity
- Tag-based filtering
- Search results with content previews

#### 📊 **Note Organization**
- Sidebar with all notes and metadata
- Chronological organization with creation dates
- Active note highlighting
- Quick navigation between notes

#### 🎨 **Theme Support**
- Light and dark theme toggle
- Consistent theming across all components
- Paper-like texture for comfortable reading
- Responsive design for all screen sizes

#### 🤖 **AI Integration**
- One-click auto-tagging for notes
- AI-powered note analysis with custom prompts
- Integration with Ollama for local AI processing
- Analysis results with thinking process visibility

### Graph Visualization

#### 📊 **Interactive Note Graph**
- **D3.js-Powered Visualization**: Smooth, interactive graph showing note relationships
- **Smart Node Positioning**: Isolated notes stay near center, connected notes spread outward
- **Connection-Based Sizing**: Node size reflects number of connections to other notes
- **Tag-Based Relationships**: Notes connected by shared tags with weighted connections
- **Color-Coded Groups**: Different colors for different note clusters/topics

#### 🎮 **Interactive Features**
- **Zoom and Pan**: Scroll to zoom, drag to pan around the graph
- **Node Interaction**: Click nodes to navigate directly to notes
- **Hover Tooltips**: See note titles, tags, and connection counts
- **Drag Nodes**: Manually position nodes by dragging
- **Filter Controls**: Filter by tag, minimum connections, and maximum nodes displayed

#### 🎯 **Graph Controls**
- **Filter Panel**: 
  - Filter by specific tags
  - Set minimum connection threshold
  - Limit maximum nodes displayed
- **Zoom Controls**: Zoom in/out buttons and fit-to-view
- **View Options**: Toggle between full graph view and embedded preview
- **Reset View**: Return to optimal view with one click

#### 📱 **Responsive Design**
- **Desktop Experience**: Full-featured graph with all controls and interactions
- **Mobile Friendly**: Simplified interface optimized for touch devices
- **Embedded Mode**: Compact graph view for the main dashboard
- **Performance Optimized**: Efficient rendering for large note collections

#### 🔗 **Graph Navigation**
- **Direct Navigation**: Click any node to jump to that note
- **Context Preservation**: Graph remembers position when returning
- **Visual Feedback**: Selected and hovered nodes are highlighted
- **Breadcrumb Integration**: Easy return to main interface

### Web Interface Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl/Cmd + S` | Save current note |
| `Ctrl/Cmd + N` | Create new note |
| `Ctrl/Cmd + /` | Toggle theme |
| `Escape` | Close modals |

### Web Server Configuration

The web server supports additional configuration options:

```bash
# Enable custom CSS
ml-notes config set webui-custom-css true

# Set default theme
ml-notes config set webui-theme dark
```

### API Endpoints

The web server also exposes REST API endpoints:

- `GET /api/v1/notes` - List all notes
- `POST /api/v1/notes` - Create new note
- `GET /api/v1/notes/{id}` - Get specific note
- `PUT /api/v1/notes/{id}` - Update note
- `DELETE /api/v1/notes/{id}` - Delete note
- `POST /api/v1/notes/search` - Search notes
- `GET /api/v1/tags` - List all tags
- `GET /api/v1/graph` - Get graph data for visualization
- `POST /api/v1/auto-tag/suggest/{id}` - Get AI tag suggestions

## 📚 CLI Usage

### Managing Notes

#### Add a Note
```bash
# Interactive mode (opens your editor)
ml-notes add -t "Title"

# With content directly
ml-notes add -t "Title" -c "Content"

# With tags
ml-notes add -t "Project Notes" -c "Important project details" --tags "project,important,todo"

# From stdin
echo "Content" | ml-notes add -t "Title"

# Use specific editor
ml-notes add -t "Code Review" --editor-cmd "code --wait"
```

#### List Notes
```bash
# List recent notes
ml-notes list

# With pagination
ml-notes list --limit 10 --offset 20

# Short format (ID and title only)
ml-notes list --short
```

#### Get a Note
```bash
ml-notes get <note-id>
```

#### Edit a Note
```bash
# Edit note in your default editor (includes tags)
ml-notes edit 123

# Edit title only
ml-notes edit -t 123

# Edit content only
ml-notes edit -c 123

# Use a specific editor
ml-notes edit -e "code --wait" 123
```

**Note**: When editing in your editor, the note format includes tags:
```
Title: Your Note Title
Tags: tag1, tag2, tag3
---
Your note content goes here
```

#### Delete Notes
```bash
# Delete a single note
ml-notes delete 123

# Delete multiple notes
ml-notes delete 123 456 789

# Skip confirmation prompt
ml-notes delete -f 123

# Delete all notes (use with extreme caution!)
ml-notes delete --all
```

### Tag Management

#### Managing Tags
```bash
# List all tags in the system
ml-notes tags list

# Add tags to an existing note
ml-notes tags add 123 --tags "urgent,important"

# Remove specific tags from a note
ml-notes tags remove 123 --tags "old,outdated"

# Replace all tags for a note
ml-notes tags set 123 --tags "research,ai,final"

# Remove all tags from a note
ml-notes tags set 123 --tags ""
```

#### Tag Search
```bash
# Search for notes with specific tags
ml-notes search --tags "project,important"

# Search for notes with any of the specified tags
ml-notes search --tags "research,ai,ml"

# Combine with other search options
ml-notes search --tags "project" --limit 5 --short
```

### Searching

#### Text Search
```bash
# Search in titles and content
ml-notes search "golang"

# Limit results
ml-notes search --limit 5 "machine learning"

# Short format showing only IDs and titles
ml-notes search --short "programming"
```

#### Vector Search
```bash
# Semantic similarity search (returns most similar note by default)
ml-notes search --vector "neural networks and deep learning"

# Get top 5 most similar notes
ml-notes search --vector --limit 5 "machine learning concepts"

# Finds related notes even without exact matches
ml-notes search --vector "AI concepts"
```

#### Tag Search
```bash
# Search by single tag
ml-notes search --tags "research"

# Search by multiple tags (finds notes with ANY of these tags)
ml-notes search --tags "project,important,urgent"

# Combine tag search with limits and formatting
ml-notes search --tags "ai,ml" --limit 10 --short

# Tag-only search (no text query needed)
ml-notes search --tags "todo"
```

### 🔄 Updating

Keep ml-notes up to date with the built-in update system that automatically downloads and installs the latest releases from GitHub.

#### Basic Updates
```bash
# Update to the latest stable release
ml-notes update

# Check for updates without installing (dry run)
ml-notes update --dry-run

# Force update even if already on latest version
ml-notes update --force
```

#### Version-Specific Updates
```bash
# Update to a specific version
ml-notes update --version v1.2.3

# Include pre-release versions
ml-notes update --prerelease

# Update to latest pre-release
ml-notes update --prerelease --force
```

#### Update Process
The updater performs these steps safely:
1. **Version Check** - Compares current version with GitHub releases
2. **Platform Detection** - Automatically selects the right binary for your OS/architecture
3. **Download** - Downloads the update with progress indication
4. **Verification** - Verifies download integrity (when checksums available)
5. **Backup** - Creates a backup of your current binary
6. **Installation** - Atomically replaces the binary
7. **Release Notes** - Shows what's new in the updated version

#### Update Features
- 🔒 **Safe Updates**: Atomic replacement with automatic backup
- 📊 **Progress Display**: Real-time download progress with emojis
- 🎯 **Platform Detection**: Automatically selects correct binary for your system
- 📋 **Release Notes**: Shows changes and new features after update
- 🔍 **Dry Run Mode**: Preview updates without installing
- ⚡ **Force Updates**: Re-install current version if needed
- 🚀 **Pre-release Support**: Access beta versions and early features

#### Troubleshooting Updates
```bash
# Check current version
ml-notes --version

# Force re-download if update failed
ml-notes update --force

# Verify update worked
ml-notes --version
```

**Note**: Updates require internet connectivity and sufficient disk space. The original binary is preserved as a backup during the update process.

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

### AI-Powered Analysis

#### Basic Analysis
```bash
# Analyze a single note
ml-notes analyze 123

# Analyze multiple notes together
ml-notes analyze 1 2 3

# Analyze all notes
ml-notes analyze --all

# Analyze recent notes
ml-notes analyze --recent 10
```

#### Custom Analysis Prompts
```bash
# Focus on technical aspects
ml-notes analyze 123 -p "Focus on technical implementation details"

# Analyze all notes with a specific tag
ml-notes search --tags "research" --analyze -p "Summarize research findings"

# Extract insights and patterns
ml-notes analyze --all -p "What are the recurring themes and patterns?"

# Comparative analysis
ml-notes analyze 1 2 3 -p "Compare and contrast these approaches"

# Business-focused analysis
ml-notes analyze --recent 5 -p "What business opportunities are mentioned?"
```

#### Search with Analysis
```bash
# Analyze search results
ml-notes search --analyze "machine learning"

# Custom analysis of search results
ml-notes search --analyze -p "Focus on practical applications" "algorithms"

# Show both analysis and detailed results
ml-notes search --analyze --show-details "python"
```

#### Analysis Features
- **Reasoning Visibility**: See the AI's thought process with formatted thinking sections
- **Custom Prompts**: Target analysis with specific questions or focus areas  
- **Multi-Note Analysis**: Analyze relationships and patterns across multiple notes
- **Search Integration**: Analyze search results for deeper insights

#### Configuration
```bash
# Enable/disable analysis features
ml-notes config set enable-summarization true

# Set the analysis model
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
- **add_note** - Create a new note with title, content, and optional tags
- **get_note** - Retrieve a specific note by ID (includes tags)
- **update_note** - Modify existing note title, content, or tags
- **delete_note** - Remove a note from the database
- **list_notes** - List notes with pagination support (shows tags)

#### Tag Management
- **list_tags** - List all tags in the system
- **update_note_tags** - Update tags for a specific note

#### Search Capabilities
- **search_notes** - Search using vector similarity, text matching, or tag search
  - Supports semantic vector search, keyword search, and tag-based search
  - Configurable result limits
  - Automatically uses best search method
  - Tag search with comma-separated tag lists

### Resources

The MCP server exposes these resources:
- `notes://recent` - Get the most recently created notes
- `notes://stats` - Database statistics and configuration
- `notes://config` - Current ML Notes configuration
- `notes://note/{id}` - Access specific note by ID

### Prompts

Pre-configured prompts for common tasks:
- **search_notes** - Structured search prompt with query parameters
- **analyze_notes** - Generate AI-powered analysis of your note collection

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
│   ├── serve.go     # Web server command
│   ├── init.go      # Init configuration
│   ├── config.go    # Config management
│   ├── reindex.go   # Reindex embeddings
│   ├── detect.go    # Detect dimensions
│   └── mcp.go       # MCP server command
├── internal/         # Internal packages
│   ├── api/         # Web server and API endpoints
│   ├── autotag/     # AI-powered auto-tagging
│   ├── config/      # Configuration management
│   ├── database/    # Database operations
│   ├── embeddings/  # Embedding generation
│   ├── logger/      # Logging utilities
│   ├── mcp/         # MCP server implementation
│   ├── models/      # Data models
│   ├── search/      # Search implementation
│   └── summarize/   # AI analysis features
├── web/              # Web interface assets
│   ├── templates/   # HTML templates
│   │   ├── index.html    # Main web interface
│   │   └── graph.html    # Graph visualization page
│   └── static/      # Static web assets
│       ├── css/     # Stylesheets
│       │   ├── styles.css    # Main styles
│       │   └── themes.css    # Theme definitions
│       └── js/      # JavaScript
│           └── app.js        # Main web application
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