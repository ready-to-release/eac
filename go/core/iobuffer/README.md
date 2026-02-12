# iobuffer

I/O utility types for bounded buffering and memory-safe stream handling.

## Key Types

- **`LimitedBuffer`** -- io.Writer with byte limit to prevent memory exhaustion from unbounded output

## Patterns

- Hard limit enforcement: Write returns error when limit is exceeded
- Total tracking: Records total bytes written including partial writes
- String conversion: Provides String() for easy content access

## Internal Structure

| File | Responsibility |
| --- | --- |
| limited_buffer.go | LimitedBuffer with size-capped Write implementation |
| limited_buffer_test.go | Test coverage for limit enforcement and partial writes |

## Dependencies

None. This is a leaf package with no internal dependencies.

## Role in System

`iobuffer` provides memory-safe I/O primitives used by container runtime adapters and external process execution. The `LimitedBuffer` prevents OOM crashes when capturing stdout/stderr from Docker containers or long-running processes that may produce unbounded output.

## Code Health

### Tech Debt
- None identified

### Pain Points
- None identified

### Optimization Opportunities
- None identified
