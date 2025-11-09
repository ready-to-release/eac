# Versioning Schemes

## Overview

This repository uses **Semantic Versioning 2.0.0** (SemVer) for all modules. Each module is versioned independently based on the semantic commit messages used.

## Semantic Versioning Format

All versions follow the format: `MAJOR.MINOR.PATCH`

```text
MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]

Examples:
1.0.0
2.3.5
1.0.0-alpha.1
2.1.3-beta.2+20250124
```

### Version Components

#### MAJOR (X.0.0)

**When to Increment**: Breaking changes that are not backward compatible

**Triggers**:

- Commit type: `feat!:` or `fix!:` (with breaking change marker)
- `BREAKING CHANGE:` in commit body or footer
- API changes that break existing integrations
- Removal of features or functions
- Changes to data formats that require migration

**Examples**:

```text
src-mcp-vscode: feat!: change tool response format to new schema
vscode-extension: refactor!: remove deprecated command API

BREAKING CHANGE: The tool response format has changed from
{status, data} to {success, result, error}
```

#### MINOR (0.X.0)

**When to Increment**: New features added in a backward-compatible manner

**Triggers**:

- Commit type: `feat:`
- New functionality added
- New endpoints, tools, or commands added
- New optional parameters added
- Enhancements that don't break existing usage

**Examples**:

```text
src-mcp-github: feat: add repository search tool
vscode-extension: feat: add commit history viewer
```

#### PATCH (0.0.X)

**When to Increment**: Backward-compatible bug fixes and minor improvements

**Triggers**:

- Commit type: `fix:`
- Commit type: `perf:`
- Bug fixes
- Performance improvements
- Security patches (non-breaking)
- Documentation updates (optional)
- Refactoring (internal changes only)

**Examples**:

```text
src-mcp-github: fix: handle null responses from GitHub API
src-mcp-commands: perf: optimize command execution performance
```

### No Version Change

These commit types do not increment version numbers:

- `docs:` - Documentation-only changes
- `chore:` - Build process, tooling, dependencies
- `style:` - Code style, formatting (no functional change)
- `test:` - Adding or updating tests
- `ci:` - CI/CD configuration changes
- `build:` - Build system changes

## Module-Specific Versioning

Each module maintains its own independent version number:

### MCP Servers (Go Modules)

**Modules**:

- `src-mcp-github`
- `src-mcp-commands`

**Version Location**:

- Not stored in `go.mod` (Go uses git tags)
- Version tracked via git tags: `src-mcp-github/v1.2.3`
- Version may be embedded in build using ldflags

**Tagging Convention**:

```bash
# Tag format: <module-name>/v<version>
git tag src-mcp-github/v1.2.0
git tag src-mcp-commands/v2.0.1
```

### VSCode Extension

**Module**: `vscode-extension`

**Version Location**: `.vscode/extensions/vscode-ext-commit/package.json`

**Update Process**:

1. Commit changes with semantic commit message
2. Update `version` field in `package.json`
3. Create git tag: `vscode-extension/v<version>`

**Example**:

```json
{
  "name": "vscode-ext-commit",
  "version": "1.3.0",
  ...
}
```

### Documentation

**Module**: `docs`

**Versioning**: Documentation typically follows the highest version of the modules it documents, but does not have its own version number.

## Version Increment Rules

### Automatic Version Calculation

When creating a commit, the version increment is determined by:

1. **Parse commit message** for semantic type
2. **Identify affected module** from file paths
3. **Determine increment type**:
   - `feat!:` or `BREAKING CHANGE:` → MAJOR
   - `feat:` → MINOR
   - `fix:` or `perf:` → PATCH
   - Other types → No increment

### Multi-Module Changes

When changes affect multiple modules, each module gets its own version increment:

**Example**: Adding a new feature that requires both server and extension changes

