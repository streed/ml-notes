# ML Notes System Design

## Architecture Overview

ML Notes is a command-line interface for intelligent note management with AI-powered search and analysis capabilities. The system combines traditional text search with semantic search via lil-rag and LLM-based analysis.

## Core Components

### 1. CLI Layer (`cmd/`)

The command-line interface built using the Cobra framework provides the user-facing API:

```
cmd/
├── root.go          # Application initialization and global config
├── add.go           # Note creation with editor integration
├── analyze.go       # AI-powered analysis with custom prompts
├── config.go        # Configuration management
├── delete.go        # Note deletion with safety features
├── edit.go          # Note editing with change detection
├── get.go           # Note retrieval
├── init.go          # Initial setup and configuration
├── list.go          # Note listing with pagination
├── mcp.go           # MCP server for LLM integration
└── search.go        # Unified search interface
```

**Key Design Principles:**
- **Single Responsibility**: Each command handles one primary function
- **Error Handling**: Comprehensive error messages and graceful degradation
- **User Experience**: Intuitive flags and helpful documentation
- **Extensibility**: Easy to add new commands and features

### 2. Business Logic (`internal/`)

The internal packages implement core functionality:

```
internal/
├── config/          # Configuration management
├── database/        # Data persistence layer
├── errors/          # Custom error types
├── lilrag/          # Lil-rag client integration
├── logger/          # Structured logging
├── mcp/            # Model Context Protocol server
├── models/          # Data models and repository pattern
├── search/          # Lil-rag semantic search
└── summarize/       # AI analysis and summarization
```

#### Configuration System (`internal/config/`)

Manages application configuration with JSON persistence:

```go
type Config struct {
    DatabasePath        string `json:"database_path"`
    DataDirectory       string `json:"data_directory"`
    OllamaEndpoint      string `json:"ollama_endpoint"`
    LilRagURL          string `json:"lilrag_url"`
    Debug               bool   `json:"debug"`
    SummarizationModel  string `json:"summarization_model"`
    EnableSummarization bool   `json:"enable_summarization"`
    EnableAutoTagging   bool   `json:"enable_auto_tagging"`
    MaxAutoTags         int    `json:"max_auto_tags"`
    Editor              string `json:"editor"`
}
```

**Features:**
- XDG Base Directory compliance
- Automatic defaults and validation
- Version tracking for migrations
- Environment variable overrides

#### Database Layer (`internal/database/` & `internal/models/`)

SQLite-based persistence with tag support:

```sql
-- Core notes table
CREATE TABLE notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Tags table for note organization
CREATE TABLE tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    note_id INTEGER NOT NULL,
    tag TEXT NOT NULL,
    FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE
);
```

**Repository Pattern:**
- Abstracted database operations
- Transaction support for consistency
- Automatic timestamp management
- Cascading delete operations

#### Lil-Rag Integration (`internal/search/` & `internal/lilrag/`)

Semantic search implementation using lil-rag service:

```go
type LilRagSearch struct {
    client   *lilrag.Client
    noteRepo *models.NoteRepository
}

func (lrs *LilRagSearch) SearchSimilarWithNamespace(query string, namespace string, limit int) ([]*models.Note, error) {
    // 1. Create namespace-aware search request
    mlNamespace := lrs.createNamespace(namespace)
    
    // 2. Perform semantic search via lil-rag
    results, err := lrs.client.SearchWithNamespace(query, mlNamespace, limit)
    
    // 3. Convert to notes
    return lrs.convertResultsToNotes(results)
}
```

**Features:**
- Semantic similarity via lil-rag service
- Project-based namespacing for search isolation
- Automatic "ml-notes" prefix to prevent collisions
- Integration with traditional text search

#### AI Analysis (`internal/summarize/`)

LLM-powered analysis using Ollama:

```go
type Summarizer struct {
    cfg         *config.Config
    model       string
    maxTokens   int
    temperature float32
}

func (s *Summarizer) SummarizeNoteWithPrompt(note *models.Note, customPrompt string) (*SummaryResult, error) {
    // Custom prompt handling for targeted analysis
    if customPrompt != "" {
        prompt = fmt.Sprintf(`Please analyze the following note with this specific focus: %s
        
%s

Analysis:`, customPrompt, content)
    }
    
    return s.callOllama(prompt)
}
```

