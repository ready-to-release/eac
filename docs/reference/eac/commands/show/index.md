# Show Commands

Show commands display repository information in human-readable formats optimized for interactive terminal use.

They provide formatted tables, lists, and text designed for visual consumption rather than programmatic processing.

**Key Characteristics**:

- Human-readable formatted output
- Tables with aligned columns
- Colorized status indicators
- Designed for terminal display
- Interactive exploration

**When to use**: During interactive development and troubleshooting when you need to quickly understand repository state.

**For automation**: Use [get commands](../get/index.md) instead, which provide JSON output.

## Commands in this Category

| Command                                                  | Purpose                                               |
| -------------------------------------------------------- | ----------------------------------------------------- |
| [show](./show.md)                                        | Base show command                                     |
| [show approval-comments](./approval-comments.md)         | Display PR approval comments in human-readable format |
| [show artifacts](./artifacts.md)                         | Display artifacts with status                         |
| [show books](./books.md)                                 | Display all book configurations                       |
| [show build-summary](./build-summary.md)                 | Generate pretty build summary                         |
| [show build-times](./build-times.md)                     | Show build timing analysis                            |
| [show changelog](./changelog.md)                         | Display changelog entries in human-readable format    |
| [show config](./config.md)                               | Display all loaded configurations                     |
| [show dependencies](./dependencies.md)                   | Show dependency graph in table format                 |
| [show environments](./environments.md)                   | Show all environment contracts                        |
| [show files](./files.md)                                 | Show repository files with module ownership           |
| [show files-changed](./files-changed.md)                 | Show changed files with module ownership              |
| [show files-staged](./files-staged.md)                   | Show staged files with module ownership               |
| [show help](./help.md)                                   | Display help information                              |
| [show modules](./modules.md)                             | Display all module contracts in table                 |
| [show component-kinds](./component-kinds.md)             | Show all component types grouped by count             |
| [show release-notes](./release-notes.md)                 | Display release notes in human-readable format        |
| [show specs](./specs.md)                                 | Display specifications for a release                  |
| [show suite](./suite.md)                                 | Display detailed test suite information               |
| [show test-summary](./test-summary.md)                   | Generate pretty test summary                          |
| [show test-timings](./test-timings.md)                   | Show test timing analysis                             |
| [show tests](./tests.md)                                 | Show all tests in table                               |
| [show valid-commands](./valid-commands.md)               | Show all valid commands in table                      |
| [show approve-summary](./approve-summary.md)             | Generate release approval summary                     |
| [show ci-results](./ci-results.md)                       | Display CI results                                    |
| [show ci-summary](./ci-summary.md)                       | Generate CI workflow summary for a module             |
| [show components](./components.md)                       | Display all components grouped by module              |
| [show dependency-ci-summary](./dependency-ci-summary.md) | Generate dependency CI check summary                  |
| [show deps-setup-summary](./deps-setup-summary.md)       | Generate dependencies setup summary                   |
| [show ghosts](./ghosts.md)                               | Display ghost files                                   |
| [show lint-summary](./lint-summary.md)                   | Generate lint summary for a module                    |
| [show release-summary](./release-summary.md)             | Generate release summary from layers JSON             |
| [show scan-summary](./scan-summary.md)                   | Generate security scan summary for a module           |
| [show test-results](./test-results.md)                   | Display test results with coverage                    |
| [show trigger-summary](./trigger-summary.md)             | Generate release trigger summary                      |
| [show units](./units.md)                                 | Display units of work for a framework                 |
| [show workspaces](./workspaces.md)                       | List all workspaces and their status                  |

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

## Best Practices

### Interactive Use

1. **Start with show modules** to understand repository structure
2. **Use show commands for exploration** rather than scripting
3. **Pipe large output to less** for easier navigation
4. **Use grep for filtering** when you know what you're looking for

### When to Use get Instead

Use [get commands](../get/index.md) when you need to:

- Process output with jq or other tools
- Cache results for repeated queries
- Integrate with CI/CD pipelines
- Parse data programmatically

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

- [Get Commands](../get/index.md) - JSON output for automation
- [Output Formats](../../overview/output-formats.md) - Understanding output types
- [Command Taxonomy](../../overview/command-taxonomy.md) - Command organization
- [Show Commands Guide](../../../../how-to-guides/eac/commands/getting-started/explore-your-repository.md) - How-to guide
