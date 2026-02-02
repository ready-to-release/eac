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
eac show approval-comments my-module

# or explicitly
eac show approval-comments my-module unreleased

# Include all review states (APPROVED, CHANGES_REQUESTED, COMMENTED)
eac show approval-comments my-module --include-all-reviews
```

**Output:**

```markdown
# PR Approvals: my-module (Unreleased)

**Summary:** 3 PRs, 5 approvals

| PR | Title | Reviewer | Review State | Reviewed At |
|----|-------|----------|--------------|-------------|
| #123 | Add new feature | @reviewer1 | APPROVED | 2025-12-10 |
| #123 | Add new feature | @reviewer2 | APPROVED | 2025-12-10 |
| #124 | Update spec | @reviewer3 | APPROVED | 2025-12-11 |

## PR Details

### PR #123: Add new feature

**Description:**

This PR adds a new feature to the module with enhanced
functionality for processing specifications.

**Merge Commit Message:**

Merge pull request #123 from user/feature-branch

Added new feature implementation with tests

---

### PR #124: Update spec

**Description:**

Updated specification files for the module

**Merge Commit Message:**

Merge pull request #124 from user/spec-update

Updated module specification

---
```

## View Approvals for Latest Release

Show approval comments included in the most recent release:

```bash
eac show approval-comments my-module latest
```

## View Approvals for Specific Version

```bash
eac show approval-comments my-module 1.2.3
```

## Query from Different Branches

By default, commands query from the trunk branch (usually `main`). Use `--branch` to query from other branches:

```bash
# Query from main branch (default)
eac show approval-comments my-module

# Query from current branch (useful when working in feature branches)
eac show approval-comments my-module --branch HEAD

# Query from specific branch
eac show approval-comments my-module --branch develop
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
eac get approval-comments my-module --as-json

# Include all review states
eac get approval-comments my-module --include-all-reviews --as-json
```

### As YAML (default)

```bash
eac get approval-comments my-module

# or explicitly
eac get approval-comments my-module --as-yaml

# Include all review states
eac get approval-comments my-module --include-all-reviews --as-yaml
```

### As TOML

```bash
eac get approval-comments my-module --as-toml

# Include all review states
eac get approval-comments my-module --include-all-reviews --as-toml
```

## Common Use Cases

### 1. Count Total Approvals

```bash
eac get approval-comments my-module --as-json | jq '.total_approvals'
```

**Example Output:**

```text
5
```

### 2. List All Reviewers

```bash
eac get approval-comments my-module --as-json | jq -r '.approvals[].reviewer' | sort -u
```

**Example Output:**

```text
reviewer1
reviewer2
reviewer3
reviewer4
```

### 3. Count Approvals Per PR

```bash
eac get approval-comments my-module --as-json | jq '.approvals | group_by(.pr_number) | map({pr: .[0].pr_number, title: .[0].pr_title, approvals: length})'
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
eac get approval-comments my-module --as-json | jq '.approvals[] | select(.reviewer == "reviewer1")'
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
eac get approval-comments my-module --as-json | jq '.approvals | group_by(.pr_number) | map(select(length > 1) | {pr: .[0].pr_number, title: .[0].pr_title, approvals: length})'
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
eac get approval-comments my-module --as-json | jq '.approvals | group_by(.pr_number) | map(select(any(.[]; .reviewer == "reviewer1")) | {pr: .[0].pr_number, title: .[0].pr_title})'
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
eac get approval-comments my-module --include-all-reviews --as-json | jq '.approvals | group_by(.review_state) | map({state: .[0].review_state, count: length})'
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
eac get approval-comments my-module --include-all-reviews --as-json | jq '.approvals | group_by(.pr_number) | map(select(any(.[]; .review_state == "CHANGES_REQUESTED")) | {pr: .[0].pr_number, title: .[0].pr_title})'
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
echo "- Total PRs: $(eac get approval-comments my-module latest --as-json | jq '.total_prs')"
echo "- Total Approvals: $(eac get approval-comments my-module latest --as-json | jq '.total_approvals')"
echo ""
echo "### Reviewers:"
eac get approval-comments my-module latest --as-json | jq -r '.approvals[].reviewer' | sort -u | sed 's/^/- /'
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

---

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
eac show approval-comments my-module

# Show all review states
eac show approval-comments my-module --include-all-reviews
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

---

## File Location Requirements

- Specification files must be in `specs/<module>/` directory
- Files must have `.feature` extension
- Files must be part of merged pull requests
- PRs must have at least one APPROVED review

## See Also

- [show approval-comments Reference](../../../../reference/eac/commands/show/approval-comments.md) - Complete command reference
- [get approval-comments Reference](../../../../reference/eac/commands/get/approval-comments.md) - JSON/YAML output reference
- [View Specifications](./view-specifications.md) - View spec files for a release
- [View Changelog and Release Notes](./view-changelog-release-notes.md) - View changes and notes
