# Create Pull Request

{{ page_breadcrumb() }}

## What You'll Accomplish

Generate a pull request with AI-written description that summarizes all branch changes.

## Prerequisites

- Feature branch with commits
- Branch pushed to remote
- GitHub CLI (`gh`) installed (or use GitHub web UI)
- AI provider configured

## Steps

### 1. Push Your Branch

```bash
git push origin feature/authentication
```

**What happens**: Branch is available on remote for PR creation

### 2. Generate PR with AI Description

```bash
r2r eac create pr
```

**What happens**:

- AI analyzes all commits in branch
- Generates comprehensive PR description
- Creates PR using GitHub CLI

### 3. Review PR

Visit the PR URL provided and review:

- Title
- Description
- Changed files
- Checks status

## Manual PR Creation

```bash
# Generate description only
r2r eac create pr --dry-run > pr-description.md

# Use with gh pr create
gh pr create --title "Add JWT authentication" --body-file pr-description.md
```

## Example Scenario

You've completed a feature and want to create a PR:

```bash
# Ensure branch is pushed
git push origin feature/add-jwt-auth

# Create PR with AI description
r2r eac create pr

# Output:
# Analyzing 5 commits...
# Generating PR description...
# Creating pull request...
# ✓ PR created: https://github.com/org/repo/pull/123
#
# Title: feat: Add JWT authentication support
#
# Description:
# ## Summary
# - Implement JWT token generation and validation
# - Add refresh token support
# - Include middleware for protected routes
#
# ## Changes
# - Added auth package with JWT utilities
# - Implemented login/logout endpoints
# - Added authentication tests
```

## Common Issues

| Problem | Solution |
|---------|----------|
| "No commits to analyze" | Ensure branch has commits vs base |
| "gh not found" | Install GitHub CLI or use --dry-run |
| Branch not pushed | Run `git push` first |

## Next Steps

- [Merge Workspace Changes](./merge-workspace-changes.md) → Merge after approval

## Related Commands

- [`create pr`](../../../reference/commands/create/pr.md) - Full command reference
- [`create squash-message`](../../../reference/commands/create/squash-message.md) - Generate squash message

{{ diataxis_footer() }}
