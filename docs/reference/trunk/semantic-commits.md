# Semantic Commit Messages

## Overview

This repository follows the **Conventional Commits** specification for all commit messages. Semantic commits enable:

- Automatic version bumping based on commit type
- Automated changelog generation
- Clear communication of intent and impact
- Consistent commit history

## Commit Message Format

```text
<module>: <type>[optional !]: <description>

[optional body]

[optional footer(s)]
```

### Structure

1. **Module prefix**: Which module is affected (see [repository-layout.md](repository-layout.md))
2. **Type**: The kind of change being made
3. **Breaking change marker** (!): Indicates a breaking change
4. **Description**: Short summary of the change (lowercase, no period)
5. **Body**: Detailed explanation (optional)
6. **Footer**: Metadata like breaking changes, issues (optional)

## Commit Types

### Primary Types (Version Affecting)

#### `feat:` - New Feature (MINOR bump)

**Use when**: Adding new functionality, features, or capabilities

**Version Impact**: Increments MINOR version (1.2.3 → 1.3.0)

**Examples**:

```
mcp-vscode: feat: add commit-message-generate tool
vscode-ext: feat: add toolbar button for quick commits
mcp-github: feat: support pull request creation
```

#### `fix:` - Bug Fix (PATCH bump)

**Use when**: Fixing a bug, defect, or incorrect behavior

**Version Impact**: Increments PATCH version (1.2.3 → 1.2.4)

**Examples**:

```
mcp-pwsh: fix: handle null output from commands
vscode-ext: fix: prevent duplicate button registration
mcp-docs: fix: correct markdown parsing for code blocks
```

#### `perf:` - Performance Improvement (PATCH bump)

**Use when**: Improving performance without changing functionality

**Version Impact**: Increments PATCH version (1.2.3 → 1.2.4)

**Examples**:

```
mcp-github: perf: cache API responses for 5 minutes
mcp-vscode: perf: optimize file watching with debounce
```

#### `feat!:` or `fix!:` - Breaking Change (MAJOR bump)

**Use when**: Making incompatible API changes or breaking existing functionality

**Version Impact**: Increments MAJOR version (1.2.3 → 2.0.0)

**Requirements**: Must include `BREAKING CHANGE:` footer explaining the impact

**Examples**:

```
mcp-vscode: feat!: change tool response schema

BREAKING CHANGE: Tool responses now use {success, result, error}
format instead of {status, data}. Update all tool implementations.
```

```
vscode-ext: refactor!: remove deprecated executeCommand API

BREAKING CHANGE: The executeCommand method has been removed.
Use the new invokeAction method instead.
```

### Secondary Types (No Version Change)

#### `docs:` - Documentation

**Use when**: Only documentation changes, no code changes

**Version Impact**: None

**Examples**:

```
docs: update quick start guide with new screenshots
docs: add troubleshooting section for MCP servers
mcp-vscode: docs: improve tool documentation in README
```

#### `style:` - Code Style

**Use when**: Formatting, whitespace, missing semicolons (no functional change)

**Version Impact**: None

**Examples**:

```
mcp-pwsh: style: format code with gofmt
vscode-ext: style: apply prettier formatting
```

#### `refactor:` - Refactoring

**Use when**: Code changes that neither fix bugs nor add features

**Version Impact**: None (unless marked as breaking)

**Examples**:

```
mcp-github: refactor: extract HTTP client to separate package
vscode-ext: refactor: reorganize command handlers
```

#### `test:` - Tests

**Use when**: Adding or updating tests

**Version Impact**: None

**Examples**:

```
mcp-vscode: test: add unit tests for commit parser
vscode-ext: test: add integration tests for button actions
```

#### `chore:` - Maintenance

**Use when**: Changes to build process, dependencies, tooling

**Version Impact**: None

**Examples**:

```
chore: update Go dependencies
chore: upgrade TypeScript to 5.x
infra: chore: update Docker base image
```

#### `ci:` - CI/CD

**Use when**: Changes to CI/CD configuration

**Version Impact**: None

**Examples**:

