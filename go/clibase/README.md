# clibase

Shared CLI infrastructure providing command framework, orchestration, output rendering,
and service initialization for all EAC CLI commands.

Module moniker: `clibase` | Dependencies: `core`

## Package Index

| Package                         | Purpose                                                                    |
| ------------------------------- | -------------------------------------------------------------------------- |
| [ansi](./ansi/)                 | ANSI escape sequence filtering for clean text output                       |
| [caching](./caching/)           | Incremental change detection and per-item content-addressable caching      |
| [capacity](./capacity/)         | Cross-process capacity management using weighted semaphores                |
| [cmdframework](./cmdframework/) | Shared framework for orchestrated CLI commands (build, test, scan, lint)   |
| [display](./display/)           | Interface definitions for TUI console implementations                      |
| [environment](./environment/)   | Runtime environment detection for CLI commands                             |
| [flags](./flags/)               | Composable flag system providing reusable flag sets for commands           |
| [initsummary](./initsummary/)   | Data structures and formatters for command initialization summaries        |
| [locktracker](./locktracker/)   | Lock tracking and visualization registry with observability                |
| [locking](./locking/)           | File-based distributed locking with lock tracking integration              |
| [orchestrator](./orchestrator/) | Parallel execution engine for build, test, lint, and scan commands         |
| [output](./output/)             | Console output formatting for non-TUI execution mode                       |
| [registry](./registry/)         | Command registration and dispatch for the CLI                              |
| [render](./render/)             | Multi-format output rendering (markdown tables, JSON, YAML, TOML)          |
| [services](./services/)         | Service initialization layer bridging CLI commands to core domain adapters |
| [template](./template/)         | Go template rendering utilities for generating files from templates        |
| [testrunners](./testrunners/)   | Registry-based test type dispatch for gotest, godog, tscucumber, mocha     |
| [testutil](./testutil/)         | Test utilities and helpers for the CLI test suite                          |

## Architecture Notes

The clibase module sits between the pure-domain `core` module and the command
implementations in `eac`. It provides the execution machinery: `cmdframework`
defines the command lifecycle, `orchestrator` handles parallel work unit dispatch,
and `services` wires up core adapters for command consumption. Output flows through
`render` and `display` for format-aware presentation, while `locking` and `capacity`
manage cross-process coordination for concurrent builds and tests.

Key dependency flows within the module:

- **Framework Layer**: `cmdframework`, `flags`, `registry` define the command structure
  and are consumed by all command implementations
- **Execution Layer**: `orchestrator` uses `capacity`, `locking`, and `locktracker`
  to manage parallel work with cross-process safety
- **Output Layer**: `render`, `display`, `output`, and `ansi` form the presentation
  pipeline from structured data to terminal output
- **Services Layer**: `services` initializes core adapters and is the primary bridge
  point between clibase and the core module
- **Services Layer**: `services` also routes external tool calls through the tool registry
  for consistent executor-mode behavior
