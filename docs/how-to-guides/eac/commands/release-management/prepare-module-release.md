# Prepare Module Release

## What You'll Accomplish

Prepare and release a new version of a deployable module following the changelog-driven release workflow. This guide shows both the manual steps you perform and the automated CI steps that happen after merge.

> **Release Pattern Selection**: This guide describes the **CDe (Continuous Deployment)** pattern that auto-deploys from main after merge. For regulated environments requiring formal approval (GxP, financial, safety-critical), see [Release Workflow Variants](./release-workflow-variants.md) for the **RA (Release Approval)** pattern that uses release branches.

## Prerequisites

**Before starting**:

- ✅ Have permission to create pull requests
- ✅ Understand your repository's release process

**Module requirements**:

- ✅ Module has pending changes to release
- ✅ CI is passing for HEAD commit
- ✅ All quality gates met

---

## Workflow Overview

The release process is split into two phases:

### 🧑 Manual Phase (You Perform)

1. Check pending changes
2. Generate changelog
3. Review and validate changelog
4. Verify CI status
5. Commit changelog changes
6. Create and merge PR

### 🤖 Automated Phase (CI Performs)

<!-- markdownlint-disable MD029 -->

7. Detect changelog changes
8. Create git tag
9. Build and publish release

<!-- markdownlint-enable MD029 -->

---

## Manual Phase: Your Actions

These are the steps **you** perform as a developer.

### Step 1: Check Pending Changes

**Action Type**: 🧑 Manual

```bash
r2r release pending my-module
```

**Output**:

```yaml
moniker: my-module
has_changes: true
current_version: 1.2.3
next_version: 1.2.4
change_summary:
  feat: 2
  fix: 1
```

**What this means**:

- Module **has changes** that should be released
- Current version is `1.2.3`
- Next version will be `1.2.4` (patch bump for fixes)
- Contains 2 new features and 1 bug fix

**If no changes**: Skip release - nothing to do.

---

### Step 2: Generate Changelog

**Action Type**: 🧑 Manual

```bash
r2r release this my-module
```

**What happens**:

- Analyzes commits since `my-module/1.2.3` tag
- Generates changelog entries from conventional commits
- Calculates version bump (patch/minor/major for SemVer, date for CalVer)
- Updates `release/my-module/CHANGELOG.md` (path discovered via module contract)

**Note**: Changelog preparation is the same for both CDe and RA patterns. Workflows diverge after this step.

**Output**:

```text
Analyzing commits since my-module/1.2.3...
Found 3 commits:
  - feat: add config validation
  - feat: support custom templates
  - fix: handle empty files correctly

Version bump: patch (2 feat + 1 fix = patch in pre-1.0)
Next version: 1.2.4

Updated: release/my-module/CHANGELOG.md
```

---

### Step 3: Review Generated Changelog

**Action Type**: 🧑 Manual

```bash
cat release/my-module/CHANGELOG.md
```

**Check for**:

- ✅ Correct version number (`1.2.4`)
- ✅ Today's date
- ✅ All changes categorized correctly
- ✅ Clear, user-friendly descriptions
- ✅ Breaking changes documented (if any)

**Edit if needed**:

```bash
# Edit manually if descriptions need improvement
code release/my-module/CHANGELOG.md
```

**Example changelog excerpt**:

```markdown
# Changelog

All notable changes to **r2r CLI** will be documented in this file.

## [1.2.4] - 2026-01-09

### Added
- feat: add config validation for module contracts
- feat: support custom templates in documentation generation

### Fixed
- fix: handle empty files correctly in changelog parser

## [1.2.3] - 2026-01-08

...
```

---

### Step 4: Validate Changelog Format

**Action Type**: 🧑 Manual

```bash
r2r validate release my-module
```

**Output** (success):

```text
✓ Changelog exists: release/my-module/CHANGELOG.md
✓ Valid header format
✓ Valid version format: 1.2.4
✓ Versions in descending order
✓ No duplicate versions
✓ Valid date format: 2026-01-09
```

**Output** (failure):

```text
✗ Invalid version format: v1.2.4 (should be 1.2.4)
✗ Date format invalid: 2026-1-9 (should be 2026-01-09)
```

**If validation fails**: Fix errors and validate again.

---

### Step 5: Verify CI Status

**Action Type**: 🧑 Manual

```bash
r2r release check-ci --workflow ci-my-module.yml --commit $(git rev-parse HEAD)
```

