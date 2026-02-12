# Command Dispatcher

> Migrated from `go/cli/eac/README.md`.

Go-based command dispatcher with auto-discovery and intelligent shell completion.

## Purpose

This module provides a command dispatcher that automatically discovers and routes commands. Commands are organized hierarchically (e.g., `show files`, `list commands`) with intelligent tab completion and automatic help text generation.

## Usage

### Quick Start (Recommended)

**PowerShell (Windows):**

```powershell
# One-time setup per session
.\importer.ps1

# Use commands with tab completion
run <TAB>              # Shows: commit, describe, list, show
run show <TAB>         # Shows: files, modules, component-kinds
run show files <TAB>   # Shows: changed, staged
run show files         # Executes: show all tracked files (markdown table)
run show modules       # Executes: show modules (markdown table)
run commit             # Executes: show staged changes by module (markdown)
run show               # Shows help for 'show' subcommands
```

**Bash (Linux/macOS):**

```bash
./run.sh <command> [subcommand] [args...]
```

### Direct Execution

Run directly from this directory:

```bash
cd go/cli/eac
go run . <command> [subcommand] [args...]
go run . list commands          # List all available commands
go run . show files             # Show repository files
go run . show                   # Show help for 'show' subcommands
```

## Available Commands

Commands are automatically discovered and registered. To see all available commands:

```bash
go run . list commands
```

Current commands include:

- **`commit`** - Show staged changes with their module mappings for AI commit message generation
- **`get changelog`** - Get changelog data in structured format (YAML/JSON/TOML)
- **`get commands`** - Output structured command information for shell integration
- **`get release-notes`** - Get release notes data in structured format (YAML/JSON/TOML)
- **`show changelog`** - Display changelog entries in human-readable markdown format
- **`show help`** - Show all available commands
- **`show files`** - Show all tracked repository files with module ownership
- **`show files-staged`** - Show only staged files with module ownership
- **`show files-changed`** - Show only modified/unstaged files with module ownership
- **`show modules`** - Show all module contracts in the repository (markdown table)
- **`show component-kinds`** - Show module types grouped by count (markdown table)
- **`show release-notes`** - Display release notes in human-readable markdown format

### Parent Commands (Implicit Help)

Running a parent command without a subcommand shows available subcommands:

```bash
go run . show          # Shows: files, modules, component-kinds
go run . show files    # Shows: changed, staged (plus executes show files)
go run . get           # Shows: commands, modules, etc.
```

## Architecture

### Command Discovery

Commands auto-register themselves using an `init()` function:

```go
func init() {
    Register("command name", FunctionName)
}
```

The dispatcher:

1. Loads all `.go` files in the package
2. Each file's `init()` registers its command
3. Commands are automatically available for routing and completion

### Command Routing

The main dispatcher (`main.go`):

- Tries longest match first for nested commands
- Falls back to parent command help if no exact match
- Provides structured error messages

### Shell Integration

**PowerShell** (`scripts/pwsh/go-invoker/go.psm1`):

- Calls `go run . get commands` to get command structure
- Provides intelligent tab completion for all command levels
- Caches command structure for performance

## Creating New Commands

### 1. Create a New Command File

```go
// Command: show files-new
// Description: Show new file report
package main

import (
    "fmt"
)

func init() {
    Register("show files-new", ShowFilesNew)
}

func ShowFilesNew() int {
    fmt.Println("=== New File Report ===")
    // Your implementation here
    return 0
}
```

### 2. That's It

The command is automatically:

- Discovered and registered
- Available via `go run . show files-new`
- Included in tab completion
- Listed in `list commands` output
- Shown in parent command help (`go run . show`)

### Command Naming Conventions

- Use **spaces** for multi-word commands: `"show files"` not `"show-files"`
- Use hierarchical naming: `"show files"`, `"show modules"`, `"list commands"`
- Parent prefixes (like `show`, `list`) automatically provide help text
- Keep names descriptive and consistent

### Function Signature

All command functions must match:

```go
type CommandFunc func() int
```

Return `0` for success, non-zero for errors.

## PowerShell Integration Details

### Setup

```powershell
# Import module
Import-Module .\scripts\pwsh\go-invoker\go.psm1 -Force

# Create 'run' alias
New-RunAlias
```

### Tab Completion

The module provides intelligent completion:

- `run <TAB>` -> shows root commands
- `run show <TAB>` -> shows subcommands under 'show'
- `run list <TAB>` -> shows subcommands under 'list'

Completion data comes from `go run . get commands` which outputs:

```json
{
  "commands": [
    {
      "name": "show files",
      "parts": ["show", "files"],
      "description": "Show repository files with module ownership",
      "parent": "show",
      "is_leaf": true
    }
  ],
  "tree": {
    "": ["describe", "list", "show"],
    "show": ["files", "modules", "component-kinds"],
    "show files": ["changed", "staged"],
    "list": ["commands"],
    "describe": ["commands"]
  }
}
```

## Dependencies

Commands can import:

- `github.com/ready-to-release/eac/go/core/contracts/*` - Module contracts
- `github.com/ready-to-release/eac/go/core/repository/*` - Repository operations
- `github.com/ready-to-release/eac/go/clibase/render` - Markdown table rendering
- Standard library packages
- Any other internal packages

## Module Contract

Defined in `contracts/modules/0.1.0/eac-commands.yml`:

```yaml
moniker: "eac-commands"
name: "Go command dispatcher with auto-discovery"
type: "go"
source:
  root: "go/cli/eac"
  includes:
    - "go.sum"
    - "go.mod"
    - "**.go"
    - "*.go"
```

## Design Philosophy

- **Self-registering**: Commands register themselves, no central registry
- **Auto-discovery**: New commands work immediately without configuration
- **Hierarchical**: Support nested command structures naturally
- **Helpful**: Automatic help text for parent commands
- **Shell-friendly**: Intelligent tab completion for PowerShell and Bash
