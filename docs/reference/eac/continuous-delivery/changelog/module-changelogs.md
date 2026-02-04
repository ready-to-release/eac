# Module Changelogs

Reference for module-level changelogs.

## Overview

Each deployable module maintains its own changelog to track module-specific changes.

Module changelogs are independent of the repository changelog and follow the same Keep a Changelog format.

**Format:** [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/)

**Versioning:** [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html) or CalVer

## Module Changelog Locations

Module changelog locations are defined in `.eac/repository.yml` under `versioning.changelog` and follow the **release type convention**.

See [Understanding Release Types](../release-types.md) for complete details.

### Location Rules by Release Type

| Release Type | Changelog Location              | Example                               |
| ------------ | ------------------------------- | ------------------------------------- |
| `published`  | `release/<module>/CHANGELOG.md` | `release/r2r-cli/CHANGELOG.md`        |
| `bundle`     | `release/<module>/CHANGELOG.md` | `release/r2r-eac-bundle/CHANGELOG.md` |
| `internal`   | `<module-root>/CHANGELOG.md`    | `go/cli/eac/CHANGELOG.md`        |
| `none`       | No changelog                    | N/A                                   |

### Pattern 1: Release Directory (Published/Bundle)

Changelog in dedicated release directory for public-facing releases:

```yaml
modules:
  - moniker: r2r-cli
    versioning:
      scheme: SemVer
      changelog: release/r2r-cli/CHANGELOG.md
      release_type: published
```

**Full path:** `release/r2r-cli/CHANGELOG.md`

**Usage:** Published modules and bundles

**Examples:**

- `release/r2r-cli/CHANGELOG.md` (published)
- `release/ext-eac/CHANGELOG.md` (published)
- `release/docs/CHANGELOG.md` (published)
- `release/books/CHANGELOG.md` (published)
- `release/r2r-eac-bundle/CHANGELOG.md` (bundle)

### Pattern 2: Module Root (Internal)

Changelog in the module's root directory for internal modules:

```yaml
modules:
  - moniker: eac-commands
    versioning:
      scheme: SemVer
      changelog: go/cli/eac/CHANGELOG.md
      release_type: internal
```

**Full path:** `go/cli/eac/CHANGELOG.md`

**Usage:** Internal implementation modules

**Examples:**

- `go/cli/eac/CHANGELOG.md` (internal)
- `go/mcp/commands/CHANGELOG.md` (internal)
- `scripts/CHANGELOG.md` (internal - r2r-installer)
- `typescript/vscode-commit/CHANGELOG.md` (internal)

## Module Changelog Structure

### Header

```markdown
# Changelog

All notable changes to **{module name}** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
```

**Module name:** Human-readable module name (not moniker)

**Examples:**

- `r2r CLI`
- `EAC Commands Library`
- `EAC Core Library`

### Version Format

Module versions are independent and follow their own numbering:

```markdown
## [1.2.0] - 2025-12-01

### Added

- feat(cli): add version command

### Changed

- refactor(commands): improve error messages

### Fixed

- fix(build): correct platform detection
```

## What Goes in Module Changelog

### Module-Specific Changes

Changes affecting only this module:

```markdown
### Added

- feat(commands): add build command with cross-platform support
- feat(commands): add test suite execution

### Changed

- refactor(parser): simplify YAML parsing logic
- perf(resolver): optimize dependency resolution

### Fixed

- fix(cli): correct exit code on errors
```

**Scope:** Changes to code, tests, or configuration in the module's directory

### API Changes

Changes to the module's public API:

```markdown
### Added

- feat(api): add GetModules() function

### Changed

- **BREAKING:** Rename Config to Configuration
- **BREAKING:** Change BuildModule signature

### Deprecated

- Deprecated GetModule() in favor of GetModules()
```

**Importance:** Document API changes clearly for library consumers

### Dependency Updates

Updates to module's direct dependencies:

```markdown
### Changed

- chore(deps): update cobra to v1.8.0
- chore(deps): update go-yaml to v3.0.1
```

**Scope:** Only dependencies in this module's `go.mod` or `package.json`

## What Does NOT Go in Module Changelog

### Repository-Level Changes

Changes to CI/CD, build system, or multi-module refactoring:

**Belongs in repository changelog:**

```markdown
# CHANGELOG.md (repository)

### Changed

- ci(multi-module): containerize CI workflows
- refactor(multi-module): reorganize module structure
```

> **Note:** This does NOT belong in the module changelog.

### Other Module Changes

Changes to other modules:

**Belongs in that module's changelog:**

```markdown
# go/core/CHANGELOG.md

### Added

- feat(core): add new validation function
```

> **Note:** This does NOT belong in eac-commands changelog.

### Dependency Changes

Shared dependency updates managed at repository level:

**Belongs in repository changelog:**

```markdown
### Changed

- chore(deps): update Go to 1.21 across all modules
```

**Exception:** Module-specific dependency updates belong in module changelog

## Module Versioning

### Versioning Schemes

Modules use either Semantic Versioning or Calendar Versioning:

#### Semantic Versioning (SemVer)

**Format:** `MAJOR.MINOR.PATCH`

**Modules using SemVer:**

- `r2r-cli` - CLI application
- `ext-eac` - Docker extension
- `eac-commands` - Go library (not independently versioned)
- `eac-core` - Go library (not independently versioned)

**Example versions:**

- `1.0.0` - Initial release
- `1.1.0` - New features
- `1.1.1` - Bug fix
- `2.0.0` - Breaking changes

#### Calendar Versioning (CalVer)

**Format:** `YYYY.MM.DD` or `YYYY.MM.MINOR`

