# Trunk-Based Development

## Overview

This repository follows **trunk-based development** practices, where all developers work on a single main branch (trunk) with short-lived feature branches and frequent integration.

## Core Principles

### 1. Single Source of Truth

- The `main` branch is always the source of truth
- All changes are integrated into `main` frequently (daily or multiple times per day)
- The `main` branch is always in a releasable state

### 2. Short-Lived Feature Branches

- Feature branches live for hours or at most 1-2 days
- Changes are kept small and focused
- Branches are deleted immediately after merging

### 3. Continuous Integration

- Every commit triggers automated builds and tests
- Developers pull from `main` frequently to stay up-to-date
- Merge conflicts are minimized through frequent integration

### 4. Incremental Development

- Large features are broken down into smaller, shippable increments
- Use feature flags for incomplete features rather than long-lived branches
- Each increment adds value independently

## Commit Workflow

### Daily Development Flow

1. **Start of Day**: Pull latest from `main`

   ```bash
   git checkout main
   git pull origin main
   ```

2. **Create Feature Branch** (optional, for small changes)

   ```bash
   git checkout -b feature/short-description
   ```

3. **Make Small, Focused Changes**
   - Keep changes under 200-400 lines when possible
   - One logical change per commit
   - Commit frequently to local branch

4. **Integrate Regularly**
   - Pull from `main` every few hours
   - Rebase or merge to stay current

   ```bash
   git fetch origin
   git rebase origin/main
   ```

5. **Push to Main**
   - Ensure all tests pass locally
   - Push directly to `main` or merge feature branch
   - Delete feature branch immediately after merge

### Commit Frequency

- **Minimum**: At least once per day
- **Recommended**: Multiple times per day
- **Ideal**: After each logical unit of work is complete

## Timeline Expectations

### Short-Lived Changes (Preferred)

- **Duration**: Hours to 1 day
- **Scope**: Bug fixes, small features, refactoring
- **Process**: Direct commits to `main` or quick feature branch

### Medium Changes

- **Duration**: 1-2 days maximum
- **Scope**: Larger features broken into increments
- **Process**: Feature branch with daily rebases, merged in pieces

### Large Features

- **Duration**: Multiple weeks
- **Scope**: Major new functionality
- **Process**:
  - Break into multiple small PRs
  - Use feature flags to hide incomplete work
  - Integrate each piece into `main` as completed
  - Never let branch diverge more than 2 days from `main`

## Best Practices

### DO

- Commit early and often
- Keep changes small and focused
- Write clear, semantic commit messages
- Pull from `main` multiple times per day
- Run tests before pushing
- Use feature flags for incomplete features
- Delete branches immediately after merge

### DON'T

- Keep feature branches open for more than 2 days
- Make large, sweeping changes in a single commit
- Let your branch diverge significantly from `main`
- Push broken code to `main`
- Keep "work in progress" branches for later
- Create long-lived development branches

## Conflict Resolution

When conflicts occur:

1. Pull latest from `main`
2. Resolve conflicts locally
3. Run full test suite
4. Commit resolution
5. Push immediately

## Emergency Fixes

For critical bugs requiring immediate fixes:

1. Create hotfix branch from `main`
2. Make minimal change to fix issue
3. Test thoroughly
4. Merge directly to `main` with priority
5. Tag the commit if needed for tracking

## Versioning in Trunk-Based Development

- Version numbers are managed through semantic commits
- Each merge to `main` may increment version automatically
- Releases are created by tagging commits on `main`
- See [Versioning](../../reference/continuous-delivery/versioning.md) for detailed version management

## References

- [Semantic Commits Guide](../../reference/continuous-delivery/semantic-commits.md)
- [Repository Layout](../../reference/continuous-delivery/repository-layout.md)
- [Versioning](../../reference/continuous-delivery/versioning.md)
