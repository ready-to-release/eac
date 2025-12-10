# Get Commands Reference

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac get commands`

**Purpose**: Output structured JSON metadata for all registered commands, including hierarchical relationships and available module monikers.

**Use Cases**:

- Dynamic tool discovery for MCP servers
- Shell completion script generation
- Command introspection and automated validation
- AI assistant integration

## Command Syntax

```bash
r2r eac get commands
```

**Parameters**: None (this command takes no flags or arguments)

**Output**: JSON to stdout

**Exit Codes**:

- `0`: Success
- `1`: Error (e.g., JSON encoding failure)

## Output Schema

The command outputs a JSON object with three top-level fields:

```typescript
{
  commands: CommandInfo[],
  tree: Record<string, string[]>,
  modules: string[]
}
```

### CommandInfo Schema

Each command in the `commands` array has the following structure:

```typescript
interface CommandInfo {
  name: string;        // Full command name (e.g., "show modules")
  parts: string[];     // Command parts split by space
  description: string; // Short description
  parent: string;      // Parent command name (empty for root)
  is_leaf: boolean;    // True if executable command
  args: string;        // Argument completion type
}
```

### Tree Schema

Maps parent commands to their child subcommands:

```typescript
type Tree = Record<string, string[]>;
// Key: Parent command name (empty string "" for root commands)
// Value: Array of child command names
```

### Modules Schema

Array of available module monikers (sorted alphabetically):

```typescript
type Modules = string[];
```

## Field Definitions

### CommandInfo Fields

| Field         | Type     | Description                                              | Example              |
| ------------- | -------- | -------------------------------------------------------- | -------------------- |
| `name`        | string   | Full command name                                        | `"show modules"`     |
| `parts`       | string[] | Command parts split by space                             | `["show", "modules"]`|
| `description` | string   | Short description of the command                         | `"Display all module contracts in a human-readable table"` |
| `parent`      | string   | Parent command name (empty string for root commands)     | `"show"`             |
| `is_leaf`     | boolean  | True if this is an executable command (not just a group) | `true`               |
| `args`        | string   | Argument completion type hint                            | `"modules"`, `"files"`, `""` |

### Argument Completion Types

The `args` field indicates what type of completion should be offered:

| Value      | Meaning                               |
| ---------- | ------------------------------------- |
| `""`       | No arguments or no completion needed  |
| `"modules"`| Expects module monikers               |
| `"files"`  | Expects file paths                    |

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
    "r2r-cli",
    "eac-commands",
    "eac-core",
    "src-mcp"
  ]
}
```

## Usage Examples

### Basic Queries

```bash
# Get all commands
r2r eac get commands | jq '.commands'

# Count total commands
r2r eac get commands | jq '.commands | length'

# List all root commands
r2r eac get commands | jq '.tree[""]'

# Get all available modules
r2r eac get commands | jq '.modules'
```

### Filtering Commands

```bash
# Find commands without descriptions
r2r eac get commands | jq '.commands[] | select(.description == "") | .name'

# Find root commands
r2r eac get commands | jq '.commands[] | select(.parent == "") | .name'

# Find executable commands only
r2r eac get commands | jq '.commands[] | select(.is_leaf == true) | .name'

# Find commands that accept module arguments
r2r eac get commands | jq '.commands[] | select(.args == "modules") | .name'

# Find commands under specific parent
r2r eac get commands | jq '.commands[] | select(.parent == "work") | .name'
```

### Tree Navigation

```bash
# List all parent commands
r2r eac get commands | jq '.tree | keys[]'

# Get subcommands of 'show'
r2r eac get commands | jq -r '.tree["show"][]'

# Display full hierarchy
r2r eac get commands | jq -r '
  .tree | to_entries[] |
  "Parent: \(.key)\n  Children: \(.value | join(", "))\n"
'
```

### Documentation Generation

