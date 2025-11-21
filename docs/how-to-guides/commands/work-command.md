# Work Command

**Problem**: Work on multiple features in parallel with separate AI agent sessions while keeping clean git history.

**Solution**: Use `work` to create isolated workspaces (git worktrees) for each feature.

## Key Benefits

- Multiple features simultaneously with separate AI sessions
- Linear git history with no context switching
- Safe isolation of experiments

## Quick Start

```bash
# Create workspace
r2r eac work create feature/authentication
cd ../eac-feature-authentication

# Commit with AI messages
r2r eac work commit --all

# Sync with main
r2r eac work pull

# Complete work - Option A: PR
r2r eac work pr

# Complete work - Option B: Direct merge
r2r eac work merge
cd ../eac && git push origin main

# Clean up
r2r eac work remove
```

## Typical Workflow

### Parallel Development

```bash
# Create workspaces
r2r eac work create feature/authentication
r2r eac work create feature/api-refactor

# Start AI agents in separate terminals
cd ../eac-feature-authentication  # Terminal 1
cd ../eac-feature-api-refactor    # Terminal 2

# In each workspace
r2r eac work commit --all
r2r eac work pull
```

### List Workspaces

```bash
r2r eac work list

# Output:
# | Path                        | Branch                | Status |
# | C:\source\eac               | main                  | clean  |
# | C:\source\eac-feature-auth  | feature/authentication| dirty  |
```

## Command Reference

### work create

```bash
r2r eac work create <branch-name> [options]

# Options:
--from=<branch>    # Base branch (default: main)
--path=<path>      # Custom workspace path

# Examples:
r2r eac work create feature/authentication
r2r eac work create feature/api --from=develop
```

Default path: `../<repo-name>-<branch-name>`

### work commit

```bash
r2r eac work commit [options]

# Options:
--all, -a          # Stage all changes
--message, -m      # Custom message (skip AI)
--debug, -d        # Debug AI generation

# Examples:
r2r eac work commit --all
r2r eac work commit -m "fix: resolve auth bug"
```

AI analyzes changes and generates semantic messages following project conventions.

### work pull

```bash
r2r eac work pull [options]

# Options:
--target=<branch>  # Target branch (default: main)
--autostash        # Auto-stash uncommitted changes
--no-fetch         # Skip remote fetch

# Handle conflicts:
git add <files>
git rebase --continue  # or --abort
```

### work merge

```bash
r2r eac work merge [options]

# Options:
--target=<branch>      # Target branch (default: main)
--no-squash            # Regular merge (preserve commits)
--keep-worktree        # Don't remove workspace
--debug, -d            # Debug AI generation

# What happens:
# 1. Validates clean workspace
# 2. Switches to target and updates
# 3. Squash merges into one commit
# 4. AI generates commit message
# 5. Removes workspace (unless --keep-worktree)
```

**Why squash merge?** Clean history, one commit per feature, easier reverts.

After merge, push from main workspace:

```bash
cd ../eac
git push origin main
```

### work pr

```bash
r2r eac work pr [options]

# Options:
--target=<branch>  # Target branch (default: main)
--title <title>    # Custom PR title
--debug, -d        # Debug AI generation

# Requirements:
# - GitHub CLI (gh) installed
# - No uncommitted changes
# - At least one commit ahead
```

Generates PR with AI title, summary, commits, diff stats, and test plan.

### work remove

```bash
r2r eac work remove [branch] [options]

# Options:
--keep-branch      # Keep local branch
--delete-remote    # Delete remote branch too
--force, -f        # Force remove with uncommitted changes

# Default: removes workspace directory and local branch, preserves remote
```

## Integration Patterns

### PR Workflow (Team Collaboration)

Use for: team work, code review, CI/CD, open source

```bash
r2r eac work create feature/auth
cd ../eac-feature-auth
# Develop...
r2r eac work commit --all
r2r eac work pull
r2r eac work pr
# After PR merged:
r2r eac work remove
```

### Direct Merge Workflow (Solo/Trusted)

Use for: solo work, repository ownership, internal tooling

```bash
r2r eac work create feature/auth
cd ../eac-feature-auth
# Develop...
r2r eac work commit --all
r2r eac work pull
r2r eac work merge
cd ../eac && git push origin main
```

### Hybrid Workflow

Features via PR, hotfixes merged directly:

```bash
# Feature
r2r eac work create feature/new-api
r2r eac work pr

# Hotfix
r2r eac work create hotfix/critical-bug
r2r eac work merge && cd ../eac && git push origin main
```

## Best Practices

- **One feature per workspace**: Separate workspaces for separate concerns
- **Sync regularly**: `r2r eac work pull` before major sessions
- **Commit frequently**: Better AI messages for focused commits
- **Choose merge strategy**: Squash for features, regular for detailed history
- **Clean up**: Remove completed workspaces regularly

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Branch already exists | `git branch -D feature/name` |
| Uncommitted changes | `r2r eac work commit --all` or `r2r eac work pull --autostash` |
| Rebase conflicts | Resolve, then `git add <files> && git rebase --continue` |
| gh CLI not found | Install: `winget install GitHub.cli` or `brew install gh`, then `gh auth login` |
| Branch not up to date | `r2r eac work pull` |

## Advanced Usage

```bash
# Different base branches
r2r eac work create feature/api --from=develop
r2r eac work pull --target=develop
r2r eac work merge --target=develop

# Custom locations
r2r eac work create feature/auth --path=/custom/location

# Preserve workspace after merge
r2r eac work merge --keep-worktree

# Debug AI generation
r2r eac work commit --debug
```

## AI Agent Integration

Run multiple AI agents simultaneously in different workspaces without conflicts.

**AI-powered commits** analyze:
- Changed files and module ownership
- Project commit conventions
- Semantic commit structure

Example output:
```
feat(src-core): implement JWT authentication system

Add secure authentication with JWT tokens:
- Token generation and validation
- Password hashing with bcrypt
- Session management middleware

Files modified:
- src/core/auth.go (new)
- src/core/auth_test.go (new)
```

## Summary

1. **Create**: `r2r eac work create feature/name`
2. **Develop**: Work with AI coding agent
3. **Commit**: `r2r eac work commit --all`
4. **Sync**: `r2r eac work pull`
5. **Complete**: `r2r eac work pr` OR `r2r eac work merge` then push
6. **Clean up**: `r2r eac work remove`
