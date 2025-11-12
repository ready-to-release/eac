# CLAUDE_TEMP.md - MCP Commands Integration Guide

## Session Initialization with MCP Commands

**IMPORTANT**: This temporary configuration demonstrates how to leverage the MCP (Model Context Protocol) command tools available in this project.

## Available MCP Command Tools

This project provides a comprehensive set of MCP commands for managing a modular monorepo architecture. You MUST use these tools when performing operations related to modules, architecture, dependencies, and documentation.

### Module Discovery & Information

- `mcp__commands__get-modules` - Get all module contracts in the repository
- `mcp__commands__show-modules` - Show all module contracts with detailed information
- `mcp__commands__show-moduletypes` - Show all module types grouped by count
- `mcp__commands__get-files` - Get repository files with their module ownership
- `mcp__commands__show-files` - Show repository files with their module ownership
- `mcp__commands__show-files-changed` - Show changed (modified, unstaged) files with their module ownership
- `mcp__commands__show-files-staged` - Show staged files with their module ownership

### Dependency Management

- `mcp__commands__get-dependencies` - Get module dependency graph in structured format
- `mcp__commands__show-dependencies` - Show module dependency graph in a human-readable table format
- `mcp__commands__validate-dependencies` - Validate module dependencies from go.mod files against contracts
- `mcp__commands__get-execution-order` - Get execution order for specific modules based on dependencies
- `mcp__commands__get-changed-modules` - Get modules affected by changed files

### Architecture Documentation (Structurizr)

- `mcp__commands__design` - Architecture documentation tools using Structurizr (main command)
- `mcp__commands__design-list` - List available modules with architecture documentation
- `mcp__commands__design-new` - Create a new architecture workspace for a module
- `mcp__commands__design-validate` - Validate workspace files using Structurizr CLI
- `mcp__commands__design-export` - Export workspace DSL content
- `mcp__commands__design-serve` - Start or stop Structurizr server for a module
- `mcp__commands__design-add-container` - Add a container to a module's architecture
- `mcp__commands__design-add-relationship` - Add a relationship between containers in a module's architecture

### Project Documentation (MkDocs)

- `mcp__commands__docs` - Project documentation tools using MkDocs (main command)
- `mcp__commands__docs-serve` - Start or stop MkDocs server

### Build & Test Operations

- `mcp__commands__build-module` - Build a module by its moniker using type-based dispatch
- `mcp__commands__build-modules` - Build multiple modules in sequence and collect results in a build run directory
- `mcp__commands__test-module` - Test a module by its moniker using type-based dispatch
- `mcp__commands__test-modules` - Test multiple modules in sequence and collect results in a test run directory
- `mcp__commands__pipeline-run` - Execute module pipelines respecting dependencies

### Templates

- `mcp__commands__templates-list` - List all placeholder variables found in template files
- `mcp__commands__templates-install` - Install templates from a Git repository with value replacements

### Git Operations

- `mcp__commands__commit-ai` - Generate commit message using AI with staged changes and module mappings

## Usage Guidelines

### When to Use MCP Commands

**ALWAYS use MCP commands when**:
1. Exploring or querying module information
2. Checking module dependencies or execution order
3. Working with architecture documentation (Structurizr)
4. Building or testing modules
5. Validating module contracts or dependencies
6. Generating commit messages for module changes
7. Managing project documentation

**Example Workflow**:
```
User: "Show me all the modules in the project"
Claude: [Uses mcp__commands__show-modules]

User: "What modules depend on the core module?"
Claude: [Uses mcp__commands__show-dependencies]

User: "Build the authentication module"
Claude: [Uses mcp__commands__build-module with args: "authentication"]

User: "What files changed and which modules are affected?"
Claude: [Uses mcp__commands__show-files-changed and mcp__commands__get-changed-modules]
```

### Command Arguments

Many MCP commands accept arguments via the `args` parameter. Pass arguments as strings:

```javascript
// Example: Building a specific module
mcp__commands__build-module({ args: "my-module-moniker" })

// Example: Filtering dependencies
mcp__commands__show-dependencies({ args: "--format json" })
```

### Integration with Existing Workflow

1. **Before making changes**: Use `mcp__commands__get-files` and `mcp__commands__show-dependencies` to understand the module structure
2. **During development**: Use `mcp__commands__build-module` and `mcp__commands__test-module` to validate changes
3. **Before committing**: Use `mcp__commands__show-files-staged` and `mcp__commands__commit-ai` to create proper commit messages

### Architecture Documentation Workflow

When working with module architecture:

1. List available modules: `mcp__commands__design-list`
2. Validate existing documentation: `mcp__commands__design-validate`
3. Create new workspace: `mcp__commands__design-new`
4. Add containers: `mcp__commands__design-add-container`
5. Define relationships: `mcp__commands__design-add-relationship`
6. Preview: `mcp__commands__design-serve`

### Advantages of Using MCP Commands

1. **Consistent**: Commands understand the project's module contract system
2. **Type-aware**: Dispatches to appropriate handlers based on module types
3. **Validated**: Enforces module contracts and dependencies
4. **Integrated**: Works with git, build systems, and documentation tools
5. **Efficient**: Purpose-built for this modular monorepo architecture

## Best Practices

1. **Prefer MCP commands over manual file operations** when working with module metadata
2. **Chain commands** to build comprehensive understanding (e.g., get-modules → get-dependencies → build-modules)
3. **Validate before modifying** using validation commands
4. **Use show-* commands** for human-readable output, **use get-* commands** for structured data
5. **Check changed modules** before running builds or tests to scope work appropriately

## Example Session Flow

```
# Understand the project structure
Claude: [Uses mcp__commands__show-modules]
Claude: [Uses mcp__commands__show-moduletypes]

# Check current changes
Claude: [Uses mcp__commands__show-files-changed]
Claude: [Uses mcp__commands__get-changed-modules]

# Validate dependencies
Claude: [Uses mcp__commands__validate-dependencies]

# Build affected modules in correct order
Claude: [Uses mcp__commands__get-execution-order with changed modules]
Claude: [Uses mcp__commands__build-modules with ordered modules]

# Run tests
Claude: [Uses mcp__commands__test-modules]

# Generate commit message
Claude: [Uses mcp__commands__commit-ai]
```

---

**Note**: This is a temporary configuration file for demonstration purposes. The main project instructions remain in `CLAUDE.md`.
