# View PR Approval Comments
## What You'll Accomplish

Learn how to view and analyze PR approval comments for specification-related pull requests using the `show approval-comments` and `get approval-comments` commands.

## Prerequisites

- Repository with EAC configuration
- Module with specification files in `specs/<module>/` directory
- GitHub repository with merged pull requests
- **Optional**: GitHub CLI (`gh`) installed and authenticated for full functionality

## Install GitHub CLI (Recommended)

The commands work best with GitHub CLI installed:

```bash
# Install GitHub CLI
# See: https://cli.github.com/

# Authenticate
gh auth login
```

**Note:** Commands degrade gracefully if `gh` is unavailable but will show limited data.

## View Approvals for Unreleased PRs

Show approval comments for PRs merged since the last release:

```bash
# Human-readable markdown output (only APPROVED reviews)
r2r eac show approval-comments ext-eac

# or explicitly
r2r eac show approval-comments ext-eac unreleased

# Include all review states (APPROVED, CHANGES_REQUESTED, COMMENTED)
r2r eac show approval-comments ext-eac --include-all-reviews
```

**Output:**
```markdown
# PR Approvals: ext-eac (Unreleased)

**Summary:** 3 PRs, 5 approvals

| PR | Title | Reviewer | Review State | Reviewed At |
|----|-------|----------|--------------|-------------|
| #123 | Add new feature | @reviewer1 | APPROVED | 2025-12-10 |
| #123 | Add new feature | @reviewer2 | APPROVED | 2025-12-10 |
| #124 | Update spec | @reviewer3 | APPROVED | 2025-12-11 |

## PR Details

### PR #123: Add new feature

**Description:**

This PR adds a new feature to the eac-commands module with enhanced
functionality for processing specifications.

**Merge Commit Message:**

Merge pull request #123 from user/feature-branch

Added new feature implementation with tests

---

### PR #124: Update spec

**Description:**

Updated specification files for r2r-cli module

**Merge Commit Message:**

Merge pull request #124 from user/spec-update

Updated r2r-cli specification

---
```

## View Approvals for Latest Release

Show approval comments included in the most recent release:

```bash
r2r eac show approval-comments ext-eac latest
```

## View Approvals for Specific Version

```bash
r2r eac show approval-comments ext-eac 0.0.7
```

## Query from Different Branches

By default, commands query from the trunk branch (usually `main`). Use `--branch` to query from other branches:

```bash
# Query from main branch (default)
r2r eac show approval-comments ext-eac

# Query from current branch (useful when working in feature branches)
r2r eac show approval-comments ext-eac --branch HEAD

# Query from specific branch
r2r eac show approval-comments ext-eac --branch develop
```

**When to use this:**
- Working in a feature branch and want to see approvals relative to that branch
- Comparing approvals across different branches
- CI/CD pipelines running on non-main branches

**Note:** This fixes the issue where running commands from a feature branch would fail with "unknown revision" errors because tags don't exist on the feature branch. Now by default, commands query from the main branch where releases live.

## Export Structured Data

### As JSON

```bash
# Only APPROVED reviews (default)
r2r eac get approval-comments ext-eac --as-json

# Include all review states
r2r eac get approval-comments ext-eac --include-all-reviews --as-json
```

### As YAML (default)

```bash
r2r eac get approval-comments ext-eac

# or explicitly
r2r eac get approval-comments ext-eac --as-yaml

# Include all review states
r2r eac get approval-comments ext-eac --include-all-reviews --as-yaml
```

### As TOML

```bash
r2r eac get approval-comments ext-eac --as-toml

# Include all review states
r2r eac get approval-comments ext-eac --include-all-reviews --as-toml
```

## Common Use Cases

### 1. Count Total Approvals

```bash
r2r eac get approval-comments ext-eac --as-json | jq '.total_approvals'
```

**Example Output:**
```
5
```

### 2. List All Reviewers

```bash
r2r eac get approval-comments ext-eac --as-json | jq -r '.approvals[].reviewer' | sort -u
```