**Features:**
- Custom prompt support
- Thinking tag formatting (`<think>...</think>`)
- Multi-note analysis
- Search result summarization
- Model switching and configuration

#### Lil-Rag Client (`internal/lilrag/`)

JSON-RPC client for lil-rag service integration:

```go
type Client struct {
    baseURL string
    client  *http.Client
}

func (c *Client) IndexDocumentWithNamespace(docID string, content, namespace string) error {
    req := IndexRequest{
        DocID:     docID,
        Content:   content,
        Namespace: namespace,
    }
    
    return c.makeRequest("index_document", req, nil)
}
```

**Features:**
- JSON-RPC communication with lil-rag
- Namespace-aware document indexing
- Concurrent search operations
- Connection pooling and error handling

## Data Flow

### Note Creation Flow

```
User Input → CLI Validation → Editor Integration → Content Processing → Database Storage → Lil-rag Indexing (with namespace)
```

1. **User Input**: Title and optional content via CLI or editor
2. **CLI Validation**: Input sanitization and validation
3. **Editor Integration**: Optional editor workflow with templates
4. **Content Processing**: Text formatting and preparation
5. **Database Storage**: Transactional note creation
6. **Vector Indexing**: Async embedding generation and storage

### Search Flow

```
Query → Search Strategy → Text/Vector Search → Result Ranking → Optional Analysis → Output Formatting
```

1. **Query Processing**: Parse search terms and flags
2. **Search Strategy**: Choose text vs vector search
3. **Execution**: Perform database queries
4. **Ranking**: Sort by relevance (distance/score)
5. **Analysis**: Optional AI-powered analysis
6. **Output**: Formatted results with metadata

### Analysis Flow

```
Note Selection → Content Preparation → Prompt Generation → LLM Processing → Response Formatting → Output
```

1. **Note Selection**: Single note, multiple notes, or search results
2. **Content Preparation**: Combine and format note content
3. **Prompt Generation**: Build analysis prompt (default or custom)
4. **LLM Processing**: Send to Ollama with parameters
5. **Response Formatting**: Process thinking tags and format output
6. **Output**: Structured analysis with reasoning

## Integration Architecture

### MCP Server Integration

Model Context Protocol server for LLM integration:

```go
type MCPServer struct {
    noteRepo     *models.NoteRepository
    vectorSearch *search.VectorSearch
    config       *config.Config
}

// Available tools for LLMs
var tools = []mcp.Tool{
    {Name: "add_note", Description: "Create a new note"},
    {Name: "search_notes", Description: "Search notes with text or vector similarity"},
    {Name: "get_note", Description: "Retrieve specific note by ID"},
    {Name: "list_notes", Description: "List notes with pagination"},
    {Name: "update_note", Description: "Modify existing note"},
    {Name: "delete_note", Description: "Remove note from database"},
}
```

### Editor Integration

Multi-editor support with intelligent detection:

```go
func openEditor(filename string) error {
    // Priority order:
    // 1. --editor-cmd flag
    // 2. Config editor setting
    // 3. $EDITOR environment variable
    // 4. $VISUAL environment variable
    // 5. Auto-detection
    
    editorCmd := determineEditor()
    return executeEditor(editorCmd, filename)
}
```

## Security Considerations

### Data Security
- **Local Storage**: All data stored locally, no cloud dependencies
- **Configuration Security**: Secure file permissions (0600)
- **No Credential Storage**: No hardcoded secrets or API keys
- **User-Controlled**: User manages Ollama endpoint and models

### Input Validation
- **SQL Injection Prevention**: Parameterized queries
- **File Path Validation**: Secure file operations
- **Command Injection Prevention**: Sanitized shell commands
- **Buffer Overflow Protection**: Length limits on inputs

### Privacy
- **Local Processing**: Vector generation and search happen locally
- **Optional Cloud**: Ollama can run locally or remote (user choice)
- **No Telemetry**: No usage tracking or analytics
- **User Control**: Complete control over data and processing

## Performance Characteristics

### Search Performance

| Operation | Complexity | Notes |
|-----------|------------|-------|
| Text Search | O(n log n) | SQLite FTS if available |
| Semantic Search | O(log n) | Lil-rag service with vector indexing |
| Note Retrieval | O(1) | Primary key lookup |
| Document Indexing | O(1) | Asynchronous via lil-rag |

