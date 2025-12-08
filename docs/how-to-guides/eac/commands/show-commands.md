# Show Commands

**Problem**: Understanding repository structure, modules, dependencies, and files requires manual exploration.

**Solution**: Use `show` commands to display repository information in human-readable tables for quick insights and interactive debugging.

## Key Benefits

- Quick repository insights with formatted tables
- Human-readable output optimized for terminal display
- Module discovery and navigation
- Dependency visualization
- File ownership tracking
- Interactive exploration and debugging

## Available Show Commands

| Command               | Description                     | Output Format |
| --------------------- | ------------------------------- | ------------- |
| `show modules`        | Module table with details       | Table         |
| `show dependencies`   | Dependency graph visualization  | Tree/Graph    |
| `show files`          | File ownership by module        | Table         |
| `show moduletypes`    | Module type distribution        | List          |
| `show tests`          | Test suites table               | Table         |
| `show environments`   | Environment configurations      | Table         |
| `show config`         | EAC configuration               | Table         |
| `show files-changed`  | Changed files with ownership    | Table         |
| `show files-staged`   | Staged files with ownership     | Table         |

> **For automation and scripting**: See [Get Commands Reference](../../../reference/commands/get-commands-reference.md) for JSON output and scripting examples.

## Common Use Cases

### show modules

Display module table with details.

```bash
r2r eac show modules

# Output:
# ┌─────────────────┬──────────────┬────────────┬─────────────┐
# │ Module          │ Type         │ Files      │ Dependencies│
# ├─────────────────┼──────────────┼────────────┼─────────────┤
# │ eac-commands    │ go-commands  │ 45         │ 2           │
# │ eac-core        │ go-library   │ 32         │ 0           │
# │ src-auth        │ go-library   │ 28         │ 1           │
# │ docs            │ mkdocs-site  │ 156        │ 0           │
# └─────────────────┴──────────────┴────────────┴─────────────┘
```

**Use this when:**
- Exploring what modules exist in the repository
- Understanding module characteristics (type, size, dependencies)
- Getting a quick overview of the codebase structure

### show dependencies

Display dependency graph.

```bash
r2r eac show dependencies

# Output:
# Module Dependencies:
#
# r2r-cli
#   └─→ eac-commands
#       └─→ eac-core
#
# src-auth
#   └─→ eac-core
#
# src-api
#   ├─→ eac-core
#   └─→ src-auth
```

**Use this when:**
- Understanding module relationships
- Identifying circular dependencies
- Planning refactoring or architectural changes
- Debugging build order issues

### show files

Display file ownership by module.

```bash
r2r eac show files

# Output:
# ┌────────────────────────────────────┬─────────────────┐
# │ File                               │ Module          │
# ├────────────────────────────────────┼─────────────────┤
# │ go/eac/commands/impl/work/work.go  │ eac-commands    │
# │ go/eac/core/repository/repo.go     │ eac-core        │
# │ src/auth/jwt/token.go              │ src-auth        │
# │ docs/index.md                      │ docs            │
# └────────────────────────────────────┴─────────────────┘
#
# Total: 2,690 files across 15 modules
```

**Use this when:**
- Finding which module owns a specific file
- Browsing file organization
- Verifying module boundaries

**For large repositories**: Consider using `show files-changed` or `show files-staged` to see only relevant files.

### show moduletypes

Display module type distribution.

```bash
r2r eac show moduletypes

# Output:
# Module Types:
#
# go-library       : 8 modules
# go-commands      : 3 modules
# go-cli           : 1 module
# mkdocs-site      : 1 module
# specifications   : 1 module
# contracts        : 1 module
#
# Total: 15 modules across 6 types
```

**Use this when:**
- Understanding repository composition
- Identifying module patterns
- Planning new modules (following existing patterns)

### show tests

Display test suites.

```bash
r2r eac show tests

# Output:
# ┌──────────────┬────────────┬────────────┬─────────────┐
# │ Suite        │ Module     │ Tests      │ Status      │
# ├──────────────┼────────────┼────────────┼─────────────┤
# │ integration  │ r2r-cli    │ 12         │ passing     │
# │ e2e          │ src-api    │ 8          │ passing     │
# │ smoke        │ src-auth   │ 3          │ passing     │
# └──────────────┴────────────┴────────────┴─────────────┘
```

**Use this when:**
- Checking test coverage across modules
- Verifying test suite status
- Finding which modules have tests

