# ML Notes Usage Guide

## Table of Contents

1. [Quick Start](#quick-start)
2. [Web Interface](#web-interface)
3. [Core Commands](#core-commands)
4. [Configuration](#configuration)
5. [Note Management](#note-management)
6. [Search & Analysis](#search--analysis)
7. [AI-Powered Features](#ai-powered-features)
8. [Integration Features](#integration-features)
9. [Advanced Usage](#advanced-usage)
10. [Troubleshooting](#troubleshooting)

## Web Interface

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

The web server provides both a user interface and REST API endpoints for integration with other tools.

### Web UI Features

#### 📝 **Smart Markdown Editor**

The web interface features an advanced split-pane markdown editor with intelligent behavior:

**Focus-Based Behavior**:
- Editor starts hidden, showing only the rendered preview
- Click or focus on the editing area to reveal the editor pane
- Editor expands to 50/50 split when active
- Automatically returns to preview-only when focus leaves editor area

**Real-Time Preview**:
- Live markdown rendering as you type
- Synchronized scrolling between editor and preview
- Smooth transitions and animations for pane resizing

**Advanced Scroll Features**:
- **Bidirectional Sync**: Scrolling in either pane automatically syncs the other
- **Cursor Following**: Preview automatically follows your cursor position
- **Auto-Scroll**: When content extends beyond editor height, preview scrolls to bottom
- **Smart Positioning**: Maintains view context when switching between panes

**Manual Resize**:
- Drag the resize handle to manually adjust pane widths
- Double-click resize handle to toggle between 50/50 and preview-only
- Width constraints (15%-85%) prevent extreme layouts
- Smooth CSS transitions for all resize operations

#### 🏷️ **Tag Management**

**Visual Tag Interface**:
- Current tags displayed as removable badges
- Click the × on any tag to remove it instantly
- Tag input field with comma-separated support
- Real-time tag updates with visual feedback

**Tag Operations**:
```bash
# Add tags to existing note via web UI
1. Navigate to note editor
2. Use tag input field: "tag1, tag2, tag3"
3. Save note to apply changes

# Remove tags visually
1. Click × button on any tag badge
2. Changes marked as unsaved until saved
```

**AI-Powered Auto-Tagging**:
- One-click auto-tag button (🏷️ Auto-tag)
- Uses configured AI model to suggest relevant tags
- Merges suggestions with existing tags (no duplicates)
- Provides feedback on number of tags added

#### 🔍 **Integrated Search**

**Real-Time Search**:
- Search as you type with immediate results
- Automatically uses lil-rag semantic search when available
- Falls back to text search if lil-rag service unavailable
- Project-scoped results based on current directory
- Results show note previews and metadata

**Search Features**:
- **Semantic Search**: AI-powered similarity search using lil-rag service
- **Text Search**: Traditional keyword matching in titles and content
- **Tag Filtering**: Filter notes by selecting tags from dropdown
- **Result Limits**: Configurable result limits (default 20)
- **Project Isolation**: Automatically scoped to current project namespace

**Search Interface**:
```bash
# Web UI search workflow
1. Type query in search box at top of page
2. Results appear instantly in sidebar
3. Click any result to navigate to that note
4. Use tag filter dropdown for tag-based filtering
```

#### 📊 **Note Organization**

**Sidebar Navigation**:
- All notes listed chronologically (newest first)
- Note previews with title, content snippet, and date
- Active note highlighted for context
- Tag display for each note

**Note Management**:
- **Create**: Click "New Note" button or use Ctrl/Cmd+N
- **Edit**: Click any note to open in editor
- **Save**: Auto-save every 30 seconds, or manual save with Ctrl/Cmd+S
- **Delete**: Delete button with confirmation modal
- **Navigate**: Click note title or sidebar items

#### 📊 **Graph Visualization**

**Interactive Note Graph**:
- **D3.js-Powered**: Smooth, interactive visualization of note relationships
- **Tag-Based Connections**: Notes connected by shared tags with weighted links
- **Smart Layout**: Isolated notes near center, connected notes spread outward
- **Node Sizing**: Node size reflects number of connections to other notes

**Graph Interactions**:
- **Zoom and Pan**: Mouse wheel to zoom, drag to pan around graph
- **Node Navigation**: Click any node to jump directly to that note
- **Drag Nodes**: Manually reposition nodes by dragging
- **Hover Info**: Tooltips show note titles, tags, and connection counts

**Graph Controls**:
- **Filter Panel**: 
  - Filter by specific tags
  - Set minimum connection threshold
  - Limit maximum nodes displayed
- **Zoom Controls**: Zoom in/out buttons and fit-to-view
- **Reset View**: Return to optimal view with one click

**Accessing Graph View**:
```bash
# From web interface
1. Click "Graph View" in navigation
2. Use filters to focus on specific notes
3. Click nodes to navigate to notes
4. Return to main interface via breadcrumb navigation
```

#### 🎨 **Theme Support**

**Theme Features**:
- Light and dark theme toggle (🌙/☀️ button)
- Consistent theming across all components
- Automatic theme persistence in browser storage
- Keyboard shortcut: Ctrl/Cmd+/ to toggle

**Theme Characteristics**:
- **Light Theme**: Clean, paper-like texture for comfortable reading
- **Dark Theme**: Easy on eyes with proper contrast ratios
- **Responsive**: Optimized for all screen sizes and devices
- **Smooth Transitions**: Animated theme switching

#### 🤖 **AI Integration**

**Analysis Features**:
- **One-Click Analysis**: Analyze button (🧠) for current note
- **Custom Prompts**: Specify analysis focus and questions
- **Write-Back Options**: Append analysis to current note or create new note
- **Thinking Process**: See AI reasoning steps alongside final analysis

**Auto-Tagging**:
- **Smart Suggestions**: AI analyzes content to suggest relevant tags
- **Merge Logic**: Combines suggestions with existing tags
- **Model Integration**: Uses configured Ollama models for processing
- **Feedback**: Clear notifications about number of tags added

### Web Interface Keyboard Shortcuts

| Shortcut | Action | Context |
|----------|--------|---------|
| `Ctrl/Cmd + S` | Save current note | Editor |
| `Ctrl/Cmd + N` | Create new note | Global |
| `Ctrl/Cmd + P` | Toggle preview mode | Editor |
| `Ctrl/Cmd + /` | Toggle theme | Global |
| `Escape` | Close modals | Modal dialogs |
| `Enter` | Trigger search | Search input |

### Web Server Configuration

**Server Options**:
```bash
# Basic server startup
ml-notes serve                          # localhost:21212
ml-notes serve --host 0.0.0.0          # all interfaces
ml-notes serve --port 3000             # custom port

# Configuration options
ml-notes config set webui-custom-css true
ml-notes config set webui-theme dark
```

**API Integration**:
The web server exposes REST API endpoints that can be used by other applications:

- `GET /api/v1/notes` - List all notes
- `POST /api/v1/notes` - Create new note
- `GET /api/v1/notes/{id}` - Get specific note
- `PUT /api/v1/notes/{id}` - Update note
- `DELETE /api/v1/notes/{id}` - Delete note
- `POST /api/v1/notes/search` - Search notes
- `GET /api/v1/tags` - List all tags
- `GET /api/v1/graph` - Get graph data
- `POST /api/v1/auto-tag/suggest/{id}` - AI tag suggestions

### Web Interface Best Practices

1. **Efficient Editing**:
   - Use focus-based editor behavior for distraction-free writing
   - Let auto-scroll handle cursor tracking during long typing sessions
   - Take advantage of real-time preview for formatting feedback

2. **Tag Organization**:
   - Use consistent tag naming conventions
   - Leverage auto-tagging for initial tag suggestions
   - Remove irrelevant tags immediately using visual tag badges

3. **Search Strategy**:
   - Use semantic vector search for conceptual queries
   - Use tag filtering for categorical organization
   - Combine search methods for comprehensive results

4. **Semantic Search**:
   - Use semantic search for conceptual queries and related content discovery
   - Leverage project namespacing for isolated search results
   - Combine with traditional text search for comprehensive coverage

5. **Graph Navigation**:
   - Use graph view to discover unexpected note relationships
   - Filter graph by tags to focus on specific topics
   - Click nodes for quick navigation between related notes

6. **Performance Tips**:
   - Auto-save reduces risk of data loss
   - Theme persistence improves user experience
   - Keyboard shortcuts speed up common operations

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
| `tags` | Manage note tags | `ml-notes tags list` |
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
ml-notes config set lilrag-url "http://localhost:12121"
ml-notes config set enable-auto-tagging true
ml-notes config set max-auto-tags 5
ml-notes config set debug true

# Available configuration keys:
# - data-dir: Data directory location
# - ollama-endpoint: Ollama API endpoint  
# - lilrag-url: Lil-Rag service endpoint for semantic search
# - summarization-model: Model for AI analysis
# - enable-summarization: Enable/disable analysis features
# - enable-auto-tagging: Enable/disable AI auto-tagging
# - max-auto-tags: Maximum auto-generated tags per note
# - editor: Default editor for note editing
# - debug: Enable debug logging
```

## Note Management

### Creating Notes

```bash
# Interactive mode (opens editor)
ml-notes add -t "Meeting Notes"

# With content directly
ml-notes add -t "Quick Note" -c "Important reminder"

# With tags
ml-notes add -t "Project Notes" -c "Meeting details" --tags "project,meeting,important"

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

## Tag Management

### Understanding Tags

Tags are comma-separated labels that help organize and categorize your notes. They provide a powerful way to group related content and enable efficient filtering and search.

**Tag Format**: Tags are comma-separated strings (e.g., `"research,ai,important"`)

**Tag Features**:
- Case-sensitive (though you should maintain consistency)
- Can contain spaces, but avoid leading/trailing spaces
- Automatically deduplicated (no duplicate tags per note)
- Searchable and filterable
- Display in note listings and details

### Creating Notes with Tags

```bash
# Create note with single tag
ml-notes add -t "Research Paper" -c "Important findings" --tags "research"

# Create note with multiple tags
ml-notes add -t "Project Meeting" -c "Team discussion" --tags "project,meeting,team,Q4"

# Tags with spaces (use quotes)
ml-notes add -t "Learning Notes" --tags "machine learning,deep learning,neural networks"
```

### Managing Tags

```bash
# List all tags in the system
ml-notes tags list

# Add tags to existing note
ml-notes tags add 123 --tags "urgent,high-priority"

# Remove specific tags from note
ml-notes tags remove 123 --tags "outdated,old"

# Replace ALL tags for a note
ml-notes tags set 123 --tags "updated,final,complete"

# Remove all tags from a note
ml-notes tags set 123 --tags ""
```

### Tag Search

```bash
# Search notes by single tag
ml-notes search --tags "research"

# Search notes with any of multiple tags (OR operation)
ml-notes search --tags "project,meeting,important"

# Combine with other options
ml-notes search --tags "research" --limit 5 --short

# Tag-only search (no text query needed)
ml-notes search --tags "todo,urgent"
```

### Tag Organization Strategies

#### By Project
```bash
ml-notes add -t "API Design" -c "REST API specifications" --tags "project-alpha,api,backend"
ml-notes add -t "UI Mockups" -c "User interface designs" --tags "project-alpha,ui,frontend"
```

#### By Priority
```bash
ml-notes add -t "Critical Bug" -c "System crash analysis" --tags "urgent,bug,critical"
ml-notes add -t "Feature Request" -c "Nice to have feature" --tags "low-priority,enhancement"
```

#### By Category
```bash
ml-notes add -t "Learning Notes" -c "Python concepts" --tags "learning,python,programming"
ml-notes add -t "Meeting Notes" -c "Team standup" --tags "meeting,team,standup"
```

#### By Status
```bash
ml-notes add -t "Task List" -c "Things to do" --tags "todo,active"
ml-notes add -t "Completed Task" -c "Finished work" --tags "done,completed"
```

### Advanced Tag Usage

#### Hierarchical-like Tags
While ml-notes doesn't have hierarchical tags, you can simulate them:
```bash
ml-notes add -t "Backend API" --tags "work,backend,api,sprint-1"
ml-notes add -t "Database Design" --tags "work,backend,database,sprint-1"
```

#### Time-based Tags
```bash
ml-notes add -t "Q4 Goals" --tags "goals,2024,q4,planning"
ml-notes add -t "January Review" --tags "review,2024,january,retrospective"
```

#### Context Tags
```bash
ml-notes add -t "Home Office Setup" --tags "personal,office,productivity,home"
ml-notes add -t "Work Project" --tags "work,client-abc,project,deadline"
```

### Editing Notes with Tags

When editing notes with `ml-notes edit <id>`, the editor format includes tags:

```
Title: Your Note Title
Tags: tag1, tag2, tag3
---
Your note content goes here.

You can modify the title and tags directly in the editor.
```

**Tips for editing tags**:
- Keep tags on one line after "Tags: "
- Use comma separation: `tag1, tag2, tag3`
- Leave "Tags: " empty to remove all tags
- Spaces around commas are automatically trimmed

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

### Semantic Search

```bash
# Semantic similarity search using lil-rag service
ml-notes search "neural networks"

# Get multiple similar results (semantic search is the default when lil-rag is available)
ml-notes search --limit 5 "deep learning"

# Semantic search finds conceptually related content even without exact keyword matches
ml-notes search "AI concepts"

# Force text-only search when you need exact matches
ml-notes search --text-only "exact phrase"
```

### Tag Search

```bash
# Search by single tag
ml-notes search --tags "research"

# Search by multiple tags (finds notes with ANY of these tags)
ml-notes search --tags "project,urgent,important"

# Tag search with result formatting
ml-notes search --tags "meeting" --short --limit 10

# Tag-only search (no text query required)
ml-notes search --tags "todo"
```

### Combined Search Strategies

```bash
# Text search with tag filtering (not yet implemented)
# ml-notes search "python" --tags "learning,tutorial"

# Use multiple search methods separately
ml-notes search "machine learning"             # Semantic search (default)
ml-notes search --text-only "python"           # Force text search  
ml-notes search --tags "coding,python"         # Tag search
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

### Lil-Rag Integration

```bash
# Configure lil-rag service endpoint
ml-notes config set lilrag-url "http://localhost:12121"

# Test lil-rag connectivity
curl http://localhost:12121/health

# Notes are automatically indexed in lil-rag when created/updated
# Search automatically uses semantic search when lil-rag is available
# Project namespacing ensures search isolation between different projects
```

### Workflow Examples

#### Research Workflow
```bash
# 1. Collect information
ml-notes add -t "Research: Topic X" < research_notes.txt

# 2. Search related content
ml-notes search "Topic X concepts"

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
ml-notes search "machine learning concepts"

# 3. Test understanding
ml-notes analyze --recent 5 -p "Explain these concepts as if teaching someone else"
```

### Performance Tips

```bash
# Use pagination for large datasets
ml-notes list --limit 50 --offset 100

# Use semantic search for conceptual queries (default)
ml-notes search "concepts and ideas"

# Use text search for exact matches
ml-notes search --text-only "specific phrase or term"

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

#### Lil-Rag Service Issues
```bash
# Test lil-rag connectivity
curl http://localhost:12121/health

# Update lil-rag endpoint
ml-notes config set lilrag-url "http://your-server:12121"

# Check if lil-rag is responding
ml-notes --debug search "test query"
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

# Check lil-rag service status
curl http://localhost:12121/health

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
6. **Search Strategy**: Use semantic search for concepts, text search for exact matches
7. **Note Organization**: Consider using prefixes like "Meeting:", "Research:", etc.

---

For more detailed information on specific commands, use `ml-notes [command] --help`.