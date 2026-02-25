# Create Pull Request

## What You'll Accomplish

Generate a pull request with AI-written description that summarizes all branch changes.

## Prerequisites

- topic branch with commits
- Branch pushed to remote
- GitHub CLI (`gh`) installed (or use GitHub web UI)
- AI provider configured

## How It Works

The command uses AI to generate comprehensive PR descriptions:

- **Commit Analysis**: AI analyzes all commits in the branch to understand the full scope of changes
- **Summary Generation**: Generates PR title and structured description with summary and changes
- **Markdown Format**: AI generates properly formatted markdown for GitHub PR descriptions
- **Validation**: Output is validated and retried automatically if needed
- **Integration**: Creates the PR directly using GitHub CLI with the generated content

This ensures PR descriptions are comprehensive, well-formatted, and accurately describe all branch changes.

## Steps

### 1. Push Your Branch

```bash
git push origin feature/authentication
```

**What happens**: Branch is available on remote for PR creation

### 2. Generate PR with AI Description

```bash
eac create pr
```

**What happens**:

- AI analyzes all commits in branch
- Generates comprehensive PR description with validation
- Creates PR using GitHub CLI with the generated title and description

### 3. Review PR

Visit the PR URL provided and review:

- Title
- Description
- Changed files
- Checks status

## Alternative: Manual PR Creation with GitHub CLI

If you prefer to create the PR manually using GitHub CLI:

```bash
# Use gh pr create with your own title and description
gh pr create --title "Add JWT authentication" --body "Your PR description"

# Or create interactively
gh pr create
```

**Note**: The `create pr` command currently creates the PR directly. For manual control, use GitHub CLI (`gh`) or the GitHub web interface.

## Example Scenario

You've completed a feature and want to create a PR:

```bash
# Ensure branch is pushed
git push origin feature/add-jwt-auth

# Create PR with AI description
eac create pr

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

| Problem                 | Solution                            |
| ----------------------- | ----------------------------------- |
| "No commits to analyze" | Ensure branch has commits vs base   |
| "gh not found"          | Install GitHub CLI or use --dry-run |
| Branch not pushed       | Run `git push` first                |

## Next Steps

- [Merge Workspace Changes](./merge-workspace-changes.md) → Merge after approval

## Related Commands

- [`create pr`](../../../../reference/eac/commands/create/pr.md) - Full command reference
- [`get squash-message`](../../../../reference/eac/commands/get/squash-message.md) - Generate squash message
