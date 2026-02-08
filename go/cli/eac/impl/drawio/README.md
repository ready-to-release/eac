# drawio

Provides commands for editing DrawIO diagram files stored as `.drawio.png` (PNG images with embedded XML metadata). Operations include decode, encode, embed, render, create, and info, all executed via a drawio-oci Docker container.

## Key Types

- **`ContainerProvider`** -- Function type returning a `ContainerPort` for Docker execution
- **`limitedBuffer`** -- Size-limited `io.Writer` preventing memory exhaustion from Docker output

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

The `drawio` package enables LLM-powered diagram editing in `eac-cli` by providing a complete toolchain for manipulating DrawIO diagrams stored as `.drawio.png` files. Its decode/encode/embed pipeline allows programmatic extraction and modification of diagram XML, while the render command produces viewable PNG images, supporting the documentation workflow alongside the `update docs` command.

## Code Health

### Tech Debt
- Global mutable `var defaultContainerProvider` (drawio.go:48) used as package-level singleton for Docker container access
- No unit tests for subcommand files: decode.go, encode.go, embed.go, render.go, create.go, info.go
- drawio_test.go (119 lines) only covers path translation and image name utilities

### Pain Points
- Each subcommand file (decode, encode, embed, render, create, info) follows a similar structure of flag parsing, path resolution, and Docker command execution but cannot be tested without Docker
- `limitedBuffer` is defined inline in drawio.go; a near-identical implementation exists in design/helper/validator.go

### Optimization Opportunities
- Deduplicate `limitedBuffer` into a shared utility package (high feasibility, identical implementation in two packages)
- Add mock-based unit tests for subcommand orchestration logic by injecting a test `ContainerProvider` (moderate feasibility, `SetContainerProvider` already exists for this purpose)
