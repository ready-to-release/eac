# Work Command Workflow

**Problem**: You need to work on multiple features in parallel with separate AI coding agent sessions while keeping your git history clean.

**Solution**: Use the `work` command to create isolated workspaces (git worktrees) for each feature, allowing parallel development with independent AI agent sessions.

---

## Overview

The `work` command provides workspace management for parallel development using git worktrees. Each workspace is an isolated working directory on a separate branch, perfect for running multiple AI coding agent sessions simultaneously.

**Key benefits:**

- Work on multiple features simultaneously
- Run separate AI agent sessions per feature
- Keep git history linear and clean
- Avoid context switching with `git checkout`
- Safely isolate experiments and features

---

## Quick Start

### Create a workspace

```bash
r2r eac work create feature/authentication
```

Creates workspace in: `../eac-feature-authentication`

### Start your AI agent in the workspace

```bash
cd ../eac-feature-authentication
# Start your AI coding agent (Claude Code, Cursor, Aider, etc.)
```

### Commit changes with AI-generated messages

```bash
r2r eac work commit --all
```

### Sync with main

```bash
r2r eac work pull
```

Rebases your commits onto the latest main branch.

### Complete the work

**Option A: Create pull request** (for team collaboration)

```bash
r2r eac work pr
```

**Option B: Merge locally and push** (direct integration)

```bash
# Merge to local main
r2r eac work merge

# Push from main workspace
cd ../eac
git push origin main
```

### Clean up

```bash
r2r eac work remove
```

---

## Typical Workflow

### Scenario: Two parallel features

**Repository:** `everything-as-code/eac`
**Main workspace:** `C:\source\eac` (main branch)

### 1. Create workspaces

```bash
# Authentication feature
r2r eac work create feature/authentication

# API refactoring
r2r eac work create feature/api-refactor
```

### 2. Start AI agents in each workspace

**Terminal 1:**
```bash
cd ../eac-feature-authentication
# Start AI agent session 1
```

**Terminal 2:**
```bash
cd ../eac-feature-api-refactor
# Start AI agent session 2
```

Each agent works in isolation without conflicts.

### 3. Commit changes

In each workspace:

```bash
r2r eac work commit --all
```

AI analyzes changes and generates semantic commit messages following your project conventions.

### 4. Sync with main

```bash
r2r eac work pull
```

Output shows rebase progress:
```
Rebasing feature/authentication onto origin/main
  Current branch: 3 commits ahead of main
  main branch: 2 new commits

✓ Rebased feature/authentication onto origin/main
```

### 5. List all workspaces

```bash
r2r eac work list

# Output:
# | Path                              | Branch                | Status |
# | --------------------------------- | --------------------- | ------ |
# | C:\source\eac                     | main                  | clean  |
# | C:\source\eac-feature-auth        | feature/authentication| dirty  |
# | C:\source\eac-feature-api-refactor| feature/api-refactor  | clean  |
```

### 6. Complete feature (Option A: Pull Request)

```bash
cd ../eac-feature-authentication
r2r eac work pull  # Ensure up to date
r2r eac work pr

# Output:
# ✓ Pull request created: https://github.com/everything-as-code/eac/pull/123
```

### 7. Complete feature (Option B: Direct merge)

```bash
cd ../eac-feature-authentication
r2r eac work pull  # Ensure up to date
r2r eac work merge

# Automatically:
# - Switches to main
# - Squash merges commits
# - Generates AI commit message
# - Removes workspace
#
# ✓ Merged feature/authentication into main (squash)

# Push from main workspace
cd ../eac
git push origin main
```

### 8. Clean up after PR merge

Once PR is merged externally:

```bash
cd ../eac-feature-api-refactor
r2r eac work remove
```

---

## Command Reference

### work create

Create new workspace for parallel development.

```bash
r2r eac work create <branch-name> [options]
```

**Options:**
- `--from=<branch>` - Base branch (default: `main`)
- `--path=<path>` - Custom workspace path

**Examples:**

```bash
# Create from main
r2r eac work create feature/authentication

# Create from develop
r2r eac work create feature/api --from=develop

# Custom path
r2r eac work create bugfix/123 --path=../custom-path
```

Default path: `../<repo-name>-<branch-name>`

### work list

List all workspaces and their status.

```bash
r2r eac work list [--verbose|-v]
```

**Status:**
- `clean` - No uncommitted changes
- `dirty` - Has uncommitted changes

### work commit

Commit changes with AI-generated messages.

