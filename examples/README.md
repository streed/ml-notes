# ML Notes HTTP API Integration Examples

This directory contains examples for integrating ML Notes with external tools via the HTTP API.

## Quick Start

1. **Start the HTTP API server:**
   ```bash
   ./ml-notes serve --host 0.0.0.0 --port 8080
   ```

2. **Test the API:**
   ```bash
   curl http://localhost:8080/api/v1/health
   ```

## Available Integrations

### 🌐 OpenWebUI Integration

**File:** `openwebui_integration.py`

Provides custom functions for OpenWebUI that enable:
- Creating notes with auto-tagging
- Searching notes with vector similarity
- Listing recent notes
- Bulk auto-tagging of notes

**Setup:**
1. Copy the functions from the Python file
2. Paste them into your OpenWebUI Functions section
3. Update the `ML_NOTES_BASE_URL` to match your setup
4. Start using in chat: "Create a note about machine learning"

### 🦙 Ollama Tool Integration

**File:** `ollama_tool_integration.json`

JSON configuration for Ollama tools integration that provides:
- Function definitions for all note operations
- Endpoint mappings for the HTTP API
- Usage examples for common scenarios

**Features:**
- Vector-powered search for RAG applications
- Intelligent auto-tagging with AI
- Full CRUD operations on notes
- Statistics and monitoring

## API Endpoints

### Core Operations
- `GET /api/v1/notes` - List notes
- `POST /api/v1/notes` - Create note (with auto-tagging)
- `GET /api/v1/notes/{id}` - Get specific note
- `PUT /api/v1/notes/{id}` - Update note
- `DELETE /api/v1/notes/{id}` - Delete note

### Search & Discovery
- `POST /api/v1/notes/search` - Search with vector/text
- `GET /api/v1/tags` - List all tags

### AI Features
- `POST /api/v1/auto-tag/suggest/{id}` - Get tag suggestions
- `POST /api/v1/auto-tag/apply` - Batch auto-tagging

### System
- `GET /api/v1/health` - Health check
- `GET /api/v1/stats` - Database statistics
- `GET /api/v1/config` - Configuration
- `GET /api/v1/docs` - API documentation

## Example API Calls

### Create a note with auto-tagging:
```bash
curl -X POST http://localhost:8080/api/v1/notes \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Machine Learning Basics",
    "content": "Neural networks, deep learning, and AI fundamentals",
    "auto_tag": true
  }'
```

### Search notes:
```bash
curl -X POST http://localhost:8080/api/v1/notes/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "machine learning",
    "limit": 5,
    "use_vector": true
  }'
```

### Auto-tag recent notes:
```bash
curl -X POST http://localhost:8080/api/v1/auto-tag/apply \
  -H "Content-Type: application/json" \
  -d '{
    "recent": 5,
    "apply": true,
    "overwrite": false
  }'
```

## Configuration

### Server Configuration
```bash
# Start on all interfaces (for network access)
./ml-notes serve --host 0.0.0.0 --port 8080

# Start on localhost only (local access)
./ml-notes serve --host localhost --port 8080
```

### Security Considerations
- The API currently has no authentication (suitable for local/trusted networks)
- CORS is enabled for all origins (configure restrictively in production)
- Consider adding authentication for production deployments

### Integration Tips

1. **For RAG Applications:**
   - Use vector search for semantic similarity
   - Auto-tag notes for better organization
   - Use search results as context for LLM responses

2. **For Note Management:**
   - Enable auto-tagging for intelligent organization
   - Use batch operations for efficiency
   - Monitor via health and stats endpoints

3. **For Development:**
   - Use the `/docs` endpoint for API reference
   - Check `/health` for service status
   - Monitor `/stats` for usage metrics

## Troubleshooting

### Server won't start:
- Check if port is already in use: `netstat -tulpn | grep 8080`
- Verify ml-notes database is accessible
- Check logs with `--debug` flag

### API calls fail:
- Verify server is running: `curl http://localhost:8080/api/v1/health`
- Check network connectivity and firewall settings
- Ensure correct Content-Type headers for POST requests

### Auto-tagging not working:
- Verify Ollama is running and accessible
- Check auto-tagging is enabled: `./ml-notes config show`
- Ensure summarization model is configured

For more help, see the main ML Notes documentation or check the server logs.