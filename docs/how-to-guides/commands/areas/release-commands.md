# Release Commands

Command reference for EAC's release management system.

## Quick Reference

| Command               | Description                                       |
| --------------------- | ------------------------------------------------- |
| `release-calver`      | Generate a CalVer tag for a module                |
| `release-changelog`   | Generate or update changelog from commits         |
| `release-this`        | Finalize changelog and prepare module for release |
| `release-pending`     | Check if module has pending changes for release   |
| `release-tag-pending` | Check for changelog versions without git tags     |
| `release-check-ci`    | Check CI status for a commit before releasing     |
| `release-get-version` | Extract latest version from changelog             |
| `release-validate`    | Validate changelog format and structure           |
| `release-r2r-cli`     | Create a semver tag for r2r-cli                   |

---

## release-calver

Generate a calendar-versioned tag for a module.

### Synopsis

```bash
r2r eac release-calver <prefix> [options]
```

### Description

Creates CalVer tags in the format `prefix/YYYY.MM.DD` or `prefix/YYYY.MM.DD.N` for multiple releases on the same day. Automatically detects existing tags and increments the suffix.

### Arguments

| Argument | Required | Description                           |
| -------- | -------- | ------------------------------------- |
| `prefix` | Yes      | Tag prefix (typically module moniker) |

### Flags

| Flag        | Short | Type | Default | Description                            |
| ----------- | ----- | ---- | ------- | -------------------------------------- |
| `--create`  |       | bool | `false` | Create the git tag locally             |
| `--push`    |       | bool | `false` | Push tag to remote (requires --create) |
| `--dry-run` | `-n`  | bool | `false` | Show what tag would be created         |

### Examples

```bash
# Preview next tag
r2r eac release-calver auth --dry-run

# Create tag locally
r2r eac release-calver auth --create

# Create and push tag
r2r eac release-calver auth --create --push

# Just display next tag
r2r eac release-calver api
```

### Output

```text
Analyzing existing tags for prefix: auth
Found existing tags:
  - auth/2025.11.29
  - auth/2025.11.30

Next tag: auth/2025.12.01

✓ Tag created: auth/2025.12.01
✓ Tag pushed to remote: origin

Release complete!
```

### Tag Format

- First release today: `auth/2025.12.01`
- Second release today: `auth/2025.12.01.1`
- Third release today: `auth/2025.12.01.2`

### Exit Codes

| Code | Description                        |
| ---- | ---------------------------------- |
| 0    | Tag created/displayed successfully |
| 1    | Error creating tag                 |
| 2    | Tag already exists                 |

---

## release-changelog

Generate or update changelog from commit history.

### Synopsis

```bash
r2r eac release-changelog [module] [options]
```

### Description

Analyzes commits since the last release and generates/updates the CHANGELOG.md file. Uses AI to categorize changes and write clear descriptions.

### Arguments

| Argument | Required | Description                                            |
| -------- | -------- | ------------------------------------------------------ |
| `module` | No       | Module to generate changelog for (defaults to current) |

### Flags

| Flag        | Short | Type   | Default      | Description                |
| ----------- | ----- | ------ | ------------ | -------------------------- |
| `--from`    | `-f`  | string | `<last-tag>` | Starting commit/tag        |
| `--to`      | `-t`  | string | `HEAD`       | Ending commit/tag          |
| `--version` | `-v`  | string | -            | Version for new section    |
| `--dry-run` | `-n`  | bool   | `false`      | Preview without writing    |
| `--debug`   | `-d`  | bool   | `false`      | Save AI generation details |

### Examples

```bash
# Generate changelog for module
r2r eac release-changelog eac-commands

# Generate with specific version
r2r eac release-changelog eac-commands --version 2025.12.01

# Generate from specific range
r2r eac release-changelog --from v1.0.0 --to v1.1.0

# Preview changes
r2r eac release-changelog --dry-run
```

### Output

```text
Generating changelog for eac-commands...

Analyzing commits from auth/2025.11.30 to HEAD...
  ✓ 12 commits found
  ✓ 4 features, 6 fixes, 2 chores

Updating CHANGELOG.md...

## [2025.12.01] - 2025-12-01

### Added
- New authentication middleware for API endpoints
- Support for JWT token refresh

### Fixed
- Resolved race condition in token validation
- Fixed memory leak in session management

### Changed
- Improved error messages for auth failures

✓ Changelog updated: CHANGELOG.md
```

