# ML Notes API & Integration Guide

## Table of Contents

1. [MCP Server API](#mcp-server-api)
2. [Claude Desktop Integration](#claude-desktop-integration)
3. [Project Namespacing](#project-namespacing)  
4. [Editor Integration](#editor-integration)
5. [Shell Integration](#shell-integration)
6. [Custom Extensions](#custom-extensions)
7. [Configuration API](#configuration-api)
8. [Error Handling](#error-handling)

## MCP Server API

The Model Context Protocol (MCP) server enables LLMs to interact with ml-notes programmatically.

### Starting the MCP Server

```bash
# Start the MCP server (communicates via stdio)
ml-notes mcp

# The server automatically handles JSON-RPC communication
# and integrates with the user's ml-notes configuration
```

### Available Tools

#### 1. add_note

Add a new note to the database.

**Parameters:**
```json
{
  "title": "string (required)",
  "content": "string (required)",
  "tags": "string (optional) - Comma-separated tags for the note"
}
```

**Example:**
```json
{
  "name": "add_note",
  "arguments": {
    "title": "Meeting Notes",
    "content": "Discussed project timeline and deliverables.",
    "tags": "meeting,project,planning"
  }
}
```

**Response:**
```json
{
  "content": [
    {
      "type": "text", 
      "text": "Created note with ID 123: Meeting Notes\\nTags: meeting, project, planning"
    }
  ]
}
```

#### 2. search_notes

Search notes using text or vector similarity.

**Parameters:**
```json
{
  "query": "string (optional if tags provided)",
  "method": "text|vector (optional, default: text)",
  "limit": "number (optional, default: 10)",
  "tags": "string (optional) - Comma-separated tags to search for"
}
```

**Example:**
```json
{
  "name": "search_notes",
  "arguments": {
    "query": "machine learning algorithms",
    "method": "vector",
    "limit": 5
  }
}
```

**Tag Search Example:**
```json
{
  "name": "search_notes",
  "arguments": {
    "tags": "research,ai,machine-learning"
  }
}
```

**Response:**
```json
{
  "content": [
    {
      "type": "text",
      "text": "Found 3 matching notes:\n\n[ID: 45] Neural Networks Guide [Tags: ai, deep-learning]\nCreated: 2024-01-15\nPreview: Introduction to neural networks and deep learning...\n\n[ID: 67] ML Algorithms Overview [Tags: ml, algorithms, research]\nCreated: 2024-01-20\nPreview: Comprehensive guide to machine learning algorithms..."
    }
  ]
}
```

#### 3. get_note

Retrieve a specific note by ID.

**Parameters:**
```json
{
  "id": "number (required)"
}
```

**Example:**
```json
{
  "name": "get_note", 
  "arguments": {
    "id": 123
  }
}
```

**Response:**
```json
{
  "content": [
    {
      "type": "text",
      "text": "ID: 123\nTitle: Meeting Notes\nTags: meeting, project, planning\nCreated: 2024-01-15 14:30:00\nUpdated: 2024-01-15 14:30:00\n\nContent:\nDiscussed project timeline and deliverables.\n- Phase 1: Research (2 weeks)\n- Phase 2: Development (4 weeks)\n- Phase 3: Testing (1 week)"
    }
  ]
}
```

#### 4. list_notes

List notes with pagination support.

**Parameters:**
```json
{
  "limit": "number (optional, default: 10)",
  "offset": "number (optional, default: 0)"
}
```

**Example:**
```json
{
  "name": "list_notes",
  "arguments": {
    "limit": 5,
    "offset": 0
  }
}
```

#### 5. update_note

Modify an existing note.

**Parameters:**
```json
{
  "id": "number (required)",
  "title": "string (optional)",
  "content": "string (optional)",
  "tags": "string (optional) - Comma-separated tags to set for the note"
}
```

**Example:**
```json
{
  "name": "update_note",
  "arguments": {
    "id": 123,
    "title": "Updated Meeting Notes",
    "content": "Added follow-up action items.",
    "tags": "meeting,project,follow-up,updated"
  }
}
```

#### 6. delete_note

Remove a note from the database.

**Parameters:**
```json
{
  "id": "number (required)"
}
```

**Example:**
```json
{
  "name": "delete_note",
  "arguments": {
    "id": 123
  }
}
```

#### 7. list_tags

List all tags in the system.

**Parameters:**
None required.

**Example:**
```json
{
  "name": "list_tags",
  "arguments": {}
}
```

**Response:**
```json
{
  "content": [
    {
      "type": "text",
      "text": "Found 5 tags:\n\n1. meeting (Created: 2024-01-15)\n2. project (Created: 2024-01-15)\n3. research (Created: 2024-01-16)\n4. urgent (Created: 2024-01-17)\n5. planning (Created: 2024-01-17)"
    }
  ]
}
```

#### 8. update_note_tags

Update the tags for a specific note.

**Parameters:**
```json
{
  "id": "number (required)",
  "tags": "string (required) - Comma-separated tags to set for the note"
}
```

**Example:**
```json
{
  "name": "update_note_tags",
  "arguments": {
    "id": 123,
    "tags": "research,updated,final"
  }
}
```

**Response:**
```json
{
  "content": [
    {
      "type": "text",
      "text": "Successfully updated tags for note 123\nTags: research, updated, final"
    }
  ]
}
```

### Available Resources

#### notes://recent

Get the most recently created notes.

**URI:** `notes://recent`

**Response:**
```json
{
  "contents": [
    {
      "uri": "notes://recent",
      "mimeType": "text/plain",
      "text": "Recent Notes:\n\n[ID: 145] Project Planning\nCreated: 2024-01-20 09:15:00\n\n[ID: 144] Research Findings\nCreated: 2024-01-19 16:30:00"
    }
  ]
}
```

#### notes://stats

Get database statistics and configuration.

**URI:** `notes://stats`

**Response:**
```json
{
  "contents": [
    {
      "uri": "notes://stats", 
      "mimeType": "application/json",
      "text": "{\n  \"total_notes\": 145,\n  \"vector_search_enabled\": true,\n  \"embedding_model\": \"nomic-embed-text:v1.5\",\n  \"vector_dimensions\": 768,\n  \"database_size\": \"2.4 MB\"\n}"
    }
  ]
}
```

#### notes://config

Get current ml-notes configuration.

**URI:** `notes://config`

#### notes://note/{id}

Access a specific note by ID.

**URI:** `notes://note/123`

### Available Prompts

#### search_notes

Pre-configured search prompt with parameters.

**Name:** `search_notes`

**Arguments:**
```json
{
  "query": "string",
  "search_type": "text|vector|both"
}
```

#### analyze_notes

Generate AI analysis of note collection.

**Name:** `analyze_notes`

**Arguments:**
```json
{
  "focus": "string (analysis focus or question)",
  "scope": "recent|all|specific_ids"
}
```

## Claude Desktop Integration

### Configuration Setup

1. **Locate Claude Desktop config file:**
   - **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - **Windows**: `%APPDATA%\\Claude\\claude_desktop_config.json`
   - **Linux**: `~/.config/claude/claude_desktop_config.json`

2. **Add ml-notes MCP server:**
```json
{
  "mcpServers": {
    "ml-notes": {
      "command": "ml-notes",
      "args": ["mcp"],
      "env": {
        "ML_NOTES_CONFIG_DIR": "/path/to/custom/config",
        "ML_NOTES_DATA_DIR": "/path/to/custom/data"
      }
    }
  }
}
```

3. **Restart Claude Desktop**

### Usage Examples

#### Basic Note Management
```
# Ask Claude to:
"Create a note about today's meeting with the development team"
"Search for all notes about machine learning"  
"Show me my most recent notes"
"Update note 123 to include the new deadline"
```

#### Advanced Analysis
```
# Ask Claude to:
"Analyze all my research notes and identify common themes"
"Search for project-related notes and summarize the current status"
"Find notes about Python and create a learning roadmap"
"What patterns emerge from my meeting notes over the last month?"
```

#### Integration Workflows
```
# Ask Claude to:
"Create a note from this email thread I'm sharing"
"Search my notes for anything related to this code problem"
"Analyze my learning notes and suggest what to study next"
"Find all project notes and create a status report"
```

## Project Namespacing

ML Notes automatically isolates notes and search results by project directory using intelligent namespacing with lil-rag.

### How It Works

**Automatic Project Detection**: ML Notes uses the current working directory name to create project-scoped namespaces:

```bash
# Working in /home/user/my-awesome-project
ml-notes search --vector "machine learning"
# Searches namespace: ml-notes-my-awesome-project

# Working in /home/user/research-notes  
ml-notes search --vector "machine learning"
# Searches namespace: ml-notes-research-notes
```

### Namespace Benefits

**Cross-Project Isolation**: 
- Notes from different projects never contaminate search results
- Each project maintains its own semantic search space
- Clean separation between work, personal, and research notes

**Consistent Experience**:
- Same commands work across all projects
- No manual namespace management required
- Automatic fallback to text search when lil-rag unavailable

### Configuration

**Lil-Rag Service Setup**:
```bash
# Install lil-rag service
go install github.com/stillmatic/lil-rag@latest

# Start service (required for namespace features)
lil-rag serve --port 12121

# Configure ML Notes
ml-notes config set lilrag-url http://localhost:12121
```

**Namespace Format**: `ml-notes-{project-directory-name}`

### API Integration

When using the MCP server or REST API, namespacing automatically applies based on the working directory where `ml-notes mcp` or `ml-notes serve` was started.

**MCP Server with Project Context**:
```json
{
  "mcpServers": {
    "ml-notes-project1": {
      "command": "ml-notes",
      "args": ["mcp"],
      "cwd": "/path/to/project1"
    },
    "ml-notes-project2": {
      "command": "ml-notes", 
      "args": ["mcp"],
      "cwd": "/path/to/project2"
    }
  }
}
```

### Debugging Namespaces

**View Current Namespace**:
```bash
# Enable debug mode to see namespace usage
ml-notes --debug search --vector "test"
# Output shows: "Searching lil-rag for: test (namespace: ml-notes-myproject)"
```

**Verify Service**:
```bash
# Check lil-rag connectivity  
curl http://localhost:12121/health

# View debug logs
ml-notes --debug add -t "Test" -c "Testing namespace"
```

## Editor Integration

### Supported Editors

ML Notes integrates with any text editor through configurable commands:

```bash
# Popular editor configurations
ml-notes config set editor "code --wait"          # VS Code
ml-notes config set editor "vim"                  # Vim
ml-notes config set editor "emacs -nw"            # Emacs (terminal)
ml-notes config set editor "nano"                 # Nano
ml-notes config set editor "subl --wait"          # Sublime Text
ml-notes config set editor "atom --wait"          # Atom
```

### Editor Detection Priority

1. `--editor-cmd` flag (highest priority)
2. Configuration `editor` setting
3. `$EDITOR` environment variable
4. `$VISUAL` environment variable  
5. Auto-detection (vim, vi, nano, emacs, code, subl, atom)

### Custom Editor Integration

For complex editor setups:

```bash
# Multi-argument commands
ml-notes config set editor "code --wait --new-window"

# Script-based integration
ml-notes config set editor "/usr/local/bin/my-editor-wrapper"

# Per-command override
ml-notes add -t "Title" --editor-cmd "vim +startinsert"
ml-notes edit 123 -e "emacs -nw"
```

### Editor Templates

When creating notes, ml-notes provides templates:

```markdown
# Note Title

[Write your note content here]

<!-- 
  Save and close the editor when done.
  To cancel, exit without saving.
-->
```

## Shell Integration

### Command Completion

Generate shell completions for better CLI experience:

```bash
# Bash
ml-notes completion bash > /etc/bash_completion.d/ml-notes
source /etc/bash_completion.d/ml-notes

# Zsh
ml-notes completion zsh > "${fpath[1]}/_ml-notes"
# or
ml-notes completion zsh > ~/.zsh/completions/_ml-notes

# Fish
ml-notes completion fish > ~/.config/fish/completions/ml-notes.fish

# PowerShell
ml-notes completion powershell > ml-notes.ps1
```

### Environment Variables

```bash
# Configuration overrides
export ML_NOTES_CONFIG_DIR="$HOME/.config/ml-notes"
export ML_NOTES_DATA_DIR="$HOME/.local/share/ml-notes"  
export ML_NOTES_DEBUG="true"
export ML_NOTES_OLLAMA_ENDPOINT="http://localhost:11434"

# Editor configuration
export EDITOR="code --wait"
export VISUAL="vim"
```

### Shell Aliases

Useful aliases for common operations:

```bash
# ~/.bashrc or ~/.zshrc
alias mn='ml-notes'
alias mna='ml-notes add'
alias mns='ml-notes search'
alias mnl='ml-notes list'
alias mnc='ml-notes config show'

# Advanced aliases
alias mnq='ml-notes search --vector'          # Quick semantic search
alias mnan='ml-notes analyze'                 # Quick analysis
alias mnr='ml-notes list --limit 5'          # Recent notes
alias mne='ml-notes edit'                     # Quick edit
```

### Scripting Integration

```bash
#!/bin/bash
# Example: Batch note creation from files

for file in *.txt; do
    title=$(basename "$file" .txt)
    content=$(cat "$file")
    ml-notes add -t "$title" -c "$content"
done

# Example: Daily note creation
today=$(date +"%Y-%m-%d")
ml-notes add -t "Daily Notes - $today"

# Example: Search and analyze workflow
results=$(ml-notes search --vector "project planning" --limit 5)
if [[ -n "$results" ]]; then
    echo "Found relevant notes, generating analysis..."
    ml-notes search --analyze -p "What are the next steps?" "project planning"
fi
```

## Custom Extensions

### Database Access

For advanced integrations, access the SQLite database directly:

```bash
# Database location
DB_PATH="$HOME/.local/share/ml-notes/notes.db"

# Example: Export notes to JSON
sqlite3 "$DB_PATH" << EOF
.mode json
.once notes_export.json
SELECT id, title, content, created_at, updated_at FROM notes;
EOF

# Example: Custom analytics
sqlite3 "$DB_PATH" << EOF
SELECT 
    DATE(created_at) as date,
    COUNT(*) as notes_created
FROM notes 
GROUP BY DATE(created_at)
ORDER BY date DESC;
EOF
```

### Configuration Access

```bash
# Configuration file location
CONFIG_FILE="$HOME/.config/ml-notes/config.json"

# Parse configuration with jq
OLLAMA_ENDPOINT=$(jq -r '.ollama_endpoint' "$CONFIG_FILE")
EMBEDDING_MODEL=$(jq -r '.embedding_model' "$CONFIG_FILE")

echo "Using Ollama at: $OLLAMA_ENDPOINT"
echo "Embedding model: $EMBEDDING_MODEL"
```

### API Wrappers

Create language-specific wrappers:

```python
# Python wrapper example
import subprocess
import json

class MLNotes:
    def add_note(self, title, content):
        cmd = ['ml-notes', 'add', '-t', title, '-c', content]
        result = subprocess.run(cmd, capture_output=True, text=True)
        return result.returncode == 0
    
    def search_notes(self, query, vector=False, limit=10):
        cmd = ['ml-notes', 'search']
        if vector:
            cmd.append('--vector')
        cmd.extend(['--limit', str(limit), query])
        
        result = subprocess.run(cmd, capture_output=True, text=True)
        return result.stdout if result.returncode == 0 else None
    
    def analyze_note(self, note_id, prompt=None):
        cmd = ['ml-notes', 'analyze', str(note_id)]
        if prompt:
            cmd.extend(['-p', prompt])
        
        result = subprocess.run(cmd, capture_output=True, text=True)
        return result.stdout if result.returncode == 0 else None

# Usage
notes = MLNotes()
notes.add_note("Python Integration", "Testing the Python wrapper")
results = notes.search_notes("Python", vector=True)
analysis = notes.analyze_note(123, "Focus on technical details")
```

## Configuration API

### Programmatic Configuration Management

```bash
# Get all configuration
ml-notes config show --format json > current_config.json

# Set multiple configurations
ml-notes config set ollama-endpoint "http://gpu-server:11434"
ml-notes config set embedding-model "nomic-embed-text:v1.5"
ml-notes config set editor "code --wait"

# Environment-based configuration
export ML_NOTES_OLLAMA_ENDPOINT="http://localhost:11434"
export ML_NOTES_DEBUG="true"
ml-notes search "test"  # Uses environment variables
```

### Configuration Validation

```bash
# Test configuration
ml-notes detect-dimensions  # Validates embedding model
ml-notes config show        # Shows current configuration
ml-notes --debug search "test"  # Debug mode for troubleshooting
```

## Error Handling

### Common Error Codes

| Exit Code | Description | Action |
|-----------|-------------|--------|
| 0 | Success | Operation completed |
| 1 | General error | Check error message |
| 2 | Configuration error | Check config with `ml-notes config show` |
| 3 | Database error | Check data directory permissions |
| 4 | Network error | Check Ollama connection |
| 5 | Validation error | Check input parameters |

### Error Response Format

MCP server errors follow JSON-RPC format:

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32000,
    "message": "Note not found",
    "data": {
      "note_id": 999,
      "details": "No note exists with ID 999"
    }
  },
  "id": 1
}
```

### Error Handling Best Practices

1. **Check exit codes** in scripts
2. **Parse error messages** for actionable information  
3. **Use debug mode** for troubleshooting
4. **Validate configuration** before operations
5. **Handle network errors** gracefully
6. **Test with minimal examples** when debugging

### Debugging Tools

```bash
# Enable debug mode
ml-notes --debug <command>
ml-notes config set debug true

# Check system status
ml-notes config show
ml-notes detect-dimensions

# Test connectivity
curl http://localhost:11434/api/tags

# Validate database
ls -la ~/.local/share/ml-notes/
sqlite3 ~/.local/share/ml-notes/notes.db ".tables"

# Test MCP server
echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | ml-notes mcp
```

---

This API documentation covers the major integration points for ml-notes. For specific use cases or additional integration needs, refer to the comprehensive usage guide and system design documentation.