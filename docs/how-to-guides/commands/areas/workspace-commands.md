# Workspace Commands

Command reference for EAC's workspace management system.

## Quick Reference

| Command           | Description                                                |
| ----------------- | ---------------------------------------------------------- |
| `work create`     | Create a new workspace for parallel development            |
| `work commit`     | Commit changes with AI-generated commit messages           |
| `work pull`       | Sync workspace with latest main via rebase                 |
| `work merge`      | Merge workspace changes back to main (squash by default)   |
| `work pr`         | Create pull request with AI-generated description          |
| `work remove`     | Remove workspace and optionally delete associated branches |
| `show-workspaces` | List all workspaces and their status                       |

---

## work create

Create a new git worktree-based workspace for parallel development.

### Synopsis

```bash
r2r eac work create <branch-name> [options]
```

### Description

Creates an isolated workspace using git worktrees, enabling parallel development on multiple features without branch switching. Each workspace has:

- Its own working directory
- Dedicated branch
- Independent file state

### Arguments

| Argument      | Required | Description             |
| ------------- | -------- | ----------------------- |
| `branch-name` | Yes      | Name for the new branch |

### Flags

| Flag     | Short | Type   | Default              | Description                |
| -------- | ----- | ------ | -------------------- | -------------------------- |
| `--from` | `-f`  | string | `main`               | Base branch to create from |
| `--path` | `-p`  | string | `../<repo>-<branch>` | Custom workspace path      |

### Examples

```bash
# Create feature workspace
r2r eac work create feature/authentication

# Create from develop branch
r2r eac work create feature/api --from develop

# Custom workspace location
r2r eac work create feature/auth --path /workspaces/auth

# Bugfix workspace
r2r eac work create bugfix/login-error
```

### Output

```text
Creating workspace for feature/authentication...

  ✓ Branch created: feature/authentication
  ✓ Worktree created: ../eac-feature-authentication

Workspace ready!

  cd ../eac-feature-authentication

To start working:
  1. cd ../eac-feature-authentication
  2. Make your changes
  3. r2r eac work commit --all
```

### Exit Codes

| Code | Description                    |
| ---- | ------------------------------ |
| 0    | Workspace created successfully |
| 1    | Error creating workspace       |
| 2    | Branch already exists          |
| 3    | Path already exists            |

---

## work commit

Commit changes with AI-generated commit messages.

### Synopsis

```bash
r2r eac work commit [options]
```

### Description

Stages and commits changes using AI to generate semantic commit messages. The AI analyzes:

- Changed files and their diffs
- Module ownership
- Recent commit history
- Project conventions

### Flags

| Flag        | Short | Type   | Default | Description                     |
| ----------- | ----- | ------ | ------- | ------------------------------- |
| `--all`     | `-a`  | bool   | `false` | Stage all changes before commit |
| `--message` | `-m`  | string | -       | Use custom message (skip AI)    |
| `--debug`   | `-d`  | bool   | `false` | Save AI generation details      |

### Examples

```bash
# Commit with AI message (staged files only)
r2r eac work commit

# Stage all and commit
r2r eac work commit --all

# Use custom message
r2r eac work commit -m "fix: resolve login bug"

# Debug AI generation
r2r eac work commit --all --debug
```

### Output

```text
Generating commit message...

Analyzing changes...
  ✓ 3 files changed
  ✓ Module: eac-commands
  ✓ Change type: feature

Generated message:
────────────────────────────────────────
feat(eac-commands): add authentication middleware

Implement JWT-based authentication middleware for API endpoints:
- Token validation and parsing
- User context injection
- Error handling for invalid tokens

Files modified:
- go/eac/commands/impl/auth/middleware.go
- go/eac/commands/impl/auth/middleware_test.go
- go/eac/commands/impl/auth/token.go
────────────────────────────────────────

✓ Commit created: abc1234
```

### Exit Codes

| Code | Description                 |
| ---- | --------------------------- |
| 0    | Commit created successfully |
| 1    | Error creating commit       |
| 2    | No changes to commit        |
| 3    | AI generation failed        |

---

## work pull

Sync workspace with latest changes from the target branch via rebase.

### Synopsis

```bash
r2r eac work pull [options]
```

### Description

Fetches latest changes from remote and rebases the workspace branch onto the target branch. Keeps your branch up-to-date with main while maintaining linear history.

### Flags

