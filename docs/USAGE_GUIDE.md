# ML Notes Usage Guide

## Table of Contents

1. [Quick Start](#quick-start)
2. [Core Commands](#core-commands)
3. [Configuration](#configuration)
4. [Note Management](#note-management)
5. [Search & Analysis](#search--analysis)
6. [AI-Powered Features](#ai-powered-features)
7. [Integration Features](#integration-features)
8. [Advanced Usage](#advanced-usage)
9. [Troubleshooting](#troubleshooting)

## Quick Start

### Installation & Setup

```bash
# 1. Build and install
make build && sudo make install

# 2. Initialize configuration
ml-notes init --interactive

# 3. Create your first note
ml-notes add -t "My First Note" -c "This is my first note content"

# 4. List your notes
ml-notes list

# 5. Search your notes
ml-notes search "first note"
```

## Core Commands

### Essential Commands Summary

| Command | Purpose | Example |
|---------|---------|---------|
| `init` | Set up ml-notes | `ml-notes init -i` |
| `add` | Create new notes | `ml-notes add -t "Title"` |
| `list` | View all notes | `ml-notes list --limit 10` |
| `get` | Retrieve specific note | `ml-notes get 123` |
| `edit` | Modify existing notes | `ml-notes edit 123` |
| `delete` | Remove notes | `ml-notes delete 123` |
| `search` | Find notes | `ml-notes search "query"` |
| `analyze` | AI-powered analysis | `ml-notes analyze 123` |
| `config` | Manage settings | `ml-notes config show` |

## Configuration

### Initial Setup

```bash
# Interactive setup (recommended)
ml-notes init --interactive

# Quick setup with defaults
ml-notes init

# Custom setup
ml-notes init \
  --data-dir ~/my-notes \
  --ollama-endpoint http://localhost:11434 \
  --summarization-model llama3.2:latest \
  --enable-summarization
```

### Configuration Management

```bash
# View current configuration
ml-notes config show

# Update specific settings
ml-notes config set editor "code --wait"
ml-notes config set ollama-endpoint "http://localhost:11434"
ml-notes config set embedding-model "nomic-embed-text:v1.5"
ml-notes config set debug true

# Available configuration keys:
# - data-dir: Data directory location
# - ollama-endpoint: Ollama API endpoint  
# - embedding-model: Model for vector embeddings
# - vector-dimensions: Embedding vector size
# - enable-vector: Enable/disable vector search
# - debug: Enable debug logging
# - summarization-model: Model for AI analysis
# - enable-summarization: Enable/disable analysis features
# - editor: Default editor for note editing
```

## Note Management

### Creating Notes

```bash
# Interactive mode (opens editor)
ml-notes add -t "Meeting Notes"

# With content directly
ml-notes add -t "Quick Note" -c "Important reminder"

# From stdin
echo "Note content" | ml-notes add -t "Stdin Note"

# Use specific editor
ml-notes add -t "Code Review" --editor-cmd "code --wait"
```

### Viewing Notes

```bash
# List recent notes
ml-notes list

# Paginated listing
ml-notes list --limit 20 --offset 40

# Short format (ID and title only)
ml-notes list --short

# Get specific note
ml-notes get 123
```

### Editing Notes

```bash
# Edit entire note
ml-notes edit 123

# Edit title only
ml-notes edit -t 123

# Edit content only  
ml-notes edit -c 123

# Use specific editor
ml-notes edit -e "vim" 123
```

### Deleting Notes

```bash
# Delete single note (with confirmation)
ml-notes delete 123

# Delete multiple notes
ml-notes delete 123 456 789

# Skip confirmation
ml-notes delete -f 123

# Delete all notes (DANGEROUS!)
ml-notes delete --all
```

## Search & Analysis

### Text Search

```bash
# Basic text search
ml-notes search "machine learning"

# Limit results
ml-notes search --limit 5 "algorithms"

# Short format results
ml-notes search --short "python"
```

### Vector Search

```bash
# Semantic similarity search (returns top match)
ml-notes search --vector "neural networks"

# Get multiple similar results
ml-notes search --vector --limit 5 "deep learning"

# Vector search finds semantically related content even without exact matches
ml-notes search --vector "AI concepts"
```

### AI-Powered Analysis

```bash
# Analyze single note
ml-notes analyze 123

# Analyze with custom prompt
ml-notes analyze 123 -p "Focus on technical implementation details"

# Analyze multiple notes
ml-notes analyze 1 2 3 -p "Compare these approaches"

# Analyze all notes
ml-notes analyze --all -p "What are the recurring themes?"

# Analyze recent notes
ml-notes analyze --recent 10

# Use specific model
ml-notes analyze --model gemma3:12b 123
```

### Search + Analysis

```bash
# Analyze search results
ml-notes search --analyze "machine learning"

# Custom analysis of search results  
ml-notes search --analyze -p "Focus on practical applications" "algorithms"

# Show both analysis and detailed results
ml-notes search --analyze --show-details "python"
```

## AI-Powered Features

### Custom Analysis Prompts

ML Notes supports custom prompts for targeted analysis:

```bash
# Technical focus
ml-notes analyze 123 -p "Explain the technical architecture and implementation"

# Business focus  
ml-notes analyze 123 -p "What are the business implications and ROI?"

# Comparative analysis
ml-notes analyze 1 2 3 -p "Compare and contrast these three approaches"

# Pattern extraction
ml-notes analyze --all -p "What patterns and themes emerge across all notes?"

# Problem-solving focus
ml-notes analyze --recent 20 -p "What problems are mentioned and what solutions are proposed?"
```

### Analysis Output Format

Analysis includes both the AI's reasoning process and final insights:

```
📝 Analysis:
[Final summary and insights]

🤔 Analysis Process:
──────────────────────────────────────────────────
[AI's internal reasoning and thought process]  
──────────────────────────────────────────────────

[Additional analysis content]
```

## Integration Features

### Editor Integration

ML Notes integrates with your preferred text editor:

```bash
# Configure default editor
ml-notes config set editor "code --wait"

# Editor precedence order:
# 1. --editor-cmd flag
# 2. Config editor setting  
# 3. $EDITOR environment variable
# 4. $VISUAL environment variable
# 5. Auto-detection (vim, nano, emacs, code, etc.)
```

### MCP Server Integration

For integration with Claude Desktop and other LLM clients:

```bash
# Start MCP server
ml-notes mcp

# Add to Claude Desktop config:
# ~/.config/claude/claude_desktop_config.json
{
  "mcpServers": {
    "ml-notes": {
      "command": "ml-notes",
      "args": ["mcp"]
    }
  }
}
```

### Shell Integration

```bash
# Generate shell completions
ml-notes completion bash > /etc/bash_completion.d/ml-notes
ml-notes completion zsh > ~/.zsh/completions/_ml-notes
ml-notes completion fish > ~/.config/fish/completions/ml-notes.fish
```

## Advanced Usage

### Vector Search Optimization

```bash
# Detect optimal embedding dimensions
ml-notes detect-dimensions

# Reindex after model changes
ml-notes reindex

# Configure vector search
ml-notes config set embedding-model "nomic-embed-text:v1.5"  
ml-notes config set vector-dimensions 768
ml-notes config set enable-vector true
```

### Workflow Examples

#### Research Workflow
```bash
# 1. Collect information
ml-notes add -t "Research: Topic X" < research_notes.txt

# 2. Search related content
ml-notes search --vector "Topic X concepts"

# 3. Analyze patterns
ml-notes search --analyze -p "What are the key research findings?" "Topic X"

# 4. Generate insights
ml-notes analyze --all -p "What research gaps exist in this area?"
```

#### Meeting Notes Workflow
```bash
# 1. Create meeting note
ml-notes add -t "Team Meeting $(date +%Y-%m-%d)"

# 2. Search previous meetings
ml-notes search "team meeting" --limit 5

# 3. Analyze meeting patterns
ml-notes search --analyze -p "What action items and decisions were made?" "meeting"
```

#### Learning Workflow
```bash
# 1. Add learning notes
ml-notes add -t "Learning: Machine Learning" -c "Key concepts..."

# 2. Search related topics
ml-notes search --vector "machine learning concepts"

# 3. Test understanding
ml-notes analyze --recent 5 -p "Explain these concepts as if teaching someone else"
```

### Performance Tips

```bash
# Use pagination for large datasets
ml-notes list --limit 50 --offset 100

# Prefer vector search for semantic queries
ml-notes search --vector "concepts and ideas"

# Use text search for exact matches
ml-notes search "specific phrase or term"

# Batch operations where possible
ml-notes delete 1 2 3 4 5  # vs individual deletes
```

## Troubleshooting

### Common Issues

#### Configuration Issues
```bash
# Check configuration
ml-notes config show

# Debug mode
ml-notes --debug <command>

# Reset configuration
rm ~/.config/ml-notes/config.json
ml-notes init
```

#### Ollama Connection Issues
```bash
# Test Ollama connectivity
curl http://localhost:11434/api/tags

# Update endpoint
ml-notes config set ollama-endpoint "http://your-server:11434"

# Check model availability
ml-notes detect-dimensions
```

#### Vector Search Issues
```bash
# Dimension mismatch - reindex
ml-notes detect-dimensions
ml-notes reindex

# Disable vector search temporarily
ml-notes config set enable-vector false
```

#### Editor Issues
```bash
# Configure editor explicitly
ml-notes config set editor "nano"

# Test with specific editor
ml-notes edit -e "vim" 123

# Check available editors
which vim vi nano emacs code
```

### Debug Information

```bash
# Enable debug logging
ml-notes config set debug true

# Or per-command
ml-notes --debug search "query"

# Check paths and configuration
ml-notes config show
```

### Performance Issues

```bash
# Check database size
ls -lh ~/.local/share/ml-notes/notes.db

# Reindex if search is slow
ml-notes reindex

# Use pagination for large result sets
ml-notes list --limit 20
ml-notes search --limit 10 "query"
```

## Best Practices

1. **Regular Backups**: Backup `~/.local/share/ml-notes/` periodically
2. **Consistent Titles**: Use descriptive, searchable titles
3. **Tag-like Keywords**: Include relevant keywords in content
4. **Custom Prompts**: Use specific analysis prompts for better insights
5. **Editor Setup**: Configure your preferred editor for better experience
6. **Search Strategy**: Use vector search for concepts, text search for specifics
7. **Note Organization**: Consider using prefixes like "Meeting:", "Research:", etc.

---

For more detailed information on specific commands, use `ml-notes [command] --help`.