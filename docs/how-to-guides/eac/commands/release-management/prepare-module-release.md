# Prepare Module Release

## What You'll Accomplish

Prepare and release a new version of your module following a complete pre-release checklist with changelog generation, CI validation, and git tagging.

## Prerequisites

### Required Setup

- Module has pending changes to release
- CI is passing for HEAD commit
- You have permission to create git tags

## Steps

### 1. Check What Needs Releasing

```bash
r2r eac release pending
```

**What happens**: Shows modules with unreleased commits

If no changes, you're done! Otherwise continue.

### 2. Generate Changelog

```bash
r2r eac release changelog
```

**What happens**: Analyzes commits since last release, generates `CHANGELOG.md` with version and changes

### 3. Review and Edit Changelog

```bash
# Review generated changelog
cat CHANGELOG.md

# Edit if needed
nano CHANGELOG.md
```

### 4. Validate Changelog Format

```bash
r2r eac validate release
```

**What happens**: Checks changelog follows format standards

Fix any validation errors and validate again.

### 5. Verify CI Passes

```bash
r2r eac release check-ci $(git rev-parse HEAD)
```

**What happens**: Confirms CI is green for current commit

If CI is failing, fix issues before releasing.

### 6. Create Release

```bash
r2r eac release this
```

**What happens**: Creates git tag for release and finalizes changelog

## Example Scenario

Releasing version 1.2.0 of src-auth module:

```bash
# Check pending changes
r2r eac release pending
# Module src-auth has 5 unreleased commits

# Generate changelog
r2r eac release changelog
# Generated CHANGELOG.md with v1.2.0

# Validate format
r2r eac validate release
# ✓ Changelog format valid

# Check CI status
r2r eac release check-ci $(git rev-parse HEAD)
# ✓ CI passing for commit abc123

# Create release
r2r eac release this
# ✓ Created tag: src-auth/v1.2.0
# ✓ Finalized changelog

# Push tag to remote
git push origin src-auth/v1.2.0
```

## Common Issues

| Problem | Solution |
|---------|----------|
| "No pending changes" | Nothing to release |
| "CI not passing" | Fix failures before releasing |
| "Changelog invalid" | Fix format errors |
| "Tag already exists" | Version already released |

## Next Steps

- [Generate Changelog](./generate-changelog.md) → More changelog details
- [Check CI Before Release](./check-ci-before-release.md) → CI validation

## Related Commands

- [`release pending`](../../../../reference/commands/release/pending.md) - Check pending changes
- [`release changelog`](../../../../reference/commands/release/changelog.md) - Generate changelog
- [`release this`](../../../../reference/commands/release/this.md) - Create release
- [`release check-ci`](../../../../reference/commands/release/check-ci.md) - Verify CI
