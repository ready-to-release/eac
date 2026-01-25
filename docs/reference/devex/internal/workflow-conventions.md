# Workflow Conventions

Git workflow and contribution guidelines for EAC monorepo contributors.

## Branching Strategy

The EAC repository uses **trunk-based development**:

- **`main`** - The trunk; always deployable
- **Feature branches** - Short-lived branches for changes
- **No long-lived branches** - Merge within days, not weeks

## Branch Naming

```
<type>/<short-description>

Examples:
  feat/add-scan-command
  fix/config-loading-error
  docs/update-getting-started
  refactor/simplify-module-loader
```

**Types**:

| Type | Purpose |
| ---- | ------- |
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `refactor` | Code restructuring |
| `test` | Test additions/fixes |
| `chore` | Maintenance tasks |

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Examples**:

```
feat(commands): add scan-zap command for DAST scanning

fix(core): resolve config loading race condition

docs(reference): update command cheat sheet

test(eac-commands): add integration tests for build command
```

**Scopes** (optional but helpful):

- `commands` - EAC commands
- `core` - eac-core module
- `cli` - R2R CLI
- `docs` - Documentation
- `ci` - CI/CD workflows

## Pull Request Workflow

### 1. Create Branch

```bash
git checkout main
git pull origin main
git checkout -b feat/my-feature
```

### 2. Make Changes

- Keep commits atomic and focused
- Run tests locally: `r2r eac test <module>`
- Validate contracts: `r2r eac validate`

### 3. Push and Create PR

```bash
git push -u origin feat/my-feature
```

Then create PR via GitHub or:

```bash
r2r eac create-pr
```

### 4. PR Requirements

Before merge, PRs must:

- [ ] Pass all CI checks
- [ ] Have passing tests
- [ ] Pass contract validation
- [ ] Have descriptive title and description
- [ ] Be reviewed (if required)

### 5. Merge

- Use **squash merge** for feature branches
- Delete branch after merge

## Code Review Guidelines

**For Authors**:

- Keep PRs small and focused
- Provide context in description
- Respond to feedback promptly

**For Reviewers**:

- Review within 24 hours
- Be constructive and specific
- Approve when ready, request changes when needed

## CI/CD Pipeline

Every push triggers:

1. **Contract Validation** - Schema and reference checks
2. **Build** - Compile all affected modules
3. **Test** - Run test suites
4. **Lint** - Code style checks

View CI status:

```bash
r2r eac pipeline-status
```

## Release Process

Releases are triggered by changelog updates:

1. Update module's `CHANGELOG.md`
2. Add version entry with changes
3. Merge to main
4. CI creates tag and release

Check release readiness:

```bash
r2r eac release-check-ci <module>
```

## Working with Modules

### Find Affected Modules

```bash
# See which modules changed
r2r eac get-changed-modules

# Build only changed modules
r2r eac build --changed
```

### Module Dependencies

```bash
# View dependency graph
r2r eac show-dependencies

# Build with dependencies
r2r eac build <module> --deps
```

## Best Practices

1. **Small, focused changes** - Easier to review and merge
2. **Test before pushing** - Catch issues early
3. **Update docs with code** - Keep documentation current
4. **Validate contracts** - Run `r2r eac validate` before committing
5. **Use conventional commits** - Enables automated changelog generation

## Troubleshooting

### CI Failing

```bash
# Run same checks locally
r2r eac validate
r2r eac test <module>
r2r eac update-lint <module>
```

### Merge Conflicts

```bash
# Rebase on main
git fetch origin
git rebase origin/main

# Or merge main into branch
git merge origin/main
```

### Stale Branch

If branch is old:

```bash
git checkout main
git pull
git checkout feat/my-feature
git rebase main
```

## See Also

- [Commit Messages](../../../explanation/continuous-delivery/workflow/commit-messages.md) - Detailed commit guidelines
- [Trunk-Based Development](../../../explanation/continuous-delivery/workflow/trunk-based-development.md) - Branching philosophy
- [CI Workflows](../../eac/continuous-delivery/workflows/) - CI/CD documentation
