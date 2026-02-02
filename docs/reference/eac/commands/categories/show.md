# Show Commands

## Overview

Show commands display repository information in human-readable formats optimized for interactive terminal use.

They provide formatted tables, lists, and text designed for visual consumption rather than programmatic processing.

**Key Characteristics**:

- Human-readable formatted output
- Tables with aligned columns
- Colorized status indicators
- Designed for terminal display
- Interactive exploration

**When to use**: During interactive development and troubleshooting when you need to quickly understand repository state.

**For automation**: Use [get commands](./get.md) instead, which provide JSON output.

## All Show Commands

<!-- book:category-commands show -->

## Common Patterns

### Table Output

Most show commands display data as formatted tables:

```bash
$ eac show modules
┌───────────────┬─────────────┬────────────────────┬──────┐
│ Moniker       │ Type        │ Path               │ Files│
├───────────────┼─────────────┼────────────────────┼──────┤
│ eac-commands  │ go-commands │ go/cli/eac    │   45 │
│ eac-core      │ go-library  │ go/core        │   32 │
│ src-auth      │ go-library  │ go/src/auth        │   18 │
└───────────────┴─────────────┴────────────────────┴──────┘
```

**Features**:

- Aligned columns
- Headers with separators
- Auto-truncated long values
- Sorted by relevance

### Status Indicators

Commands use symbols and colors for status:

```text
✓ Success    (green)
✗ Failed     (red)
⚠ Warning    (yellow)
⋯ Running    (blue)
⊘ Skipped    (gray)
```

### Report Format

Summary commands generate markdown-compatible reports:

```markdown
# Test Summary: module-name (suite-name)

## Results
✓ Passed: 45
✗ Failed: 2
Total: 47 (95.7% pass rate)

## Performance
- Average: 0.45s
- Total: 21.6s
```

## show vs get Duality

Many show commands have corresponding `get` commands that provide the same information in JSON format:

| show command          | get command          | Use Case              |
| --------------------- | -------------------- | --------------------- |
| `show modules`        | `get modules`        | Module information    |
| `show dependencies`   | `get dependencies`   | Dependency graph      |
| `show files`          | `get files`          | File ownership        |
| `show config`         | `get config`         | Configuration         |
| `show tests`          | `get tests`          | Test information      |
| `show environments`   | `get environments`   | Environment contracts |
| `show build-times`    | `get build-times`    | Build performance     |
| `show test-timings`   | `get test-timings`   | Test performance      |
| `show suite <name>`   | `get suite <name>`   | Test suite details    |
| `show artifacts <m>`  | `get artifacts <m>`  | Build artifacts       |
| `show valid-commands` | `get valid-commands` | Command list          |

**Rule**: Use `show` for interactive terminal use, `get` for scripts and automation.

## Common Workflows

### Exploring the Repository

```bash
# Start with modules
eac show modules

# Understand dependencies
eac show dependencies

# See file organization
eac show files

# Check configuration
eac show config
```

### Checking Status

```bash
# See what's changed
eac show files-changed

# Check workspaces
eac show workspaces

# View test status
eac show tests
```

### Reviewing Results

```bash
# Build summary
eac show build-summary eac-commands

# Test summary
eac show test-summary src-auth acceptance

# Performance analysis
eac show build-times
eac show test-timings
```

### Getting Help

```bash
# General help
eac show help

# List all commands
eac show valid-commands

# Command-specific help
eac help <command>
```

## Usage Examples

### Module Discovery

```bash
# All modules
eac show modules

# Module types
eac show component-types

# Dependency graph
eac show dependencies
```

### File Investigation

```bash
# All files (large output)
eac show files

# Changed files only
eac show files-changed

# Staged files only
eac show files-staged
```

### Test Analysis

```bash
# All tests
eac show tests

# Specific suite
eac show suite acceptance

# Test performance
eac show test-timings

# Test summary for CI
eac show test-summary src-auth acceptance
```

### Environment Information

```bash
# Environments
eac show environments

# Documentation books
eac show books

# Git worktrees
eac show workspaces
```

## Output Customization

### Piping to less

For large output, pipe to `less`:

```bash
eac show files | less
eac show tests | less
```

### Filtering with grep

Filter output with grep:

```bash
# Find specific module
eac show modules | grep "src-auth"

# Find failed tests
eac show tests | grep "✗"

# Find go-library modules
eac show modules | grep "go-library"
```

### Saving Output

Save formatted output to files:

```bash
# Save to file
eac show modules > modules.txt

# Append to log
eac show test-summary src-auth acceptance >> test-report.log
```

## Performance Notes

### Fast Commands

These commands execute quickly (< 1s):

- `show modules`
- `show component-types`
- `show dependencies`
- `show config`
- `show workspaces`
- `show environments`
- `show books`

### Moderate Commands

These commands may take a few seconds:

- `show files` (loads ~2,690 files)
- `show tests` (scans all test files)
- `show build-times` (parses build logs)
- `show test-timings` (parses test logs)

### Expensive Commands

Avoid in tight loops:

- `show files` - Loads all files (~19k tokens)
- Use `show files-changed` or `show files-staged` instead when possible

## Best Practices

### Interactive Use

1. **Start with show modules** to understand repository structure
2. **Use show commands for exploration** rather than scripting
3. **Pipe large output to less** for easier navigation
4. **Use grep for filtering** when you know what you're looking for

### When to Use get Instead

Use [get commands](./get.md) when you need to:

- Process output with jq or other tools
- Cache results for repeated queries
- Integrate with CI/CD pipelines
- Parse data programmatically

### Combining Commands

```bash
# Find module, then show its dependencies
MODULE=$(eac show modules | grep "src-auth" | awk '{print $1}')
eac show dependencies | grep "$MODULE"

# Better: use get commands for this
eac get dependencies | jq ".dependencies[\"src-auth\"]"
```

## Common Issues

### Output Too Large

**Problem**: Command output fills the screen

**Solution**: Pipe to `less` or filter with `grep`

```bash
eac show files | less
eac show tests | grep "@auth"
```

### Colors Not Working

**Problem**: No color in output

**Solution**: Ensure terminal supports colors, or check if output is piped

```bash
# Force colors (if supported)
export FORCE_COLOR=1
eac show modules
```

### Table Misaligned

**Problem**: Table columns don't align

**Solution**: Ensure terminal is wide enough or use narrower output

```bash
# Use get command for narrow terminals
eac get modules | jq -r '.modules[] | "\(.moniker): \(.type)"'
```

## See Also

- [Get Commands](./get.md) - JSON output for automation
- [Output Formats](../overview/output-formats.md) - Understanding output types
- [Command Taxonomy](../overview/command-taxonomy.md) - Command organization
- [Show Commands Guide](../../../../how-to-guides/eac/commands/getting-started/explore-your-repository.md) - How-to guide
