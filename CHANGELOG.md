# Changelog

All notable changes to ML Notes will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of ML Notes
- Command-line interface for note management
- Vector similarity search using sqlite-vec
- Ollama integration for embeddings
- Nomic embedding model support with proper formatting
- Configuration management system
- Debug mode for troubleshooting
- Automatic dimension detection and correction
- Reindexing capability for vector updates
- Interactive initialization wizard
- Comprehensive documentation

### Features
- `add` - Add new notes with title and content
- `list` - List notes with pagination
- `get` - Retrieve specific notes by ID
- `search` - Text and vector similarity search
- `init` - Initialize configuration
- `config` - Manage configuration settings
- `reindex` - Rebuild vector embeddings
- `detect-dimensions` - Detect embedding model dimensions

### Technical
- Built with Go 1.22+
- SQLite database with sqlite-vec extension
- Embedded vector database (no external dependencies)
- Support for multiple embedding models
- Automatic dimension mismatch handling
- Cross-platform support (Linux, macOS, Windows)

## [1.0.0] - TBD

### Added
- First stable release
- Production-ready vector search
- Comprehensive test coverage
- CI/CD pipeline with GitHub Actions
- Installation script for easy setup
- Makefile for building and installation
- Contributing guidelines
- MIT License

### Changed
- Improved error handling and logging
- Optimized vector search performance
- Enhanced configuration validation

### Fixed
- Dimension mismatch handling
- Ollama connection resilience
- Database initialization issues

## Development Versions

### [0.9.0-beta] - Development
- Beta testing phase
- Core functionality implementation
- Initial vector search capabilities

### [0.1.0-alpha] - Development
- Initial prototype
- Basic CRUD operations
- SQLite integration

---

## Version History Format

### [VERSION] - YYYY-MM-DD

#### Added
- New features

#### Changed
- Changes in existing functionality

#### Deprecated
- Soon-to-be removed features

#### Removed
- Removed features

#### Fixed
- Bug fixes

#### Security
- Security updates