# Work Commands

{{ page_breadcrumb() }}

## Overview

Work commands manage parallel development workspaces using **git worktrees**. They enable multiple topic branches to be worked on simultaneously without switching contexts, providing isolated development environments for each feature.

**Key Characteristics**:

- Git worktree-based isolation
- Parallel feature development
- AI-powered commit messages
- Workspace lifecycle management
- Safe merge and sync operations

**When to use**: When you need to work on multiple features simultaneously or want isolated development environments without branch switching.

## All Work Commands

| Command                                  | Purpose                 | Use Case                                  |
| ---------------------------------------- | ----------------------- | ----------------------------------------- |
| [work create](../work/create.md)         | Create new workspace    | Start new feature in isolated environment |
| [work commit](../work/commit.md)         | Commit with AI messages | Commit changes with generated messages    |
| [work pull](../work/pull.md)             | Sync with main          | Update workspace with latest changes      |
| [work merge](../work/merge.md)           | Merge to main           | Complete feature and integrate            |
| [work remove](../work/remove.md)         | Remove workspace        | Clean up after merge                      |
| [show workspaces](../show/workspaces.md) | List all workspaces     | See all active workspaces                 |

## Git Worktrees Explained

### What are Git Worktrees?

Git worktrees allow you to have multiple working directories attached to the same repository. Each worktree is checked out to a different branch, enabling parallel development without switching branches.

### Traditional Workflow (Branch Switching)

```bash
# Work on feature A
git checkout feature-a
# ... make changes ...
git add . && git commit -m "WIP"

# Switch to feature B (context switch)
git checkout feature-b
# ... make changes ...

# Switch back to feature A (context switch again)
git checkout feature-a
```

**Problems**:

- Constant context switching
- Need to commit/stash before switching
- Can't compare features side-by-side
- Builds get invalidated on each switch

### Worktree Workflow

```bash
# Create workspace for feature A
r2r eac work create feature/auth
# Work in: ~/project/work/feature-auth/

# Create workspace for feature B (in different directory)
r2r eac work create feature/api
# Work in: ~/project/work/feature-api/

# Both features available simultaneously!
```

**Benefits**:

- No context switching
- Independent build caches
- Compare features side-by-side
- Run multiple servers simultaneously

## Common Workflows

### Starting a New Feature

```bash
# Create workspace
r2r eac work create feature/user-authentication

# Change to workspace directory
cd ../work/feature-user-authentication

# Develop feature
# ... write code ...

# Commit with AI message
r2r eac work commit --all
```

### Working on Multiple Features

```bash
# Feature 1: Authentication
r2r eac work create feature/auth
cd ../work/feature-auth
# ... develop ...

# Feature 2: API (separate workspace)
cd ~/project
r2r eac work create feature/api
cd ../work/feature-api
# ... develop ...

# Both workspaces remain active
r2r eac show workspaces
```

### Syncing with Main

```bash
# Update workspace with latest main changes
r2r eac work pull

# Rebases current workspace on latest main
# Resolves conflicts if any
```

### Completing a Feature

```bash
# Commit final changes
r2r eac work commit --all

# Merge to main (squash by default)
r2r eac work merge

# Remove workspace
r2r eac work remove feature/auth
```

## Workspace Lifecycle

```text
1. CREATE    →  2. DEVELOP  →  3. SYNC     →  4. MERGE    →  5. REMOVE
work create     work commit     work pull      work merge      work remove
   ↓               ↓               ↓              ↓                ↓
New worktree   AI messages    Rebase on      Squash to       Clean up
created        generated      main           main            worktree
```

### Detailed Lifecycle

**1. Create** (`work create feature/my-feature`)

- Creates git worktree
- Creates branch from main
- Sets up isolated directory

**2. Develop** (`work commit --all`)

- Make code changes
- Generate AI commit messages
- Commit to topic branch

**3. Sync** (`work pull`)

- Fetch latest main
- Rebase topic branch
- Resolve conflicts if needed

**4. Merge** (`work merge`)

- Squash commits
- Generate squash message
- Merge to main
- Push to remote

**5. Remove** (`work remove`)

- Remove worktree
- Optionally delete branch
- Clean up workspace

## Command Details

### work create

Create isolated workspace for new feature:

```bash
# From main
r2r eac work create feature/authentication

# From specific branch
r2r eac work create hotfix/security --from release/v1.2.0
```

**What it does**:

