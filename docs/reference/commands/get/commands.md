# Get Commands Reference

<!-- book:cmd get commands -->

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