### Memory Usage
- **Embedding Cache**: Configurable embedding caching
- **Database Connections**: Connection pooling
- **Vector Storage**: Efficient float32 storage in SQLite
- **Streaming**: Large content handled via streaming

### Scalability Limits
- **Database Size**: SQLite practical limit ~1TB
- **Vector Dimensions**: Configurable, typically 384-1536
- **Concurrent Access**: Single-writer, multi-reader SQLite
- **Memory**: Scales with embedding cache size

## Error Handling Strategy

### Error Types
```go
// Custom error types with context
var (
    ErrNoteNotFound     = errors.New("note not found")
    ErrInvalidDimensions = errors.New("invalid vector dimensions")
    ErrOllamaConnection = errors.New("ollama connection failed")
    ErrConfigLoad       = errors.New("configuration load failed")
)
```

### Error Propagation
1. **Low-level Errors**: Database and API errors with context
2. **Business Logic Errors**: Validation and constraint violations  
3. **User Errors**: Clear messages with actionable guidance
4. **System Errors**: Infrastructure and dependency failures

### Graceful Degradation
- **Vector Search Fallback**: Fall back to text search if vector fails
- **Analysis Fallback**: Continue without AI features if Ollama unavailable
- **Editor Fallback**: Terminal input if editor fails
- **Configuration Fallback**: Use defaults if config corrupted

## Extension Points

### Adding New Commands
1. Create command file in `cmd/`
2. Implement Cobra command structure
3. Add to root command initialization
4. Include help documentation and examples

### Adding New Analysis Types
1. Extend `Summarizer` interface
2. Add prompt templates
3. Implement response processing
4. Add CLI flags and configuration

### Adding New Search Methods
1. Implement `Searcher` interface
2. Add configuration options
3. Integrate with unified search command
4. Add performance optimizations

### Adding New Integrations
1. Define integration interface
2. Implement protocol handling
3. Add configuration management
4. Create documentation and examples

## Configuration Management

### Configuration Hierarchy
1. **Command-line Flags**: Highest priority
2. **Environment Variables**: Middle priority  
3. **Configuration File**: Default priority
4. **Built-in Defaults**: Lowest priority

### Configuration Validation
- **Type Validation**: Ensure correct data types
- **Range Validation**: Numeric ranges and constraints
- **Dependency Validation**: Related setting consistency
- **Migration Support**: Version-based configuration updates

### Environment Variables
```bash
ML_NOTES_DATA_DIR        # Override data directory
ML_NOTES_OLLAMA_ENDPOINT # Override Ollama endpoint
ML_NOTES_DEBUG           # Enable debug mode
ML_NOTES_CONFIG_DIR      # Override config directory
```

## Testing Strategy

### Unit Tests
- **Business Logic**: Comprehensive coverage of internal packages
- **Command Tests**: CLI command functionality and edge cases
- **Integration Tests**: Database and API interactions
- **Mock Dependencies**: External service mocking (Ollama)

### Test Structure
```
internal/
├── config/
│   └── config_test.go    # Configuration loading and validation
├── database/
│   └── db_test.go        # Database operations and transactions
├── embeddings/
│   └── embeddings_test.go # Embedding generation and formatting
└── summarize/
    └── summarize_test.go  # Analysis and thinking tag processing
```

### Test Data Management
- **Test Fixtures**: Consistent test data
- **Temporary Databases**: Isolated test environments
- **Mock Servers**: HTTP test servers for API testing
- **Cleanup**: Automated test environment cleanup

## Deployment Considerations

### Build Requirements
- **Go Version**: 1.22+ for latest features
- **Platform Support**: Linux, macOS, Windows
- **External Services**: Lil-rag service for semantic search (optional)
- **Dependencies**: No external runtime dependencies

### Installation Methods
1. **Source Build**: `make build && sudo make install`
2. **Release Binaries**: Pre-built platform binaries
3. **Package Managers**: Future APT/Homebrew support
4. **Container**: Docker image for containerized deployments

### Runtime Requirements
- **SQLite**: Built into Go binary
- **Lil-rag Service**: External service for semantic search (configurable)
- **Ollama**: Optional, user-configurable endpoint for AI features
- **File System**: Read/write access to config and data directories

---

This design supports the current feature set while maintaining extensibility for future enhancements. The modular architecture allows for independent development and testing of components while ensuring reliable integration.