# eac

Entry point and command dispatcher for the `eac` CLI. Bootstraps dependency
providers, resolves the command path from `os.Args`, and dispatches to the
matching command implementation.

## Command Registration Flow

EAC uses an **init-based registration pattern** where command packages
self-register at startup through import side-effects:

1. Each command package (under `go/commands/`) has a `register.go` with an
   `init()` that calls `registry.RegisterProvider(Commands)`.
2. Import files in this package (`imports_build.go`, `imports_lint.go`, etc.)
   use blank imports to trigger those `init()` functions.
3. `allcmds.BuildRegistry()` collects all registered providers and populates
   a `CommandRegistry`.
4. `main()` calls `dispatch()`, which resolves the longest-matching command
   name from `os.Args` and executes it.

Build tags control conditional inclusion (e.g. `//go:build !lite`), allowing
minimal CLI variants.

## Directory Layout

```text
go/cli/eac/                     # This package — CLI entry point
├── main.go                     # Dispatch, help printing, panic recovery
├── register.go                 # Calls allcmds.BuildRegistry()
├── allcmds/                    # BuildRegistry() — collects all providers
├── imports_build.go            # Blank-import go/commands/build
├── imports_lint.go             # Blank-import go/commands/lint
├── imports_repository.go       # Blank-import go/commands/repository
├── imports_scan.go             # Blank-import go/commands/scan
├── imports_test_command.go     # Blank-import go/commands/test
├── imports_update.go           # Blank-import go/commands/update
├── imports_deploy.go           # Blank-import go/commands/deploy
├── imports_ai.go               # Blank-import AI command packages
├── imports_adapters_*.go       # Blank-import language adapters
├── container_init.go           # Wires Docker adapter via init()
├── tui_init.go                 # Registers TUI factories via init()
└── custom_renderers.go         # Blank-imports custom render extensions

go/commands/                    # Command implementations
├── base/                       # Shared command infrastructure
├── build/                      # Build commands
│   ├── register.go             # init() → registry.RegisterProvider(Commands)
│   ├── build.go                # Commands() + build command logic
│   └── framework.go            # cmdframework integration
├── lint/                       # Lint commands
├── scan/                       # Security scan commands
├── test/                       # Test commands
├── deploy/                     # Deploy commands
├── update/                     # Update commands
└── repository/                 # Repository commands (get, show, create, ...)
```

## Command Interfaces

Commands implement interfaces from `contracts/core/0.1.0/command.go`:

- **`CommandPort`** — metadata only (`Name()`, `Metadata()`); used for
  parent/group commands that have subcommands but no direct action.
- **`SimpleCommandPort`** — extends `CommandPort` with
  `Execute(ctx, *CommandRequest) int`; used for most leaf commands.

Each command package exposes a `Commands() []core.CommandPort` function
returning all commands in the group. The `register.go` file passes this
to `registry.RegisterProvider` inside `init()`.

## Implementing a New Command

1. Create a package under `go/commands/<group>/`.
2. Implement `CommandPort` or `SimpleCommandPort`.
3. Add a `Commands()` function returning all commands in the package.
4. Add `register.go` with `func init() { registry.RegisterProvider(Commands) }`.
5. Add a blank-import file in `go/cli/eac/` (e.g. `imports_<group>.go`).

Command names use spaces for nesting (e.g. `"show files changed"`). The
dispatcher resolves the longest matching prefix automatically.

## Patterns

- **Init-based wiring**: `container_init.go`, `tui_init.go`, and `imports_*.go`
  use blank imports and `init()` to register adapters before `main` runs.
- **Longest-match routing**: `resolveCommand()` walks `os.Args` from longest
  to shortest to resolve nested commands like `show files changed`.
- **Panic recovery**: top-level `defer/recover` in `main()` prints full stack
  traces with platform-correct line endings.
- **Provider bootstrap**: `bootstrapProviders()` wires the git remote provider,
  tool system, and report dependencies before any command runs.

## Dependencies

- `clibase/registry` — command registration and lookup
- `clibase/executor` — command execution
- `core/config` — repository configuration loading
- `core/git` — go-git based remote URL resolution
- `core/domain/reports` — report generation with GitHub data
- `core/environments` — environment variable constants
- `core/tool` — tool system initialization
- `adapters/docker` — container provider wired via `init()`
- `adapters/tui/parallel` — TUI console for build/test/lint/scan
- `adapters/tui/selector` — TUI selector for interactive commands
- `adapters/gh` — GitHub CLI adapter for report commands

## Role in System

This package is the `main` binary of the `eac` module. It contains no
business logic itself; it wires adapters and delegates to command
implementations under `go/commands/`. All command packages self-register via
blank imports in `imports_*.go`, keeping the dispatcher decoupled from
individual commands.

## Code Health

### Pain Points

- Test file `main_test.go` is gated behind `L1 && ov` build tags, making it easy to skip inadvertently.
