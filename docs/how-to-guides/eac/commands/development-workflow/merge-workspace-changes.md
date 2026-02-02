# Merge Workspace Changes

## What You'll Accomplish

Merge your topic branch back to main using squash merge, then clean up the workspace.

## Prerequisites

- topic branch with completed work
- PR approved (if using PR workflow)
- Tests passing

## Steps

### 1. Sync with Latest Main

```bash
eac work pull
```

**What happens**: Rebases your branch on latest main to avoid conflicts

### 2. Merge to Main

```bash
eac work merge
```

**What happens**:

- Switches to main branch
- Squashes all commits into one
- Generates squash commit message with AI
- Merges to main
- Deletes topic branch

### 3. Verify Merge

```bash
git log -1
```

**What happens**: Shows the squashed commit on main

### 4. Clean Up Workspace

```bash
eac work remove feature/authentication
```

**What happens**: Removes worktree and associated files

## Merge Options

```bash
# Merge without squash (preserve commits)
eac work merge --no-squash

# Merge without deleting worktree
eac work merge --keep-worktree

# Merge to different target branch
eac work merge --target develop
```

**Note**: Squash commit messages are always AI-generated based on all commits in the branch. Manual message customization is not supported.

## Example Scenario

You've completed JWT authentication feature:

```bash
# Sync with main first
eac work pull
# Rebasing feature/add-jwt-auth onto main...
# ✓ Up to date

# Merge to main
eac work merge
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
eac work remove add-jwt-auth
# ✓ Removed worktree
```

## Common Issues

| Problem                 | Solution                      |
| ----------------------- | ----------------------------- |
| Merge conflicts         | Resolve conflicts, then retry |
| "Branch not up to date" | Run `work pull` first         |
| Tests failing           | Fix tests before merging      |

## Next Steps

- [Create Feature Workspace](./create-feature-workspace.md) → Start next feature

## Related Commands

- [`work merge`](../../../../reference/eac/commands/work/merge.md) - Merge workspace
- [`work pull`](../../../../reference/eac/commands/work/pull.md) - Sync with main
- [`work remove`](../../../../reference/eac/commands/work/remove.md) - Clean up workspace