| Flag          | Short | Type   | Default | Description                    |
| ------------- | ----- | ------ | ------- | ------------------------------ |
| `--target`    | `-t`  | string | `main`  | Target branch to sync with     |
| `--autostash` |       | bool   | `false` | Auto-stash uncommitted changes |
| `--no-fetch`  |       | bool   | `false` | Skip fetching from remote      |

### Examples

```bash
# Sync with main
r2r eac work pull

# Sync with develop
r2r eac work pull --target develop

# Auto-stash uncommitted changes
r2r eac work pull --autostash

# Skip fetch (use local only)
r2r eac work pull --no-fetch
```

### Output

```text
Syncing workspace with main...

  ✓ Fetched latest from origin
  ✓ Rebasing onto main...
  ✓ Rebase successful

Your branch is up to date with main.
  Commits ahead: 3
```

### Conflict Handling

```text
Syncing workspace with main...

  ✓ Fetched latest from origin
  ✗ Rebase conflict detected

Conflicts in:
  - go/eac/commands/impl/auth/middleware.go

To resolve:
  1. Edit conflicted files
  2. git add <resolved-files>
  3. git rebase --continue

Or abort with: git rebase --abort
```

### Exit Codes

| Code | Description                                 |
| ---- | ------------------------------------------- |
| 0    | Sync successful                             |
| 1    | Rebase conflicts (manual resolution needed) |
| 2    | Fetch failed                                |

---

## work merge

Merge workspace changes back to the target branch.

### Synopsis

```bash
r2r eac work merge [options]
```

### Description

Squash merges all workspace commits into a single commit on the target branch. Generates an AI commit message summarizing all changes. Optionally removes the workspace after successful merge.

### Flags

| Flag              | Short | Type   | Default | Description                        |
| ----------------- | ----- | ------ | ------- | ---------------------------------- |
| `--target`        | `-t`  | string | `main`  | Target branch to merge into        |
| `--no-squash`     |       | bool   | `false` | Regular merge (preserve commits)   |
| `--keep-worktree` |       | bool   | `false` | Don't remove workspace after merge |
| `--debug`         | `-d`  | bool   | `false` | Debug AI message generation        |

### Examples

```bash
# Squash merge to main (default)
r2r eac work merge

# Merge to develop
r2r eac work merge --target develop

# Keep all commits (no squash)
r2r eac work merge --no-squash

# Keep workspace after merge
r2r eac work merge --keep-worktree
```

### Output

```text
Merging workspace to main...

Validating workspace...
  ✓ No uncommitted changes
  ✓ Branch is up to date with main
  ✓ 5 commits to merge

Generating merge commit message...
  ✓ Analyzing all commits
  ✓ Message generated

Merging...
  ✓ Switched to main
  ✓ Squash merge complete
  ✓ Commit created: def5678

Cleaning up...
  ✓ Worktree removed
  ✓ Branch deleted

✓ Merge complete!

To push: git push origin main
```

### Exit Codes

| Code | Description                        |
| ---- | ---------------------------------- |
| 0    | Merge successful                   |
| 1    | Error during merge                 |
| 2    | Uncommitted changes (commit first) |
| 3    | Branch not up to date (pull first) |

---

## work pr

Create a pull request with AI-generated description.

### Synopsis

```bash
r2r eac work pr [options]
```

### Description

Pushes the workspace branch and creates a GitHub pull request. Uses AI to generate:

- PR title from changes
- Summary of all commits
- Test plan suggestions

Requires GitHub CLI (`gh`) to be installed and authenticated.

### Flags

| Flag       | Short | Type   | Default | Description               |
| ---------- | ----- | ------ | ------- | ------------------------- |
| `--target` | `-t`  | string | `main`  | Target branch for PR      |
| `--title`  |       | string | -       | Custom PR title (skip AI) |
| `--draft`  |       | bool   | `false` | Create as draft PR        |
| `--debug`  | `-d`  | bool   | `false` | Debug AI generation       |

### Examples

```bash
# Create PR to main
r2r eac work pr

# PR to develop
r2r eac work pr --target develop

# Custom title
r2r eac work pr --title "Add authentication system"

# Create as draft
r2r eac work pr --draft
```

### Output

```text
Creating pull request...

Pushing branch...
  ✓ Pushed feature/authentication to origin

Generating PR description...
  ✓ Analyzing 5 commits
  ✓ Description generated

Creating PR...
  ✓ PR created: #123

Pull Request: https://github.com/org/repo/pull/123

Title: feat: Add JWT authentication system

Summary:
  - Implement JWT token validation
  - Add authentication middleware
  - Create user context injection
  - Add comprehensive tests
```

