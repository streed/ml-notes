# Release Process

This document outlines the release process for ML Notes.

## Versioning

We follow [Semantic Versioning](https://semver.org/):
- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions
- **PATCH** version for backwards-compatible bug fixes

## Release Checklist

### Pre-release

1. **Update version number** in relevant files
2. **Update CHANGELOG.md** with release notes
3. **Run full test suite**:
   ```bash
   make test
   make lint
   ```
4. **Build all platforms locally**:
   ```bash
   make build-all
   ```
5. **Test installation script**:
   ```bash
   ./install.sh --version local
   ```

### Creating a Release

1. **Create and push a tag**:
   ```bash
   git tag -a v1.0.0 -m "Release version 1.0.0"
   git push origin v1.0.0
   ```

2. **Create GitHub Release**:
   - Go to [Releases page](https://github.com/streed/ml-notes/releases)
   - Click "Draft a new release"
   - Select the tag you created
   - Title: "ML Notes v1.0.0"
   - Generate release notes
   - Add highlights and breaking changes

3. **GitHub Actions will automatically**:
   - Run tests on multiple platforms
   - Build binaries for all platforms
   - Upload release artifacts
   - Update release page

### Post-release

1. **Verify release artifacts**:
   - Download and test each platform binary
   - Test installation script with new version

2. **Update documentation**:
   - Update README if needed
   - Update wiki/documentation site

3. **Announce release**:
   - Social media
   - Community forums
   - Mailing list (if applicable)

## Release Notes Template

```markdown
# ML Notes v1.0.0

## 🎉 Highlights

- Major feature or improvement
- Another significant change

## ✨ New Features

- Feature 1 (#PR)
- Feature 2 (#PR)

## 🐛 Bug Fixes

- Fix 1 (#PR)
- Fix 2 (#PR)

## 🔧 Improvements

- Improvement 1 (#PR)
- Improvement 2 (#PR)

## 📚 Documentation

- Doc update 1 (#PR)
- Doc update 2 (#PR)

## ⚠️ Breaking Changes

- Breaking change description

## 🔄 Migration Guide

If there are breaking changes, provide migration instructions.

## 📦 Installation

### Using the install script:
\```bash
curl -sSL https://raw.githubusercontent.com/streed/ml-notes/main/install.sh | bash
\```

### Manual download:
Download the appropriate binary for your platform from the [releases page](https://github.com/streed/ml-notes/releases/tag/v1.0.0).

## 🙏 Contributors

Thanks to all contributors who made this release possible!

@contributor1, @contributor2, ...

## 📊 Stats

- X commits
- Y files changed
- Z contributors

**Full Changelog**: https://github.com/streed/ml-notes/compare/v0.9.0...v1.0.0
```

## Platform Support

| Platform | Architecture | Binary Name |
|----------|-------------|-------------|
| Linux | AMD64 | ml-notes-linux-amd64 |
| Linux | ARM64 | ml-notes-linux-arm64 |
| macOS | AMD64 | ml-notes-darwin-amd64 |
| macOS | ARM64 | ml-notes-darwin-arm64 |
| Windows | AMD64 | ml-notes-windows-amd64.exe |

## Troubleshooting Releases

### Build Failures

If GitHub Actions build fails:
1. Check the workflow logs
2. Ensure all tests pass locally
3. Verify Go version compatibility

### Missing Artifacts

If release artifacts are missing:
1. Check GitHub Actions completed successfully
2. Manually build and upload if needed:
   ```bash
   make release
   # Upload files from dist/release/
   ```

### Installation Script Issues

Test the installation script locally:
```bash
./install.sh --platform linux-amd64 --version v1.0.0
```

## Security Releases

For security updates:
1. Follow responsible disclosure
2. Prepare patches privately
3. Release simultaneously with disclosure
4. Update all supported versions
5. Notify users via security advisory

## Rollback Procedure

If a release has critical issues:
1. Mark the release as pre-release on GitHub
2. Create a patch release with fixes
3. Update installation script to skip bad version
4. Notify users of the issue

## Automation

The release process is largely automated via GitHub Actions:
- Tests run on every push
- Binaries build on tag push
- Release artifacts upload automatically

See `.github/workflows/ci.yml` for details.