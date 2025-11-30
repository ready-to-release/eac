# Get Commands (Describe Commands)

**Problem**: Shell integration and MCP servers need programmatic access to command metadata for dynamic tool discovery and completion.

**Solution**: Use `get commands` to output structured JSON containing all command information, hierarchical relationships, and available module monikers.

## Key Benefits

- Dynamic tool discovery for MCP servers
- Shell completion script generation
- Command introspection and documentation
- Automated command registry validation
- Integration with AI assistants

## Command Overview

```bash
r2r eac get commands
```

**Purpose**: Output structured command information in JSON format for shell integration and programmatic access.

**Output**: JSON object with command metadata, hierarchical tree, and module monikers.

**No flags or arguments**: This command takes no parameters.

## Output Format

The command outputs a JSON object with three main sections:

### 1. Commands Array

Array of `CommandInfo` objects, each containing:

| Field         | Type     | Description                                              |
| ------------- | -------- | -------------------------------------------------------- |
| `name`        | string   | Full command name (e.g., "show modules")                 |
| `parts`       | string[] | Command parts split by space (e.g., ["show", "modules"]) |
| `description` | string   | Short description of the command                         |
| `parent`      | string   | Parent command name (empty for root commands)            |
| `is_leaf`     | boolean  | True if this is an executable command                    |
| `args`        | string   | Argument completion type (e.g., "modules", "files")      |

### 2. Tree Map

Object mapping parent commands to their child subcommands:

- Key: Parent command name (empty string for root)
- Value: Array of child command names

### 3. Modules Array

Array of available module monikers for completion (sorted alphabetically).

## Example Output

```json
{
  "commands": [
    {
      "name": "show modules",
      "parts": ["show", "modules"],
      "description": "Display all module contracts in a human-readable table",
      "parent": "show",
      "is_leaf": true,
      "args": ""
    },
    {
      "name": "build",
      "parts": ["build"],
      "description": "Build one or more modules by moniker",
      "parent": "",
      "is_leaf": true,
      "args": "modules"
    },
    {
      "name": "work create",
      "parts": ["work", "create"],
      "description": "Create a new workspace for parallel development",
      "parent": "work",
      "is_leaf": true,
      "args": ""
    }
  ],
  "tree": {
    "": ["build", "test", "show", "get-modules", "work", "specs", "help"],
    "show": ["modules", "dependencies", "files", "tests"],
    "work": ["create", "list", "remove", "commit", "merge", "pull", "pr"],
    "specs": ["create", "validate"]
  },
  "modules": [
    "contracts",
    "docs",
    "scripts",
    "specs",
    "src-cli",
    "src-commands",
    "src-core",
    "src-mcp"
  ]
}
```

## Integration Points

### 1. MCP Server Integration

The MCP server (`src/mcp/commands/main.go`) uses `get commands` for dynamic tool discovery:

```go
// getCommandTools discovers commands by calling "describe commands"
func getCommandTools() []Tool {
    tree := describeCommands()
    var tools []Tool

    for _, cmd := range tree.Commands {
        // Convert command name to kebab-case for tool name
        toolName := strings.ReplaceAll(cmd.Name, " ", "-")

        tools = append(tools, Tool{
            Name:        toolName,
            Description: cmd.Description,
            // ... additional fields
        })
    }

    return tools
}
```

**Benefits:**

- Commands are automatically exposed as MCP tools
- No manual tool registration required
- Changes to commands automatically reflected in MCP
- AI assistants can discover all available commands

### 2. Shell Completion Scripts

PowerShell completion uses `get commands` for dynamic command discovery:

```powershell
# From scripts/pwsh/go-invoker/go.psm1
function Get-CommandStructure {
    # Calls 'go run . get commands' to get JSON structure
    Push-Location $commandsPath
    try {
        $jsonOutput = & go run . get commands 2>$null
        if ($LASTEXITCODE -eq 0) {
            # Store in environment variable for session persistence
            $env:SRC_COMMANDS_DESCRIBE = $jsonOutput
            return $jsonOutput | ConvertFrom-Json
        }
    }
    finally {
        Pop-Location
    }
}
```

**Benefits:**

- Tab completion for all commands
- Subcommand discovery
- Module name completion
- Session-level caching for performance

### 3. Command Registry Validation

The output can be used to validate command registration:

```bash
# Check for commands without descriptions
r2r eac get commands | jq '.commands[] | select(.description == "") | .name'

# Count total commands
r2r eac get commands | jq '.commands | length'

# Find root commands
r2r eac get commands | jq '.commands[] | select(.parent == "") | .name'

# List all parent commands
r2r eac get commands | jq '.tree | keys[]'
```

## Use Cases

### 1. Generate Command Documentation

```bash
# Extract all commands with descriptions
r2r eac get commands | jq -r '.commands[] | "\(.name): \(.description)"'

# Output:
# build: Build one or more modules by moniker
# show modules: Display all module contracts in a human-readable table
# work create: Create a new workspace for parallel development
# ...
```

### 2. Build Completion Helpers

```bash
# Get all commands under 'show'
r2r eac get commands | jq -r '.tree["show"][]'

# Output:
# modules
# dependencies
# files
# tests
```

### 3. Module Completion

```bash
# Get all available modules for completion
r2r eac get commands | jq -r '.modules[]'

# Output:
# contracts
# docs
# scripts
# specs
# src-cli
# src-commands
# src-core
# src-mcp
```

### 4. Command Discovery