**Modules using CalVer:**

- `docs` - Documentation site
- `books` - PDF documentation

**Example versions:**

- `2025.12.01` - Released December 1, 2025
- `2025.12.02` - Updated December 2, 2025

### Version Independence

Each module versions independently:

**Example:**

```text
Repository: 0.0.2
r2r-cli: 1.2.0
eac-commands: (not versioned - library)
docs: 2025.12.01
ext-eac: 0.3.0
```

**Implication:** Module versions don't align with repository version

## Version Links

Footer links use module-specific tag format:

### SemVer Modules

```markdown
[Unreleased]: https://github.com/ready-to-release/eac/compare/r2r-cli/1.2.0...HEAD
[1.2.0]: https://github.com/ready-to-release/eac/compare/r2r-cli/1.1.0...r2r-cli/1.2.0
[1.1.0]: https://github.com/ready-to-release/eac/compare/r2r-cli/1.0.0...r2r-cli/1.1.0
[1.0.0]: https://github.com/ready-to-release/eac/releases/tag/r2r-cli/1.0.0
```

**Tag format:** `{moniker}/{version}`

**Examples:**

- `r2r-cli/1.2.0`
- `ext-eac/0.3.0`

### CalVer Modules

```markdown
[Unreleased]: https://github.com/ready-to-release/eac/compare/docs/2025.12.01...HEAD
[2025.12.01]: https://github.com/ready-to-release/eac/releases/tag/docs/2025.12.01
```

**Tag format:** `{moniker}/{YYYY.MM.DD}`

**Examples:**

- `docs/2025.12.01`
- `books/2025.12.01`

## Updating Module Changelog

### During Development

Add changes to Unreleased section as you work:

```markdown
## [Unreleased]

### Added

- feat(commands): add new validate command

### Fixed

- fix(parser): handle empty files correctly
```

### During Release

1. **Move Unreleased to version section:**

   ```markdown
   ## [1.3.0] - 2025-12-15

   ### Added

   - feat(commands): add new validate command

   ### Fixed

   - fix(parser): handle empty files correctly

   ## [Unreleased]
   ```

2. **Update footer links:**

   ```markdown
   [Unreleased]: https://github.com/ready-to-release/eac/compare/r2r-cli/1.3.0...HEAD
   [1.3.0]: https://github.com/ready-to-release/eac/compare/r2r-cli/1.2.0...r2r-cli/1.3.0
   ```

### Automated Generation

Generate module changelog from commits:

```bash
# Generate changelog for a module
eac release changelog r2r-cli

# Update existing changelog
eac release changelog r2r-cli --update

# Preview without writing
eac release changelog r2r-cli --dry-run
```

**Filtering:**

- Only includes commits affecting the module's files
- Filters by module scope in commit messages
- Groups by change type
- Formats according to Keep a Changelog

## Release Process

When releasing a module version:

### Step 1: Finalize Changelog

```bash
# Review and finalize changelog
eac release this r2r-cli
```

**Actions:**

- Moves Unreleased changes to new version section
- Adds release date
- Updates footer links
- Validates format

### Step 2: Commit Changes

```bash
git add release/r2r-cli/CHANGELOG.md
git commit -m "docs(r2r-cli): update changelog for v1.3.0"
```

### Step 3: Create Tag

```bash
git tag r2r-cli/1.3.0
git push origin r2r-cli/1.3.0
```

### Step 4: Release Workflow

Tag push triggers release workflow:

```yaml
# .github/workflows/release-r2r-cli.yaml
on:
  push:
    tags:
      - "r2r-cli/*"
```

**Actions:**

- Builds release artifacts
- Creates GitHub release
- Uploads binaries/assets

## Module Changelog Inventory

### Deployable Modules with Changelogs

| Module            | Changelog Location                          | Versioning | Current Version |
| ----------------- | ------------------------------------------- | ---------- | --------------- |
| r2r-cli           | `release/r2r-cli/CHANGELOG.md`              | SemVer     | (varies)        |
| ext-eac           | `release/ext-eac/CHANGELOG.md`              | SemVer     | (varies)        |
| docs              | `release/docs/CHANGELOG.md`                 | CalVer     | (varies)        |
| books             | `release/books/CHANGELOG.md`                | CalVer     | (varies)        |
| vscode-commit     | `typescript/vscode-commit/CHANGELOG.md`     | SemVer     | (varies)        |

### Supporting Modules (No Independent Changelog)

Supporting modules don't maintain separate changelogs as they're not independently versioned:

- eac-core
- eac-ai
- eac-commands
- eac-specs

**Note:** Changes to supporting modules are documented in:

- Repository changelog (if affecting multiple modules)
- Dependent module changelogs (if affecting specific module)

## Querying Module Changelogs

### Get Latest Version

```bash
# Get latest version from module changelog
eac release get-version r2r-cli
```

**Output:** `1.3.0` (latest version from CHANGELOG.md)

### Check Pending Release

```bash
# Check if module has unreleased changes
eac release pending r2r-cli
```

**Output:**

- `true` - Unreleased section has content
- `false` - Unreleased section is empty

### Validate Changelog

```bash
# Validate changelog format and structure
eac validate release r2r-cli
```

**Checks:**

- Keep a Changelog format compliance
- Valid version numbers
- ISO 8601 dates
- Properly formatted links

## References

- Module contracts: `.eac/repository.yml`
- [Format Specification](./format-specification.md) - Keep a Changelog format
- [Repository Changelog](./repository-changelog.md) - Repository-level conventions
- [Versioning](./versioning.md) - Semantic versioning and CalVer
- [Release Workflows](../workflows/release-workflows.md) - Release automation