### Exit Codes

| Code | Description                      |
| ---- | -------------------------------- |
| 0    | Changelog generated successfully |
| 1    | Error generating changelog       |
| 2    | No commits found                 |

---

## release-this

Finalize changelog and prepare module for release.

### Synopsis

```bash
r2r eac release-this <module> [options]
```

### Description

Comprehensive release preparation command that:

1. Validates CI status
2. Updates changelog with version
3. Creates release commit
4. Creates and pushes CalVer tag

### Arguments

| Argument | Required | Description               |
| -------- | -------- | ------------------------- |
| `module` | Yes      | Module moniker to release |

### Flags

| Flag              | Short | Type | Default | Description                 |
| ----------------- | ----- | ---- | ------- | --------------------------- |
| `--skip-ci-check` |       | bool | `false` | Skip CI status verification |
| `--no-push`       |       | bool | `false` | Don't push tag to remote    |
| `--dry-run`       | `-n`  | bool | `false` | Preview release steps       |
| `--debug`         | `-d`  | bool | `false` | Enable debug output         |

### Examples

```bash
# Full release workflow
r2r eac release-this eac-commands

# Preview release steps
r2r eac release-this eac-commands --dry-run

# Release without pushing
r2r eac release-this eac-commands --no-push

# Skip CI check (use with caution)
r2r eac release-this eac-commands --skip-ci-check
```

### Output

```text
Preparing release for eac-commands...

Step 1: Checking CI status...
  ✓ All workflows passing for commit abc123

Step 2: Updating changelog...
  ✓ Version 2025.12.01 finalized
  ✓ Changelog updated

Step 3: Creating release commit...
  ✓ Commit created: def456

Step 4: Creating release tag...
  ✓ Tag created: eac-commands/2025.12.01

Step 5: Pushing to remote...
  ✓ Changes pushed to origin
  ✓ Tag pushed to origin

✓ Release complete: eac-commands/2025.12.01
```

### Exit Codes

| Code | Description                    |
| ---- | ------------------------------ |
| 0    | Release completed successfully |
| 1    | Release failed                 |
| 2    | CI checks failing              |
| 3    | Uncommitted changes present    |

---

## release-pending

Check if module has pending changes for release.

### Synopsis

```bash
r2r eac release-pending [module]
```

### Description

Analyzes commits since the last release tag to determine if there are unreleased changes that warrant a new release.

### Arguments

| Argument | Required | Description                       |
| -------- | -------- | --------------------------------- |
| `module` | No       | Module to check (defaults to all) |

### Flags

| Flag     | Short | Type | Default | Description    |
| -------- | ----- | ---- | ------- | -------------- |
| `--json` |       | bool | `false` | Output as JSON |

### Examples

```bash
# Check all modules
r2r eac release-pending

# Check specific module
r2r eac release-pending eac-commands

# JSON output
r2r eac release-pending --json
```

### Output

```text
Checking for pending releases...

Module: eac-commands
  Last release: eac-commands/2025.11.30
  Commits since: 5
  Has pending: Yes

  Pending changes:
    - feat: add new validation command
    - fix: resolve config parsing issue
    - chore: update dependencies

Module: eac-core
  Last release: eac-core/2025.12.01
  Commits since: 0
  Has pending: No

Summary:
  Modules with pending releases: 1
  Modules up to date: 1
```

### Exit Codes

| Code | Description            |
| ---- | ---------------------- |
| 0    | No pending releases    |
| 1    | Pending releases found |

---

## release-tag-pending

Check for changelog versions without corresponding git tags.

### Synopsis

```bash
r2r eac release-tag-pending [module]
```

### Description

Compares changelog versions against git tags to find versions that have changelog entries but no corresponding tag.

### Arguments

| Argument | Required | Description                       |
| -------- | -------- | --------------------------------- |
| `module` | No       | Module to check (defaults to all) |

### Flags

| Flag     | Short | Type | Default | Description    |
| -------- | ----- | ---- | ------- | -------------- |
| `--json` |       | bool | `false` | Output as JSON |

### Examples

```bash
# Check all modules
r2r eac release-tag-pending

# Check specific module
r2r eac release-tag-pending eac-commands

# JSON output
r2r eac release-tag-pending --json
```

### Output

