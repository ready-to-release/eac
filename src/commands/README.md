# src/commands

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
run <TAB>              # Shows: commit-ai, describe, list, show
run show <TAB>         # Shows: files, modules, moduletypes
run show files <TAB>   # Shows: changed, staged
run show files         # Executes: show all tracked files (markdown table)
run show modules       # Executes: show modules (markdown table)
run commit-ai          # Executes: show staged changes by module (markdown)
run show               # Shows help for 'show' subcommands
```

**Bash (Unix/Linux/macOS):**

```bash
./run.sh <command> [subcommand] [args...]
```

### Direct Execution

Run directly from this directory:

```bash
cd src/commands
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

- **`commit-ai`** - Show staged changes with their module mappings for AI commit message generation
- **`describe commands`** - Output structured command information for shell integration
- **`list commands`** - List all available commands
- **`show files`** - Show all tracked repository files with module ownership
- **`show files staged`** - Show only staged files with module ownership
- **`show files changed`** - Show only modified/unstaged files with module ownership
- **`show modules`** - Show all module contracts in the repository (markdown table)
- **`show moduletypes`** - Show module types grouped by count (markdown table)

### Parent Commands (Implicit Help)

Running a parent command without a subcommand shows available subcommands:

```bash
go run . show          # Shows: files, modules, moduletypes
go run . show files    # Shows: changed, staged (plus executes show files)
go run . list          # Shows: commands
go run . describe      # Shows: commands
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

- Calls `go run . describe commands` to get command structure
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

- `run <TAB>` → shows root commands
- `run show <TAB>` → shows subcommands under 'show'
- `run list <TAB>` → shows subcommands under 'list'

Completion data comes from `go run . describe commands` which outputs:

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
    "show": ["files", "modules", "moduletypes"],
    "show files": ["changed", "staged"],
    "list": ["commands"],
    "describe": ["commands"]
  }
}
```

## Special Commands

### commit-ai

The `commit-ai` command is designed to help generate meaningful commit messages by analyzing staged changes grouped by module:

```bash
# Stage some changes
git add .

# View changes organized by module
run commit-ai
```

Output includes:
1. **Summary Table** - All staged files with line counts and module mappings
2. **Changes by Module** - Grouped view showing which modules are affected
3. **Unmapped Files** - Files not belonging to any module contract

This is particularly useful for multi-module mono-repositories where you want to understand the scope of your changes before committing.

## Output Formats

All commands output formatted markdown tables for human readability and machine parsing.

### Markdown Tables

All commands use the `render` package to output clean markdown tables:

**File commands:**
- `show files` - Table with File, Modules columns
- `show files staged` - Table with File, Modules columns
- `show files changed` - Table with File, Modules columns

**Module commands:**
- `show modules` - Table with Moniker, Type, Root Path columns
- `show moduletypes` - Table with Module Type, Count columns (includes footer with total)

**Special commands:**
- `commit-ai` - Multiple tables (summary, by-module breakdown, unmapped files)

### Piping and Parsing

Tables can still be piped and parsed in shells:

```powershell
# PowerShell examples
run show files | Select-String "src-commands"
run show modules | Select-String "go-library"

# Count lines (subtract header/separator)
(run show files | Measure-Object).Count - 2
```

```bash
# Bash examples
./run.sh show files | grep "src-commands"
./run.sh show modules | grep "go-library"
```

## Dependencies

Commands can import:

- `github.com/ready-to-release/eac/src/contracts/*` - Module contracts
- `github.com/ready-to-release/eac/src/repository/*` - Repository operations
- `github.com/ready-to-release/eac/src/commands/render` - Markdown table rendering
- Standard library packages
- Any other internal packages

## Module Contract

Defined in `contracts/modules/0.1.0/src-commands.yml`:

```yaml
moniker: "src-commands"
name: "Go command dispatcher with auto-discovery"
type: "go-library"
source:
  root: "src/commands"
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