```
ci: add GitHub Actions workflow for testing
ci: configure automatic releases on tag push
```

#### `build:` - Build System

**Use when**: Changes to build configuration or scripts

**Version Impact**: None

**Examples**:

```
build: update webpack configuration
build: add production build optimization
```

#### `revert:` - Revert

**Use when**: Reverting a previous commit

**Version Impact**: Depends on what's being reverted

**Format**: `revert: <header of reverted commit>`

**Examples**:

```
revert: mcp-vscode: feat: add commit-message-generate tool

This reverts commit abc123def456. The feature caused performance
issues in large repositories.
```

## Module Prefixes

Based on [repository-layout.md](repository-layout.md), use these module prefixes:

| Prefix | Module | Location |
|--------|--------|----------|
| `mcp-pwsh:` | PowerShell MCP Server | `src/mcp/pwsh/` |
| `mcp-docs:` | Documentation MCP Server | `src/mcp/docs/` |
| `mcp-github:` | GitHub MCP Server | `src/mcp/github/` |
| `mcp-vscode:` | VSCode MCP Server | `src/mcp/vscode/` |
| `vscode-ext:` | VSCode Extension | `.vscode/extensions/claude-mcp-vscode/` |
| `infra:` | Infrastructure | `automation/`, `containers/` |
| `docs:` | Documentation | `docs/`, `*.md` files |
| `config:` | Configuration | `.mcp.json`, `.gitignore`, etc. |
| `contracts:` | Contracts | `contracts/` |

### Multi-Module Commits

**Preferred Approach**: Create separate commits for each module

```
Commit 1: mcp-vscode: feat: add commit-analyze tool
Commit 2: vscode-ext: feat: integrate commit analyzer in UI
```

**Alternative**: Use comma-separated prefixes for tightly coupled changes

```
mcp-vscode,vscode-ext: feat: add commit analysis feature
```

## Description Guidelines

### DO

- Use imperative mood ("add feature" not "added feature")
- Start with lowercase
- No period at the end
- Keep under 72 characters
- Be specific and concise

### DON'T

- Use past tense ("added" or "adding")
- Capitalize first letter (unless proper noun)
- End with period
- Be vague ("fix stuff", "update code")
- Include implementation details (save for body)

### Good Examples

```
feat: add commit message validation
fix: prevent crash on empty repository
perf: optimize file parsing by 40%
docs: clarify installation requirements
```

### Bad Examples

```
feat: Added commit message validation.    ❌ (past tense, capitalized, period)
fix: Fix bug                              ❌ (too vague)
perf: Performance improvements            ❌ (not specific)
Update the docs                           ❌ (missing type)
```

## Commit Body

Optional detailed explanation of the change.

### When to Use Body

- Complex changes requiring context
- Explaining "why" not just "what"
- Multiple related changes
- Migration instructions

### Format

- Wrap at 72 characters
- Blank line after description
- Use bullet points for multiple items
- Reference issues or PRs

### Example

```
mcp-github: feat: add repository search tool

Add new tool for searching GitHub repositories with advanced filters.
Supports filtering by:
- Language
- Stars count
- Last updated date
- License type

Implements #45
```

## Commit Footer

Optional metadata at the end of the commit message.

### Common Footers

#### Breaking Changes

```
BREAKING CHANGE: <description of breaking change>
```

#### Issue References

```
Fixes #123
Closes #456
Refs #789
```

#### Co-authors

```
Co-authored-by: Name <email@example.com>
```

#### Reviewed By

```
Reviewed-by: Name <email@example.com>
```

### Example with Multiple Footers

```
mcp-vscode: feat!: migrate to async tool interface

All tools now return Promises to support async operations.
This enables better error handling and streaming responses.

BREAKING CHANGE: Synchronous tool return values are no longer
supported. All tools must return Promise<ToolResult>.

Migration guide:
- Change `return {data}` to `return Promise.resolve({data})`
- Add async/await to tool implementations
- Update error handling to use Promise rejection

Fixes #123
Reviewed-by: John Doe <john@example.com>
```

