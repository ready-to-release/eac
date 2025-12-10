# Building and Testing Your Changes

{{ page_breadcrumb() }}

**Status:** Placeholder - Content coming soon

**Prerequisites:** [Your First Module](../getting-started/first-module.md), [Understanding Test Suites](../getting-started/understanding-test-suites.md)

## Planned Content

This tutorial teaches you how to efficiently build and test only the modules affected by your changes, following best practices for fast feedback loops.

### What You'll Learn

- Identify which modules are affected by your changes
- Build only affected modules (avoid rebuilding everything)
- Run appropriate test suites for different scenarios
- Validate module dependencies and contracts
- Debug test failures efficiently
- Understand build artifacts and dependencies

### Tutorial Structure

1. **Understanding what changed**
   - Check git status: `r2r show files changed`
   - See which modules are affected: `r2r get changed-modules`
   - Understand module ownership of files

2. **Building affected modules**
   - Build single module: `r2r build <module>`
   - Build changed modules: `r2r build $(r2r get changed-modules)`
   - Understand build dependencies: `r2r get build-deps <module>`
   - View build artifacts: `r2r show artifacts <module>`

3. **Running tests efficiently**
   - Fast feedback: `r2r test <module>` (commit suite, L0-L2)
   - Full validation: `r2r test <module> --suite acceptance`
   - Understand test selection and filtering

4. **Validating dependencies**
   - Check module contracts: `r2r validate dependencies`
   - Check file ownership: `r2r validate module-files`
   - Validate module hierarchy: `r2r validate module-hierarchy`

5. **Pre-commit validation**
   - Run all validations: `r2r validate`
   - Check specifications: `r2r validate specs`
   - Validate Go modules: `r2r validate go-tidy`
   - Set up pre-commit hooks

6. **Debugging test failures**
   - View test summary: `r2r show test-summary <module>`
   - List all failures: `r2r test debug`
   - Analyze test timings: `r2r show test-timings`
   - Re-run failed tests

### Example Workflow

The tutorial will walk through a realistic scenario:

1. You modify `go/eac-commands/cmd/build.go`
2. Check impact: `r2r show files changed` shows eac-commands affected
3. Build: `r2r build eac-commands`
4. Test fast: `r2r test eac-commands` (2 minutes, L0-L2)
5. Test thorough: `r2r test eac-commands --suite acceptance` (5 minutes)
6. Validate all: `r2r validate` (dependency checks, contracts, specs)
7. Commit with confidence

### Key Concepts Covered

- Module dependency graph and build order
- Incremental builds (only what changed)
- Test suite selection for fast feedback
- Validation checks before committing
- Build artifacts and their purposes
- Efficient debugging workflow

### Best Practices

- Always check `r2r show files changed` before building
- Run commit suite locally, acceptance suite in CI
- Validate before committing (use pre-commit hooks)
- Build and test only affected modules
- Use `r2r test debug` to quickly find failures

### Integration with Git Workflow

```bash
# Make changes
vim go/eac-commands/cmd/build.go

# Check impact
r2r show files changed

# Build and test affected modules
r2r build eac-commands
r2r test eac-commands

# Validate everything
r2r validate

# Commit (or use r2r work commit for AI-generated message)
git add .
git commit -m "feat(eac-commands): improve build command error handling"
```

### Next Steps

After completing this tutorial, you'll have an efficient build and test workflow. Continue to [Working with Git Worktrees](./working-with-worktrees.md) to learn parallel development with isolated workspaces.

{{ diataxis_footer() }}
