# Generate Changelog

## What You'll Accomplish

Create changelog from Git commits with automatic version detection and formatting. This is a **manual step** in the automated release workflow.

## Prerequisites

- Git repository with commits
- Module has changes since last release
- Conventional commit messages used

## Steps

### 1. Generate Changelog

**Action Type**: 🧑 Manual

```bash
r2r release this my-module
```

**What happens**:

- Analyzes commits since last release (tag `my-module/1.2.3`)
- Determines version bump (patch/minor/major for SemVer, date for CalVer)
- Generates/updates `release/my-module/CHANGELOG.md`

**Path discovery**: Command finds changelog via module contract (`versioning.changelog` in `.r2r/eac/repository.yml`), defaulting to `release/<module>/CHANGELOG.md`. See [Understanding the Release Folder](./understanding-release-folder.md) for details.

### 2. Review Generated Changelog

**Action Type**: 🧑 Manual

```bash
cat release/my-module/CHANGELOG.md
```

**What happens**: View generated changelog content

### 3. Edit if Needed

**Action Type**: 🧑 Manual

```bash
code release/my-module/CHANGELOG.md
```

**What happens**: Manually refine entries if descriptions need improvement

### 4. Validate Format

**Action Type**: 🧑 Manual

```bash
r2r validate release my-module
```

**What happens**: Checks changelog follows Keep a Changelog format standards

## Changelog Sections

Generated changelog includes:

- **Added** - New features
- **Changed** - Changes to existing features
- **Deprecated** - Soon-to-be removed features
- **Removed** - Removed features
- **Fixed** - Bug fixes
- **Security** - Security updates

## Example Scenario

Preparing my-module release after multiple commits:

```bash
# Generate changelog for my-module
r2r release this my-module

# Output:
moniker: my-module
current_version: 1.2.3
next_version: 1.2.4
change_summary:
  feat: 2
  fix: 1

Updated: release/my-module/CHANGELOG.md

# Review
cat release/my-module/CHANGELOG.md

# ## [1.2.4] - 2026-01-09
#
# ### Added
# - feat: add config validation for module contracts
# - feat: support custom templates in documentation
#
# ### Fixed
# - fix: handle empty files correctly in changelog parser
#
# ## [1.2.3] - 2026-01-08
# ...

# Validate format
r2r validate release my-module
# Output:
# ✓ Changelog exists: release/my-module/CHANGELOG.md
# ✓ Valid header format
# ✓ Valid version format: 1.2.4
# ✓ Versions in descending order

# Get version for verification
r2r release get-version my-module
# Output: 1.2.4
```

## Version Detection

Version is determined by commit types (for SemVer modules):

- `feat:` → Minor bump (1.2.3 → 1.3.0) *or Patch in pre-1.0 versions*
- `fix:` → Patch bump (1.2.3 → 1.2.4)
- `BREAKING CHANGE:` → Major bump (1.3.0 → 2.0.0)

**For CalVer modules**: Version is today's date (YYYY.MMDD)

**Note**: This step is the same for both CDe (Continuous Deployment) and RA (Release Approval) patterns. See [Release Workflow Variants](./release-workflow-variants.md) to understand how workflows diverge after changelog preparation.

## Next Steps

- [Release Workflow Variants](./release-workflow-variants.md) → Choose CDe or RA pattern
- [Prepare Module Release](./prepare-module-release.md) → Complete CDe workflow
- [Understanding the Release Folder](./understanding-release-folder.md) → Learn folder structure

## Related Commands

- [`release changelog`](../../../../reference/commands/release/changelog.md) - Generate changelog
- [`release get-version`](../../../../reference/commands/release/get-version.md) - Extract version
- [`validate release`](../../../../reference/commands/validate/release.md) - Validate format
