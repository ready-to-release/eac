# Release Command

**Problem**: Manually creating and managing version tags is error-prone and inconsistent, especially when coordinating releases with CI/CD workflows.

**Solution**: Use `release` commands to automate versioning with CalVer or SemVer, validate CI status, and create properly formatted git tags.

## Key Benefits

- Automated calendar versioning (CalVer) with auto-incrementing suffixes
- Semantic versioning (SemVer) validation for tool releases
- CI/CD integration to verify workflow success before tagging
- Consistent tag naming across all modules
- Dry-run support for safe preview
- Git tag creation and push automation

## Quick Start

```bash
# Create CalVer tag for a module (dry-run first)
r2r eac release calver auth --dry-run
r2r eac release calver auth --create --push

# Check CI status before release
r2r eac release check-ci --workflow=ci --commit=abc123

# Release src-cli with SemVer
r2r eac release src-cli 1.2.0 --dry-run
r2r eac release src-cli 1.2.0
```

## Command Reference

### release calver

Generate calendar-versioned tags in the format `prefix/YYYY.MM.DD` or `prefix/YYYY.MM.DD.N` for duplicates.

```bash
r2r eac release calver <prefix> [options]

# Options:
--create               # Create the git tag locally (default: false)
--push                 # Push tag to remote (requires --create, default: false)
--dry-run              # Show what tag would be created without creating it

# Examples:
r2r eac release calver auth --dry-run          # Preview tag
r2r eac release calver auth --create           # Create tag locally
r2r eac release calver auth --create --push    # Create and push tag
r2r eac release calver api                     # Just display next tag
```

**Tag Format:**

- First release today: `auth/2025.11.30`
- Second release today: `auth/2025.11.30.1`
- Third release today: `auth/2025.11.30.2`

**Behavior:**

- Automatically detects existing tags for the same prefix and date
- Increments suffix (`.N`) when multiple releases happen on the same day
- Uses current date in UTC
- Validates that tag doesn't already exist before creating
- Skips creation if tag would be duplicate

### release check-ci

Check CI workflow status before releasing to ensure all checks pass.

```bash
r2r eac release check-ci [options]

# Required options:
--workflow <name>      # GitHub workflow name (e.g., "ci", "build-and-test")
--commit <sha>         # Full or short commit SHA to check

# Optional:
--timeout <seconds>    # Maximum time to wait for workflow (default: 600)
--interval <seconds>   # Polling interval in seconds (default: 30)

# Examples:
r2r eac release check-ci --workflow=ci --commit=abc123def
r2r eac release check-ci --workflow=build --commit=HEAD --timeout=300
r2r eac release check-ci --workflow=test --commit=$(git rev-parse HEAD)
```

**Requirements:**

- GitHub CLI (`gh`) must be installed and authenticated
- Repository must have GitHub Actions workflows configured
- Commit must exist in the repository

**Behavior:**

- Queries GitHub API for workflow runs associated with the commit
- Waits for in-progress workflows to complete (up to timeout)
- Checks all workflow jobs for success status
- Returns exit code 0 if all checks pass, 1 if any fail
- Provides detailed status updates during polling

**Use cases:**

- Pre-release validation in CI/CD pipelines
- Manual release workflows requiring approval
- Hotfix releases requiring immediate CI confirmation

### release src-cli

Create a semantic version tag for the src-cli tool in the format `src-cli/x.y.z`.

```bash
r2r eac release src-cli <version> [options]

# Options:
--dry-run              # Show what would happen without creating tag
--push                 # Push tag to remote (default: true)

# Examples:
r2r eac release src-cli 1.2.0 --dry-run        # Preview release
r2r eac release src-cli 1.2.0                  # Create and push tag
r2r eac release src-cli 1.2.0 --push=false     # Create tag locally only
r2r eac release src-cli 2.0.0-beta.1           # Pre-release version
```

**Version Format:**

Must follow SemVer specification:

- `MAJOR.MINOR.PATCH` (e.g., `1.2.0`)
- Pre-release: `MAJOR.MINOR.PATCH-prerelease` (e.g., `1.2.0-beta.1`)
- Build metadata: `MAJOR.MINOR.PATCH+build` (e.g., `1.2.0+20130313144700`)

