# Get approval-comments

<!-- book:cmd get approval-comments -->

## Special Keywords

| Keyword      | Description                    | Output                       |
| ------------ | ------------------------------ | ---------------------------- |
| `latest`     | Most recent released version   | Approvals from that release  |
| `unreleased` | Pending unreleased changes     | Approvals since last release |
| _(omit)_     | Same as `unreleased` (default) | Approvals since last release |

## Bundle Modules

For **container/bundle modules** with dependencies, approvals are **aggregated from all dependent modules**.

**Example:** `ext-eac` depends on `eac-commands` and `r2r-cli`:

```bash
r2r eac get approval-comments ext-eac --as-json
```

Returns approvals for PRs containing spec files from:

- `specs/eac-commands/` (dependency)
- `specs/r2r-cli/` (dependency)
- `specs/ext-eac/` (if any)

This provides a **complete view** of all PR approvals included in the release bundle.

**Regular modules** (without dependencies) only return approvals from their own `specs/<module>/` directory.

## Data Structure

**Default output (only APPROVED reviews):**

```yaml
module: "ext-eac"
version: "0.0.7"
total_prs: 3
total_approvals: 5
prs:  # All PRs found (with spec files), even if they have no reviews
  - number: 123
    title: "Add new feature"
    author: "developer1"
    body: "This PR adds a new feature to the eac-commands module"
    merged_at: "2025-12-10T11:00:00Z"
    merge_commit_message: |
      Merge pull request #123 from user/feature-branch

      Added new feature implementation
    files:
      - "specs/eac-commands/new-feature.feature"
      - "go/eac/commands/impl/test.go"
    reviews:
      - author: "reviewer1"
        state: "APPROVED"
        submitted_at: "2025-12-10T10:30:00Z"
      - author: "reviewer2"
        state: "APPROVED"
        submitted_at: "2025-12-10T10:45:00Z"
approvals:
  - pr_number: 123
    pr_title: "Add new feature"
    pr_author: "developer1"
    pr_body: "This PR adds a new feature to the eac-commands module"
    merge_message: |
      Merge pull request #123 from user/feature-branch

      Added new feature implementation
    reviewer: "reviewer1"
    review_state: "APPROVED"
    reviewed_at: "2025-12-10T10:30:00Z"
    spec_files:
      - "specs/eac-commands/new-feature.feature"
    merged_at: "2025-12-10T11:00:00Z"
  - pr_number: 123
    pr_title: "Add new feature"
    pr_author: "developer1"
    pr_body: "This PR adds a new feature to the eac-commands module"
    merge_message: |
      Merge pull request #123 from user/feature-branch

      Added new feature implementation
    reviewer: "reviewer2"
    review_state: "APPROVED"
    reviewed_at: "2025-12-10T10:45:00Z"
    spec_files:
      - "specs/eac-commands/new-feature.feature"
    merged_at: "2025-12-10T11:00:00Z"
```

**With `--include-all-reviews` flag (all review states):**

```yaml
module: "ext-eac"
version: "0.0.7"
total_prs: 3
total_approvals: 7  # Includes all review types
approvals:
  - pr_number: 123
    pr_title: "Add new feature"
    pr_author: "developer1"
    pr_body: "This PR adds a new feature to the eac-commands module"
    merge_message: |
      Merge pull request #123 from user/feature-branch

      Added new feature implementation
    reviewer: "reviewer1"
    review_state: "APPROVED"
    reviewed_at: "2025-12-10T10:30:00Z"
    spec_files:
      - "specs/eac-commands/new-feature.feature"
    merged_at: "2025-12-10T11:00:00Z"
  - pr_number: 123
    pr_title: "Add new feature"
    pr_author: "developer1"
    pr_body: "This PR adds a new feature to the eac-commands module"
    merge_message: |
      Merge pull request #123 from user/feature-branch

      Added new feature implementation
    reviewer: "reviewer2"
    review_state: "CHANGES_REQUESTED"
    reviewed_at: "2025-12-10T09:15:00Z"
    spec_files:
      - "specs/eac-commands/new-feature.feature"
    merged_at: "2025-12-10T11:00:00Z"
  - pr_number: 124
    pr_title: "Update spec"
    pr_author: "developer2"
    pr_body: "Updated specification files for r2r-cli module"
    merge_message: |
      Merge pull request #124 from user/spec-update

      Updated r2r-cli specification
    reviewer: "reviewer3"
    review_state: "COMMENTED"
    reviewed_at: "2025-12-11T14:20:00Z"
    spec_files:
      - "specs/r2r-cli/update.feature"
    merged_at: "2025-12-11T15:00:00Z"
```

## How It Works

1. **Extract PR Numbers**: Scans git commits for PR references (e.g., `#123` or `Merge pull request #123`)
2. **Fetch PR Data**: Uses GitHub CLI (`gh pr view`) to get PR details, files, and reviews
3. **Filter Spec PRs**: Only includes PRs that contain `.feature` files in `specs/<module>/`
4. **Extract Approvals**: By default, returns only reviews with state `APPROVED`. Use `--include-all-reviews` to include all review states (APPROVED, CHANGES_REQUESTED, COMMENTED)

## Requirements

- **GitHub CLI (`gh`)**: Optional but recommended - command degrades gracefully if unavailable
- **Authentication**: `gh` must be authenticated (`gh auth login`)
- **Repository**: Must be a GitHub repository with pull requests

## Error Handling

| Error                        | Exit Code | Solution                                                                         |
| ---------------------------- | --------- | -------------------------------------------------------------------------------- |
| `module not found`           | 1         | Verify module with `show modules`                                                |
| `version not found`          | 1         | List versions with `get changelog <module> --as-json \| jq '.versions[].number'` |
| `no released versions found` | 1         | Normal if no releases yet                                                        |
| `gh command failed`          | Skips PR  | Install and authenticate GitHub CLI                                              |

## See Also

- [show approval-comments](../show/approval-comments.md) - Human-readable markdown output
- [How-To Guide](../../../how-to-guides/eac/commands/release-management/view-approval-comments.md)
