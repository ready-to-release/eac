# cmd

Cobra command tree defining all CLI subcommands, argument parsing, and execution flows.

## Key Types

- **`RootCmd`** -- Base Cobra command with PersistentPreRunE for logging and argument parsing
- **`RunCmd`** -- Runs extension containers with I/O streaming and signal handling
- **`InteractiveCmd`** -- Starts shell-attached interactive container sessions
- **`InstallCmd`** -- Adds extensions to config and pulls their Docker images
- **`InitCmd`** -- Creates `.clie/clie.yml` configuration scaffolding
- **`Release`** -- GitHub release metadata for self-update operations

## Key Functions

- **`Execute`** -- Entry point called from main; handles `--version` shortcut and runs root command
- **`GetParsedCommand`** -- Singleton accessor for parsed command structure via `sync.Once`
- **`CreateExtensionAliases`** -- Dynamically registers Cobra commands for configured extensions
- **`addExtensionToConfig`** -- Resolves latest SHA tag and writes extension to YAML config

## Patterns

- Parsed command singleton: `sync.Once` ensures EBNF parsing happens exactly once per invocation
- Extension aliases: Configured extensions registered as top-level commands (e.g., `clie pwsh` instead of `clie run pwsh`)
- Argument boundary: `DisableFlagParsing` on run command allows pass-through to container
- Lazy config loading: Help functions temporarily suppress logging during config load attempts
- Signal handling: SIGINT/SIGTERM captured with graceful container shutdown and cleanup

## Internal Structure

| File                    | Responsibility                                                    |
| ----------------------- | ----------------------------------------------------------------- |
| root.go                 | Root Cobra command, PersistentPreRunE, log-level setup, Execute() |
| run.go                  | Run command: container create, attach, I/O stream, wait, cleanup  |
| interactive.go          | Interactive command: shell-attached mode with docker attach/exec  |
| install.go              | Install command: add to config, resolve SHA, pull images          |
| init.go                 | Init command: create `.clie/clie.yml`, `--delete-configs` flag    |
| version.go              | Version command: build info from `debug.ReadBuildInfo` + ldflags  |
| validate.go             | Validate command: JSON Schema validation via embedded contract    |
| verify.go               | Verify command: Docker, GitHub auth, config prerequisite checks   |
| update.go               | Update self command: download latest release from GitHub API      |
| cleanup.go              | Cleanup command: remove old images, stopped containers, prune     |
| list.go                 | List command: discover extensions from GHCR with cache            |
| metadata.go             | Metadata command: retrieve extension metadata via container exec  |
| parsed_command.go       | Global parsed command singleton with accessor functions           |
| extension_aliases.go    | Dynamic Cobra command registration for configured extensions      |
| test_parsed_command.go  | Integration tests for parsed command (build-tagged `L1`)          |

## Dependencies

- `internal/cache` -- Registry cache for list command
- `internal/conf` -- Configuration loading and validation
- `internal/command-parser` -- EBNF-based argument parsing
- `internal/docker` -- Container lifecycle operations
- `internal/extensions` -- Extension image installation
- `internal/github` -- Registry client for install/list
- `internal/logging` -- Leveled logging
- `internal/validator` -- JSON Schema validation for validate command
- `internal/version` -- Version metadata for version command
- `internal/envconsts` -- Environment variable constants

## Role in System

The cmd package defines the entire user-facing command surface of the clie CLI. Each file corresponds to a distinct subcommand. The run command is the primary execution path, handling container creation, I/O multiplexing via Docker's stdcopy, signal propagation, and exit code forwarding.

## Code Health

### Tech Debt

- `run.go:269` -- Accesses `parsedCmd.ArgumentBoundary` without nil-guard; `parsedCmd` can be nil from the `nolint:errcheck` on line 251.
- `install.go:316-329` -- SHA extraction parses warning messages with fragile string splitting (`strings.Split(msg, "'")`) rather than using a structured return value from `ValidatePinnedExtensions`.
- `test_parsed_command.go` -- File uses build tag `L1` but is not a `_test.go` file by convention; its test functions are included in the production build unless tag-filtered.

### Pain Points

- `install.go:182-409` -- `addExtensionToConfig` is 227 lines handling config creation, registry lookup, SHA resolution, and YAML write in a single function.
- `run.go` -- Signal handler goroutine (line 394-429) calls `os.Exit(130)` after a brief sleep, potentially racing with cleanup operations.

### Optimization Opportunities

- The `install.go` SHA resolution reuses `ValidatePinnedExtensions` (designed for CI validation) to extract SHA tags from formatted warning messages. A dedicated `GetLatestSHA(extensionName)` function would be cleaner and more direct.