**Example Output:**
```
reviewer1
reviewer2
reviewer3
reviewer4
```

### 3. Count Approvals Per PR

```bash
r2r eac get approval-comments ext-eac --as-json | jq '.approvals | group_by(.pr_number) | map({pr: .[0].pr_number, title: .[0].pr_title, approvals: length})'
```

**Example Output:**
```json
[
  {
    "pr": 123,
    "title": "Add new feature",
    "approvals": 2
  },
  {
    "pr": 124,
    "title": "Update spec",
    "approvals": 1
  }
]
```

### 4. Filter Approvals by Specific Reviewer

```bash
r2r eac get approval-comments ext-eac --as-json | jq '.approvals[] | select(.reviewer == "reviewer1")'
```

**Example Output:**
```json
{
  "pr_number": 123,
  "pr_title": "Add new feature",
  "reviewer": "reviewer1",
  "review_state": "APPROVED",
  "reviewed_at": "2025-12-10T10:30:00Z"
}
```

### 5. List PRs with Multiple Approvals

```bash
r2r eac get approval-comments ext-eac --as-json | jq '.approvals | group_by(.pr_number) | map(select(length > 1) | {pr: .[0].pr_number, title: .[0].pr_title, approvals: length})'
```

**Example Output:**
```json
[
  {
    "pr": 123,
    "title": "Add new feature",
    "approvals": 2
  }
]
```

### 6. Find PRs Approved by Specific User

```bash
r2r eac get approval-comments ext-eac --as-json | jq '.approvals | group_by(.pr_number) | map(select(any(.[]; .reviewer == "reviewer1")) | {pr: .[0].pr_number, title: .[0].pr_title})'
```

**Example Output:**
```json
[
  {
    "pr": 123,
    "title": "Add new feature"
  },
  {
    "pr": 125,
    "title": "Fix bug"
  }
]
```

### 7. Count Reviews by State

Count how many reviews of each type exist:

```bash
r2r eac get approval-comments ext-eac --include-all-reviews --as-json | jq '.approvals | group_by(.review_state) | map({state: .[0].review_state, count: length})'
```

**Example Output:**
```json
[
  {
    "state": "APPROVED",
    "count": 5
  },
  {
    "state": "CHANGES_REQUESTED",
    "count": 2
  },
  {
    "state": "COMMENTED",
    "count": 1
  }
]
```

### 8. Find PRs with Changes Requested

Identify PRs that need attention:

```bash
r2r eac get approval-comments ext-eac --include-all-reviews --as-json | jq '.approvals | group_by(.pr_number) | map(select(any(.[]; .review_state == "CHANGES_REQUESTED")) | {pr: .[0].pr_number, title: .[0].pr_title})'
```

**Example Output:**
```json
[
  {
    "pr": 123,
    "title": "Add new feature"
  }
]
```

### 9. Generate Compliance Report

Create a summary for regulatory compliance:

```bash
echo "## PR Approval Summary"
echo ""
echo "- Total PRs: $(r2r eac get approval-comments ext-eac latest --as-json | jq '.total_prs')"
echo "- Total Approvals: $(r2r eac get approval-comments ext-eac latest --as-json | jq '.total_approvals')"
echo ""
echo "### Reviewers:"
r2r eac get approval-comments ext-eac latest --as-json | jq -r '.approvals[].reviewer' | sort -u | sed 's/^/- /'
```

**Example Output:**
```markdown
## PR Approval Summary

- Total PRs: 3
- Total Approvals: 5

### Reviewers:
- reviewer1
- reviewer2
- reviewer3
- reviewer4
```

## Troubleshooting

### "module not found" Error

**Problem:** The module moniker is invalid or doesn't exist.

**Solution:** List available modules:
```bash
r2r eac show modules
```

### "version not found" Error

**Problem:** The specified version doesn't exist in the changelog.

**Solution:** List available versions:
```bash
r2r eac get changelog ext-eac --as-json | jq -r '.versions[].number'
```

### "no released versions found" Error

**Problem:** Using `latest` keyword but module has no releases yet.