```text
Checking for untagged changelog versions...

Module: eac-commands
  Changelog versions:
    - 2025.12.01 ✗ (no tag)
    - 2025.11.30 ✓
    - 2025.11.29 ✓

  Missing tags: 1

Module: eac-core
  All versions tagged ✓

Summary:
  Modules with missing tags: 1

Create missing tags with:
  r2r eac release-calver eac-commands --create --push
```

### Exit Codes

| Code | Description                      |
| ---- | -------------------------------- |
| 0    | All changelog versions have tags |
| 1    | Missing tags found               |

---

## release-check-ci

Check CI status for a commit before releasing.

### Synopsis

```bash
r2r eac release-check-ci [options]
```

### Description

Queries GitHub Actions to verify all CI workflows pass for a specific commit. Waits for in-progress workflows to complete.

### Flags

| Flag         | Short | Type   | Default | Description                  |
| ------------ | ----- | ------ | ------- | ---------------------------- |
| `--workflow` | `-w`  | string | `ci`    | Workflow name to check       |
| `--commit`   | `-c`  | string | `HEAD`  | Commit SHA to check          |
| `--timeout`  | `-t`  | int    | `600`   | Maximum wait time in seconds |
| `--interval` |       | int    | `30`    | Polling interval in seconds  |

### Examples

```bash
# Check default CI workflow for HEAD
r2r eac release-check-ci

# Check specific workflow and commit
r2r eac release-check-ci --workflow=ci --commit=abc123

# Check with custom timeout
r2r eac release-check-ci --commit=HEAD --timeout=300

# Check current HEAD
r2r eac release-check-ci --commit=$(git rev-parse HEAD)
```

### Output (Success)

```text
Checking workflow 'ci' for commit abc123...

Workflow run found: #1234
Status: in_progress

Polling every 30 seconds (timeout: 600s)...
⏳ Waiting... (30s elapsed)
⏳ Waiting... (60s elapsed)

✓ Workflow completed successfully

All jobs passed:
  ✓ build (2m 34s)
  ✓ test (4m 12s)
  ✓ validate (1m 5s)

Safe to proceed with release.
```

### Output (Failure)

```text
Checking workflow 'ci' for commit abc123...

Workflow run found: #1234
Status: completed
Conclusion: failure

✗ Workflow failed

Failed jobs:
  ✗ test (4m 12s) - exit code 1

Do not proceed with release.
```

### Exit Codes

| Code | Description                   |
| ---- | ----------------------------- |
| 0    | All workflows passed          |
| 1    | One or more workflows failed  |
| 2    | Timeout waiting for workflows |
| 3    | Commit or workflow not found  |

---

## release-get-version

Extract latest version from changelog.

### Synopsis

```bash
r2r eac release-get-version [module]
```

### Description

Parses the CHANGELOG.md file to extract the most recent version number.

### Arguments

| Argument | Required | Description                           |
| -------- | -------- | ------------------------------------- |
| `module` | No       | Module to check (defaults to current) |

### Flags

| Flag     | Short | Type | Default | Description       |
| -------- | ----- | ---- | ------- | ----------------- |
| `--all`  | `-a`  | bool | `false` | List all versions |
| `--json` |       | bool | `false` | Output as JSON    |

### Examples

```bash
# Get latest version
r2r eac release-get-version eac-commands

# List all versions
r2r eac release-get-version eac-commands --all

# JSON output
r2r eac release-get-version --json
```

### Output

```text
Latest version: 2025.12.01
```

### Exit Codes

| Code | Description                    |
| ---- | ------------------------------ |
| 0    | Version extracted successfully |
| 1    | No version found               |
| 2    | Changelog not found            |

---

## release-validate

Validate changelog format and structure.

### Synopsis

```bash
r2r eac release-validate [module]
```

### Description

Validates CHANGELOG.md files for:

- Keep a Changelog format compliance
- Version ordering
- Date format
- Required sections

### Arguments

| Argument | Required | Description                          |
| -------- | -------- | ------------------------------------ |
| `module` | No       | Module to validate (defaults to all) |

### Flags

| Flag       | Short | Type | Default | Description      |
| ---------- | ----- | ---- | ------- | ---------------- |
| `--strict` |       | bool | `false` | Fail on warnings |

### Examples

```bash
# Validate all changelogs
r2r eac release-validate

# Validate specific module
r2r eac release-validate eac-commands

# Strict validation
r2r eac release-validate --strict
```

### Output (Success)