```bash
r2r eac work commit [options]
```

**Options:**
- `--all`, `-a` - Stage all changes before committing
- `--message <msg>`, `-m <msg>` - Custom message (skip AI)
- `--debug`, `-d` - Debug AI generation

**Examples:**

```bash
# AI-generated message
r2r eac work commit --all

# Custom message
r2r eac work commit -m "fix: resolve auth bug"
```

### work pull

Sync workspace with main via rebase.

```bash
r2r eac work pull [options]
```

**Options:**
- `--target=<branch>` - Target branch (default: `main`)
- `--autostash` - Auto-stash uncommitted changes
- `--no-fetch` - Skip remote fetch

**Examples:**

```bash
# Rebase onto latest main
r2r eac work pull

# Auto-stash uncommitted work
r2r eac work pull --autostash
```

**Conflict handling:**

If conflicts occur:

```bash
# Resolve files, then:
git add <files>
git rebase --continue

# Or abort:
git rebase --abort
```

### work merge

Merge workspace to main (squash by default).

```bash
r2r eac work merge [options]
```

**Options:**
- `--target=<branch>` - Target branch (default: `main`)
- `--no-squash` - Regular merge (preserve commits)
- `--keep-worktree` - Don't remove workspace after merge
- `--debug`, `-d` - Debug AI message generation

**Examples:**

```bash
# Squash merge to main (recommended)
r2r eac work merge

# Regular merge
r2r eac work merge --no-squash

# Keep workspace after merge
r2r eac work merge --keep-worktree
```

**What happens:**
1. Validates workspace is clean and up to date
2. Switches to target branch and updates it
3. Squash merges all commits into one
4. Uses AI to generate comprehensive commit message
5. Removes workspace (unless `--keep-worktree`)

**After merge:**

```bash
# Push from main workspace
cd ../eac
git push origin main
```

**Why squash merge?**
- Clean main branch history
- One logical commit per feature
- Easier to revert features
- Reduces git log noise

### work pr

Create pull request with AI-generated description.

```bash
r2r eac work pr [options]
```

**Options:**
- `--target=<branch>` - Target branch (default: `main`)
- `--title <title>` - Custom PR title
- `--debug`, `-d` - Debug AI generation

**Requirements:**
- GitHub CLI (`gh`) installed and authenticated
- No uncommitted changes
- At least one commit ahead of target

**Examples:**

```bash
# Create PR to main
r2r eac work pr

# Custom title
r2r eac work pr --title "Add authentication"
```

**Generated PR includes:**
- AI-generated title and summary
- List of all commits
- Diff statistics
- Test plan checklist

### work remove

Remove workspace and optionally delete branches.

```bash
r2r eac work remove [branch] [options]
```

**Options:**
- `--keep-branch` - Keep local branch
- `--delete-remote` - Delete remote branch too
- `--force`, `-f` - Force remove with uncommitted changes

**Examples:**

```bash
# Remove current workspace
r2r eac work remove

# Remove specific workspace
r2r eac work remove feature/old

# Delete remote branch too
r2r eac work remove --delete-remote
```

**Default behavior:**
- Removes workspace directory
- Deletes local branch
- Preserves remote branch

---

## Integration Patterns

### Option 1: Pull Request Workflow (Team Collaboration)

**Use when:**
- Working with a team
- Code review required
- CI/CD must run before merge
- Working on open source projects

**Flow:**

```bash
# 1. Create and develop
r2r eac work create feature/auth
cd ../eac-feature-auth
# Work with AI agent...

# 2. Commit and sync
r2r eac work commit --all
r2r eac work pull

# 3. Create PR
r2r eac work pr
# ✓ PR created: https://github.com/everything-as-code/eac/pull/123

# 4. After PR is merged externally, clean up
r2r eac work remove
```

### Option 2: Direct Merge Workflow (Solo/Trusted Contributors)

**Use when:**
- Working solo
- You own the repository
- Direct push access to main
- Internal tooling/scripts

**Flow:**

```bash
# 1. Create and develop
r2r eac work create feature/auth
cd ../eac-feature-auth
# Work with AI agent...

# 2. Commit and sync
r2r eac work commit --all
r2r eac work pull

# 3. Merge to local main
r2r eac work merge
# Automatically switches to main and merges

# 4. Push from main workspace
cd ../eac
git push origin main

# Workspace is auto-removed after successful merge
```

### Option 3: Hybrid Workflow

**Use when:**
- Most features go through PR
- Hotfixes merged directly
- Mixed team/solo work

