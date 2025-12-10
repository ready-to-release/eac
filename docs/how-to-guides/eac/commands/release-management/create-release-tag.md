# Create Release Tag

{{ page_breadcrumb() }}

## What You'll Accomplish

Create git tag for release using proper version format (CalVer for modules, SemVer for CLI).

## Prerequisites

- Changelog finalized and validated
- CI passing for commit
- Version determined

## Steps

### 1. Get Release Version

```bash
VERSION=$(r2r eac release get-version)
echo $VERSION
```

**What happens**: Extracts latest version from CHANGELOG.md

### 2. Create Release

```bash
r2r eac release this
```

**What happens**:

- Creates git tag with version
- Finalizes changelog
- Updates release metadata

### 3. Push Tag to Remote

```bash
git push origin $VERSION
```

**What happens**: Tag is available on remote for CI/releases

### 4. Verify Tag Created

```bash
git tag -l | grep $VERSION
```

**What happens**: Confirms tag exists locally

## Manual Tag Creation

For modules using CalVer:

```bash
# Generate CalVer tag
TAG=$(r2r eac release generate-module-calver src-auth)
echo $TAG
# src-auth/2025.12.09

# Create tag
git tag -a $TAG -m "Release $TAG"
```

For CLI using SemVer:

```bash
# Create SemVer tag
r2r eac release r2r-cli v1.2.3
```

## Example Scenario

Creating release tag for src-auth module:

```bash
# Check changelog version
r2r eac release get-version
# src-auth/2025.12.09

# Create release tag
r2r eac release this

# Output:
# Creating release for src-auth...
# ✓ Created tag: src-auth/2025.12.09
# ✓ Finalized CHANGELOG.md
# ✓ Updated release metadata

# Push to remote
git push origin src-auth/2025.12.09

# Verify on GitHub
gh release view src-auth/2025.12.09
```

## Version Formats

**CalVer (Modules)**:

- Format: `module-name/YYYY.MM.DD`
- Example: `src-auth/2025.12.09`

**SemVer (CLI)**:

- Format: `vMAJOR.MINOR.PATCH`
- Example: `v1.2.3`

## Common Issues

| Problem | Solution |
|---------|----------|
| "Tag already exists" | Version already released |
| "CI not passing" | Wait for CI or fix issues |
| Wrong version format | Check CalVer vs SemVer usage |

## Next Steps

- [Prepare Module Release](./prepare-module-release.md) → Full workflow

## Related Commands

- [`release this`](../../../../reference/commands/release/this.md) - Create release
- [`release generate-module-calver`](../../../../reference/commands/release/generate-module-calver.md) - Generate CalVer
- [`release r2r-cli`](../../../../reference/commands/release/r2r-cli.md) - Release CLI

{{ diataxis_footer() }}
