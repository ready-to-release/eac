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

## References

- [Semantic Commits Guide](./semantic-commits.md)
- [Trunk-Based Development](../../explanation/continuous-delivery/trunk-based-development.md)
