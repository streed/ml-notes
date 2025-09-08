# GitHub Workflow Behavior

## Updated CI Workflow Behavior

### Pull Requests to `main`
- ✅ **Tests**: Run Go tests with race detection and coverage
- ✅ **Linting**: Run golangci-lint for code quality checks  
- ❌ **Build**: No binaries are built for PRs

### Pushes to `main`
- ✅ **Tests**: Run Go tests with race detection and coverage
- ✅ **Linting**: Run golangci-lint for code quality checks
- ✅ **Build**: Build binaries for all supported platforms (Linux, macOS, Windows)

## Release Workflows

### Automatic Release (`release.yml`)
- **Trigger**: Push to `main` branch
- **Condition**: Only builds if changes detected since last tag
- **Output**: Creates GitHub release with binaries and packages

### Manual Release (`manual-release.yml`)
- **Trigger**: Manual workflow dispatch
- **Output**: Creates GitHub release with binaries and packages

## Summary

This configuration ensures that:
1. **Pull requests are fast** - Only run tests and linting, no time-consuming builds
2. **Resource efficient** - No unnecessary binary builds during development 
3. **Quality maintained** - All code changes are still tested and linted before merge
4. **Releases work** - Binaries are still built when changes are merged to main
5. **Manual control** - Manual releases are available when needed

The build job now includes the condition:
```yaml
if: github.event_name == 'push' && github.ref == 'refs/heads/main'
```

This ensures binaries are only built when code is merged into the main branch, not during the PR review process.