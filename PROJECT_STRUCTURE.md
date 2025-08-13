# ML Notes Project Structure

## Repository Layout

```
ml-notes/
├── .github/
│   └── workflows/
│       └── ci.yml              # GitHub Actions CI/CD pipeline
├── cmd/                        # CLI Commands
│   ├── add.go                  # Add note command with editor support
│   ├── analyze.go              # AI-powered analysis with custom prompts
│   ├── config.go               # Configuration management
│   ├── delete.go               # Delete notes with safety features
│   ├── detect.go               # Dimension detection
│   ├── edit.go                 # Edit notes with change detection
│   ├── get.go                  # Get note by ID
│   ├── init.go                 # Initialize configuration
│   ├── list.go                 # List notes with pagination
│   ├── mcp.go                  # MCP server for LLM integration
│   ├── reindex.go              # Reindex embeddings
│   ├── root.go                 # Root command and app initialization
│   └── search.go               # Search notes (text and vector)
├── internal/                   # Internal packages (not exposed)
│   ├── config/
│   │   └── config.go           # Configuration structures and management
│   ├── database/
│   │   └── db.go               # SQLite database operations
│   ├── embeddings/
│   │   └── embeddings.go       # Embedding generation (Ollama/fallback)
│   ├── errors/
│   │   └── errors.go           # Custom error types and handling
│   ├── logger/
│   │   └── logger.go           # Debug and logging utilities
│   ├── mcp/
│   │   └── server.go           # Model Context Protocol server
│   ├── models/
│   │   └── note.go             # Note model and repository
│   ├── search/
│   │   └── vector_search.go    # Vector similarity search implementation
│   └── summarize/
│       ├── summarize.go        # AI analysis and summarization
│       └── summarize_test.go   # Comprehensive test suite
├── .gitignore                  # Git ignore patterns
├── docs/                       # Documentation
│   ├── DESIGN.md               # System design and architecture  
│   ├── USAGE_GUIDE.md          # Comprehensive usage documentation
│   ├── delete-command.md       # Delete command documentation
│   └── edit-command.md         # Edit command documentation
├── examples/
│   └── mcp_search_demo.md      # MCP integration examples
├── CHANGELOG.md                # Version history and changes
├── CONTRIBUTING.md             # Contribution guidelines
├── go.mod                      # Go module dependencies
├── go.sum                      # Go module checksums
├── install.sh                  # Installation script
├── LICENSE                     # MIT License
├── main.go                     # Application entry point
├── Makefile                    # Build automation
├── PROJECT_STRUCTURE.md        # This file
├── README.md                   # Main documentation
└── RELEASES.md                 # Release process documentation
```

## Key Components

### CLI Layer (`cmd/`)
- Implements Cobra commands
- Handles user input and output
- Orchestrates business logic

### Business Logic (`internal/`)
- **config**: Application configuration management
- **database**: SQLite and sqlite-vec integration
- **embeddings**: Text-to-vector conversion with Nomic formatting
- **logger**: Structured logging with debug support
- **models**: Data structures and database operations
- **search**: Vector similarity search algorithms

### Build & Deploy
- **Makefile**: Automates building, testing, and installation
- **install.sh**: User-friendly installation script
- **.github/workflows**: Automated CI/CD with GitHub Actions

## Technology Stack

- **Language**: Go 1.22+
- **Database**: SQLite with sqlite-vec extension
- **CLI Framework**: Cobra
- **Embeddings**: Ollama with Nomic models
- **Build**: Make, CGO
- **CI/CD**: GitHub Actions

## Data Flow

1. **User Input** → CLI Command
2. **Command Processing** → Business Logic
3. **Text Processing** → Embedding Generation
4. **Database Operations** → SQLite/sqlite-vec
5. **Search Operations** → Vector Similarity
6. **Results** → Formatted Output

## Configuration

Configuration stored in `~/.config/ml-notes/config.json`:
- Data directory location
- Ollama endpoint
- Embedding model settings
- Vector dimensions
- Debug mode

## Installation Paths

- **Binary**: `/usr/local/bin/ml-notes`
- **Config**: `~/.config/ml-notes/`
- **Data**: `~/.local/share/ml-notes/`

## Development Workflow

1. **Setup**: `make deps`
2. **Build**: `make build`
3. **Test**: `make test`
4. **Install**: `make install`
5. **Release**: `make release`

## Security Considerations

- No hardcoded credentials
- User-specific configuration
- Local data storage
- Optional network access (Ollama)

## Performance Features

- Embedded sqlite-vec for fast vector search
- Lazy loading of embeddings
- Efficient L2 distance calculations
- Pagination support for large datasets

## Extensibility

- Pluggable embedding providers
- Configurable vector dimensions
- Multiple search algorithms
- Extensible CLI commands