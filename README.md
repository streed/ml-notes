# ML Notes

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-blue.svg)](https://golang.org)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)
[![Release](https://github.com/streed/ml-notes/actions/workflows/release.yml/badge.svg)](https://github.com/streed/ml-notes/actions/workflows/release.yml)

A powerful note-taking application with semantic vector search capabilities, powered by SQLite and lil-rag for intelligent search. Available as both a command-line interface (CLI) and a modern desktop application.

## ✨ Features

### 🖥️ **Dual Interface Options**
- 📱 **Desktop Application** - Modern native desktop app built with Wails framework
- 💻 **Command-Line Interface** - Full-featured CLI for terminal enthusiasts and automation

### 📝 **Core Functionality**
- 📝 **Complete Note Management** - Create, edit, delete, and organize notes with powerful tools
- 🌐 **Website Import** - Import web pages as notes with headless browser support and image URL preservation
- 🌐 **Modern Web Interface** - Beautiful, responsive web UI with real-time markdown preview and graph visualization
- 🏷️ **Smart Tagging System** - Organize notes with tags, search by tags, and manage tag collections
- 🔍 **Triple Search Methods** - Semantic vector search, traditional text search, and tag-based search
- 📊 **Interactive Graph Visualization** - Explore relationships between notes with D3.js-powered graph views

### 🚀 **Performance & Integration**
- 🚀 **Fast & Lightweight** - Built with Go and SQLite for maximum performance
- 🔌 **Lil-Rag Integration** - Advanced semantic search with project-aware namespacing
- 📊 **Smart Search Isolation** - Project-scoped search prevents cross-contamination between different note collections
- 🧠 **AI-Powered Analysis** - Deep analysis with custom prompts and reasoning visibility
- ✏️ **Advanced Editor Features** - Split-pane markdown editor with synchronized scrolling and focus-based behavior

### 🛠️ **Developer Features**
- 🛠️ **Highly Configurable** - Customize everything from storage paths to AI models
- 🐛 **Debug Support** - Built-in debugging for troubleshooting and development
- 🤖 **MCP Server** - Model Context Protocol server for LLM integration (Claude Desktop)
- 🔄 **Smart Reindexing** - Automatic vector index management and optimization

## 🚀 Binaries Overview

ML Notes provides two distinct binaries to suit different user preferences and workflows:

### 📟 CLI Binary (`ml-notes-cli`)

The command-line interface provides full access to all ML Notes functionality through terminal commands.

**Key Features:**
- 🖥️ **Pure Terminal Interface** - Works entirely in your terminal without GUI dependencies
- 🤖 **Automation Friendly** - Perfect for scripts, CI/CD pipelines, and automated workflows
- 🌐 **Web Server Mode** - Built-in web server for browser-based note management
- 🔧 **Configuration Management** - Initialize and manage settings via command line
- 📊 **MCP Server** - Run as Model Context Protocol server for LLM integration
- 🚀 **Lightweight** - Minimal resource usage, fast startup

**Use Cases:**
- Daily note-taking via terminal commands
- Automated note processing and analysis
- Server deployments and headless environments
- Integration with text editors and development workflows
- Claude Desktop integration via MCP server

**Example Commands:**
```bash
# Initialize configuration
ml-notes-cli init --interactive

# Add a note
ml-notes-cli add -t "Meeting Notes" -c "Important project updates"

# Search notes
ml-notes-cli search --vector "machine learning concepts"

# Start web interface
ml-notes-cli serve --port 21212

# Run as MCP server
ml-notes-cli mcp
```

### 🖥️ Desktop Application (`ml-notes`)

A modern desktop application built with the Wails framework, providing a native GUI experience.

**Key Features:**
- 🖱️ **Native Desktop UI** - Modern desktop interface with native OS integration
- 📱 **Cross-Platform** - Runs on Linux, macOS, and Windows with native look and feel
- 🎨 **Rich Interface** - Advanced UI components, drag-and-drop, and visual interactions
- 🔄 **Real-Time Updates** - Live note editing and instant search results
- 📊 **Visual Analytics** - Built-in graph visualization and interactive charts
- 💾 **Local Storage** - All data stored locally with offline capability

**Use Cases:**
- Visual note-taking and organization
- Interactive graph exploration and analysis
- Desktop productivity workflows
- Users who prefer GUI over command-line interfaces
- Rich text editing with live preview

**Technical Details:**
- Built with [Wails v2](https://wails.io/) framework
- Go backend with modern web frontend
- Native system integration (notifications, file dialogs, etc.)
- Embedded web server for hybrid functionality
- Shared codebase with CLI for consistent feature parity

### 🔄 Choosing the Right Binary

| Feature | CLI (`ml-notes-cli`) | Desktop (`ml-notes`) |
|---------|---------------------|---------------------|
| **Interface** | Terminal-based | Native desktop GUI |
| **Resource Usage** | Minimal | Moderate |
| **Automation** | Excellent | Limited |
| **Visual Features** | Web-based | Native desktop |
| **Headless Mode** | ✅ Yes | ❌ No |
| **Cross-Platform** | ✅ Yes | ✅ Yes |
| **MCP Server** | ✅ Yes | ❌ No |
| **Web Interface** | ✅ Built-in | ✅ Embedded |

**Recommendation:**
- Use **CLI** for automation, server deployments, terminal workflows, and MCP integration
- Use **Desktop** for visual note-taking, interactive exploration, and GUI-based workflows
- Both binaries can coexist and share the same configuration and data files

## 📋 Table of Contents

- [Binaries Overview](#binaries-overview)
  - [CLI Binary (ml-notes-cli)](#cli-binary-ml-notes-cli)
  - [Desktop Application (ml-notes)](#desktop-application-ml-notes)
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
- CGO support (for SQLite integration)
- (Optional) [Lil-Rag](https://github.com/stillmatic/lil-rag) service for enhanced semantic search
- (Optional) [Ollama](https://ollama.ai) for AI-powered features

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
# Open http://localhost:21212 in your browser
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
| `lilrag-url` | Lil-Rag service endpoint | `http://localhost:12121` |
| `summarization-model` | Model for AI analysis | `llama3.2:latest` |
| `enable-summarization` | Enable/disable analysis features | `true` |
| `enable-auto-tagging` | Enable/disable AI auto-tagging | `true` |
| `max-auto-tags` | Maximum auto-generated tags | `5` |
| `editor` | Default editor for note editing | Auto-detect |
| `debug` | Enable debug logging | `false` |

### Lil-Rag Integration

ML Notes uses [Lil-Rag](https://github.com/stillmatic/lil-rag) for advanced semantic search with project-aware namespacing:

**Project Isolation**: Each project directory gets its own search namespace (e.g., `ml-notes-myproject`), preventing search results from mixing between different projects.

**Setup Lil-Rag Service**:
```bash
# Install and run lil-rag service
go install github.com/stillmatic/lil-rag@latest
lil-rag serve --port 12121

# Configure ML Notes to use lil-rag
ml-notes config set lilrag-url http://localhost:12121
```

**Benefits**:
- **Project Scoping**: Search only within your current project's notes
- **Cross-Project Isolation**: Notes from different projects don't contaminate search results  
- **Automatic Namespacing**: Project directory name automatically determines search scope
- **Fallback Support**: Gracefully falls back to text search when lil-rag is unavailable

### Managing Configuration

```bash
# View current configuration
ml-notes config show

# Update settings
ml-notes config set ollama-endpoint http://localhost:11434
ml-notes config set lilrag-url http://localhost:12121
ml-notes config set summarization-model llama3.2:latest
ml-notes config set enable-auto-tagging true
ml-notes config set max-auto-tags 5
ml-notes config set editor "code --wait"
ml-notes config set debug true
```

## 🌐 Web Interface

ML Notes includes a modern, responsive web interface that provides an intuitive way to manage your notes, visualize relationships, and edit content with a powerful markdown editor.

### Starting the Web Server

```bash
# Start the web server on default port (21212)
ml-notes serve

# Start on specific host and port
ml-notes serve --host 0.0.0.0 --port 3000

# Access the web interface
open http://localhost:21212
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

#### Import Website Content
```bash
# Import a website as a new note
ml-notes import-url https://blog.example.com/article

# Import with custom tags
ml-notes import-url https://docs.example.com/guide --tags "docs,reference,tutorial"

# Import with AI auto-tagging
ml-notes import-url https://example.com/post --auto-tag

# Import with custom timeout for slow-loading sites
ml-notes import-url https://heavy-site.com --timeout 60s
```

**Features:**
- **Headless Browser**: Uses Chrome to render JavaScript and dynamic content
- **Smart Content Extraction**: Prioritizes main content areas (article, main) while filtering out navigation, ads, and sidebars
- **Image URL Preservation**: Converts relative image URLs to absolute URLs while maintaining external/CDN links
- **Markdown Conversion**: Clean HTML-to-markdown conversion with proper formatting
- **Security-First**: Uses secure browser settings with SSL validation for live websites

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

### Advanced Features

#### Project-Scoped Search
ML Notes automatically isolates search results by project directory:
```bash
# In /home/user/project1
ml-notes search --vector "machine learning"  # Searches ml-notes-project1 namespace

# In /home/user/project2  
ml-notes search --vector "machine learning"  # Searches ml-notes-project2 namespace
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
      "command": "ml-notes-cli",
      "args": ["mcp"]
    }
  }
}
```

**Important:** Use `ml-notes-cli` (the CLI binary) for MCP server functionality, not the desktop `ml-notes` binary.

3. Restart Claude Desktop

### Available Tools

The MCP server provides the following tools to LLMs:

#### Note Management (6 tools)
- **add_note** - Create new notes with optional tags
- **get_note** - Retrieve specific notes by ID
- **update_note** - Modify existing notes and tags
- **delete_note** - Remove notes from database
- **list_notes** - List notes with pagination

#### Tag Management (2 tools)
- **list_tags** - View all available tags
- **update_note_tags** - Manage note tags

#### AI-Powered Features (2 tools)
- **suggest_tags** - AI-powered tag suggestions
- **auto_tag_note** - Automatically apply AI-generated tags

#### Enhanced Search Capabilities
- **search_notes** - Enhanced search with multiple modes:
  - **Vector search** - Semantic similarity using lil-rag
  - **Text search** - Traditional keyword search
  - **Tag search** - Filter by comma-separated tags
  - **Auto mode** - Intelligent search type selection
  - Configurable output format (summaries or full content)
  - Smart result limits (max 100, type-specific defaults)

### Resources (6 available)

The MCP server exposes these comprehensive resources:
- **notes://recent** - Most recently created notes with metadata
- **notes://note/{id}** - Individual note access by ID (supports URI templates)
- **notes://tags** - Complete tag listing with creation timestamps
- **notes://stats** - Comprehensive database statistics and metrics
- **notes://config** - System configuration and capability information
- **notes://health** - Service health and availability status monitoring

### Prompts (2 available)

Pre-configured interaction templates:
- **search_notes** - Structured search interactions with flexible parameters
- **summarize_notes** - Generate analysis and summaries of note collections

### Starting the MCP Server

```bash
# Start MCP server (for use with LLM clients)
ml-notes-cli mcp

# The server communicates via stdio for Claude Desktop integration
# Debug mode can be enabled with --debug flag
ml-notes-cli --debug mcp
```

### Recent Enhancements (v1.1.0)

The MCP server has been significantly enhanced with:

- **Enhanced Search**: New `search_type` parameter with auto/vector/text/tags modes
- **Flexible Output**: `show_content` parameter for full content vs. previews
- **Better Validation**: Input validation with enums and parameter constraints
- **Rich Resources**: New URI template support for individual note access
- **Health Monitoring**: Comprehensive health and status checking
- **Improved Logging**: Detailed debug logging for troubleshooting
- **Smart Defaults**: Intelligent parameter defaults based on search type
- **Error Handling**: Enhanced error messages and validation

**Compatibility**: Uses mcp-go v0.39.1 with latest MCP protocol support.

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

**Lil-Rag Connection:**
```bash
# Check lil-rag service is running
curl http://localhost:12121/health

# Update endpoint if needed
ml-notes config set lilrag-url http://your-lilrag:12121

# Service falls back to text search if lil-rag unavailable
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
# Enable debug mode to see namespace usage
ml-notes --debug search --vector "test"
ml-notes --debug <command>
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Lil-Rag](https://github.com/stillmatic/lil-rag) - Advanced semantic search service
- [Ollama](https://ollama.ai) - Local LLM inference
- [SQLite](https://sqlite.org) - Reliable embedded database
- [Cobra](https://github.com/spf13/cobra) - CLI framework

## 📮 Support

- 🐛 [Report bugs](https://github.com/streed/ml-notes/issues)
- 💡 [Request features](https://github.com/streed/ml-notes/issues)
- 📖 [Documentation](https://github.com/streed/ml-notes/wiki)

---

Made with ❤️ by the ML Notes community
