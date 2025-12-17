# Changelog Format Specification
Detailed specification of the Keep a Changelog format used in the repository.

## Overview

All changelogs in this repository follow the [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/) format, which provides a standardized structure for documenting changes.

**Standard URL:** [https://keepachangelog.com/en/1.1.0/](https://keepachangelog.com/en/1.1.0/)

## Format Standard

Keep a Changelog is based on these principles:

1. **Human-readable** - Written for humans, not machines
2. **Chronological** - Newest changes first
3. **Categorized** - Changes grouped by type
4. **Versioned** - Each release has a version entry
5. **Linked** - Versions link to diffs in version control

## Changelog Structure

### Header

```markdown
# Changelog

All notable changes to **{project/module name}** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
```

**Components:**

- Title: `# Changelog`
- Scope statement: What the changelog covers
- Format attribution: Links to Keep a Changelog
- Versioning statement: Links to Semantic Versioning

### Unreleased Section

```markdown
## [Unreleased]
```

**Purpose:** Tracks work in progress not yet in a release

**Usage:**

- Add changes here as they're developed
- Move to versioned section when releasing
- Helps draft release notes

**Content:** Same categories as version sections (Added, Changed, etc.)

### Version Sections

```markdown
## [x.y.z] - YYYY-MM-DD
```

**Format:**

- Version number in brackets: `[x.y.z]`
- Release date: ISO 8601 format (`YYYY-MM-DD`)
- Sorted newest first

**Example:**

```markdown
## [1.2.0] - 2025-12-01

## [1.1.0] - 2025-11-15

## [1.0.0] - 2025-10-01
```

### Change Type Categories

Each version section contains categorized changes:

#### Added

New features or functionality.

```markdown
### Added
- feat(scope): new feature description
- feat(scope): another new feature
```

**Examples:**

- New commands or tools
- New API endpoints
- New configuration options
- New documentation sections

#### Changed

Changes to existing functionality.

```markdown
### Changed
- refactor(scope): refactored implementation
- chore(scope): updated dependencies
```

**Examples:**

- API changes (breaking or non-breaking)
- Configuration format changes
- Dependency updates
- Performance improvements
- Behavior modifications

#### Deprecated

Features marked for removal in future versions.

```markdown
### Deprecated
- Deprecated old-command in favor of new-command
```

**Examples:**

- Features scheduled for removal
- APIs being phased out
- Configuration options being replaced

#### Removed

Features removed in this version.

```markdown
### Removed
- Removed deprecated feature-name
```

**Examples:**

- Deleted features
- Removed APIs
- Eliminated configuration options

#### Fixed

Bug fixes.

```markdown
### Fixed
- fix(scope): fixed bug description
```

**Examples:**

- Bug fixes
- Crash fixes
- Error handling improvements
- Incorrect behavior corrections

#### Security

Security improvements and vulnerability fixes.

```markdown
### Security
- Fixed SQL injection vulnerability in query builder
```

**Examples:**

- Vulnerability patches
- Security enhancements
- Dependency security updates

### Footer Links

```markdown
[Unreleased]: https://github.com/owner/repo/compare/v1.2.0...HEAD
[1.2.0]: https://github.com/owner/repo/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/owner/repo/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/owner/repo/releases/tag/v1.0.0
```

**Format:**

- Links version headers to GitHub compare URLs
- Unreleased links to diff from latest version to HEAD
- Version links to diff from previous version
- First version links to release tag

## Complete Example

```markdown
# Changelog

All notable changes to **r2r-cli** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- feat(cli): add version command for displaying build information

## [1.1.0] - 2025-12-01

### Added
- feat(commands): add build command for cross-platform compilation
- feat(commands): add test command with suite support

### Changed
- refactor(core): improve module dependency resolution
- chore(deps): update Go to 1.21

### Fixed
- fix(build): correct artifact path resolution on Windows

## [1.0.0] - 2025-11-15

### Added
- feat(cli): initial release with basic command structure
- feat(commands): implement get and show commands
- docs: add comprehensive user guide

[Unreleased]: https://github.com/owner/repo/compare/r2r-cli/1.1.0...HEAD
[1.1.0]: https://github.com/owner/repo/compare/r2r-cli/1.0.0...r2r-cli/1.1.0
[1.0.0]: https://github.com/owner/repo/releases/tag/r2r-cli/1.0.0
```

## Commit Message Conventions

Changes are derived from commit messages using conventional commit format:

### Conventional Commit Format

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

**Components:**

- `type` - Change type (feat, fix, docs, etc.)
- `scope` - Module or component affected
- `description` - Brief description of change

### Type Mapping

| Commit Type | Changelog Category               |
| ----------- | -------------------------------- |
| `feat`      | Added                            |
| `fix`       | Fixed                            |
| `refactor`  | Changed                          |
| `perf`      | Changed                          |
| `docs`      | Changed (or omitted if doc-only) |
| `test`      | Changed (or omitted)             |
| `chore`     | Changed                          |
| `ci`        | Changed (or omitted)             |
| `build`     | Changed                          |
| `revert`    | Changed or Removed               |
| `security`  | Security                         |

### Commit Examples

```bash
# Maps to Added
git commit -m "feat(commands): add build command with cross-platform support"

# Maps to Fixed
git commit -m "fix(parser): correct error handling in YAML parser"

# Maps to Changed
git commit -m "refactor(core): simplify module dependency resolution"

# Maps to Changed
git commit -m "chore(deps): update dependencies to latest versions"

# Maps to Security
git commit -m "security(auth): patch authentication bypass vulnerability"
```

## Changelog Generation

Changelogs can be generated automatically from commit history:

```bash
# Generate changelog for a module
r2r eac release changelog <moniker>

# Update existing changelog with new commits
r2r eac release changelog <moniker> --update
```

**Process:**

1. Reads commit history since last version
2. Parses conventional commit messages
3. Groups commits by type
4. Formats entries according to Keep a Changelog
5. Updates CHANGELOG.md file

## Best Practices

### Writing Good Changelog Entries

**Good:**

```markdown
- Added support for cross-platform binary builds
- Fixed crash when processing empty configuration files
- Improved module dependency resolution performance by 40%
```

**Bad:**

```markdown
- Stuff (too vague)
- Updated code (not descriptive)
- Fixed bug (which bug?)
```

### Entry Guidelines

1. **Be specific** - Describe what changed, not how
2. **Be concise** - One line per change
3. **Use present tense** - "Add" not "Added"
4. **Lead with verb** - Start with action word
5. **Reference issues** - Link to issue numbers when applicable
6. **Group related changes** - Combine similar changes

### What to Include

**Include:**

- New features
- Breaking changes
- Bug fixes
- Security fixes
- Deprecations
- Major dependency updates

**Exclude:**

- Code style changes
- Internal refactoring (unless user-visible)
- Test updates (unless affecting behavior)
- Documentation typos
- Build script changes

### Breaking Changes

Highlight breaking changes prominently:

```markdown
### Changed
- **BREAKING:** Renamed config field `oldName` to `newName`
- **BREAKING:** Changed API endpoint from `/old` to `/new`
```

**Prefix with `BREAKING:`** to make them immediately visible

### Yanked Releases

Mark pulled releases as yanked:

```markdown
## [1.1.0] - 2025-12-01 [YANKED]

### Security
- Fixed critical security vulnerability

*This version was yanked due to a regression in the build system.*
```

## Validation

Validate changelog format:

```bash
# Validate changelog structure
r2r eac validate release

# Check version format
r2r eac validate release-version <version>
```

**Validation checks:**

- Follows Keep a Changelog format
- Versions are valid semver
- Dates are ISO 8601 format
- Links are properly formatted
- Unreleased section exists

## References

- [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) - Format specification
- [Semantic Versioning](./versioning.md) - Version numbering
- [Repository Changelog](./repository-changelog.md) - Repository-level conventions
- [Module Changelogs](./module-changelogs.md) - Module-level conventions