```bash
# Find all commands that work with modules
r2r eac get commands | jq '.commands[] | select(.args == "modules") | .name'

# Find leaf commands (executable)
r2r eac get commands | jq '.commands[] | select(.is_leaf == true) | .name'

# Find commands with specific parent
r2r eac get commands | jq '.commands[] | select(.parent == "work") | .name'
```

### 5. Hierarchical Command Tree

```bash
# Display full command hierarchy
r2r eac get commands | jq -r '
  .tree | to_entries[] |
  "Parent: \(.key)\n  Children: \(.value | join(", "))\n"
'

# Output:
# Parent:
#   Children: build, test, show, get-modules, work, specs, help
#
# Parent: show
#   Children: modules, dependencies, files, tests
#
# Parent: work
#   Children: create, list, remove, commit, merge, pull, pr
```

## Advanced Usage

### Command Analysis

```bash
# Count commands by parent
r2r eac get commands | jq '
  [.commands[] | .parent] |
  group_by(.) |
  map({parent: .[0], count: length}) |
  sort_by(.count) |
  reverse
'
```

### Generate Help Index

```bash
# Create markdown list of all commands
r2r eac get commands | jq -r '
  .commands[] |
  "- **\(.name)**: \(.description)"
' > command-index.md
```

### Validate Command Structure

```bash
# Check for orphaned parent references
r2r eac get commands | jq '
  .commands[] |
  select(.parent != "" and (.parent | in($tree) | not)) |
  .name
'

# Find multi-word root commands (potential issues)
r2r eac get commands | jq '.commands[] | select(.parent == "" and (.parts | length > 1)) | .name'
```

## Output Characteristics

### Performance

- **Fast**: Command metadata is pre-registered in memory
- **Lightweight**: JSON output is typically < 10KB
- **Cacheable**: Output is stable unless commands change
- **No side effects**: Read-only operation

### Stability

- **Deterministic**: Same commands always produce same output
- **Sorted**: Module list is alphabetically sorted
- **Versioned**: Output format is stable across versions

### Error Handling

The command returns:

- **Exit code 0**: Success, JSON written to stdout
- **Exit code 1**: Error (e.g., JSON encoding failure)

Errors are written to stderr:

```text
Error encoding JSON: <error message>
```

## Integration Best Practices

### Caching

Cache the output for performance:

```bash
# Shell script caching
if [ -z "$COMMAND_CACHE" ]; then
    export COMMAND_CACHE=$(r2r eac get commands)
fi

# Use cached data
echo "$COMMAND_CACHE" | jq '.modules[]'
```

### Error Handling

Always check exit codes:

```bash
if OUTPUT=$(r2r eac get commands 2>/dev/null); then
    echo "$OUTPUT" | jq '.commands | length'
else
    echo "Failed to get command information"
    exit 1
fi
```

### JSON Validation

Validate output before processing:

```bash
# Validate JSON structure
if ! r2r eac get commands | jq empty 2>/dev/null; then
    echo "Invalid JSON output"
    exit 1
fi
```

## Comparison with Related Commands

| Command        | Output                 | Use Case                            |
| -------------- | ---------------------- | ----------------------------------- |
| `get commands` | JSON metadata          | MCP servers, completion, automation |
| `help`         | Human-readable help    | Interactive user assistance         |
| `completion`   | Bash completion script | Shell completion installation       |

**When to use each:**

- **get commands**: Building tools, MCP integration, programmatic access
- **help**: Reading documentation, learning commands
- **completion**: Installing shell completion

## Troubleshooting

| Problem                 | Cause                     | Solution                               |
| ----------------------- | ------------------------- | -------------------------------------- |
| Empty commands array    | No commands registered    | Check command registration in `init()` |
| Missing module monikers | Repository root not found | Run from within repository             |
| Invalid JSON output     | Encoding error            | Check stderr for error message         |
| Outdated command list   | Stale cache               | Clear cache and regenerate             |

## Implementation Details

### Command Registration

Commands register themselves using structured comments:

```go
// Command: show modules
// Short: Display all module contracts in a human-readable table
package show

func init() {
    registry.Register(ShowModules)
}
```

The `get commands` implementation:

1. Reads all registered commands from `registry.GetCommandRegistry()`
2. Extracts command metadata from registration
3. Builds hierarchical tree structure
4. Loads module monikers from contracts
5. Outputs complete JSON structure

### Module Discovery

Module monikers are loaded from module contracts:

```go
func loadModuleMonikers() []string {
    workspaceRoot, _ := repository.GetRepositoryRoot("")
    moduleReport, _ := reports.GetModuleContracts(workspaceRoot)

    var monikers []string
    for _, module := range moduleReport.Registry.All() {
        monikers = append(monikers, module.Moniker)
    }

    sort.Strings(monikers) // Ensure consistent ordering
    return monikers
}
```

## Summary

**Command:** `get commands`

**Purpose:** Output structured command metadata for shell integration and MCP servers

**Output Format:** JSON with three sections:

1. `commands`: Array of command metadata
2. `tree`: Hierarchical parent-child mapping
3. `modules`: Available module monikers

**Primary Use Cases:**

- MCP server dynamic tool discovery
- Shell completion script generation
- Command registry validation
- Automated documentation generation

**Integration Points:**

- MCP server (`src/mcp/commands/main.go`)
- PowerShell completion (`scripts/pwsh/go-invoker/go.psm1`)
- Command validation scripts
- AI assistant tool discovery

**Best Practice:** Cache the output when used repeatedly to avoid redundant execution.

> **Note:** The command was renamed from `describe commands` to `get commands` as part of the verb-first command restructuring.
