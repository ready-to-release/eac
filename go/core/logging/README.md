# logging

Unified structured logging system with dual-sink output (console + rolling file),
TUI integration, component-scoped loggers, and YAML-driven configuration.

## Key Types

- **`Logger`** -- Zap-backed logger with dual-output and file handle lifecycle
- **`ComponentLogger`** -- Per-package logger with auto-inferred component names
- **`Config`** -- Logger configuration (command, workspace, debug, file flags)
- **`LoggingConfig`** -- YAML-driven sink and target configuration
- **`SinkConfig`** -- Per-sink levels, formatter, and rolling settings
- **`TargetConfig`** -- Command-specific extra log file targets
- **`FormatterType`** -- Output format enum (raw, timestamped, JSON)
- **`ExecutionContext`** -- Runtime context detection (CLI vs Docker)

## Patterns

- Global singleton: `Initialize`/`Get`/`L` manage a process-wide logger instance
- Component inference: `C()` uses `runtime.Caller` to derive component name from call site
- Config layering: Contract defaults merged with `.eac/logging.yml` user overrides
- Tee cores: Console, file, target, and TUI cores composed via `zapcore.NewTee`

## Internal Structure

| File | Responsibility |
| --- | --- |
| logger.go | `Logger` type, `New`, global singleton (`Initialize`/`Get`/`L`) |
| component.go | `ComponentLogger`, `C()` factory, call-site inference |
| config.go | `Config` struct and builder methods |
| configure.go | `ConfigureLogging` unified entry point with TUI support |
| core.go | Zap core builders (console, file, target) |
| debug.go | Atomic debug flag, `DebugDirect`, execution context detection |
| formatters.go | Raw, timestamped, and JSON encoder implementations |
| logging_config.go | `LoggingConfig` YAML loading and merging |
| factory_defaults.go | Embedded contract defaults singleton |
| tui_core.go | `tuiCore` zapcore.Core for TUI pane output |

## Dependencies

- `core/paths` -- log file path resolution
- `core/environments` -- environment variable constants for debug init
- `contracts/core` -- embedded logging defaults from contract filesystem

## Role in System

`logging` is the observability backbone of the `core` module. Every command and
subsystem obtains loggers through this package, which routes messages to console,
rolling log files, per-module target logs, and TUI panes based on configuration.

## Code Health

### Tech Debt
- debug.go: Seven package-level mutable vars (`debugEnabled`, `stdOutput`, `debugOutput`, `executionContext`, `originalCommand`, `contextLogged`, plus `globalLogger` in logger.go) form a large implicit global surface; consider grouping into a single `debugState` struct behind a sync.Mutex
- configure.go and logger.go both construct `zap.ErrorOutput(zapcore.AddSync(io.Discard))` independently; extract a shared `defaultZapOptions()` helper

### Pain Points
- None identified

### Optimization Opportunities
- `silentLumberjackWriter` captures and restores `stdlog.Writer()` on every `Write` call; replacing with a one-time redirect at init time would remove per-write overhead (low risk, measure under concurrent load)
- The `Get()` auto-initialization path re-acquires the write lock; using `sync.Once` instead of the manual double-check pattern would simplify the logic (straightforward refactor)