## Version Increment Matrix

| Commit Type | Breaking (!) | Version Change | Example |
|-------------|-------------|----------------|---------|
| `feat:` | No | MINOR | 1.2.3 → 1.3.0 |
| `feat!:` | Yes | MAJOR | 1.2.3 → 2.0.0 |
| `fix:` | No | PATCH | 1.2.3 → 1.2.4 |
| `fix!:` | Yes | MAJOR | 1.2.3 → 2.0.0 |
| `perf:` | No | PATCH | 1.2.3 → 1.2.4 |
| `perf!:` | Yes | MAJOR | 1.2.3 → 2.0.0 |
| `refactor!:` | Yes | MAJOR | 1.2.3 → 2.0.0 |
| `docs:` | - | None | - |
| `style:` | - | None | - |
| `refactor:` | No | None | - |
| `test:` | - | None | - |
| `chore:` | - | None | - |
| `ci:` | - | None | - |
| `build:` | - | None | - |

## Complete Examples

### Example 1: Simple Feature

```
mcp-pwsh: feat: add output formatting options

Allow customization of command output format via new parameters.
```

**Version**: 1.2.0 → 1.3.0 (MINOR)

### Example 2: Bug Fix with Context

```
vscode-ext: fix: prevent duplicate toolbar buttons on reload

The extension was registering toolbar buttons multiple times when
the window was reloaded, causing visual duplication. Now properly
disposes of existing buttons before registering new ones.

Fixes #234
```

**Version**: 1.5.3 → 1.5.4 (PATCH)

### Example 3: Breaking Change

```
mcp-github: feat!: standardize error response format

All API errors now return consistent error objects with code,
message, and details fields.

BREAKING CHANGE: Error responses changed from string messages to
structured objects. Update error handling code:

Before: catch(err => console.log(err))
After:  catch(err => console.log(err.message))

Refs #567
```

**Version**: 2.3.1 → 3.0.0 (MAJOR)

### Example 4: Documentation Only

```
docs: add VSCode extension usage examples

Include screenshots and step-by-step guide for first-time users.
```

**Version**: No change

### Example 5: Multi-Module Feature

```
Commit 1:
mcp-vscode: feat: add semantic commit parser tool

Implement tool for parsing and validating semantic commit messages.
Returns structured data with module, type, breaking flag, and scope.

Commit 2:
vscode-ext: feat: integrate semantic commit parser

Add UI for generating semantic commit messages using the new parser.
Validates messages in real-time and shows version impact.

Both commits increment MINOR version of their respective modules.
```

## Commit Message Checklist

Before committing, verify:

- [ ] Module prefix matches affected files
- [ ] Type accurately reflects the change
- [ ] Breaking changes marked with `!` and footer
- [ ] Description is imperative, lowercase, no period
- [ ] Description is specific and under 72 characters
- [ ] Body explains "why" if change is complex
- [ ] Footer includes issue references if applicable
- [ ] No merge/rebase artifacts in message

## Tools for Validation

### Git Hooks

Set up `commit-msg` hook to validate format:

```bash
#!/bin/sh
# .git/hooks/commit-msg

commit_msg=$(cat "$1")
pattern="^(mcp-pwsh|mcp-docs|mcp-github|mcp-vscode|vscode-ext|infra|docs|config|contracts): (feat|fix|perf|docs|style|refactor|test|chore|ci|build|revert)(!)?:"

if ! echo "$commit_msg" | grep -Eq "$pattern"; then
  echo "ERROR: Commit message does not follow semantic commit format"
  echo "Format: <module>: <type>[!]: <description>"
  exit 1
fi
```

### VSCode Extension

The VSCode extension in this repo automatically generates semantic commits following these conventions.

## References

- [Conventional Commits 1.0.0](https://www.conventionalcommits.org/)
- [Semantic Versioning 2.0.0](https://semver.org/)
- [Repository Layout](repository-layout.md)
- [Versioning Documentation](versioning.md)
- [Angular Commit Guidelines](https://github.com/angular/angular/blob/main/CONTRIBUTING.md#commit)
