# Vector Search - Single Top Result

## Overview
Vector search in ML Notes now returns only the most similar note by default, making it more precise and focused for finding the single best match to your query.

## Behavior

### Default Limits
- **Vector Search**: Returns 1 result (the most similar note)
- **Text Search**: Returns up to 10 results

### Command Line Usage

#### Get the single most similar note (default)
```bash
ml-notes search -v "machine learning concepts"
```

#### Get multiple similar notes
```bash
ml-notes search -v "machine learning concepts" -l 5
```

#### Text search (returns multiple by default)
```bash
ml-notes search "python programming"
```

## MCP Server Behavior

The MCP server's `search_notes` tool also follows this pattern:

### Default behavior (vector search, single result):
```json
{
  "tool": "search_notes",
  "arguments": {
    "query": "neural networks"
  }
}
```

### Get multiple results:
```json
{
  "tool": "search_notes",
  "arguments": {
    "query": "neural networks",
    "limit": 5
  }
}
```

### Force text search:
```json
{
  "tool": "search_notes",
  "arguments": {
    "query": "neural networks",
    "use_vector": false
  }
}
```

## Rationale

Returning only the top result for vector search provides several benefits:

1. **Precision**: Vector search finds the semantically most similar note, so the top result is usually what you're looking for
2. **Clarity**: No need to scan through multiple results when you want the best match
3. **Performance**: Faster response times when retrieving just one result
4. **LLM Integration**: When used via MCP, LLMs get a focused result instead of having to parse multiple notes

## Customization

You can always override the default behavior:
- Use `-l N` or `--limit N` to get more results
- Use `-l 10` to get the previous default behavior
- The limit applies to both vector and text search

## Examples

### Finding a specific concept
```bash
# Returns the single most relevant note about recursion
ml-notes search -v "explain recursion in programming"
```

### Research mode - getting multiple perspectives
```bash
# Returns top 5 notes about databases
ml-notes search -v "database optimization techniques" -l 5
```

### Comparison with text search
```bash
# Text search: returns multiple partial matches
ml-notes search "python"

# Vector search: returns the single most relevant Python note
ml-notes search -v "python programming"
```