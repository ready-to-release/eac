# Versioning
Semantic versioning and calendar versioning conventions for repository and module releases.

## Overview

The repository uses two versioning schemes depending on the type of module:

- **Semantic Versioning (SemVer)** - For software libraries and applications
- **Calendar Versioning (CalVer)** - For documentation and content releases

## Semantic Versioning

**Standard:** [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html)

**Format:** `MAJOR.MINOR.PATCH`

**URL:** [https://semver.org/spec/v2.0.0.html](https://semver.org/spec/v2.0.0.html)

### Version Format

```text
MAJOR.MINOR.PATCH
  |     |     |
  |     |     +-- Patch: Backwards-compatible bug fixes
  |     +-------- Minor: Backwards-compatible functionality
  +-------------- Major: Incompatible API changes
```

**Examples:**

- `1.0.0` - Initial stable release
- `1.1.0` - New features, backwards compatible
- `1.1.1` - Bug fixes, backwards compatible
- `2.0.0` - Breaking changes, incompatible with 1.x

### Version Components

#### MAJOR version

Increment when making incompatible API changes.

**Examples:**

- Renaming public functions, types, or modules
- Removing public APIs
- Changing function signatures
- Changing configuration format (breaking)
- Changing command-line interface (breaking)

**Breaking changes require:**

- MAJOR version increment
- Documentation of migration path
- Clear changelog entry with `BREAKING:` prefix

#### MINOR version

Increment when adding functionality in a backwards-compatible manner.

**Examples:**

- Adding new functions or methods
- Adding new configuration options
- Adding new command-line flags
- Deprecating features (without removing)
- Internal refactoring (no API changes)

**Requirements:**

- Backwards compatible with previous MINOR versions in same MAJOR
- New features optional to use
- Existing functionality unchanged

#### PATCH version

Increment when making backwards-compatible bug fixes.

**Examples:**

- Fixing incorrect behavior
- Correcting error messages
- Performance improvements (no API changes)
- Security patches
- Documentation fixes

**Requirements:**

- No new features
- No API changes
- Backwards compatible with all versions in same MAJOR.MINOR

### Pre-release Versions

For versions before 1.0.0 or pre-release testing.

**Format:** `MAJOR.MINOR.PATCH-PRERELEASE+BUILD`

**Pre-release identifiers:**

- `alpha` - Early development, unstable
- `beta` - Feature complete, testing
- `rc` (release candidate) - Production ready, final testing

**Examples:**

```text
0.1.0-alpha.1      First alpha release
0.1.0-alpha.2      Second alpha release
0.1.0-beta.1       First beta release
0.1.0-beta.2       Second beta release
0.1.0-rc.1         Release candidate 1
0.1.0-rc.2         Release candidate 2
0.1.0              Stable release
```

**Precedence:**

```text
0.1.0-alpha.1 < 0.1.0-alpha.2 < 0.1.0-beta.1 < 0.1.0-rc.1 < 0.1.0
```

### Build Metadata

Optional build information (not part of version precedence).

**Format:** `MAJOR.MINOR.PATCH+BUILD`

**Examples:**

```text
1.0.0+20251201      Built on December 1, 2025
1.0.0+001           Build number 1
1.0.0+sha.abc123    Built from commit abc123
```

**Note:** Build metadata ignored for version precedence

### Version Precedence

Version comparison rules:

```text
0.1.0 < 0.2.0 < 0.2.1 < 0.3.0 < 1.0.0 < 1.1.0 < 2.0.0
```

**Rules:**

1. Compare MAJOR, then MINOR, then PATCH
2. Pre-release versions have lower precedence than normal versions
3. Build metadata doesn't affect precedence

## Pre-1.0 Development

**Special rules for 0.x versions:**

```text
0.MINOR.PATCH
```

**Characteristics:**

- **Unstable API** - Breaking changes may occur in MINOR versions
- **No compatibility guarantees** - 0.x versions may be incompatible
- **Rapid iteration** - Frequent breaking changes expected

**Version increments:**

- `0.1.0 → 0.2.0` - May include breaking changes
- `0.2.0 → 0.2.1` - Bug fixes only
- `0.x.y → 1.0.0` - First stable release

**Current status:**

- Repository: `0.0.2` (pre-1.0 development)
- r2r-cli: May be 1.x (stable)

## Calendar Versioning

**Format:** `YYYY.MM.DD` or `YYYY.MM.MINOR`

**Use cases:** Documentation, content, and time-based releases

### Date-Based Format

**Format:** `YYYY.MM.DD`

**Structure:**

- `YYYY` - Four-digit year
- `MM` - Two-digit month (01-12)
- `DD` - Two-digit day (01-31)

**Examples:**

```text
2025.12.01    December 1, 2025
2025.12.15    December 15, 2025
2026.01.01    January 1, 2026
```

**Usage:** Documentation site, PDF books

**Precedence:**

```text
2025.11.01 < 2025.12.01 < 2025.12.15 < 2026.01.01
```

### Month-Based Format

**Format:** `YYYY.MM.MINOR`

**Structure:**

- `YYYY` - Four-digit year
- `MM` - Two-digit month (01-12)
- `MINOR` - Sequential number within month (0, 1, 2, ...)

**Examples:**

```text
2025.12.0     First release in December 2025
2025.12.1     Second release in December 2025
2025.12.2     Third release in December 2025
```

**Usage:** Multiple releases per month

**Precedence:**

```text
2025.12.0 < 2025.12.1 < 2025.12.2 < 2026.01.0
```

## Module Versioning Schemes

| Module            | Versioning      | Rationale                               |
| ----------------- | --------------- | --------------------------------------- |
| repository        | SemVer          | Repository infrastructure and contracts |
| r2r-cli           | SemVer          | CLI application with API                |
| ext-eac           | SemVer          | Docker extension with interface         |
| vscode-ext-commit | SemVer          | VSCode extension with API               |
| docs              | CalVer          | Documentation site (time-based)         |
| books             | CalVer          | PDF documentation (time-based)          |
| eac-core          | (not versioned) | Supporting library                      |
| eac-commands      | (not versioned) | Supporting library                      |
| eac-ai            | (not versioned) | Supporting library                      |

## Version Validation

### Command-Line Validation

```bash
# Validate semver format
r2r eac validate release-version 1.2.3

# Validate with pre-release
r2r eac validate release-version 1.2.3-beta.1

# Validate with build metadata
r2r eac validate release-version 1.2.3+20251201
```

**Output:**

- Exit 0: Valid version
- Exit 1: Invalid version with error message

**Error examples:**

```text
Error: Invalid version format: 1.2
Expected: MAJOR.MINOR.PATCH (e.g., 1.2.3)

Error: Invalid version format: v1.2.3
Expected: 1.2.3 (no 'v' prefix)

Error: Invalid pre-release: 1.2.3-beta
Expected: 1.2.3-beta.1 (numeric suffix required)
```

### Programmatic Validation

Validation checks:

1. **Format:** Matches `MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]`
2. **Components:** All numeric (except pre-release/build identifiers)
3. **No prefix:** No 'v' prefix (use `1.0.0`, not `v1.0.0`)
4. **Pre-release format:** Alphanumeric with dots (e.g., `alpha.1`)
5. **Build format:** Alphanumeric with dots (e.g., `20251201`)

## Git Tags

### Tag Format

Tags follow the format: `{moniker}/{version}`

**Repository:**

```text
repository/0.0.1
repository/0.0.2
```

**Modules (SemVer):**

```text
r2r-cli/1.0.0
r2r-cli/1.1.0
r2r-cli/2.0.0

ext-eac/0.1.0
ext-eac/0.2.0
```

**Modules (CalVer):**

```text
docs/2025.12.01
docs/2025.12.15

books/2025.12.01
books/2026.01.01
```

### Tag Creation

```bash
# Create tag
git tag {moniker}/{version}

# Push tag (triggers release workflow)
git push origin {moniker}/{version}

# Examples
git tag r2r-cli/1.3.0
git push origin r2r-cli/1.3.0

git tag docs/2025.12.15
git push origin docs/2025.12.15
```

### Tag Listing

```bash
# List all tags
git tag

# List tags for specific module
git tag -l 'r2r-cli/*'

# List tags with dates
git tag -l --format='%(refname:short) %(creatordate:short)'
```

### Tag Deletion

```bash
# Delete local tag
git tag -d {moniker}/{version}

# Delete remote tag
git push --delete origin {moniker}/{version}

# Example
git tag -d r2r-cli/1.3.0
git push --delete origin r2r-cli/1.3.0
```

## Version Comparison

### SemVer Comparison

```text
1.0.0 < 1.0.1 < 1.1.0 < 2.0.0
```

**Algorithm:**

1. Compare MAJOR (higher wins)
2. If equal, compare MINOR (higher wins)
3. If equal, compare PATCH (higher wins)
4. If equal, compare pre-release (none > any pre-release)

### CalVer Comparison

```text
2025.11.01 < 2025.12.01 < 2025.12.15 < 2026.01.01
```

**Algorithm:**

1. Compare YYYY (higher wins)
2. If equal, compare MM (higher wins)
3. If equal, compare DD or MINOR (higher wins)

## Version Increment Commands

```bash
# Get current version from changelog
r2r eac release get-version {moniker}

# Validate version format
r2r eac validate release-version {version}

# Generate next version tag (CalVer)
r2r eac release generate-module-calver {moniker}
```

## Best Practices

### Choosing Version Numbers

**For bug fixes:**

```text
1.2.3 → 1.2.4
```

**For new features:**

```text
1.2.4 → 1.3.0
```

**For breaking changes:**

```text
1.3.0 → 2.0.0
```

**For pre-1.0:**

```text
0.1.0 → 0.2.0  (may include breaking changes)
0.2.0 → 1.0.0  (first stable release)
```

### Documentation

**In CHANGELOG.md:**

```markdown
## [2.0.0] - 2025-12-15

### Changed
- **BREAKING:** Renamed Config to Configuration
- **BREAKING:** Changed API endpoint structure
```

**In release notes:**

- Highlight breaking changes prominently
- Provide migration guide
- Document deprecated features
- List new features with examples

### Communication

**Breaking changes:**

- Announce in advance (deprecation warnings)
- Provide migration path
- Update documentation
- Consider compatibility layers

**Major releases:**

- Write migration guide
- Update tutorials and examples
- Communicate timeline
- Support previous MAJOR version temporarily

## References

- [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html)
- [Calendar Versioning](https://calver.org/)
- [Format Specification](./format-specification.md) - Changelog format
- [Repository Changelog](./repository-changelog.md) - Repository versioning
- [Module Changelogs](./module-changelogs.md) - Module versioning
- Decision Record: [DR-002 - Use Semantic Versioning](../../decision-records/dr002.md)
