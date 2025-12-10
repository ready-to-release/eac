# Merge Workspace Changes

{{ page_breadcrumb() }}

## What You'll Accomplish

Merge your feature branch back to main using squash merge, then clean up the workspace.

## Prerequisites

- Feature branch with completed work
- PR approved (if using PR workflow)
- Tests passing

## Steps

### 1. Sync with Latest Main

```bash
r2r eac work pull
```

**What happens**: Rebases your branch on latest main to avoid conflicts

### 2. Merge to Main

```bash
r2r eac work merge
```

**What happens**:

- Switches to main branch
- Squashes all commits into one
- Generates squash commit message with AI
- Merges to main
- Deletes feature branch

### 3. Verify Merge

```bash
git log -1
```

**What happens**: Shows the squashed commit on main

### 4. Clean Up Workspace

```bash
r2r eac work remove feature/authentication
```

**What happens**: Removes worktree and associated files

## Merge Options

```bash
# Merge without squash (preserve commits)
r2r eac work merge --no-squash

# Merge without deleting worktree
r2r eac work merge --keep-worktree

# Merge to different target branch
r2r eac work merge --target develop
```

**Note**: Squash commit messages are always AI-generated based on all commits in the branch. Manual message customization is not supported.

## Example Scenario

You've completed JWT authentication feature:

```bash
# Sync with main first
r2r eac work pull
# Rebasing feature/add-jwt-auth onto main...
# ✓ Up to date

# Merge to main
r2r eac work merge
# Switched to 'main'
# Generating squash message...
# Squashing 5 commits...
# ✓ Merged to main
# ✓ Deleted branch feature/add-jwt-auth

# Verify
git log -1
# commit abc123
# feat(auth): add JWT authentication support

# Clean up workspace
r2r eac work remove add-jwt-auth
# ✓ Removed worktree
```

## Common Issues

| Problem | Solution |
|---------|----------|
| Merge conflicts | Resolve conflicts, then retry |
| "Branch not up to date" | Run `work pull` first |
| Tests failing | Fix tests before merging |

## Next Steps

- [Create Feature Workspace](./create-feature-workspace.md) → Start next feature

## Related Commands

- [`work merge`](../../../../reference/commands/work/merge.md) - Merge workspace
- [`work pull`](../../../../reference/commands/work/pull.md) - Sync with main
- [`work remove`](../../../../reference/commands/work/remove.md) - Clean up workspace

{{ diataxis_footer() }}
