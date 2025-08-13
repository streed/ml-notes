# Delete Command

## Overview
The delete command allows you to remove notes from your database. It supports deleting single notes, multiple notes, or all notes with safety features to prevent accidental data loss.

## Usage

### Basic Syntax
```bash
ml-notes delete [note IDs...] [flags]
```

### Aliases
- `ml-notes rm [note IDs...]`
- `ml-notes remove [note IDs...]`

## Examples

### Delete a Single Note
```bash
ml-notes delete 123
```
Output:
```
The following notes will be deleted:
------------------------------------------------------------
  [123] My Important Note
------------------------------------------------------------
Are you sure you want to delete this note? (y/N): y
✓ Deleted note 123: My Important Note
============================================================
Successfully deleted 1 note(s).
```

### Delete Multiple Notes
```bash
ml-notes delete 10 20 30
```
Output:
```
The following notes will be deleted:
------------------------------------------------------------
  [10] First Note
  [20] Second Note
  [30] Third Note
------------------------------------------------------------
Are you sure you want to delete 3 notes? (y/N): y
✓ Deleted note 10: First Note
✓ Deleted note 20: Second Note
✓ Deleted note 30: Third Note
============================================================
Successfully deleted 3 note(s).
```

### Skip Confirmation
```bash
ml-notes delete -f 456
# OR
ml-notes delete --force 456
```
This immediately deletes the note without asking for confirmation.

### Delete All Notes
```bash
# No IDs needed with --all flag
ml-notes delete --all

# This will error - cannot mix --all with IDs
ml-notes delete --all 123  # ❌ Error!
```
Output:
```
⚠️  WARNING: This will delete ALL 42 notes!
This action cannot be undone.
============================================================
Type 'DELETE ALL 42 NOTES' to confirm: DELETE ALL 42 NOTES
✓ Deleted note 1: ...
✓ Deleted note 2: ...
[...]
============================================================
Successfully deleted all 42 notes.
```

## Safety Features

### Confirmation Prompt
By default, the delete command asks for confirmation before deleting:
- Single note: "Are you sure you want to delete this note? (y/N):"
- Multiple notes: "Are you sure you want to delete X notes? (y/N):"

### Delete All Protection
When using `--all`, you must type an exact confirmation phrase:
- The phrase includes the exact number of notes to be deleted
- Even with `--force`, delete all requires confirmation
- This prevents accidental deletion of entire database

### Invalid IDs
If you specify a note ID that doesn't exist:
- A warning is displayed
- The note is skipped
- Other valid notes are still processed

Example:
```bash
ml-notes delete 1 999 2
```
Output:
```
Warning: Note with ID 999 not found
The following notes will be deleted:
------------------------------------------------------------
  [1] First Note
  [2] Second Note
------------------------------------------------------------
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Skip confirmation prompt (except for --all) |
| `--all` | | Delete all notes (requires special confirmation) |

## Vector Index Cleanup

If vector search is enabled, deleting a note automatically:
- Removes the note's embedding from the vector index
- Cleans up the vec_notes table entry
- This happens via CASCADE constraints in the database

## Error Handling

### Note Not Found
```bash
ml-notes delete 999
```
Output:
```
Warning: Note with ID 999 not found
No valid notes to delete.
```

### Mixed Success
If some deletions fail:
```bash
ml-notes delete 1 2 3  # If note 2 fails
```
Output:
```
✓ Deleted note 1: First Note
✗ Failed to delete note 2: database error
✓ Deleted note 3: Third Note
============================================================
Deleted 2 note(s), failed to delete 1 note(s).
```

## Best Practices

1. **Always verify IDs before deletion**
   ```bash
   ml-notes get 123  # Check the note first
   ml-notes delete 123  # Then delete if correct
   ```

2. **Use list to find notes to delete**
   ```bash
   ml-notes list --limit 5
   ml-notes delete 1 2 3  # Delete based on list
   ```

3. **Be extremely careful with --all**
   - Consider backing up your database first
   - The confirmation text must be typed exactly
   - Cannot be undone

4. **Use force flag sparingly**
   - Good for scripts and automation
   - Risky for interactive use
   - Still requires confirmation for --all

## Integration with Other Commands

### Find and Delete
```bash
# Search for notes
ml-notes search "old project"

# Delete specific results
ml-notes delete 45 67 89
```

### List and Clean
```bash
# List old notes
ml-notes list --limit 50

# Delete unwanted ones
ml-notes delete 1 2 3 4 5
```

## Database Integrity

The delete operation:
- Uses database transactions for consistency
- Cascades to related tables (embeddings, etc.)
- Cannot be rolled back once confirmed
- Maintains referential integrity