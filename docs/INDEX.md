# ML Notes Documentation Index

This directory contains comprehensive documentation for the ML Notes project - a command-line interface for intelligent note management with AI-powered search and analysis.

## 📚 Documentation Overview

### Core Documentation

#### [README.md](../README.md)
**Main project documentation** - Quick start, features overview, and basic usage examples. Start here for your first experience with ML Notes.

#### [USAGE_GUIDE.md](USAGE_GUIDE.md)
**Comprehensive usage documentation** - Detailed guide covering all commands, features, and workflows. Essential for mastering ML Notes.

#### [DESIGN.md](DESIGN.md)  
**System design and architecture** - Technical deep-dive into the system architecture, data flow, and design decisions. Important for developers and advanced users.

#### [API_INTEGRATION.md](API_INTEGRATION.md)
**Integration and API documentation** - MCP server API, Claude Desktop integration, editor integration, and custom extensions.

### Command-Specific Documentation

#### [delete-command.md](delete-command.md)
**Delete command documentation** - Comprehensive guide to safely removing notes with various options and safety features.

#### [edit-command.md](edit-command.md)  
**Edit command documentation** - Complete guide to editing notes with editor integration, change detection, and reindexing.

### Project Information

#### [PROJECT_STRUCTURE.md](../PROJECT_STRUCTURE.md)
**Project structure and organization** - Overview of the codebase structure, key components, and development workflow.

#### [CONTRIBUTING.md](../CONTRIBUTING.md)
**Contribution guidelines** - How to contribute to the ML Notes project, including development setup and coding standards.

#### [CHANGELOG.md](../CHANGELOG.md)
**Version history and changes** - Detailed changelog of features, improvements, and bug fixes across versions.

## 🎯 Quick Navigation by Use Case

### New Users
1. Start with [README.md](../README.md) for installation and quick start
2. Follow the [USAGE_GUIDE.md](USAGE_GUIDE.md) for comprehensive learning
3. Check command-specific docs as needed

### Advanced Users
1. [DESIGN.md](DESIGN.md) for architectural understanding
2. [API_INTEGRATION.md](API_INTEGRATION.md) for integration projects
3. Command-specific documentation for detailed usage

### Developers
1. [PROJECT_STRUCTURE.md](../PROJECT_STRUCTURE.md) for codebase overview
2. [DESIGN.md](DESIGN.md) for system architecture
3. [CONTRIBUTING.md](../CONTRIBUTING.md) for development guidelines

### Integrators
1. [API_INTEGRATION.md](API_INTEGRATION.md) for MCP server and APIs
2. [USAGE_GUIDE.md](USAGE_GUIDE.md) for workflow examples
3. [DESIGN.md](DESIGN.md) for technical details

## 🔍 Documentation by Feature

