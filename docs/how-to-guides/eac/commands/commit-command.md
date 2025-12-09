# Commit Command

{{ page_breadcrumb() }}

**Problem**: Writing high-quality, semantic commit messages is time-consuming and requires consistency across team members, especially for complex multi-module changes.

**Solution**: Use `commit` to generate AI-powered commit messages that analyze your staged changes and follow project conventions automatically.

## Key Benefits

- Multi-phase AI generation with validation
- Parallel processing for multi-module commits
- Automatic contract validation
- Standard commit types (feat, fix, refactor, docs, chore, test, perf, style)
- Retry logic with auto-cleanup of formatting issues
- Context-aware module analysis

For technical details and all available options, see the [Command Reference](../../../reference/commands/commit-reference.md).

To understand how the AI analyzes your changes, see [AI Process Explanation](./commit-ai-process.md).

## Quick Start

```bash
# Stage your changes
git add src/auth/login.go src/auth/login_test.go

# Generate commit message
r2r eac create commit-message

# Review and create commit automatically
r2r eac create commit-message --commit
```

### Basic Command

```bash
r2r eac create commit-message [options]

# Options:
--commit, -c           # Automatically create the commit (no manual edit)
--debug, -d            # Save intermediate outputs to out/ directory

# Examples:
r2r eac create commit-message                    # Generate message, open editor
r2r eac create commit-message --commit           # Generate and commit automatically
r2r eac create commit-message --debug            # Debug AI generation process
r2r eac create commit-message --commit --debug   # Auto-commit with debug output
```

**What it does:**

1. **Context Analysis**: Analyzes git status, diff, and recent commits
2. **Summary Generation**: Creates concise change summary
3. **Module Sections**: Processes each affected module in parallel
4. **Assembly**: Combines sections into structured commit message
5. **Validation**: Validates against project contracts
6. **Auto-cleanup**: Fixes common formatting issues (up to 5 retries)
7. **Commit**: Opens editor or commits directly with `--commit`

## Common Workflows

### Standard Workflow

```bash
# 1. Make changes
vim src/auth/login.go

# 2. Stage changes
git add src/auth/

# 3. Generate commit message
r2r eac create commit-message

# 4. Review in editor, save, and commit
```

### Fast Workflow (Auto-commit)

```bash
# 1. Make changes and stage
git add .

# 2. Generate and commit in one step
r2r eac create commit-message --commit

# No editor opens - message is auto-committed
```

### Iterative Workflow

```bash
# 1. Generate initial commit
r2r eac create commit-message --commit

# 2. Realize you need to add more changes
git reset --soft HEAD~1

# 3. Stage additional changes
git add src/auth/logout.go

# 4. Generate new commit with all changes
r2r eac create commit-message --commit
```

### Multi-Module Development

```bash
# Work on multiple modules
vim src/auth/jwt.go
vim go/eac/core/validation.go
vim src/api/handlers.go

# Stage all changes
git add go/eac/auth/ go/eac/core/ go/eac/api/

# AI processes modules in parallel
r2r eac create commit-message --debug

# Review parallel processing in out/commit-message/
```

## Best Practices

### Staging Strategy

```bash
# ✅ Good: Stage related changes together
git add src/auth/jwt.go src/auth/jwt_test.go
r2r eac create commit-message --commit

# ❌ Avoid: Mixing unrelated changes
git add src/auth/jwt.go src/api/unrelated.go src/docs/random.md
r2r eac create commit-message --commit
```

### Commit Frequency

```bash
# ✅ Good: Frequent, focused commits
git add src/auth/jwt.go
r2r eac create commit-message --commit

git add src/auth/middleware.go
r2r eac create commit-message --commit

# ❌ Avoid: Massive commits with many unrelated changes
git add .
r2r eac create commit-message --commit
```

### Message Quality

The AI generates better messages when you:

- Stage related changes together
- Make focused changes to specific modules
- Include tests with implementation
- Have clear, descriptive file names
- Follow consistent coding patterns

## Troubleshooting

| Problem | Solution |
|---------|----------|
| No staged changes | Run `git add <files>` first |
| AI API error | Check API key configuration with `r2r eac init` |
| Invalid module | Verify module exists in contracts: `r2r eac show modules` |
| Validation fails | Use `--debug` to inspect outputs, check contract rules |
| Message too generic | Stage more focused changes, smaller commits |
| Wrong commit type | AI analyzes git diff - ensure changes match intent |
| Parallel processing slow | Normal for multi-module changes; AI processes in parallel |

For more troubleshooting details and advanced usage, see the [Command Reference](../../../reference/commands/commit-reference.md).

## Next Steps

- **Learn the details**: Read the [Command Reference](../../../reference/commands/commit-reference.md) for all flags, commit types, and technical configuration
- **Understand the AI**: Read [AI Process Explanation](./commit-ai-process.md) to see how the 5-phase generation works
- **Work command integration**: Use `r2r eac work commit` for workspace-aware commits
- **Custom editor**: Set `GIT_EDITOR` to your preferred editor

{{ diataxis_footer() }}