```text
Commit 1: src-mcp-vscode: feat: add commit-analyze tool
  → src-mcp-vscode: 1.2.0 → 1.3.0 (MINOR)

Commit 2: vscode-extension: feat: add commit analysis UI
  → vscode-extension: 2.1.0 → 2.2.0 (MINOR)
```

## Pre-Release Versions

For features under development or testing:

### Format

```text
<version>-<pre-release-label>.<number>

Examples:
1.3.0-alpha.1
2.0.0-beta.3
1.5.0-rc.1
```

### Pre-Release Labels

- `alpha` - Early development, unstable
- `beta` - Feature complete, testing phase
- `rc` (Release Candidate) - Final testing before release

### Incrementing Pre-Releases

```text
1.3.0-alpha.1  → First alpha of 1.3.0
1.3.0-alpha.2  → Second alpha of 1.3.0
1.3.0-beta.1   → First beta (after alpha phase)
1.3.0-rc.1     → Release candidate
1.3.0          → Final release
```

## Build Metadata

Optional build information can be added:

```text
1.3.0+20250124
1.3.0-beta.1+exp.sha.5114f85
```

Build metadata:

- Does not affect version precedence
- Used for CI/CD tracking
- Often includes: date, commit hash, build number

## Version History and Changelog

### Git Tags

All versions are tagged in git:

```bash
# List all tags for a module
git tag -l "src-mcp-vscode/*"

# Show version history
git log --oneline --decorate
```

### Changelog Generation

Changelogs are generated from commit messages:

```markdown
## [1.3.0] - 2025-01-24

### Added
- New commit-analyze tool for semantic analysis

### Fixed
- Handle null outputs from PowerShell commands

### Changed
- Improved error messages for API failures
```

## Version Compatibility

### Semantic Version Ranges

When specifying dependencies:

- `^1.2.3` - Compatible with 1.2.3 up to (but not including) 2.0.0
- `~1.2.3` - Compatible with 1.2.3 up to (but not including) 1.3.0
- `>=1.2.3 <2.0.0` - Explicit range
- `1.2.x` - Patch-level changes only

### Module Compatibility Matrix

| VSCode Extension | src-mcp-github | src-mcp-commands |
| ---------------- | -------------- | ---------------- |
| 1.x.x            | ^1.0.0         | ^1.0.0           |
| 2.x.x            | ^1.0.0         | ^1.0.0           |

## Examples

### Example 1: Bug Fix

```text
Current version: 1.2.3

Commit: src-mcp-github: fix: handle empty API responses

New version: 1.2.4 (PATCH increment)
Tag: src-mcp-github/v1.2.4
```

### Example 2: New Feature

```text
Current version: 1.2.4

Commit: src-mcp-github: feat: add pull request creation tool

New version: 1.3.0 (MINOR increment)
Tag: src-mcp-github/v1.3.0
```

### Example 3: Breaking Change

```text
Current version: 1.3.0

Commit: src-mcp-vscode: feat!: change tool interface to support async operations

BREAKING CHANGE: All tools must now return Promises instead of
synchronous values.

New version: 2.0.0 (MAJOR increment)
Tag: src-mcp-vscode/v2.0.0
```

### Example 4: Multi-Module Update

```text
Commits:
1. src-mcp-vscode: feat: add file-watch tool
   → src-mcp-vscode: 2.0.0 → 2.1.0

2. vscode-extension: feat: integrate file watching in UI
   → vscode-extension: 1.5.0 → 1.6.0
```

## Version Query Commands

```bash
# Get current version of a module
git describe --tags --match "src-mcp-vscode/*" --abbrev=0

# Get all versions for a module
git tag -l "src-mcp-vscode/*" --sort=-version:refname

# Check what version a commit belongs to
git describe --tags --match "src-mcp-vscode/*" <commit-hash>
```

## References

- [Semantic Commits Guide](./semantic-commits.md)
- [Repository Layout](./repository-layout.md)
- [Trunk-Based Development](../../explanation/continuous-delivery/trunk-based-development.md)