**Tag Format:**

- Version `1.2.0` creates tag `src-cli/1.2.0`
- Version `2.0.0-beta.1` creates tag `src-cli/2.0.0-beta.1`

**Validation:**

- Ensures version follows SemVer format
- Checks that tag doesn't already exist
- Fails if invalid version provided

## Versioning Strategies

### CalVer vs SemVer

**Use CalVer (`release calver`) when:**

- Versioning application modules (APIs, services, features)
- Release cadence is time-based (daily, weekly)
- Version communicates "when" rather than "what changed"
- Multiple releases per day are possible
- Breaking changes are managed through API versioning

**Use SemVer (`release src-cli`) when:**

- Versioning developer tools and libraries
- Version communicates compatibility (breaking, features, fixes)
- Semantic meaning is important for dependency management
- Consumers need to understand impact of upgrading
- Following standard package manager conventions

### CalVer Format Details

```text
Format: prefix/YYYY.MM.DD[.N]

Examples:
auth/2025.11.30       # First release on Nov 30, 2025
auth/2025.11.30.1     # Second release same day
auth/2025.11.30.2     # Third release same day
api/2025.12.01        # First release on Dec 1, 2025
```

**Benefits:**

- Immediately see when code was released
- No need to track version numbers
- Natural chronological ordering
- Supports multiple releases per day

### SemVer Format Details

```text
Format: MAJOR.MINOR.PATCH[-prerelease][+build]

Examples:
1.0.0                 # Major version 1, initial release
1.1.0                 # Added features, backward compatible
1.1.1                 # Bug fixes, backward compatible
2.0.0                 # Breaking changes
2.0.0-beta.1          # Pre-release (beta)
2.0.0-rc.1            # Release candidate
1.2.3+20130313        # Build metadata
```

**Increment rules:**

- **MAJOR**: Breaking changes (incompatible API changes)
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

## Typical Workflows

### Module Release (CalVer)

```bash
# 1. Complete development in workspace
cd eac-feature-auth
r2r eac work commit --all
r2r eac work pull

# 2. Merge to main
r2r eac work merge
cd ../eac

# 3. Push to trigger CI
git push origin main

# 4. Wait for CI to complete
r2r eac release check-ci --workflow=ci --commit=HEAD

# 5. Preview release tag
r2r eac release calver auth --dry-run

# 6. Create and push release tag
r2r eac release calver auth --create --push

# Output:
# Next tag: auth/2025.11.30
# ✅ Tag created: auth/2025.11.30
# ✅ Tag pushed to remote
```

### Tool Release (SemVer)

```bash
# 1. Update version in code/docs
# Edit CHANGELOG.md, update version constants, etc.

# 2. Commit version bump
git add .
git commit -m "chore: bump version to 1.2.0"
git push origin main

# 3. Wait for CI
r2r eac release check-ci --workflow=ci --commit=HEAD

# 4. Preview release
r2r eac release src-cli 1.2.0 --dry-run

# 5. Create and push release
r2r eac release src-cli 1.2.0

# 6. Verify tag
git tag -l "src-cli/*"
```

### Hotfix Release

```bash
# 1. Create hotfix workspace
r2r eac work create hotfix/critical-bug

# 2. Fix and test
cd ../eac-hotfix-critical-bug
# ... make changes ...
r2r eac test auth
r2r eac work commit --all

# 3. Merge and release immediately
r2r eac work merge
cd ../eac
git push origin main

# 4. Check CI (with shorter timeout)
r2r eac release check-ci --workflow=ci --commit=HEAD --timeout=300

# 5. Release
r2r eac release calver auth --create --push
```

### Pre-release Testing

```bash
# 1. Test release commands with dry-run
r2r eac release calver api --dry-run
# Output: Next tag would be: api/2025.11.30

# 2. Verify no conflicts
git tag -l "api/2025.11.30*"

# 3. Create locally first (test without pushing)
r2r eac release calver api --create

# 4. Verify tag
git tag -l "api/*" | tail -5

# 5. If good, push manually
git push origin api/2025.11.30

# Or delete and retry with --push
git tag -d api/2025.11.30
r2r eac release calver api --create --push
```

## Pre-release Checklist

Before running `release` commands:

