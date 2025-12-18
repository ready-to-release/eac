# Generate Changelog

## What You'll Accomplish

Create changelog from Git commits with automatic version detection and formatting.

## Prerequisites

- Git repository with commits
- Module has changes since last release

## Steps

### 1. Generate Changelog

```bash
r2r eac release changelog
```

**What happens**:

- Analyzes commits since last release
- Determines version bump (patch/minor/major)
- Generates CHANGELOG.md with sections

### 2. Review Generated Changelog

```bash
cat CHANGELOG.md
```

**What happens**: View generated changelog content

### 3. Edit if Needed

```bash
nano CHANGELOG.md
```

**What happens**: Manually refine entries if needed

### 4. Validate Format

```bash
r2r eac validate release
```

**What happens**: Checks changelog follows format standards

## Changelog Sections

Generated changelog includes:

- **Added** - New features
- **Changed** - Changes to existing features
- **Deprecated** - Soon-to-be removed features
- **Removed** - Removed features
- **Fixed** - Bug fixes
- **Security** - Security updates

## Example Scenario

Preparing release after multiple commits:

```bash
# Generate changelog
r2r eac release changelog

# Output:
# Analyzing commits since v1.1.0...
# Found 8 commits
# Detected version: v1.2.0 (minor bump)
# Generated CHANGELOG.md

# Review
cat CHANGELOG.md

# ## [1.2.0] - 2025-12-09
#
# ### Added
# - JWT authentication support
# - Refresh token endpoint
#
# ### Fixed
# - Login timeout issue
# - Password validation bug
#
# ### Changed
# - Updated authentication flow

# Validate format
r2r eac validate release
# ✓ Changelog format valid

# Get version for tagging
VERSION=$(r2r eac release get-version)
echo $VERSION
# v1.2.0
```

## Version Detection

Version is determined by commit types:

- `feat:` → Minor bump (1.1.0 → 1.2.0)
- `fix:` → Patch bump (1.1.0 → 1.1.1)
- `BREAKING CHANGE:` → Major bump (1.1.0 → 2.0.0)

## Common Issues

| Problem | Solution |
|---------|----------|
| "No changes detected" | Ensure commits since last tag |
| Version incorrect | Manually edit CHANGELOG.md |
| Format invalid | Fix format per validation errors |

## Next Steps

- [Prepare Module Release](./prepare-module-release.md) → Complete release
- [Create Release Tag](./create-release-tag.md) → Tag the release

## Related Commands

- [`release changelog`](../../../../reference/commands/release/changelog.md) - Generate changelog
- [`release get-version`](../../../../reference/commands/release/get-version.md) - Extract version
- [`validate release`](../../../../reference/commands/validate/release.md) - Validate format
