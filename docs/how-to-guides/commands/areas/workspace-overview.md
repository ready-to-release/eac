# Workspace Management

Workspace management in EAC enables parallel development using git worktrees, allowing multiple AI agent sessions to work simultaneously without conflicts.

## What is Workspace Management?

EAC's workspace system enables you to:

- **Create isolated workspaces** for each feature or task
- **Run multiple AI sessions** simultaneously without conflicts
- **Maintain clean git history** through squash merges
- **Streamline the PR workflow** with AI-generated descriptions

The system uses git worktrees to create separate working directories, each with its own branch and AI coding session.

## When to Use Workspaces

Use workspace commands when you need:

| Scenario                            | Commands      |
| ----------------------------------- | ------------- |
| Starting a new feature              | `work create` |
| Committing changes with AI messages | `work commit` |
| Syncing with main branch            | `work pull`   |
| Completing work via PR              | `work pr`     |
| Completing work via direct merge    | `work merge`  |
| Cleaning up finished work           | `work remove` |

### Common Use Cases

- **Parallel feature development** - Work on auth while API refactor continues
- **AI pair programming** - Run Claude Code in multiple workspaces
- **Experiment isolation** - Try approaches without affecting main work
- **Context switching** - Keep multiple tasks in progress cleanly

## Key Concepts

### Git Worktrees

Git worktrees allow multiple working directories from one repository:

```
C:\projects\eac\                    # Main worktree (main branch)
C:\projects\eac-feature-auth\       # Feature worktree (feature/auth branch)
C:\projects\eac-bugfix-123\         # Bugfix worktree (bugfix/123 branch)
```

Each worktree:

- Has its own working directory
- Tracks its own branch
- Shares the same git history
- Can run independent operations

### Branch-per-Workspace

Each workspace creates a dedicated branch:

```bash
r2r eac work create feature/authentication
# Creates branch: feature/authentication
# Creates directory: ../eac-feature-authentication/
```

### Merge Strategies

| Strategy                   | Use Case         | Result                |
| -------------------------- | ---------------- | --------------------- |
| **Squash merge** (default) | Features         | One clean commit      |
| **Regular merge**          | Detailed history | Preserves all commits |
| **PR workflow**            | Team review      | GitHub PR + review    |

### AI-Powered Commits

The `work commit` command:

1. Stages specified changes
2. Analyzes diff with AI
3. Generates semantic commit message
4. Creates the commit

## Workflow Overview

### Solo Development (Direct Merge)

```bash
# 1. Create workspace
r2r eac work create feature/authentication
cd ../eac-feature-authentication

# 2. Develop with AI assistant
# ... make changes ...

# 3. Commit with AI messages
r2r eac work commit --all

# 4. Sync with main
r2r eac work pull

# 5. Complete work
r2r eac work merge
cd ../eac && git push origin main

# 6. Clean up
r2r eac work remove
```

### Team Development (PR Workflow)

```bash
# 1. Create workspace
r2r eac work create feature/api-refactor
cd ../eac-feature-api-refactor

# 2. Develop
# ... make changes ...

# 3. Commit frequently
r2r eac work commit --all

# 4. Stay in sync
r2r eac work pull

# 5. Create PR
r2r eac work pr

# 6. After PR merges, clean up
r2r eac work remove
```

### Parallel Development

```bash
# Terminal 1: Feature A
r2r eac work create feature/auth
cd ../eac-feature-auth
# Start AI session here

# Terminal 2: Feature B
r2r eac work create feature/api
cd ../eac-feature-api
# Start AI session here

# Both can run simultaneously without conflicts
```

## Workspace Lifecycle

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   create    │────▶│   develop   │────▶│  complete   │
└─────────────┘     └─────────────┘     └─────────────┘
       │                   │                   │
       ▼                   ▼                   ▼
  New worktree        commit/pull          merge/pr
  New branch         Stay in sync         Clean merge
```

### States

| State      | Description            | Next Actions            |
| ---------- | ---------------------- | ----------------------- |
| **Clean**  | No uncommitted changes | commit, pull, merge, pr |
| **Dirty**  | Uncommitted changes    | commit first            |
| **Behind** | Main has new commits   | pull to sync            |
| **Ahead**  | Has commits to merge   | merge or pr             |

## Integration Points

### With AI Coding Assistants

Workspaces enable isolated AI sessions:

```bash
# Workspace 1: Claude Code session for auth
cd ../eac-feature-auth
claude  # Start Claude Code

# Workspace 2: Different Claude session for API
cd ../eac-feature-api
claude  # Independent session
```

### With Commit Command

`work commit` uses the same AI engine as `create commit-message`:

```bash
# These are equivalent:
r2r eac work commit --all
# vs
git add . && r2r eac create commit-message --commit
```

### With CI/CD

Push workspace branches for CI:

```bash
# Push for CI checks before merge
git push origin feature/auth

# After CI passes
r2r eac work merge  # or work pr
```

## Best Practices

### Do's

- **One feature per workspace** - Keep concerns separated
- **Commit frequently** - Better AI messages for focused changes
- **Sync regularly** - `work pull` before major sessions
- **Clean up promptly** - Remove completed workspaces

### Don'ts

- **Don't mix features** - Use separate workspaces
- **Don't skip sync** - Merge conflicts are harder to resolve later
- **Don't forget cleanup** - Stale workspaces consume disk space

## Next Steps

- [Workspace Configuration](workspace-configuration.md) - Configure paths and branch naming
- [Workspace Commands](workspace-commands.md) - Full command reference

## Related Areas

- [Commit Command](../commit-command.md) - AI-powered commit messages
- [Release](release-overview.md) - Versioning after merging features
- [Pipeline](pipeline-overview.md) - CI/CD for workspace branches