- [ ] All changes committed and pushed
- [ ] CI/CD workflows passing (`release check-ci`)
- [ ] Tests passing locally (`r2r eac test`)
- [ ] Validation passing (`r2r eac validate`)
- [ ] CHANGELOG updated (for SemVer releases)
- [ ] Version bumped in code (for SemVer releases)
- [ ] Documentation updated
- [ ] Breaking changes documented
- [ ] Migration guide provided (for breaking changes)

## Integration Patterns

### GitHub Actions Release Workflow

```yaml
name: Release

on:
  workflow_dispatch:
    inputs:
      module:
        description: 'Module prefix for CalVer tag'
        required: true
        type: string
      push:
        description: 'Push tag to remote'
        required: true
        type: boolean
        default: true

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0  # Full history for tags

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Build and test
        run: |
          r2r eac build
          r2r eac test
          r2r eac validate

      - name: Check CI status
        run: |
          r2r eac release check-ci \
            --workflow=ci \
            --commit=${{ github.sha }}

      - name: Create release tag
        run: |
          r2r eac release calver ${{ inputs.module }} \
            --create \
            ${{ inputs.push && '--push' || '' }}
```

### Manual Release Script

```bash
#!/bin/bash
# scripts/release.sh

set -e

MODULE=$1
if [ -z "$MODULE" ]; then
  echo "Usage: ./release.sh <module-prefix>"
  exit 1
fi

echo "🚀 Starting release for module: $MODULE"

# 1. Verify clean working directory
if [ -n "$(git status --porcelain)" ]; then
  echo "❌ Working directory is not clean"
  exit 1
fi

# 2. Pull latest
git pull origin main

# 3. Build and test
echo "🔨 Building..."
r2r eac build

echo "🧪 Testing..."
r2r eac test

echo "✅ Validating..."
r2r eac validate

# 4. Check CI
COMMIT=$(git rev-parse HEAD)
echo "⏳ Checking CI for commit $COMMIT..."
r2r eac release check-ci --workflow=ci --commit=$COMMIT

# 5. Preview release
echo "📋 Preview:"
r2r eac release calver $MODULE --dry-run

# 6. Confirm
read -p "Create and push this tag? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  r2r eac release calver $MODULE --create --push
  echo "✅ Release complete!"
else
  echo "❌ Release cancelled"
  exit 1
fi
```

### CI-Triggered Auto-Release

```yaml
name: Auto Release

on:
  push:
    branches: [main]
    paths:
      - 'modules/auth/**'

jobs:
  auto-release:
    runs-on: ubuntu-latest
    if: "contains(github.event.head_commit.message, '[release]')"
    steps:
      - uses: actions/checkout@v3

      - name: Extract module from commit
        id: module
        run: |
          # Parse commit message for module name
          MODULE=$(echo "${{ github.event.head_commit.message }}" | \
                   grep -oP '\[release:\K[^\]]+')
          echo "module=$MODULE" >> $GITHUB_OUTPUT

      - name: Wait for CI
        run: |
          r2r eac release check-ci \
            --workflow=ci \
            --commit=${{ github.sha }} \
            --timeout=600

      - name: Create release
        run: |
          r2r eac release calver ${{ steps.module.outputs.module }} \
            --create --push
```

## Best Practices

- **Always dry-run first**: Use `--dry-run` to preview before creating tags
- **Verify CI status**: Run `release check-ci` before tagging
- **Use automation**: Integrate releases into CI/CD workflows
- **Tag from main**: Always create tags from the main branch
- **Document releases**: Update CHANGELOG for important releases
- **Consistent prefixes**: Use module monikers as CalVer prefixes
- **SemVer for tools**: Use `release src-cli` for developer-facing tools
- **CalVer for services**: Use `release calver` for deployed modules
- **Push carefully**: Ensure tag is correct before pushing to remote
- **Fetch tags**: Run `git fetch --tags` before checking for duplicates

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Tag already exists | Check with `git tag -l "prefix/*"`, delete with `git tag -d <tag>` if local |
| CI workflow not found | Verify workflow name with `gh workflow list` |
| Commit not found | Ensure commit is pushed to remote, use full SHA |
| Invalid SemVer | Follow format: `MAJOR.MINOR.PATCH` (e.g., `1.2.0`) |
| gh CLI not found | Install: `winget install GitHub.cli` or `brew install gh` |
| gh not authenticated | Run `gh auth login` |
| CI timeout | Increase `--timeout` or check workflow status manually |
| Permission denied | Ensure you have push permissions to the repository |
| Workflow still running | Wait for completion or increase timeout |
| Network error | Check internet connection and GitHub status |

