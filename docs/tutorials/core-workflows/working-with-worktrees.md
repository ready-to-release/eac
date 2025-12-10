# Working with Git Worktrees

{{ page_breadcrumb() }}

**Status:** Placeholder - Content coming soon

**Prerequisites:** [Building and Testing Changes](./building-and-testing.md), Git fundamentals

## Planned Content

This tutorial teaches you how to use git worktrees for parallel development, enabling you to work on multiple features simultaneously without context switching or branch management overhead.

### What You'll Learn

- Understand git worktrees and their benefits
- Create isolated feature workspaces with `r2r work create`
- Work independently in separate worktrees
- Make commits with AI-generated messages
- Sync with main branch using `r2r work pull`
- Merge changes back with squash
- Clean up worktrees when done

### Tutorial Structure

1. **Understanding worktrees**
   - What are git worktrees?
   - Benefits: parallel work, no context switching, independent state
   - Use cases: multiple features, experimentation, code review

2. **Create a feature workspace**
   - Create worktree: `r2r work create my-feature`
   - Understand directory structure
   - Navigate to worktree: `cd ../worktrees/my-feature`
   - List all workspaces: `r2r show workspaces`

3. **Work independently**
   - Make changes in your worktree
   - Build and test as usual
   - No impact on main working directory
   - Switch between worktrees freely

4. **Making commits**
   - Stage changes: `git add .`
   - AI-generated commit: `r2r work commit`
   - Or manual commit: `git commit -m "..."`
   - View commit history

5. **Syncing with main**
   - Pull latest changes: `r2r work pull` (rebase on main)
   - Resolve conflicts if any
   - Keep your feature branch up to date

6. **Merging back to main**
   - Create PR: `r2r create pr` (AI-generated description)
   - After approval, merge: `r2r work merge` (squash by default)
   - Understand squash vs. merge strategies

7. **Cleanup**
   - Remove worktree: `r2r work remove my-feature`
   - Option to delete branch: `r2r work remove my-feature --delete-branch`
   - Clean up stale worktrees

### Example Workflow

The tutorial will demonstrate a realistic parallel development scenario:

**Main working directory:**
- Working on feature A (dashboard improvements)

**Worktree 1 (../worktrees/api-auth):**
- Working on feature B (API authentication)

**Worktree 2 (../worktrees/fix-bug-123):**
- Working on urgent bugfix

You can:
- Switch between workspaces instantly (just `cd`)
- Build and test independently
- Commit to different branches simultaneously
- No stashing or branch switching required

### Key Commands

```bash
# Create workspace
r2r work create my-feature

# Navigate to workspace
cd ../worktrees/my-feature

# Work as usual
# ... make changes ...

# Commit with AI
r2r work commit

# Sync with main
r2r work pull

# Go back to main directory
cd ../../cli

# Create PR
cd ../worktrees/my-feature
r2r create pr

# After merge, clean up
r2r work remove my-feature --delete-branch
```

### Key Concepts Covered

- Git worktrees and their advantages
- Parallel feature development
- AI-generated commit messages
- Keeping feature branches synchronized
- Pull request workflow
- Workspace lifecycle management

### Multi-Worktree Awareness

Important considerations when using worktrees:

- Each worktree is independent (files, branches, index)
- Claude Code sessions can run in parallel (one per worktree)
- User coordinates git operations across worktrees
- `.git/worktrees/` contains worktree metadata
- Some git operations affect all worktrees (tags, config)

### Best Practices

- Use worktrees for features that take multiple sessions
- Keep worktree names descriptive and short
- Clean up merged worktrees promptly
- Use `r2r show workspaces` to track active workspaces
- Sync with main regularly to avoid large conflicts

### Troubleshooting

Common issues and solutions:

- **Worktree already exists**: Use unique names or remove old worktree
- **Branch conflicts**: Use `r2r work pull` to sync with main
- **Stale worktrees**: Clean up with `r2r work remove`

### Next Steps

After completing this tutorial, you'll be able to work on multiple features in parallel. Continue to [Making Your First Release](./making-first-release.md) to learn how to prepare and release modules.

{{ diataxis_footer() }}
