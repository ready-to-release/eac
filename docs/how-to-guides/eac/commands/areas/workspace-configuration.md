# Workspace Configuration

{{ page_breadcrumb() }}

This guide covers configuration options for EAC's workspace management system, including worktree paths, branch naming, and merge settings.

## Configuration Files

| File                       | Purpose                   |
| -------------------------- | ------------------------- |
| `.r2r/eac/work/config.yml` | Workspace settings        |
| `.r2r/eac/ai/commit/`      | AI commit message prompts |
| `.r2r/eac/ai/pr/`          | AI PR description prompts |

## Workspace Settings

### Basic Configuration

`.r2r/eac/work/config.yml`:

```yaml
# Workspace directory naming
workspace:
  # Pattern for workspace directory names
  # Available variables: {repo}, {branch}, {safe-branch}
  path_pattern: "../{repo}-{safe-branch}"

  # Default base branch for new workspaces
  default_base: main

  # Auto-cleanup completed workspaces
  auto_cleanup: false

# Branch naming
branches:
  # Allowed prefixes
  prefixes:
    - feature/
    - bugfix/
    - hotfix/
    - refactor/
    - docs/
    - chore/

  # Branch name validation regex
  pattern: "^(feature|bugfix|hotfix|refactor|docs|chore)/[a-z0-9-]+$"

# Merge settings
merge:
  # Default merge strategy
  strategy: squash  # squash, merge, rebase

  # Delete branch after merge
  delete_branch: true

  # Delete worktree after merge
  delete_worktree: true
```

### Path Pattern Variables

| Variable        | Description              | Example        |
| --------------- | ------------------------ | -------------- |
| `{repo}`        | Repository name          | `eac`          |
| `{branch}`      | Full branch name         | `feature/auth` |
| `{safe-branch}` | Branch with `/` replaced | `feature-auth` |
| `{date}`        | Current date             | `2024-12-01`   |

### Examples

```yaml
# Default: ../eac-feature-auth
path_pattern: "../{repo}-{safe-branch}"

# With date: ../eac-2024-12-01-auth
path_pattern: "../{repo}-{date}-{safe-branch}"

# Custom location: /workspaces/eac-feature-auth
path_pattern: "/workspaces/{repo}-{safe-branch}"
```

## Branch Naming

### Prefix Configuration

```yaml
branches:
  prefixes:
    - feature/     # New features
    - bugfix/      # Bug fixes
    - hotfix/      # Urgent fixes
    - refactor/    # Code refactoring
    - docs/        # Documentation
    - chore/       # Maintenance tasks
    - experiment/  # Experimental work
```

### Validation Pattern

```yaml
branches:
  # Enforce lowercase, alphanumeric with hyphens
  pattern: "^(feature|bugfix|hotfix)/[a-z0-9][a-z0-9-]*[a-z0-9]$"
```

### Branch Examples

| Valid                 | Invalid               | Reason             |
| --------------------- | --------------------- | ------------------ |
| `feature/auth`        | `auth`                | Missing prefix     |
| `bugfix/fix-123`      | `bugfix/Fix-123`      | Uppercase          |
| `hotfix/urgent-patch` | `hotfix/urgent_patch` | Underscore         |
| `feature/add-api-v2`  | `feature/-api`        | Starts with hyphen |

## Merge Configuration

### Squash Merge (Default)

```yaml
merge:
  strategy: squash
  # All commits combined into one
  # AI generates combined commit message
```

Workflow:

1. All workspace commits squashed
2. AI generates summary commit message
3. Single commit on target branch

### Regular Merge

```yaml
merge:
  strategy: merge
  # Preserves all commits
  # Creates merge commit
```

Workflow:

1. All commits preserved
2. Merge commit created
3. Full history maintained

### Rebase Merge

```yaml
merge:
  strategy: rebase
  # Replays commits on target
  # Linear history
```

Workflow:

1. Commits rebased onto target
2. No merge commit
3. Linear history

### Post-Merge Cleanup

```yaml
merge:
  # Delete local branch after merge
  delete_branch: true

  # Delete worktree directory after merge
  delete_worktree: true

  # Also delete remote branch
  delete_remote: false
```

## AI Commit Configuration

### Prompt Templates

Location: `.r2r/eac/ai/commit/`

```text
.r2r/eac/ai/commit/
├── context-prompt.md       # Context gathering
├── summary-prompt.md       # Summary generation
├── module-prompt.md        # Per-module analysis
└── assembly-prompt.md      # Final message assembly
```

### Summary Prompt

```markdown
# Commit Summary Generation

## Context
Generate a concise summary of the changes.

## Changes
{{.Diff}}

## Recent Commits
{{.RecentCommits}}

## Guidelines
1. Use conventional commit format
2. Identify primary change type (feat, fix, refactor, etc.)
3. Identify primary module affected
4. Write clear, concise subject line
5. Keep under 72 characters

## Output Format
<type>(<scope>): <description>
```

### Module Prompt

