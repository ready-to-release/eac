# logging

Leveled logger with configurable sinks and multiple output formatters.

## Key Types

- **`Logger`** -- Core logger with config and mutex for thread-safe formatter access
- **`LoggingConfig`** -- Complete logging configuration with console and file sinks
- **`SinkConfig`** -- Per-sink configuration: levels, formatter type, enabled flag
- **`FormatterType`** -- Output format enum: raw, timestamped, JSON

## Key Functions

- **`InitFromEnv`** -- Initializes global logger with defaults and checks env vars for debug mode
- **`EnableDebug` / `DisableDebug`** -- Atomic toggle for debug logging
- **`SetLevel`** -- Sets minimum log level (debug, info, warn, error)
- **`Debug` / `Info` / `Warn` / `Error` / `Fatal`** -- Level-specific log functions with `f` variants
- **`SetOutput` / `ResetOutput`** -- Redirect output writers for testing
- **`LoadConfig`** -- Loads logging configuration from `.clie/clie-logging.yml`

## Patterns

- Global singleton: `sync.Once` ensures single logger initialization; `get()` lazy-creates on first use
- Atomic debug toggle: `sync/atomic` uint32 for zero-cost debug check on hot path
- Stream separation: Info goes to stdout; Debug, Warn, Error go to stderr
- Configurable formatters: Raw (clean CLI output), Timestamped (`HH:MM:SS.mmm LEVEL msg`), JSON (structured)
- Test isolation: `ResetForTesting()` resets all global state including `sync.Once`

## Internal Structure

| File      | Responsibility                                                    |
| --------- | ----------------------------------------------------------------- |
| logger.go | Logger type, global singleton, level functions, formatters, output |
| config.go | LoggingConfig, SinkConfig, FormatterType, DefaultConfig, LoadConfig |

## Dependencies

- `internal/envconsts` -- Environment variable names for debug and log level

## Role in System

The logging package is used by every other package in the clie module. It provides the standard logging interface for all user-facing messages, debug output, warnings, and errors. The raw formatter is the default for clean CLI output, while timestamped and JSON formatters can be enabled via configuration for debugging and structured log collection.

## Code Health

### Tech Debt

_None identified._

### Pain Points

_None identified._

### Optimization Opportunities

_None identified._
