# Create Feature Workspace

## What You'll Accomplish

Set up an isolated workspace for developing a new feature without affecting your main branch. Uses git worktrees to enable parallel development on multiple features.

## Prerequisites

### Required Knowledge

**New to git worktrees?** Learn these concepts first:

- [Working with Git Worktrees](../../../../tutorials/core-workflows/working-with-worktrees.md) - Understand parallel development with workspaces

### Required Setup

- Working in a git repository
- Changes in current workspace committed
- Know the feature name or branch you want to create

## Steps

### 1. Create the Workspace

```bash
r2r eac work create feature/authentication
```

**What happens**: Creates new branch `feature/authentication` and worktree in `../simply-cli-feature-authentication`

### 2. Navigate to New Workspace

```bash
cd ../simply-cli-feature-authentication
```

**What happens**: You're now in isolated workspace with dedicated file state

### 3. Verify Workspace Created

```bash
r2r eac show workspaces
```

**What happens**: Shows list of all workspaces with their branches and paths

### 4. Start Working

```bash
git status
# On branch feature/authentication
# nothing to commit, working tree clean
```

## Example Scenario

You need to add JWT authentication while continuing work on another feature:

```bash
# Create auth workspace
r2r eac work create feature/add-jwt-auth

# Move to new workspace
cd ../simply-cli-feature-add-jwt-auth

# Verify isolation
git branch
# * feature/add-jwt-auth

# Start coding
code .
```

## Common Issues

| Problem | Solution |
|---------|----------|
| "Branch already exists" | Use different branch name or delete old branch |
| "Worktree path exists" | Choose custom path with `--path` flag |
| Uncommitted changes | Commit or stash changes before creating workspace |

## Next Steps

- [Make Commits with AI](./make-commits-with-ai.md) → Commit your changes
- [Merge Workspace Changes](./merge-workspace-changes.md) → Merge back to main

## Related Commands

- [`work create`](../../../../reference/commands/work/create.md) - Full command reference
- [`show workspaces`](../../../../reference/commands/show/workspaces.md) - List all workspaces
- [`work remove`](../../../../reference/commands/work/remove.md) - Delete workspace