## Advanced Usage

### Custom Date for CalVer

```bash
# CalVer uses current date automatically, but you can preview future dates
# by creating tags manually following the format

git tag auth/2025.12.01
git push origin auth/2025.12.01
```

### Multiple Modules Release

```bash
# Release multiple modules on same day
for module in auth api gateway; do
  r2r eac release calver $module --create --push
done

# Output:
# auth/2025.11.30
# api/2025.11.30
# gateway/2025.11.30
```

### Release with CI Integration

```bash
# Function to release after CI passes
release_after_ci() {
  local module=$1
  local commit=${2:-HEAD}

  echo "Waiting for CI..."
  if r2r eac release check-ci --workflow=ci --commit=$commit; then
    echo "CI passed, creating release..."
    r2r eac release calver $module --create --push
  else
    echo "CI failed, aborting release"
    return 1
  fi
}

# Usage
release_after_ci auth HEAD
```

### Pre-release Versions

```bash
# SemVer supports pre-release identifiers
r2r eac release src-cli 2.0.0-alpha.1 --dry-run
r2r eac release src-cli 2.0.0-beta.1
r2r eac release src-cli 2.0.0-rc.1
r2r eac release src-cli 2.0.0  # Final release
```

### Tag Cleanup

```bash
# List all tags for a module
git tag -l "auth/*"

# Delete local tag
git tag -d auth/2025.11.30

# Delete remote tag
git push --delete origin auth/2025.11.30

# Fetch all tags
git fetch --tags --prune
```

## Output Examples

### Successful CalVer Release

```text
$ r2r eac release calver auth --create --push

Analyzing existing tags for prefix: auth
Found existing tags:
  - auth/2025.11.29
  - auth/2025.11.29.1

Next tag: auth/2025.11.30

✅ Tag created: auth/2025.11.30
✅ Tag pushed to remote: origin

Release complete!
```

### Duplicate CalVer Release

```text
$ r2r eac release calver auth --create

Analyzing existing tags for prefix: auth
Found existing tags:
  - auth/2025.11.30

Next tag: auth/2025.11.30.1

✅ Tag created: auth/2025.11.30.1
```

### CI Check Success

```text
$ r2r eac release check-ci --workflow=ci --commit=abc123

Checking workflow 'ci' for commit abc123...

Workflow run found: #1234
Status: in_progress

Polling every 30 seconds (timeout: 600s)...
⏳ Waiting... (30s elapsed)
⏳ Waiting... (60s elapsed)

✅ Workflow completed successfully

All jobs passed:
  ✓ build (2m 34s)
  ✓ test (4m 12s)
  ✓ validate (1m 5s)

Safe to proceed with release.
```

### CI Check Failure

```text
$ r2r eac release check-ci --workflow=ci --commit=abc123

Checking workflow 'ci' for commit abc123...

Workflow run found: #1234
Status: completed
Conclusion: failure

❌ Workflow failed

Failed jobs:
  ✗ test (4m 12s) - exit code 1

Do not proceed with release.
```

## Summary

**CalVer workflow:**

1. Complete development and merge to main
2. Wait for CI: `r2r eac release check-ci --workflow=ci --commit=HEAD`
3. Preview: `r2r eac release calver <prefix> --dry-run`
4. Release: `r2r eac release calver <prefix> --create --push`

**SemVer workflow:**

1. Update version in code/docs
2. Commit and push version bump
3. Wait for CI: `r2r eac release check-ci --workflow=ci --commit=HEAD`
4. Preview: `r2r eac release src-cli <version> --dry-run`
5. Release: `r2r eac release src-cli <version>`

**Key differences:**

- CalVer: Time-based, automatic suffix, for modules/services
- SemVer: Meaning-based, manual version, for tools/libraries
- Both: Support dry-run, CI validation, git tag automation
