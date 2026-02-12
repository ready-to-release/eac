# drawio

Provides commands for editing DrawIO diagram files stored as `.drawio.png` (PNG images with embedded XML metadata). Operations include decode, encode, embed, render, create, and info, all executed via a drawio-oci Docker container.

## Key Types

- **`ContainerProvider`** -- Function type returning a `ContainerPort` for Docker execution
- **`limitedBuffer`** -- Size-limited `io.Writer` preventing memory exhaustion from Docker output (now in `core/iobuffer`)

## Patterns

- Docker-based tooling: all diagram manipulation runs in a drawio-oci container
- Dual execution path: uses `ContainerPort` interface when available, falls back to `exec.Command`
- Path translation: converts local paths to container mount paths, handling Windows drive letters
- DinD-aware mounting: resolves host repo root for correct volume mounts in Docker-in-Docker
- Registry-based subcommand dispatch: each subcommand file calls `registry.Register` in `init()`

## Internal Structure

| File | Responsibility |
| --- | --- |
| drawio.go | Docker image management, container command execution, path translation utilities |
| decode.go | Extract and decode XML from PNG to human-readable format |
| encode.go | Encode human-readable XML back to DrawIO compressed format |
| embed.go | Write encoded XML into PNG file metadata (create or update) |
| render.go | Render DrawIO diagram to actual PNG image |
| create.go | Create new `.drawio.png` with blank canvas or provided XML content |
| info.go | Show diagram metadata (pages, elements, dimensions) |

## Dependencies

- `contracts/container-runtime` -- `ContainerPort` interface for container execution
- `adapters/docker/util` -- Docker availability check, volume path formatting, DinD detection
- `clibase/flags` -- flag validation from registry metadata
- `clibase/registry` -- command registration
- `core/logging` -- structured logging
- `core/repository` -- repository root discovery
- `core/tool` -- Docker image resolution from tool-config.yml

## Role in System

The `drawio` package enables LLM-powered diagram editing in `eac` by providing a complete toolchain for manipulating DrawIO diagrams stored as `.drawio.png` files. Its decode/encode/embed pipeline allows programmatic extraction and modification of diagram XML, while the render command produces viewable PNG images, supporting the documentation workflow alongside the `update docs` command.

## Code Health

### Tech Debt
- drawio.go (311 lines) is the largest file; contains Docker management, path translation, and command execution
- subcommands_test.go (163 lines) now provides unit tests for subcommand logic

### Pain Points
- drawio_test.go (119 lines) covers path translation utilities but subcommand orchestration relies on integration tests

### Optimization Opportunities
- None identified
