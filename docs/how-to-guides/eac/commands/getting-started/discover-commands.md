# Discovering Available Commands

{{ page_breadcrumb() }}

The EAC extension provides 147 commands organized by category. This guide shows you how to discover, explore, and use them effectively.

**What you'll learn:**

- How to list and explore available commands
- Understanding command categories and organization
- Common command patterns (`show` vs `get`)
- Most frequently used commands
- Where to find detailed command documentation

## Quick Command Discovery

### Get Help

List all available commands:

```bash
r2r eac help
```

Get help for a specific command:

```bash
r2r eac help <command>
```

Get structured command information:

```bash
r2r eac show valid-commands  # Table format
r2r eac get valid-commands   # JSON format
```

## Command Categories

EAC provides 147 commands organized into logical groups. Here are the main categories:

| Category | Commands | Description |
|----------|----------|-------------|
| **Viewing Repository** | `show`, `get` | Display repository information (modules, files, dependencies) |
| **Building & Testing** | `build`, `test` | Build modules and run test suites |
| **Development** | `create`, `update` | Generate specs, commit messages, designs, and PR descriptions |
| **Validation** | `validate` | Verify contracts, dependencies, specs, and release readiness |
| **Release Management** | `release` | Changelog, versioning, and release execution |
| **CI/CD Integration** | `pipeline` | Continuous integration orchestration and status |
| **Workspace Management** | `work` | Parallel development with git worktrees |
| **Documentation** | `serve`, `update docs` | Build and serve documentation sites |
| **Templates** | `templates` | Manage documentation and specification templates |
| **Security** | `scan` | Security scanning and compliance evidence |

## Common Command Patterns

### Show vs Get

Understanding the difference between `show` and `get` commands:

- **`show` commands** display human-readable output (tables, summaries)
- **`get` commands** return structured JSON data for scripting

**Examples:**

```bash
r2r eac show modules          # Pretty table for humans
r2r eac get modules           # JSON for scripts
```

This pattern is consistent across all information-retrieval commands, making it easy to switch between interactive use and automation.

## Most Frequently Used Commands

| Command | Purpose |
|---------|---------|
| `r2r eac show modules` | View all modules in your repository |
| `r2r eac show files-changed` | See which files changed and their module ownership |
| `r2r eac build <module>` | Build a specific module |
| `r2r eac test <module>` | Run tests for a module |
| `r2r eac validate` | Validate all repository contracts |
| `r2r eac create spec "description"` | Generate a Gherkin specification |
| `r2r eac release this` | Prepare a module for release |
| `r2r eac work create <name>` | Create a new development workspace |

## Finding More Information

### Command Reference

For detailed command documentation, see:

- **[Command Reference](../../../../reference/commands/)** - Complete alphabetical listing
- **[Getting Started Commands](../getting-started/)** - Beginner-friendly guides
- **[Development Workflow Commands](../development-workflow/)** - Daily development tasks
- **[Build, Test & Validate Commands](../build-test-validate/)** - Build and test execution
- **[Release Management Commands](../release-management/)** - Version management

### Getting Help for Specific Commands

Every command has built-in help:

```bash
r2r eac help <command>        # Show command help
r2r eac <command> --help      # Alternative syntax
```

Example:

```bash
r2r eac help show modules     # Learn about showing modules
r2r eac help test             # Learn about testing
```

## Related Guides

- **[Get Help with Commands](./get-help-with-commands.md)** - Master the help system
- **[Explore Your Repository](./explore-your-repository.md)** - Discover modules, files, and structure
- **[Setup AI Provider](./setup-ai-provider.md)** - Configure Claude, OpenAI, or Gemini

{{ diataxis_footer() }}