```text
Validating changelogs...

eac-commands/CHANGELOG.md:
  ✓ Format: Keep a Changelog 1.0.0
  ✓ Versions: 5 entries
  ✓ Dates: All valid
  ✓ Sections: Valid structure

✓ All changelogs valid
```

### Output (Errors)

```text
Validating changelogs...

eac-commands/CHANGELOG.md:
  ✗ Line 15: Invalid date format (expected YYYY-MM-DD)
  ✗ Line 23: Unknown section type "Updates"
  ⚠ Line 30: Empty section "Removed"

✗ Validation failed: 2 errors, 1 warning
```

### Exit Codes

| Code | Description                |
| ---- | -------------------------- |
| 0    | Validation passed          |
| 1    | Validation failed          |
| 2    | Warning only (strict mode) |

---

## release-r2r-cli

Create a semantic version tag for r2r-cli.

### Synopsis

```bash
r2r eac release-r2r-cli <version> [options]
```

### Description

Creates a SemVer tag specifically for the r2r-cli tool. Unlike CalVer used for modules, r2r-cli uses semantic versioning for compatibility signaling.

### Arguments

| Argument  | Required | Description                    |
| --------- | -------- | ------------------------------ |
| `version` | Yes      | Semantic version (e.g., 1.2.0) |

### Flags

| Flag        | Short | Type | Default | Description                  |
| ----------- | ----- | ---- | ------- | ---------------------------- |
| `--dry-run` | `-n`  | bool | `false` | Preview without creating tag |
| `--push`    |       | bool | `true`  | Push tag to remote           |

### Examples

```bash
# Preview release
r2r eac release-r2r-cli 1.2.0 --dry-run

# Create and push tag (default)
r2r eac release-r2r-cli 1.2.0

# Create tag locally only
r2r eac release-r2r-cli 1.2.0 --push=false

# Pre-release version
r2r eac release-r2r-cli 2.0.0-beta.1
```

### Output

```text
Creating release tag for r2r-cli...

Version: 1.2.0
Tag: r2r-cli/1.2.0

Validating...
  ✓ Version format valid
  ✓ Tag doesn't exist

Creating tag...
  ✓ Tag created: r2r-cli/1.2.0
  ✓ Tag pushed to remote

Release complete!
```

### Version Format

Must follow SemVer:

- `MAJOR.MINOR.PATCH` (e.g., `1.2.0`)
- Pre-release: `MAJOR.MINOR.PATCH-prerelease` (e.g., `1.2.0-beta.1`)
- Build metadata: `MAJOR.MINOR.PATCH+build` (e.g., `1.2.0+20251201`)

### Exit Codes

| Code | Description              |
| ---- | ------------------------ |
| 0    | Tag created successfully |
| 1    | Error creating tag       |
| 2    | Invalid version format   |
| 3    | Tag already exists       |

---

## Common Workflows

### Standard Module Release

```bash
# 1. Check for pending changes
r2r eac release-pending eac-commands

# 2. Generate changelog
r2r eac release-changelog eac-commands

# 3. Verify CI status
r2r eac release-check-ci --commit=HEAD

# 4. Create release
r2r eac release-calver eac-commands --create --push
```

### Full Release with Validation

```bash
# 1. Validate changelog format
r2r eac release-validate eac-commands

# 2. Check for untagged versions
r2r eac release-tag-pending eac-commands

# 3. Prepare and release
r2r eac release-this eac-commands
```

### Tool Release (SemVer)

```bash
# 1. Update version in code
# ... edit version constants ...

# 2. Update changelog
r2r eac release-changelog r2r-cli --version 1.2.0

# 3. Commit changes
git add . && git commit -m "chore: bump version to 1.2.0"
git push origin main

# 4. Check CI
r2r eac release-check-ci --commit=HEAD

# 5. Create release
r2r eac release-r2r-cli 1.2.0
```

### CI/CD Integration

```yaml
name: Release

on:
  workflow_dispatch:
    inputs:
      module:
        description: 'Module to release'
        required: true

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Check CI status
        run: r2r eac release-check-ci --commit=${{ github.sha }}

      - name: Create release
        run: r2r eac release-calver ${{ inputs.module }} --create --push
```

---

## Related Documentation

- [Release Overview](release-overview.md) - Concepts and versioning strategies
- [Release Configuration](release-configuration.md) - Configuration reference
- [Pipeline Commands](pipeline-commands.md) - CI/CD integration