### Note Management
- **Creating Notes**: [USAGE_GUIDE.md#creating-notes](USAGE_GUIDE.md#creating-notes)
- **Editing Notes**: [edit-command.md](edit-command.md) + [USAGE_GUIDE.md#editing-notes](USAGE_GUIDE.md#editing-notes)
- **Deleting Notes**: [delete-command.md](delete-command.md) + [USAGE_GUIDE.md#deleting-notes](USAGE_GUIDE.md#deleting-notes)
- **Listing Notes**: [USAGE_GUIDE.md#viewing-notes](USAGE_GUIDE.md#viewing-notes)
- **Tag Management**: [USAGE_GUIDE.md#tag-management](USAGE_GUIDE.md#tag-management)

### Search & Analysis
- **Text Search**: [USAGE_GUIDE.md#text-search](USAGE_GUIDE.md#text-search)
- **Vector Search**: [USAGE_GUIDE.md#vector-search](USAGE_GUIDE.md#vector-search)
- **Tag Search**: [USAGE_GUIDE.md#tag-search](USAGE_GUIDE.md#tag-search)
- **AI Analysis**: [USAGE_GUIDE.md#ai-powered-analysis](USAGE_GUIDE.md#ai-powered-analysis)
- **Custom Prompts**: [USAGE_GUIDE.md#custom-analysis-prompts](USAGE_GUIDE.md#custom-analysis-prompts)

### Configuration
- **Initial Setup**: [USAGE_GUIDE.md#initial-setup](USAGE_GUIDE.md#initial-setup)
- **Configuration Management**: [USAGE_GUIDE.md#configuration-management](USAGE_GUIDE.md#configuration-management)
- **Editor Integration**: [USAGE_GUIDE.md#editor-integration](USAGE_GUIDE.md#editor-integration)

### Integration
- **MCP Server**: [API_INTEGRATION.md#mcp-server-api](API_INTEGRATION.md#mcp-server-api)
- **Claude Desktop**: [API_INTEGRATION.md#claude-desktop-integration](API_INTEGRATION.md#claude-desktop-integration)
- **Shell Integration**: [API_INTEGRATION.md#shell-integration](API_INTEGRATION.md#shell-integration)
- **Custom Extensions**: [API_INTEGRATION.md#custom-extensions](API_INTEGRATION.md#custom-extensions)

### Technical Details
- **Architecture**: [DESIGN.md#architecture-overview](DESIGN.md#architecture-overview)
- **Data Flow**: [DESIGN.md#data-flow](DESIGN.md#data-flow)
- **Performance**: [DESIGN.md#performance-characteristics](DESIGN.md#performance-characteristics)
- **Security**: [DESIGN.md#security-considerations](DESIGN.md#security-considerations)

## 📋 Command Reference Quick Links

| Command | Primary Documentation | Additional Resources |
|---------|----------------------|---------------------|
| `init` | [USAGE_GUIDE.md#initial-setup](USAGE_GUIDE.md#initial-setup) | [README.md#quick-start](../README.md#quick-start) |
| `add` | [USAGE_GUIDE.md#creating-notes](USAGE_GUIDE.md#creating-notes) | [DESIGN.md#note-creation-flow](DESIGN.md#note-creation-flow) |
| `list` | [USAGE_GUIDE.md#viewing-notes](USAGE_GUIDE.md#viewing-notes) | [API_INTEGRATION.md#list_notes](API_INTEGRATION.md#4-list_notes) |
| `get` | [USAGE_GUIDE.md#viewing-notes](USAGE_GUIDE.md#viewing-notes) | [API_INTEGRATION.md#get_note](API_INTEGRATION.md#3-get_note) |
| `edit` | [edit-command.md](edit-command.md) | [USAGE_GUIDE.md#editing-notes](USAGE_GUIDE.md#editing-notes) |
| `delete` | [delete-command.md](delete-command.md) | [USAGE_GUIDE.md#deleting-notes](USAGE_GUIDE.md#deleting-notes) |
| `search` | [USAGE_GUIDE.md#search--analysis](USAGE_GUIDE.md#search--analysis) | [API_INTEGRATION.md#search_notes](API_INTEGRATION.md#2-search_notes) |
| `tags` | [USAGE_GUIDE.md#tag-management](USAGE_GUIDE.md#tag-management) | [API_INTEGRATION.md#list_tags](API_INTEGRATION.md#7-list_tags) |
| `analyze` | [USAGE_GUIDE.md#ai-powered-analysis](USAGE_GUIDE.md#ai-powered-analysis) | [DESIGN.md#analysis-flow](DESIGN.md#analysis-flow) |
| `config` | [USAGE_GUIDE.md#configuration](USAGE_GUIDE.md#configuration) | [API_INTEGRATION.md#configuration-api](API_INTEGRATION.md#configuration-api) |
| `mcp` | [API_INTEGRATION.md#mcp-server-api](API_INTEGRATION.md#mcp-server-api) | [README.md#mcp-server](../README.md#mcp-server) |

## 🛠️ Development Resources

### Getting Started with Development
1. [PROJECT_STRUCTURE.md](../PROJECT_STRUCTURE.md) - Understand the codebase
2. [CONTRIBUTING.md](../CONTRIBUTING.md) - Development guidelines
3. [DESIGN.md](DESIGN.md) - Architectural patterns

### Testing & Debugging
- **Test Structure**: [DESIGN.md#testing-strategy](DESIGN.md#testing-strategy)
- **Debug Tools**: [USAGE_GUIDE.md#troubleshooting](USAGE_GUIDE.md#troubleshooting)
- **Error Handling**: [API_INTEGRATION.md#error-handling](API_INTEGRATION.md#error-handling)

### Extension Development
- **Custom Commands**: [DESIGN.md#extension-points](DESIGN.md#extension-points)
- **API Wrappers**: [API_INTEGRATION.md#custom-extensions](API_INTEGRATION.md#custom-extensions)
- **Integration Patterns**: [API_INTEGRATION.md](API_INTEGRATION.md)

## 📈 Learning Path Recommendations

### Beginner Path
1. **Installation**: [README.md#installation](../README.md#installation)
2. **Quick Start**: [README.md#quick-start](../README.md#quick-start)
3. **Basic Usage**: [USAGE_GUIDE.md#core-commands](USAGE_GUIDE.md#core-commands)
4. **Configuration**: [USAGE_GUIDE.md#configuration](USAGE_GUIDE.md#configuration)

### Intermediate Path
1. **Advanced Features**: [USAGE_GUIDE.md#ai-powered-features](USAGE_GUIDE.md#ai-powered-features)
2. **Workflow Examples**: [USAGE_GUIDE.md#advanced-usage](USAGE_GUIDE.md#advanced-usage)
3. **Editor Integration**: [edit-command.md](edit-command.md)
4. **Shell Integration**: [API_INTEGRATION.md#shell-integration](API_INTEGRATION.md#shell-integration)

### Advanced Path
1. **System Design**: [DESIGN.md](DESIGN.md)
2. **MCP Integration**: [API_INTEGRATION.md#mcp-server-api](API_INTEGRATION.md#mcp-server-api)
3. **Custom Extensions**: [API_INTEGRATION.md#custom-extensions](API_INTEGRATION.md#custom-extensions)
4. **Development**: [CONTRIBUTING.md](../CONTRIBUTING.md)

## 🔧 Troubleshooting Quick Reference

### Common Issues
| Issue | Documentation | Quick Fix |
|-------|---------------|-----------|
| Configuration problems | [USAGE_GUIDE.md#troubleshooting](USAGE_GUIDE.md#troubleshooting) | `ml-notes config show` |
| Ollama connection | [USAGE_GUIDE.md#ollama-connection-issues](USAGE_GUIDE.md#ollama-connection-issues) | `curl http://localhost:11434/api/tags` |
| Vector search errors | [USAGE_GUIDE.md#vector-search-issues](USAGE_GUIDE.md#vector-search-issues) | `ml-notes detect-dimensions && ml-notes reindex` |
| Editor problems | [edit-command.md#error-handling](edit-command.md#error-handling) | `ml-notes config set editor "nano"` |
| MCP integration | [API_INTEGRATION.md#error-handling](API_INTEGRATION.md#error-handling) | Check Claude Desktop config |

### Debug Commands
```bash
ml-notes --debug <command>          # Enable debug mode
ml-notes config show               # Check configuration
ml-notes detect-dimensions         # Validate embedding setup
ml-notes mcp                      # Test MCP server
```

## 📄 Document Status

| Document | Last Updated | Status | Completeness |
|----------|-------------|--------|--------------|
| README.md | Current | ✅ Updated | Complete |
| USAGE_GUIDE.md | Current | ✅ New | Complete |
| DESIGN.md | Current | ✅ New | Complete |
| API_INTEGRATION.md | Current | ✅ New | Complete |
| delete-command.md | Previous | ✅ Current | Complete |
| edit-command.md | Previous | ✅ Current | Complete |
| PROJECT_STRUCTURE.md | Current | ✅ Updated | Complete |

---

**Note**: This documentation is actively maintained and updated with each release. For the most current information, always refer to the specific version tag or the main branch documentation.

## 📝 Contributing to Documentation

Found an error or want to improve the documentation? Please see [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines on contributing to documentation.

Key areas where contributions are welcome:
- **Examples and use cases** - Real-world usage scenarios
- **Integration tutorials** - Step-by-step integration guides  
- **Troubleshooting guides** - Solutions to common problems
- **Performance optimization** - Tips and best practices
- **Translation** - Documentation in other languages

---

*This index was generated automatically and reflects the current state of ML Notes documentation.*