# ML Notes + Lil-Rag Docker Compose Setup

This setup allows you to run both ml-notes and lil-rag together using Docker Compose.

## Prerequisites

- Docker and Docker Compose installed
- Access to a remote Ollama instance
- Git (for cloning lil-rag during build)

## Quick Start

1. **Configure environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your Ollama endpoint and other settings
   ```

2. **Start the services:**
   ```bash
   docker-compose up -d
   ```

3. **Access the applications:**
   - ML Notes Web UI: http://localhost:21212
   - Lil-Rag Web UI: http://localhost:12121/chat

## Services

### ml-notes
- **Port**: 21212
- **Health Check**: `/api/v1/health`
- **Data Volume**: `ml-notes-app-data`
- **Config Volume**: `ml-notes-app-config`

### lil-rag
- **Port**: 12121
- **Health Check**: `/health`
- **Data Volume**: `ml-notes-lil-rag-data`

## Configuration

### Environment Variables

Copy `.env.example` to `.env` and configure:

- `OLLAMA_ENDPOINT`: Your remote Ollama instance URL
- `SUMMARIZATION_MODEL`: Model to use for AI features
- Other optional settings for debugging and auto-tagging

### Volumes

Data is persisted in Docker volumes:
- `ml-notes-app-data`: ML Notes database and user data
- `ml-notes-app-config`: ML Notes configuration files
- `ml-notes-lil-rag-data`: Lil-Rag embeddings and data

## Commands

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Rebuild after code changes
docker-compose build
docker-compose up -d

# Remove everything including volumes
docker-compose down -v

# Build with custom Ollama endpoint
OLLAMA_ENDPOINT=http://192.168.1.100:11434 docker-compose up -d
```

## Verification

After starting the services, verify they're working:

```bash
# Check service status
docker-compose ps

# Test ML Notes web interface
curl http://localhost:21212/

# Test Lil-Rag API
curl http://localhost:12121/api/health

# Access the web interfaces
# ML Notes: http://localhost:21212
# Lil-Rag Chat: http://localhost:12121/chat
```

## Pre-configured Models

Both services come pre-configured with optimized models:

### ML Notes:
- **Embedding Model**: nomic-embed-text:v1.5
- **Summarization Model**: gemma3:4b
- **Chat Model**: gemma3:4b
- **Lil-Rag URL**: http://lil-rag:12121

### Lil-Rag:
- **Embedding Model**: nomic-embed-text:v1.5
- **Chat Model**: gpt-oss:20b

## Initial Setup

After starting the services, both are ready to use with no additional configuration needed. Just ensure your Ollama instance has the required models:

```bash
# On your Ollama host, pull the required models:
ollama pull nomic-embed-text:v1.5
ollama pull gemma3:4b
ollama pull gpt-oss:20b
```

## Networking

Services communicate via the `ml-notes-network` bridge network:
- ml-notes connects to lil-rag at `http://lil-rag:12121`
- Both services can connect to your external Ollama instance

## Troubleshooting

### Check service health:
```bash
docker-compose ps
```

### View logs:
```bash
docker-compose logs ml-notes
docker-compose logs lil-rag
```

### Restart services:
```bash
docker-compose restart ml-notes
docker-compose restart lil-rag
```