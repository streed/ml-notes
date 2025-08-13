# Thinking Tag Filtering

## Overview
The ML Notes summarization feature automatically filters out `<think>` and `</think>` tags from LLM responses. These tags are sometimes used by language models for internal reasoning but should not be displayed to end users.

## Implementation
The filtering is implemented in the `internal/summarize/summarize.go` file through the `cleanThinkingTags` function.

## Features

### What Gets Removed
1. **Complete thinking blocks**: Any content between `<think>` and `</think>` tags
2. **Nested thinking tags**: Handles nested or malformed tag structures
3. **Standalone tags**: Removes orphaned `<think>` or `</think>` tags
4. **Excessive whitespace**: Collapses multiple consecutive newlines left after tag removal

### Examples

#### Input:
```
This is the summary. <think>Internal reasoning here.</think> Final conclusion.
```

#### Output:
```
This is the summary.  Final conclusion.
```

#### Multi-line Example Input:
```
Start of summary.
<think>
Line 1 of thinking
Line 2 of thinking
</think>
End of summary.
```

#### Output:
```
Start of summary.

End of summary.
```

## Technical Details

### Regular Expression Pattern
The primary pattern used: `(?s)<think>.*?</think>`
- `(?s)`: Enables single-line mode (dot matches newlines)
- `.*?`: Non-greedy match of any content
- Removes the entire block including tags

### Additional Cleanup
1. Removes any remaining standalone tags
2. Collapses multiple consecutive newlines (3+ becomes 2)
3. Trims final whitespace

## Testing
Comprehensive tests are included in `internal/summarize/summarize_test.go`:
- Tests various tag configurations
- Tests nested and malformed tags
- Tests whitespace handling
- Integration test with mock Ollama server

## When This Applies
This filtering is automatically applied to:
- Note summarization (`ml-notes summarize`)
- Multi-note summarization
- Text summarization
- Any response from the Ollama API through the summarize package

## Benefits
1. **Cleaner output**: Users see only the final summary without internal reasoning
2. **Professional presentation**: Removes implementation details from user-facing content
3. **Compatibility**: Works with various LLM models that might use thinking tags
4. **Robustness**: Handles various tag formats and edge cases