```bash
# Generate command list with descriptions
r2r eac get commands | jq -r '.commands[] | "- **\(.name)**: \(.description)"'

# Create markdown command index
r2r eac get commands | jq -r '
  .commands[] |
  "- **\(.name)**: \(.description)"
' > command-index.md

# Generate completion data
r2r eac get commands | jq -r '.commands[] | "\(.name)\t\(.description)"'
```

### Analysis Queries

```bash
# Count commands by parent
r2r eac get commands | jq '
  [.commands[] | .parent] |
  group_by(.) |
  map({parent: .[0], count: length}) |
  sort_by(.count) |
  reverse
'

# Find multi-word root commands
r2r eac get commands | jq '
  .commands[] |
  select(.parent == "" and (.parts | length > 1)) |
  .name
'

# Check for orphaned parent references
r2r eac get commands | jq '
  .commands[] |
  select(.parent != "" and (.parent | in($tree) | not)) |
  .name
'
```

### Integration Patterns

```bash
# Cache for repeated use
export COMMAND_CACHE=$(r2r eac get commands)
echo "$COMMAND_CACHE" | jq '.modules[]'

# Error handling
if OUTPUT=$(r2r eac get commands 2>/dev/null); then
    echo "$OUTPUT" | jq '.commands | length'
else
    echo "Failed to get command information" >&2
    exit 1
fi

# JSON validation
if ! r2r eac get commands | jq empty 2>/dev/null; then
    echo "Invalid JSON output" >&2
    exit 1
fi
```

## Output Characteristics

### Performance

- **Fast**: Command metadata is pre-registered in memory
- **Lightweight**: JSON output typically < 10KB
- **Cacheable**: Output is stable unless commands change
- **No side effects**: Read-only operation

### Stability

- **Deterministic**: Same commands always produce same output
- **Sorted**: Module list is alphabetically sorted
- **Versioned**: Output format is stable across versions
- **Consistent**: Field names and types do not change

### Data Guarantees

- All `name` fields are unique across commands
- All `parent` values (except root "") appear as keys in `tree`
- All `modules` entries are valid module monikers
- `parts` array always equals `name.split(" ")`
- Empty parent (`""`) indicates root-level command

## Error Handling

### Success Case

```bash
$ r2r eac get commands
{"commands":[...],"tree":{...},"modules":[...]}
$ echo $?
0
```

### Error Case

```bash
$ r2r eac get commands
Error encoding JSON: <error message>
$ echo $?
1
```

Errors are written to **stderr**, JSON output to **stdout**.

### Common Issues

| Problem                 | Cause                     | Solution                               |
| ----------------------- | ------------------------- | -------------------------------------- |
| Empty commands array    | No commands registered    | Check command registration in `init()` |
| Missing module monikers | Repository root not found | Run from within repository             |
| Invalid JSON output     | Encoding error            | Check stderr for error message         |
| Outdated command list   | Stale cache               | Clear cache and regenerate             |

## Implementation Notes

### Command Registration

Commands register themselves during package initialization:

```go
// Command: show modules
// Short: Display all module contracts in a human-readable table
package show

func init() {
    registry.Register(ShowModules)
}
```

### Output Generation

The `get commands` implementation:

1. Reads registered commands from `registry.GetCommandRegistry()`
2. Extracts metadata from command registration
3. Builds hierarchical tree structure from parent relationships
4. Loads module monikers from module contracts
5. Encodes complete structure as JSON

### Module Discovery

Module monikers are discovered from the module contract registry:

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

### Integration Points

**MCP Server** (`go/eac/mcp/commands/main.go`):

- Calls `get commands` for dynamic tool discovery
- Converts command names to kebab-case tool names
- Exposes all commands as MCP tools automatically

**Shell Completion** (`scripts/pwsh/go-invoker/go.psm1`):

- Calls `get commands` for tab completion
- Caches output in `$env:SRC_COMMANDS_DESCRIBE`
- Uses tree structure for subcommand completion

**Command Validation**:

- CI scripts validate command registration
- Tests check for missing descriptions
- Documentation generation uses command metadata

---

> **Note**: The command was renamed from `describe commands` to `get commands` as part of the verb-first command restructuring.

{{ diataxis_footer() }}