**Output** (passing):

```text
✓ CI workflow: ci-my-module.yml
✓ Commit: abc123def
✓ Status: success
✓ Completed: 2 minutes ago
✓ All jobs passed
```

**Output** (failing):

```text
✗ CI workflow: ci-my-module.yml
✗ Status: failure
✗ Failed job: test
✗ View logs: https://github.com/.../actions/runs/...
```

**If CI failing**: Fix test failures before continuing.

---

### Step 6: Commit and Create PR

**Action Type**: 🧑 Manual

```bash
# Stage changelog
git add release/my-module/CHANGELOG.md

# Commit with conventional format
git commit -m "release(my-module): prepare 1.2.4 release"

# Push to branch
git push origin your-branch

# Create PR
gh pr create \
  --title "release(my-module): 1.2.4" \
  --body "Prepare my-module version 1.2.4 release

## Changes
- Add config validation
- Support custom templates
- Fix empty file handling

## Pre-release Checklist
- [x] Changelog updated
- [x] CI passing
- [x] Quality gates met
- [x] Breaking changes documented (N/A)
"
```

**What to include in PR**:

- Clear title: `release(module): version`
- Summary of changes
- Pre-release checklist
- Link to release notes

---

### Step 7: Get Approval and Merge

**Action Type**: 🧑 Manual

**Required approvals**:

- ✅ Code owner review
- ✅ CI validation (automated)
- ✅ Security scan (automated)

**Review focus**:

- Changelog accuracy
- Version number correctness
- Breaking changes clearly documented
- Quality gates all passing

**After approval**: Click "Merge pull request" → "Confirm merge"

**After merge**: Automation takes over. Your manual work is done.

---

## Automated Phase: CI Actions

These steps happen **automatically** in GitHub Actions after you merge the PR.

### Step 8: Detect Changelog Changes

**Action Type**: 🤖 CI Automation

**Workflow**: `.github/workflows/release-trigger.yml`

**What happens**:

1. Workflow detects `release/my-module/CHANGELOG.md` changed
2. Extracts latest version from changelog: `1.2.4`
3. Checks if tag `my-module/1.2.4` already exists
4. Prepares to create tag

**You don't need to do anything.**

---

### Step 9: Create Git Tag

**Action Type**: 🤖 CI Automation

**Workflow**: `.github/workflows/release-trigger.yml`

**What happens**:

```bash
# CI automatically runs:
git tag -a my-module/1.2.4 -m "Release my-module 1.2.4"
git push origin my-module/1.2.4
```

**Tag format**: `{moniker}/{version}`

- SemVer example: `my-module/1.2.4`
- CalVer example: `my-calver-module/2026.0109`

**You don't create tags manually.**

---

### Step 10: Build and Publish Release

**Action Type**: 🤖 CI Automation

**Workflow**: `.github/workflows/release-my-module.yml`

**What happens**:

1. **Verify CI passed** for the commit
2. **Build artifacts**:
   - Linux: amd64, arm64
   - macOS: amd64, arm64
   - Windows: amd64
   - UPX-compressed variants
3. **Generate attestations** (Sigstore build provenance)
4. **Create GitHub release** with artifacts
5. **Upload binaries** to release

**Typical duration**: 10-15 minutes

**Verify release created**:

```bash
# Check workflow status
gh run list --workflow=release-my-module.yml --limit 1

# View release
gh release view my-module/1.2.4
```

---

## Key Takeaways

1. **You prepare, CI publishes** - Your job is updating changelog and creating PR; CI handles tagging and publishing
2. **Changelog drives releases** - Merging updated changelog automatically triggers release workflow
3. **PR approval = production approval** - When you approve release PR, you're approving deployment
4. **Tags are automatic** - Don't create tags manually; CI creates them after merge
5. **Quality gates must pass** - All 5 gates (tests, coverage, bugs, performance, security) must be green

---

## Next Steps

- **[Release Workflow Variants](./release-workflow-variants.md)** - Learn RA pattern for regulated environments
- **[Generate Changelog](generate-changelog.md)** - Deep dive into changelog generation
- **[Understanding the Release Folder](./understanding-release-folder.md)** - Learn folder structure
- **[Check CI Before Release](check-ci-before-release.md)** - CI validation details
- **[Understanding Tag Creation](create-release-tag.md)** - Learn about automated tag creation