**Solution:** This is normal for new modules. Use `unreleased` or omit version parameter:
```bash
r2r eac show approval-comments ext-eac unreleased
```

### No Approvals Shown

**Problem:** Command shows "No PR approvals found for this version."

**Possible Causes:**
- No PRs were merged in the version range
- PRs don't contain `.feature` specification files
- No APPROVED reviews on spec-related PRs (default behavior only shows APPROVED)
- GitHub CLI not available

**Solution:** Verify PRs and spec files:
```bash
# Check recent commits for PR references
git log --oneline | grep "#"

# Check for .feature files in commits
git log --name-only | grep ".feature"

# Try including all review states (not just APPROVED)
r2r eac show approval-comments ext-eac --include-all-reviews

# Manually check a specific PR (if gh CLI available)
gh pr view 123 --json files,reviews
```

### GitHub CLI Not Available

**Problem:** `gh` command not found or not authenticated.

**Solution:**
```bash
# Install GitHub CLI
# See: https://cli.github.com/

# Authenticate
gh auth login

# Test authentication
gh auth status
```

**Graceful Degradation**: The command continues without `gh` CLI but shows limited data from git commits only.

### Bundle Modules Show Approvals from Multiple Modules

**Question:** Why do I see approvals for `eac-commands` PRs when querying `ext-eac`?

**Answer:** Container/bundle modules like `ext-eac` automatically **aggregate approvals from all their dependencies**. This is intentional and provides a complete view of all PR approvals included in the release bundle.

**Example:**
- `ext-eac` depends on `eac-commands` and `r2r-cli`
- Running `r2r eac show approval-comments ext-eac` shows approvals from all three modules
- This ensures you see the full scope of reviews for the release

**To see only a specific module's approvals:**
```bash
# Query the dependency directly
r2r eac show approval-comments eac-commands
```

## PR Detection

The command detects PRs using these patterns in commit messages:

- `Merge pull request #123 from branch`
- `feat: add feature (#123)`
- `fix: bug fix (#456)`

Only PRs containing `.feature` files in `specs/<module>/` are included.

## Review States

By default, only reviews with state `APPROVED` are shown. Other states (`CHANGES_REQUESTED`, `COMMENTED`) are filtered out.

To include all review states, use the `--include-all-reviews` flag:

```bash
# Show only APPROVED reviews (default)
r2r eac show approval-comments ext-eac

# Show all review states
r2r eac show approval-comments ext-eac --include-all-reviews
```

**Available review states:**
- **APPROVED**: Reviewer approved the changes
- **CHANGES_REQUESTED**: Reviewer requested changes before approval
- **COMMENTED**: Reviewer left comments without explicit approval/rejection

## Understanding the Output

The `show approval-comments` command displays two sections:

### 1. Summary Table

A compact table showing:
- PR number (as a link reference like #123)
- PR title
- Reviewer username
- Review state (APPROVED, CHANGES_REQUESTED, COMMENTED)
- Date reviewed

### 2. PR Details Section

Detailed information for **all PRs found** (even those with no reviews) including:
- **Description**: The PR body/description text from GitHub
- **Merge Commit Message**: The full merge commit message (headline + body)

**Important:** This section displays for ALL PRs that contain spec files, regardless of whether they have any reviews/approvals. This means you'll see PR details even when "No PR approvals found" is shown.

If a PR has no description or merge message, it will show "(No description provided)" or "(No merge message)" respectively.

## File Location Requirements

- Specification files must be in `specs/<module>/` directory
- Files must have `.feature` extension
- Files must be part of merged pull requests
- PRs must have at least one APPROVED review

## See Also

- [show approval-comments Reference](../../../../reference/commands/show/approval-comments.md) - Complete command reference
- [get approval-comments Reference](../../../../reference/commands/get/approval-comments.md) - JSON/YAML output reference
- [View Specifications](./view-specifications.md) - View spec files for a release
- [View Changelog and Release Notes](./view-changelog-release-notes.md) - View changes and notes
