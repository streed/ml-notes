# Edit Command

## Overview
The edit command allows you to modify existing notes using your preferred text editor. It automatically detects changes and reindexes notes for vector search when content is modified.

## Usage

### Basic Syntax
```bash
ml-notes edit <note-id> [flags]
```

## Examples

### Edit Full Note
```bash
ml-notes edit 123
```
This opens the note in your editor with the format:
```
Title: My Note Title
---
This is the content of my note.
Multiple lines are supported.
```

### Edit Title Only
```bash
ml-notes edit -t 123
# or
ml-notes edit --title 123
```
Opens only the title in the editor.

### Edit Content Only
```bash
ml-notes edit -c 123
# or
ml-notes edit --content 123
```
Opens only the content in the editor, preserving the title.

### Specify Editor
```bash
# Use VS Code
ml-notes edit -e "code --wait" 123

# Use nano
ml-notes edit -e nano 123

# Use vim with specific options
ml-notes edit -e "vim -c 'set wrap'" 123
```

## Editor Selection

The editor is chosen in this order of precedence:

1. `--editor` flag (if provided)
2. `$EDITOR` environment variable
3. `$VISUAL` environment variable
4. Auto-detection of common editors:
   - vim
   - vi
   - nano
   - emacs
   - code (VS Code)
   - subl (Sublime Text)
   - atom

### Setting Default Editor
```bash
# Set for current session
export EDITOR="vim"

# Set permanently (add to ~/.bashrc or ~/.zshrc)
echo 'export EDITOR="vim"' >> ~/.bashrc

# For VS Code (needs --wait flag)
export EDITOR="code --wait"
```

## Edit Format

### Full Note Format
When editing a full note, the editor shows:
```
Title: [Current Title]
---
[Current Content]
```

- The first line must start with `Title: `
- The `---` separator divides title from content
- Everything after `---` becomes the note content
- Blank lines are preserved in content

### Title-Only Format
When using `-t`, only the title text appears in the editor.

### Content-Only Format
When using `-c`, only the content text appears in the editor.

## Change Detection

The edit command uses MD5 hashing to detect changes:
- If no changes are made, the note is not updated
- Only modified notes trigger database updates
- Vector search reindexing occurs only when content changes

## Reindexing

When vector search is enabled and content changes:
1. The note is updated in the database
2. The full text (title + content) is reindexed
3. Vector embeddings are regenerated
4. The search index is updated

Output example:
```
Reindexing note for vector search...
✓ Note reindexed successfully

✓ Note updated successfully
  Title changed from: Old Title
                  to: New Title
  Content increased by 150 characters
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--title` | `-t` | Edit title only |
| `--content` | `-c` | Edit content only |
| `--editor` | `-e` | Specify editor to use |

## Error Handling

### Note Not Found
```bash
ml-notes edit 999
```
Output:
```
Error: failed to get note 999: note not found
```

### No Editor Available
If no editor is found:
```
Error: no editor found. Set $EDITOR environment variable or use --editor flag
```

### Editor Fails
If the editor crashes or returns an error:
```
Error: failed to run editor vim: exit status 1
```

## Best Practices

### 1. Configure Your Editor
Set up your preferred editor with appropriate flags:
```bash
# VS Code - must wait for close
export EDITOR="code --wait"

# Vim with line wrapping
export EDITOR="vim -c 'set wrap'"

# Emacs in terminal mode
export EDITOR="emacs -nw"
```

### 2. Use Specific Flags
When you only need to change one part:
```bash
# Quick title fix
ml-notes edit -t 42

# Update content only
ml-notes edit -c 42
```

### 3. Preview Before Editing
Check the current content first:
```bash
ml-notes get 123
ml-notes edit 123
```

### 4. Backup Important Notes
Before major edits:
```bash
# Export note first
ml-notes get 123 > backup-note-123.md

# Then edit
ml-notes edit 123
```

## Integration with Other Commands

### Search and Edit
```bash
# Find notes
ml-notes search "TODO"

# Edit specific result
ml-notes edit 45
```

### List and Edit
```bash
# List recent notes
ml-notes list --limit 5

# Edit one
ml-notes edit 3
```

### Get, Edit, and Verify
```bash
# View current state
ml-notes get 100

# Make changes
ml-notes edit 100

# Verify changes
ml-notes get 100
```

## Temporary Files

The edit command creates temporary files in the system temp directory:
- Format: `ml-notes-{note-id}-*.md` for full notes
- Format: `ml-notes-{note-id}-title-*.txt` for title only
- Format: `ml-notes-{note-id}-content-*.txt` for content only
- Files are automatically deleted after editing

## Vector Search Impact

When vector search is enabled:
- Each edit triggers reindexing
- Embeddings are regenerated using the configured model
- Search results will reflect the updated content
- This process adds a small delay after saving

## Tips and Tricks

### Quick Edits
For quick edits, use a lightweight editor:
```bash
ml-notes edit -e nano 123
```

### Syntax Highlighting
Use `.md` extension awareness:
- The temp file has `.md` extension
- Most editors will apply Markdown syntax highlighting

### Multi-Line Editing
The content section preserves:
- Line breaks
- Indentation  
- Special characters
- Unicode text

### Aborting Edits
To cancel editing without saving:
- In vim: `:q!`
- In nano: `Ctrl+X` then `N`
- In VS Code: Close without saving
- The original note remains unchanged