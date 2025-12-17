# Repository Changelog
Reference for the repository-level changelog.

## Overview

The repository-level changelog tracks changes that affect the entire repository, multiple modules, or infrastructure that spans the codebase.

**Location:** `CHANGELOG.md` (repository root)

**Format:** [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/)

**Versioning:** [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html)

**Current Version:** 0.0.2 (as of 2025-12-01)

## File Structure

### Header

```markdown
# Changelog

All notable changes to **repository** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
```

**Scope:** `repository` - Indicates repository-level changes

### Version Format

```markdown
## [0.0.2] - 2025-12-01

### Added
- feat(multi-module): add unused-steps command for spec analysis
- feat(multi-module): add dual-output logging with debug support

### Changed
- refactor(multi-module): simplify module contract structure
- refactor(multi-module): reorganize registry and consolidate dependencies

### Fixed
- fix(multi-module): remove git library import and update tests
- fix: ci
```

## What Goes in Repository Changelog

The repository changelog documents changes that:

### Infrastructure Changes

Changes to build system, CI/CD, or development tools:

```markdown
### Added
- feat(ci): add incremental CI with change detection
- feat(build): add cross-platform build support

### Changed
- ci(multi-module): containerize CI workflows with go-bvt-image
- chore(multi-module): enhance CI workflows and update build utilities
```

**Examples:**

- CI/CD workflow changes
- Build system modifications
- Development tool updates
- GitHub Actions updates

### Multi-Module Changes

Changes affecting multiple modules:

```markdown
### Changed
- refactor(multi-module): centralize EAC configuration structure
- refactor(multi-module): migrate to centralized contracts repository
```

**Examples:**

- Shared library updates
- Cross-module refactoring
- Configuration format changes
- Dependency reorganization

### Repository Configuration

Changes to repository-level configuration:

```markdown
### Added
- feat: commands for docs and design
- feat(multi-module): add ext-eac Docker extension and CI/CD workflows

### Changed
- chore(multi-module): restructure testing framework and add module isolation
```

**Examples:**

- Module contracts updates (`.r2r/eac/repository.yml`)
- Workspace configuration
- Repository-level documentation
- Project templates

### Breaking Changes

Repository-wide breaking changes:

```markdown
### Changed
- **BREAKING:** Migrate to centralized contracts repository
- **BREAKING:** Restructure module directory layout
```

**Impact:** Affects multiple modules or development workflow

## What Does NOT Go in Repository Changelog

### Module-Specific Changes

Changes affecting a single module go in that module's changelog:

**Belongs in module changelog:**

```markdown
# go/eac/commands/CHANGELOG.md

### Added
- feat(commands): add new build command
```

**Not in repository changelog**:

### Documentation Updates

Documentation changes for specific modules:

**Belongs in module changelog:**

```markdown
### Changed
- docs(cli): update installation guide
```

**Exception:** Repository-level documentation (README.md, CONTRIBUTING.md) goes in repository changelog

### Test Updates

Test changes for specific modules:

**Belongs in module changelog:**

```markdown
### Changed
- test(core): expand test coverage
```

**Exception:** Test infrastructure changes affecting all modules go in repository changelog

## Version Links

Footer links for version comparison:

```markdown
[Unreleased]: https://github.com/ready-to-release/eac/compare/repository/0.0.2...HEAD
[0.0.2]: https://github.com/ready-to-release/eac/releases/tag/repository/0.0.2
```

**Format:** `{owner}/{repo}/compare/{prefix}/{version1}...{prefix}/{version2}`

**Prefix:** `repository` - Distinguishes from module tags

**Examples:**

- Unreleased: `repository/0.0.2...HEAD`
- Version: `repository/0.0.2` (tag)
- Compare: `repository/0.0.1...repository/0.0.2`

## Version Numbering

Repository versions follow Semantic Versioning:

- **MAJOR.MINOR.PATCH** - Standard semver format
- **Current:** 0.0.2 (pre-1.0 development)

### Version Increment Rules

**PATCH (0.0.x):**

- Bug fixes
- Minor documentation updates
- Small infrastructure improvements

**MINOR (0.x.0):**

- New features or capabilities
- Significant infrastructure changes
- Non-breaking module contract updates

**MAJOR (x.0.0):**

- Breaking changes to repository structure
- Major version milestones
- Incompatible module contract changes

**Pre-1.0:** Repository is in development, expect breaking changes in minor versions

## Updating Repository Changelog

### Manual Update

Edit `CHANGELOG.md` directly:

1. **Add to Unreleased section:**

   ```markdown
   ## [Unreleased]

   ### Added
   - Your new change here
   ```

2. **When releasing, move to version section:**

   ```markdown
   ## [0.0.3] - 2025-12-15

   ### Added
   - Your new change here

   ## [Unreleased]
   ```

3. **Update footer links:**

   ```markdown
   [Unreleased]: https://github.com/ready-to-release/eac/compare/repository/0.0.3...HEAD
   [0.0.3]: https://github.com/ready-to-release/eac/compare/repository/0.0.2...repository/0.0.3
   ```

### Automated Generation

Generate changelog from commits:

```bash
# Generate changelog entries from git history
r2r eac release changelog repository

# Update existing changelog
r2r eac release changelog repository --update

# Preview without writing
r2r eac release changelog repository --dry-run
```

**Process:**

1. Reads commit history since last version
2. Filters to repository-level commits (multi-module, infrastructure, etc.)
3. Groups by change type
4. Formats according to Keep a Changelog
5. Updates CHANGELOG.md

## Release Process

When releasing a new repository version:

### Step 1: Update Changelog

```markdown
## [0.0.3] - 2025-12-15

### Added
- feat(ci): add parallel module testing

### Changed
- refactor(config): centralize configuration structure
```

### Step 2: Commit Changes

```bash
git add CHANGELOG.md
git commit -m "docs(repository): update changelog for v0.0.3"
```

### Step 3: Create Tag

```bash
git tag repository/0.0.3
git push origin repository/0.0.3
```

### Step 4: Verify Links

Check that footer links are updated:

```markdown
[Unreleased]: https://github.com/ready-to-release/eac/compare/repository/0.0.3...HEAD
[0.0.3]: https://github.com/ready-to-release/eac/compare/repository/0.0.2...repository/0.0.3
```

## Current Status

**Latest Version:** 0.0.2

**Released:** 2025-12-01

**Notable Changes in 0.0.2:**

- Added unused-steps command for spec analysis
- Added dual-output logging with debug support
- Simplified module contract structure
- Reorganized registry and consolidated dependencies
- Centralized EAC configuration structure
- Containerized CI workflows
- Enhanced testing framework

**File:** `/CHANGELOG.md`

## References

- Repository changelog file: `CHANGELOG.md`
- [Format Specification](./format-specification.md) - Keep a Changelog format
- [Module Changelogs](./module-changelogs.md) - Module-level conventions
- [Versioning](./versioning.md) - Semantic versioning rules
- [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