### show environments

Display environment configurations.

```bash
r2r eac show environments

# Output:
# ┌────────────┬─────────────────┬──────────────────────┐
# │ Environment│ Description     │ Variables            │
# ├────────────┼─────────────────┼──────────────────────┤
# │ dev        │ Development     │ DEBUG=true, PORT=3000│
# │ staging    │ Staging         │ DEBUG=false, PORT=80 │
# │ prod       │ Production      │ DEBUG=false, PORT=80 │
# └────────────┴─────────────────┴──────────────────────┘
```

**Use this when:**
- Reviewing available environments
- Checking environment configuration
- Verifying environment variables

### show config

Display EAC configuration with defaults applied in human-readable format.

```bash
r2r eac show config

# Output:
# EAC Configuration:
#
# ┌─────────────────────┬──────────────────────────────────────┐
# │ Config              │ Value                                │
# ├─────────────────────┼──────────────────────────────────────┤
# │ repository.root     │ /home/user/projects/eac              │
# │ repository.name     │ eac                                  │
# │ ai.provider         │ anthropic                            │
# │ ai.model            │ claude-sonnet-4                      │
# │ build.parallel      │ true                                 │
# │ test.timeout        │ 30m                                  │
# └─────────────────────┴──────────────────────────────────────┘
#
# Total: 6 configuration settings
```

**Use this when:**
- Reviewing EAC configuration
- Debugging configuration issues
- Verifying settings before running builds/tests

### show files-changed

Display changed files with module ownership.

```bash
r2r eac show files-changed

# Output:
# ┌────────────────────────────────────┬─────────────────┐
# │ Changed File                       │ Module          │
# ├────────────────────────────────────┼─────────────────┤
# │ go/eac/commands/impl/work/remove.go│ eac-commands    │
# │ go/eac/core/repository/repo.go     │ eac-core        │
# └────────────────────────────────────┴─────────────────┘
#
# Changed modules: eac-commands, eac-core
```

**Use this when:**
- Pre-commit review: seeing what you've changed
- Understanding impact of changes
- Identifying affected modules before building

**Tip**: This is much faster than `show files` for large repositories.

### show files-staged

Display staged files with module ownership.

```bash
r2r eac show files-staged

# Output:
# ┌────────────────────────────────────┬─────────────────┐
# │ Staged File                        │ Module          │
# ├────────────────────────────────────┼─────────────────┤
# │ src/auth/jwt/token.go              │ src-auth        │
# │ src/auth/jwt/token_test.go         │ src-auth        │
# └────────────────────────────────────┴─────────────────┘
```

**Use this when:**
- Reviewing staged changes before commit
- Verifying what will be committed
- Understanding module impact of staged changes

## Typical Workflows

### Module Discovery

```bash
# What modules exist?
r2r eac show modules

# What types are there?
r2r eac show moduletypes

# What depends on what?
r2r eac show dependencies
```

### Pre-commit Analysis

```bash
# What files changed?
r2r eac show files-changed

# What's staged?
r2r eac show files-staged

# Which modules are affected?
# (For scripting this, see get changed-modules in reference docs)
```

### File Ownership

```bash
# Who owns this file?
r2r eac show files | grep "auth/jwt/token.go"

# All files in src-auth module
# (For scripting this, see get files in reference docs)
```

### Understanding Dependencies

```bash
# See the full dependency graph
r2r eac show dependencies

# Identify potential circular dependencies
# (Look for warning messages in output)
```

## Best Practices

1. **Use show for interactive work**: These commands are optimized for human readability
2. **Use filters with large outputs**: Pipe through `grep` to find specific items
3. **Check files-changed before committing**: Understand the impact of your changes
4. **Review dependencies regularly**: Identify potential architectural issues early
5. **Combine with other commands**: Use show commands to inform your next action

## Summary

Show commands provide quick, human-readable insights into your repository:

- **`show modules`** - Overview of all modules
- **`show dependencies`** - Dependency relationships
- **`show files`** - File ownership (all files)
- **`show files-changed`** - File ownership (changed files only)
- **`show files-staged`** - File ownership (staged files only)
- **`show moduletypes`** - Module type distribution
- **`show tests`** - Test suite status
- **`show environments`** - Environment configurations
- **`show config`** - EAC configuration

**For automation and scripting**: Use `get` commands instead (see [Get Commands Reference](../../../reference/commands/get-commands-reference.md)).