**Flow:**

```bash
# Regular feature: PR workflow
r2r eac work create feature/new-api
# ... develop ...
r2r eac work pr

# Urgent hotfix: Direct merge
r2r eac work create hotfix/critical-bug
# ... fix ...
r2r eac work merge
cd ../eac && git push origin main
```

---

## Best Practices

### 1. One feature per workspace

```bash
# Good
r2r eac work create feature/authentication
r2r eac work create bugfix/login-error

# Avoid mixing features in one workspace
```

### 2. Sync regularly

```bash
# Before major work sessions
r2r eac work pull
```

Prevents large, conflict-prone rebases.

### 3. Commit frequently

```bash
# After each logical change
r2r eac work commit --all
```

AI generates better messages for small, focused commits.

### 4. Choose merge strategy per context

**Squash merge** (default) for features:
```bash
r2r eac work merge  # Clean history
```

**Regular merge** for preserving detailed history:
```bash
r2r eac work merge --no-squash
```

### 5. Clean up completed work

```bash
r2r eac work remove
```

Keeps workspace directory clean.

### 6. Use descriptive branch names

```bash
# Good
feature/user-authentication
bugfix/issue-123
refactor/api-v2

# Avoid
test, new-stuff, branch1
```

---

## Troubleshooting

### Branch already exists

```bash
# Remove old branch first
git branch -D feature/authentication
```

### Uncommitted changes detected

```bash
# Commit or autostash
r2r eac work commit --all
# or
r2r eac work pull --autostash
```

### Rebase conflicts

```bash
# Resolve conflicts in listed files
git add <resolved-files>
git rebase --continue

# Or abort
git rebase --abort
```

### gh CLI not found

Install GitHub CLI:

```bash
# Windows
winget install GitHub.cli

# macOS
brew install gh

# Authenticate
gh auth login
```

### Branch not up to date

```bash
# Error: branch not up to date with main (behind by 2 commits)
# Solution:
r2r eac work pull
```

---

## Advanced Usage

### Different base branches

```bash
# Create from develop
r2r eac work create feature/api --from=develop

# Sync with develop
r2r eac work pull --target=develop

# Merge to develop
r2r eac work merge --target=develop
```

### Custom workspace locations

```bash
r2r eac work create feature/auth --path=/custom/location
```

### Preserve workspace after merge

```bash
# Keep workspace for follow-up work
r2r eac work merge --keep-worktree
```

### Force operations

```bash
# Remove with uncommitted changes (destructive!)
r2r eac work remove --force
```

### Debug AI generation

```bash
r2r eac work commit --debug
r2r eac work merge --debug
r2r eac work pr --debug
```

---

## AI Agent Integration

### Multi-agent workflow

Run different AI agents simultaneously:

**Terminal 1:**
```bash
cd C:\source\eac
# Claude Code, Cursor, or Copilot
```

**Terminal 2:**
```bash
cd C:\source\eac-feature-auth
# Different agent or session
```

**Terminal 3:**
```bash
cd C:\source\eac-feature-api
# Another agent session
```

Each agent works independently without conflicts.

### AI-powered commits

The `work commit` command uses AI to:
- Analyze changed files and module ownership
- Follow project commit conventions
- Generate semantic commit messages
- Include module-specific context

Example:
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

---

## FAQ

**Q: Can I use regular git commands?**

A: Yes! Workspaces are normal git worktrees. All git commands work.

**Q: What's the difference between `work merge` and `work pr`?**

A:
- `work merge`: Merges directly to local main (then push manually)
- `work pr`: Creates GitHub pull request for review

**Q: Can I merge without squashing?**

A: Yes, use `--no-squash`:
```bash
r2r eac work merge --no-squash
```

**Q: How do I share workspaces?**

A: You don't. Workspaces are local. Share via PRs or pushed branches. Each developer creates their own workspaces.

**Q: What if I delete a workspace manually?**

A: Clean up git tracking:
```bash
git worktree prune
```

---

## Summary

The `work` command workflow for parallel AI-assisted development:

1. **Create**: `r2r eac work create feature/name`
2. **Develop**: Work with your AI coding agent
3. **Commit**: `r2r eac work commit --all`
4. **Sync**: `r2r eac work pull`
5. **Complete**:
   - **PR**: `r2r eac work pr`
   - **Direct**: `r2r eac work merge` then `git push` from main
6. **Clean up**: `r2r eac work remove`

This enables clean git history, multiple AI agent sessions, and streamlined parallel feature development.
