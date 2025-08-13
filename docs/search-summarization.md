# Search Result Summarization

## Overview
The search command can generate AI-powered summaries of search results, providing a concise overview instead of detailed note listings.

## Default Behavior

When using the `--summarize` flag:
1. Search results are collected as usual
2. An AI summary is generated using the configured summarization model
3. **Only the summary is displayed** (detailed results are hidden)
4. If summarization fails, detailed results are shown as fallback

## Usage Examples

### Basic Summarization
```bash
# Shows only the AI-generated summary
ml-notes search --summarize "machine learning"
```

Output:
```
Performing text search for: machine learning

Found 5 matching notes:

Generating summary of search results...
================================================================================

📝 Summary of Search Results:
--------------------------------------------------------------------------------
The search results cover various aspects of machine learning, including 
supervised learning algorithms, neural network architectures, and practical
implementation examples in Python. Key topics include decision trees, 
gradient descent optimization, and deep learning frameworks.
--------------------------------------------------------------------------------

✨ Summary generated using llama3.2:latest
   Reduced from 3542 to 287 characters (91.9% compression)
================================================================================
```

### Summary with Detailed Results
If you want both the summary AND the detailed results:

```bash
ml-notes search --summarize --show-details "python programming"
```

This will show:
1. The AI-generated summary first
2. Followed by the detailed list of all matching notes

### Vector Search with Summary
```bash
# Get summary of the most similar note
ml-notes search -v --summarize "neural networks"

# Get summary of top 5 similar notes
ml-notes search -v --summarize -l 5 "deep learning concepts"
```

## Flags

| Flag | Description |
|------|-------------|
| `--summarize` | Generate AI summary of results (hides details by default) |
| `--show-details` | Show detailed results even when summarizing |
| `-v, --vector` | Use vector search |
| `-l, --limit N` | Limit number of results |
| `-s, --short` | Show only ID and title (when not summarizing) |

## Benefits

### Clean Output
- Focus on the summary without information overload
- Ideal for quick understanding of search results
- Perfect for capturing the essence of multiple notes

### Flexibility
- Use `--show-details` when you need both summary and details
- Fallback to details if summarization fails
- Works with both text and vector search

### Use Cases

1. **Research Overview**: Quickly understand what notes you have on a topic
   ```bash
   ml-notes search --summarize "database optimization"
   ```

2. **Note Discovery**: Find related concepts across multiple notes
   ```bash
   ml-notes search -v --summarize "performance tuning"
   ```

3. **Knowledge Review**: Get a high-level view before diving into details
   ```bash
   ml-notes search --summarize --show-details "python best practices"
   ```

## Configuration

The summarization feature requires:
- Ollama running with a language model
- Summarization enabled in config (`enable_summarization: true`)
- A configured summarization model (default: `llama3.2:latest`)

## Error Handling

If summarization fails:
- An error message is displayed
- The command automatically falls back to showing detailed results
- This ensures you always get useful output

## Performance Notes

- Summarization adds processing time (typically 1-5 seconds)
- Larger result sets take longer to summarize
- The first request may be slower if the model needs to load
- Consider using `--limit` to reduce the amount of text being summarized