### Exit Codes

| Code | Description                   |
| ---- | ----------------------------- |
| 0    | PR created successfully       |
| 1    | Error creating PR             |
| 2    | Uncommitted changes           |
| 3    | GitHub CLI not found          |
| 4    | Not authenticated with GitHub |

---

## work remove

Remove workspace and optionally delete associated branches.

### Synopsis

```bash
r2r eac work remove [branch] [options]
```

### Description

Removes the workspace directory and optionally deletes local and remote branches. Use after merging or abandoning a workspace.

### Arguments

| Argument | Required | Description                                 |
| -------- | -------- | ------------------------------------------- |
| `branch` | No       | Branch name (defaults to current workspace) |

### Flags

| Flag              | Short | Type | Default | Description                           |
| ----------------- | ----- | ---- | ------- | ------------------------------------- |
| `--keep-branch`   |       | bool | `false` | Keep local branch                     |
| `--delete-remote` |       | bool | `false` | Also delete remote branch             |
| `--force`         | `-f`  | bool | `false` | Force remove with uncommitted changes |

### Examples

```bash
# Remove current workspace
r2r eac work remove

# Remove specific workspace
r2r eac work remove feature/authentication

# Keep the branch
r2r eac work remove --keep-branch

# Delete remote branch too
r2r eac work remove --delete-remote

# Force remove (discard changes)
r2r eac work remove --force
```

### Output

```text
Removing workspace...

  ✓ Worktree removed: ../eac-feature-authentication
  ✓ Local branch deleted: feature/authentication

Workspace removed successfully.
```

### Exit Codes

| Code | Description                       |
| ---- | --------------------------------- |
| 0    | Workspace removed                 |
| 1    | Error removing workspace          |
| 2    | Uncommitted changes (use --force) |
| 3    | Workspace not found               |

---

## show-workspaces

List all workspaces and their status.

### Synopsis

```bash
r2r eac show-workspaces [options]
```

### Description

Displays all git worktrees with their status, branch information, and sync state.

### Flags

| Flag     | Short | Type | Default | Description    |
| -------- | ----- | ---- | ------- | -------------- |
| `--json` |       | bool | `false` | Output as JSON |

### Examples

```bash
# List all workspaces
r2r eac show-workspaces

# JSON output
r2r eac show-workspaces --json
```

### Output

```text
Workspaces
═══════════════════════════════════════════════════════════════════

│ Path                              │ Branch                   │ Status │ Ahead │ Behind │
├───────────────────────────────────┼──────────────────────────┼────────┼───────┼────────┤
│ C:\projects\eac                   │ main                     │ clean  │ 0     │ 0      │
│ C:\projects\eac-feature-auth      │ feature/authentication   │ dirty  │ 3     │ 0      │
│ C:\projects\eac-bugfix-123        │ bugfix/login-error       │ clean  │ 1     │ 2      │

Total: 3 workspaces (1 main, 2 feature)
```

### Exit Codes

| Code | Description              |
| ---- | ------------------------ |
| 0    | Success                  |
| 1    | Error listing workspaces |

---

## Common Workflows

### Feature Development

```bash
# 1. Create workspace
r2r eac work create feature/new-api
cd ../eac-feature-new-api

# 2. Develop
# ... make changes ...

# 3. Commit frequently
r2r eac work commit --all

# 4. Stay in sync
r2r eac work pull

# 5. Create PR
r2r eac work pr

# 6. After merge, cleanup
r2r eac work remove
```

### Solo Development

```bash
# 1. Create and develop
r2r eac work create feature/quick-fix
cd ../eac-feature-quick-fix

# 2. Make changes and commit
r2r eac work commit --all

# 3. Sync and merge directly
r2r eac work pull
r2r eac work merge

# 4. Push from main
cd ../eac
git push origin main
```

### Parallel Development

```bash
# Terminal 1
r2r eac work create feature/auth
cd ../eac-feature-auth
# Start AI coding session

# Terminal 2
r2r eac work create feature/api
cd ../eac-feature-api
# Start separate AI session

# Both can work independently
```

---

## Related Documentation

- [Workspace Overview](workspace-overview.md) - Concepts and workflows
- [Workspace Configuration](workspace-configuration.md) - Configuration reference
- [Commit Command](../commit-command.md) - AI commit messages
