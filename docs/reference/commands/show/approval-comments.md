# Show approval-comments

<!-- book:cmd show approval-comments -->

## Special Keywords

| Keyword      | Description                                          |
| ------------ | ---------------------------------------------------- |
| `latest`     | Show approvals from the most recent released version |
| `unreleased` | Show approvals since last release (default)          |
| _(omit)_     | Same as `unreleased` - approvals since last release  |

## Bundle Modules

For **container/bundle modules** with dependencies, approvals are **aggregated from all dependent modules**.

**Example:** When querying `ext-eac` (which depends on `eac-commands` and `r2r-cli`):

```bash
r2r eac show approval-comments ext-eac
```

**Shows approvals from PRs containing:**

- `specs/eac-commands/` (dependency)
- `specs/r2r-cli/` (dependency)
- `specs/ext-eac/` (if any)

This provides a **complete view** of all PR approvals included in the release bundle.

**Regular modules** (without dependencies) only show approvals from their own `specs/<module>/` directory.

## Output Format

The command displays two sections:

### 1. Summary Table

Markdown-formatted table with:

- Version header (`# PR Approvals: module (version)`)
- Summary line with PR and approval counts
- Table columns: PR, Title, Reviewer, Review State, Reviewed At
- Review states: By default only APPROVED. With `--include-all-reviews`: APPROVED, CHANGES_REQUESTED, COMMENTED

### 2. PR Details Section

Detailed information for **all PRs found** (even those with no reviews) including:

- PR description/body from GitHub
- Merge commit message (headline + body)

**Note:** This section displays for ALL PRs that contain spec files, regardless of whether they have any reviews/approvals.

If a PR has no description or merge message, it displays "(No description provided)" or "(No merge message)" respectively.

## Example Output

**Default output (only APPROVED reviews):**

```markdown
# PR Approvals: ext-eac (0.0.7)

**Summary:** 3 PRs, 3 approvals

| PR   | Title           | Reviewer   | Review State | Reviewed At |
| ---- | --------------- | ---------- | ------------ | ----------- |
| #123 | Add new feature | @reviewer1 | APPROVED     | 2025-12-10  |
| #124 | Update spec     | @reviewer3 | APPROVED     | 2025-12-11  |
| #125 | Fix bug         | @reviewer1 | APPROVED     | 2025-12-12  |

## PR Details

### PR #123: Add new feature

**Description:**

This PR adds a new feature to the eac-commands module with enhanced
functionality for processing specifications.

**Merge Commit Message:**

Merge pull request #123 from user/feature-branch

Added new feature implementation

---

### PR #124: Update spec

**Description:**

Updated specification files for r2r-cli module

**Merge Commit Message:**

Merge pull request #124 from user/spec-update

Updated r2r-cli specification

---

### PR #125: Fix bug

**Description:**

(No description provided)

**Merge Commit Message:**

Merge pull request #125 from user/bugfix

Fixed bug in approval logic

---

```

**With `--include-all-reviews` flag:**

```markdown
# PR Approvals: ext-eac (0.0.7)

**Summary:** 3 PRs, 6 approvals

| PR   | Title           | Reviewer   | Review State      | Reviewed At |
| ---- | --------------- | ---------- | ----------------- | ----------- |
| #123 | Add new feature | @reviewer1 | APPROVED          | 2025-12-10  |
| #123 | Add new feature | @reviewer2 | CHANGES_REQUESTED | 2025-12-10  |
| #124 | Update spec     | @reviewer3 | APPROVED          | 2025-12-11  |
| #124 | Update spec     | @reviewer4 | COMMENTED         | 2025-12-11  |
| #125 | Fix bug         | @reviewer1 | APPROVED          | 2025-12-12  |
| #125 | Fix bug         | @reviewer5 | CHANGES_REQUESTED | 2025-12-12  |

## PR Details

### PR #123: Add new feature

**Description:**

This PR adds a new feature to the eac-commands module with enhanced
functionality for processing specifications.

**Merge Commit Message:**

Merge pull request #123 from user/feature-branch

Added new feature implementation

---

### PR #124: Update spec

**Description:**

Updated specification files for r2r-cli module

**Merge Commit Message:**

Merge pull request #124 from user/spec-update

Updated r2r-cli specification

---

### PR #125: Fix bug

**Description:**

(No description provided)

**Merge Commit Message:**

Merge pull request #125 from user/bugfix

Fixed bug in approval logic

---

```

## How It Works

1. **Extract PR Numbers**: Scans git commits for PR references (e.g., `#123` or `Merge pull request #123`)
2. **Fetch PR Data**: Uses GitHub CLI (`gh pr view`) to get PR details, files, and reviews
3. **Filter Spec PRs**: Only includes PRs that contain `.feature` files in `specs/<module>/`
4. **Extract Approvals**: By default, shows only reviews with state `APPROVED`. Use `--include-all-reviews` to show all review states (APPROVED, CHANGES_REQUESTED, COMMENTED)
5. **Multiple Approvals**: Same PR can have multiple approval rows (one per reviewer)

## Requirements

- **GitHub CLI (`gh`)**: Optional but recommended - command degrades gracefully if unavailable
- **Authentication**: `gh` must be authenticated (`gh auth login`)
- **Repository**: Must be a GitHub repository with pull requests

## Error Handling

| Error               | Exit Code | Solution                                                        |
| ------------------- | --------- | --------------------------------------------------------------- |
| `module not found`  | 1         | Verify module with `r2r eac show modules`                       |
| `version not found` | 1         | Check available versions with `r2r eac show changelog <module>` |
| `gh command failed` | Skips PR  | Install and authenticate GitHub CLI: `gh auth login`            |
| No approvals found  | 0         | Normal - shows friendly message                                 |

## Troubleshooting

### "gh command failed" Warnings

**Problem:** GitHub CLI not installed or not authenticated

**Solution:**

```bash
# Install GitHub CLI
# See: https://cli.github.com/

# Authenticate
gh auth login
```

### No Approvals Shown

**Problem:** Command shows "No PR approvals found"

**Possible Causes:**

- No PRs merged in the version range
- PRs don't contain `.feature` files
- No APPROVED reviews on spec PRs (default behavior only shows APPROVED)
- GitHub CLI unavailable

**Solution:** Verify PRs exist and contain spec files:

```bash
# Check recent commits
git log --oneline

# Check for PR references
git log --oneline | grep "#"

# Verify spec files in commits
git log --name-only | grep ".feature"

# Try including all review states
r2r eac show approval-comments ext-eac --include-all-reviews
```

## See Also

- [get approval-comments](../get/approval-comments.md) - JSON/YAML output
- [How-To Guide](../../../how-to-guides/eac/commands/release-management/view-approval-comments.md)
