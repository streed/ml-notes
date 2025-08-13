# MCP Server Search Functionality

The ML Notes MCP server provides comprehensive search capabilities through its `search_notes` tool.

## Search Tool Features

### Tool Definition
```json
{
  "name": "search_notes",
  "description": "Search for notes using vector similarity or text search",
  "inputSchema": {
    "type": "object",
    "properties": {
      "query": {
        "type": "string",
        "description": "Search query string",
        "required": true
      },
      "limit": {
        "type": "number",
        "description": "Maximum number of results",
        "default": 10
      },
      "use_vector": {
        "type": "boolean",
        "description": "Use vector search if available",
        "default": true
      }
    }
  }
}
```

## Search Capabilities

1. **Vector Search** (when enabled):
   - Uses Ollama embeddings for semantic similarity
   - Searches based on meaning, not just keywords
   - Falls back to text search if vector search fails

2. **Text Search**:
   - SQL LIKE queries on title and content
   - Case-insensitive partial matching
   - Fast and efficient for keyword searches

## Usage Examples

### Basic Search
Search for notes containing "machine learning":
```json
{
  "tool": "search_notes",
  "arguments": {
    "query": "machine learning"
  }
}
```

### Limited Results
Search with a maximum of 5 results:
```json
{
  "tool": "search_notes",
  "arguments": {
    "query": "python",
    "limit": 5
  }
}
```

### Force Text Search
Disable vector search and use text matching only:
```json
{
  "tool": "search_notes",
  "arguments": {
    "query": "database optimization",
    "use_vector": false,
    "limit": 10
  }
}
```

## Response Format

The search tool returns results in a formatted text format:
```
Found 3 notes:

1. [ID: 42] Introduction to Machine Learning
   This note covers the basics of supervised and unsupervised learning...

2. [ID: 85] Deep Learning with Python
   A comprehensive guide to implementing neural networks using TensorFlow...

3. [ID: 91] ML Model Deployment
   Best practices for deploying machine learning models in production...
```

## Integration with Claude Desktop

To use the search functionality in Claude Desktop:

1. Start the MCP server:
   ```bash
   ml-notes mcp
   ```

2. Configure Claude Desktop's `claude_desktop_config.json`:
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

3. In Claude Desktop, you can now:
   - Search for notes: "Search my notes for information about neural networks"
   - Find specific topics: "Find all notes mentioning Python programming"
   - Semantic search: "Look for notes about data analysis techniques"

## Search Behavior

- **With Vector Search Enabled**:
  1. Generates embedding for the search query
  2. Finds notes with similar embeddings using cosine similarity
  3. Returns top N most similar notes
  4. Falls back to text search if vector search fails

- **With Vector Search Disabled or Unavailable**:
  1. Performs SQL LIKE query on note titles and content
  2. Returns all matching notes (up to limit)
  3. Results ordered by creation date (newest first)

## Error Handling

The search tool handles various error scenarios:
- Missing required query parameter
- Database connection issues
- Ollama service unavailable (falls back to text search)
- No results found (returns friendly message)

## Performance Considerations

- Vector search requires Ollama to be running
- First vector search may be slower due to model loading
- Text search is always fast and doesn't require external services
- Results are truncated to 100 characters in the listing for readability