```markdown
# Module Change Analysis

## Module
- Moniker: {{.Module.Moniker}}
- Type: {{.Module.Type}}

## Changed Files
{{range .Files}}
- {{.Path}}: {{.Status}}
{{end}}

## Diff
{{.Diff}}

## Guidelines
1. Describe what changed in this module
2. Explain why the change was made
3. Note any breaking changes
4. List key files modified

## Output
Bullet points describing module changes.
```

## AI PR Configuration

### Prompt Templates

Location: `.r2r/eac/ai/pr/`

```text
.r2r/eac/ai/pr/
├── title-prompt.md         # PR title generation
├── summary-prompt.md       # PR description
└── test-plan-prompt.md     # Test plan generation
```

### PR Description Template

```markdown
# PR Description Generation

## Branch Info
- Source: {{.Branch}}
- Target: {{.TargetBranch}}
- Commits: {{.CommitCount}}

## Commits
{{range .Commits}}
- {{.Hash}}: {{.Message}}
{{end}}

## Changed Files
{{.FileStats}}

## Guidelines
1. Write clear summary of all changes
2. Group by module or feature
3. Note any breaking changes
4. Suggest testing approach

## Output Format
## Summary
<bullet points>

## Changes
<grouped by module>

## Test Plan
<testing checklist>
```

## Pull Settings

### Sync Configuration

```yaml
pull:
  # Target branch for sync
  default_target: main

  # Auto-stash uncommitted changes
  autostash: true

  # Fetch before rebase
  fetch: true

  # Rebase strategy
  strategy: rebase  # rebase, merge
```

### Conflict Handling

```yaml
pull:
  # On conflict
  on_conflict: pause  # pause, abort

  # Auto-resolve strategy
  auto_resolve: none  # none, ours, theirs
```

## Environment Variables

| Variable              | Description                   | Default        |
| --------------------- | ----------------------------- | -------------- |
| `WORK_BASE_PATH`      | Base directory for workspaces | `..`           |
| `WORK_DEFAULT_BRANCH` | Default base branch           | `main`         |
| `WORK_MERGE_STRATEGY` | Default merge strategy        | `squash`       |
| `GIT_EDITOR`          | Editor for commit messages    | System default |

## Git Configuration

### Recommended Settings

```bash
# Enable worktree support
git config --global extensions.worktreeConfig true

# Default branch name
git config --global init.defaultBranch main

# Rebase by default on pull
git config --global pull.rebase true

# Auto-stash on rebase
git config --global rebase.autoStash true
```

### Per-Workspace Git Config

Each worktree can have its own config:

```bash
# In workspace directory
git config --worktree user.email "work@example.com"
```

## Integration Settings

### GitHub CLI Integration

```yaml
pr:
  # Use GitHub CLI for PR creation
  use_gh_cli: true

  # Default PR settings
  draft: false
  assignee: "@me"
  labels: []

  # Auto-request reviewers
  reviewers: []
```

### CI Integration

```yaml
# Run CI checks before merge
ci:
  check_before_merge: true
  required_checks:
    - build
    - test
    - lint
```

## Example Configurations

### Solo Developer

```yaml
workspace:
  path_pattern: "../{repo}-{safe-branch}"
  default_base: main
  auto_cleanup: true

merge:
  strategy: squash
  delete_branch: true
  delete_worktree: true

branches:
  prefixes:
    - feature/
    - fix/
```

### Team Environment

```yaml
workspace:
  path_pattern: "../{repo}-{safe-branch}"
  default_base: develop
  auto_cleanup: false

merge:
  strategy: merge  # Preserve history for review
  delete_branch: false  # Keep for reference
  delete_worktree: true

pr:
  use_gh_cli: true
  draft: true
  reviewers:
    - "@team/reviewers"

branches:
  prefixes:
    - feature/
    - bugfix/
    - hotfix/
  pattern: "^(feature|bugfix|hotfix)/[A-Z]+-[0-9]+-[a-z0-9-]+$"  # JIRA format
```

### Multiple Projects

```yaml
workspace:
  # Separate directory per project
  path_pattern: "/workspaces/{repo}/{safe-branch}"
  default_base: main

merge:
  strategy: squash
  delete_branch: true
  delete_worktree: false  # Keep for reference
```

## Troubleshooting

| Issue                   | Cause                        | Solution                     |
| ----------------------- | ---------------------------- | ---------------------------- |
| Worktree exists         | Directory already exists     | Remove or use different name |
| Branch in use           | Branch checked out elsewhere | Switch other worktree        |
| Merge conflicts         | Diverged from target         | Resolve conflicts, continue  |
| Permission denied       | File locked                  | Close editors, retry         |
| Path too long (Windows) | Deep nesting                 | Use shorter path pattern     |

## Related Documentation

- [Workspace Overview](workspace-overview.md) - Concepts and workflows
- [Workspace Commands](workspace-commands.md) - Command reference
- [Commit Command](../commit-command.md) - AI commit messages

{{ diataxis_footer() }}