- Creates git worktree in `../work/<branch-name>/`
- Creates new branch (or checks out existing)
- Sets up isolated development environment

### work commit

Commit changes with AI-generated messages:

```bash
# Commit staged changes
r2r eac work commit

# Stage all and commit
r2r eac work commit --all

# Use manual message
r2r eac work commit --all --message "fix: resolve auth bug"
```

**What it does**:

- Generates semantic commit message using AI
- Commits changes to topic branch
- Follows conventional commit format

### work pull

Sync workspace with latest main:

```bash
r2r eac work pull
```

**What it does**:

- Fetches latest main branch
- Rebases topic branch on main
- Handles conflicts interactively

### work merge

Merge feature to main:

```bash
# Squash merge (default)
r2r eac work merge

# No squash (preserve commits)
r2r eac work merge --no-squash
```

**What it does**:

- Generates squash commit message
- Merges feature to main
- Pushes to remote
- Optionally removes workspace

### work remove

Remove workspace:

```bash
# Remove workspace only
r2r eac work remove feature/auth

# Remove workspace and branch
r2r eac work remove feature/auth --delete-branch
```

**What it does**:

- Removes git worktree
- Optionally deletes topic branch
- Cleans up workspace directory

## Best Practices

### Workspace Naming

Use descriptive, hierarchical names:

```bash
# Good
r2r eac work create feature/user-authentication
r2r eac work create fix/memory-leak
r2r eac work create refactor/database-layer

# Avoid
r2r eac work create auth
r2r eac work create fix
r2r eac work create temp
```

### Workspace Organization

```text
~/project/
├── cli/              # Main worktree (main branch)
└── work/             # Feature worktrees
    ├── feature-auth/
    ├── feature-api/
    └── fix-bug-123/
```

### Commit Frequently

```bash
# Commit small, logical changes
r2r eac work commit --all

# Don't wait until feature is complete
# Small commits = better AI messages
```

### Sync Regularly

```bash
# Sync with main daily
r2r eac work pull

# Prevents large merge conflicts
# Keeps topic branch up-to-date
```

### Clean Up

```bash
# Remove workspace after merge
r2r eac work merge
r2r eac work remove feature/auth --delete-branch

# Don't accumulate stale workspaces
```

## Common Issues

### Workspace Already Exists

**Problem**: `work create` fails because workspace exists

**Solution**: Check existing workspaces and remove if needed

```bash
# List workspaces
r2r eac show workspaces

# Remove existing workspace
r2r eac work remove feature/auth
```

### Merge Conflicts

**Problem**: `work pull` reports conflicts

**Solution**: Resolve conflicts manually

```bash
# Pull attempts automatic rebase
r2r eac work pull

# If conflicts occur:
# 1. Resolve conflicts in files
# 2. Stage resolved files: git add <files>
# 3. Continue rebase: git rebase --continue
```

### Uncommitted Changes

**Problem**: Can't switch operations with uncommitted changes

**Solution**: Commit or stash changes

```bash
# Commit changes
r2r eac work commit --all

# Or stash temporarily
git stash
# ... do other work ...
git stash pop
```

### Workspace Directory Not Found

**Problem**: Workspace directory missing

**Solution**: Recreate or remove worktree reference

```bash
# Remove broken worktree
git worktree remove feature/auth --force

# Recreate
r2r eac work create feature/auth
```

## Integration with CI/CD

### Pre-merge Validation

```bash
# Before merging, ensure CI passes
r2r eac pipeline status

# Check CI for current commit
r2r eac release check-ci $(git rev-parse HEAD)

# Then merge
r2r eac work merge
```

### Automated Workflow

```bash
#!/bin/bash
# Feature workflow script

FEATURE_NAME=$1

# Create workspace
r2r eac work create "feature/$FEATURE_NAME"

# Change directory
cd "../work/feature-$FEATURE_NAME"

# ... develop feature ...

# Commit
r2r eac work commit --all

# Sync with main
r2r eac work pull

# Run tests
r2r eac test

# Merge if tests pass
if [ $? -eq 0 ]; then
  r2r eac work merge
  r2r eac work remove "feature/$FEATURE_NAME" --delete-branch
fi
```

## See Also

- [show workspaces](../show/workspaces.md) - List all workspaces
- [create commit-message](../create/commit-message.md) - AI commit messages
- [create squash-message](../create/squash-message.md) - Squash merge messages
- [Workspace Commands Guide](../work/index.md)

{{ diataxis_footer() }}
