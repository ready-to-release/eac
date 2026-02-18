# OCI Tools Overview

The OCI (Open Container Initiative) tools module group provides containerized development tools used by EAC for build, test, documentation, and visualization workflows. These container images provide isolated, reproducible tool environments.

## Purpose

OCI tool containers enable:

- **Reproducible Builds**: Consistent tool versions across all environments
- **Dependency Isolation**: No global tool installation required
- **Cross-Platform Support**: Linux-based tools work on Windows, macOS, and Linux
- **Version Pinning**: SHA-based image tags for exact reproducibility

## Modules

The OCI tools consist of the following container images:

| OCI Tool              | Purpose                          | Base Image                   | Use Cases                       |
| --------------------- | -------------------------------- | ---------------------------- | ------------------------------- |
| **cgo-oci**           | C/Go cross-compilation toolchain | Alpine + gcc                 | CGO builds, native compilation  |
| **dotnet-oci**        | .NET SDK and runtime             | mcr.microsoft.com/dotnet/sdk | .NET builds, tests, packaging   |
| **drawio-oci**        | Draw.io diagram rendering        | Alpine + Node.js             | Architecture diagram generation |
| **git-oci**           | Git version control tools        | Alpine + git                 | Git operations in CI/CD         |
| **go-oci**            | Go SDK and tools                 | golang:alpine                | Go builds, tests, linting       |
| **gource-oci**        | Repository visualization         | Ubuntu + gource              | Repository history videos       |
| **mermaid-oci**       | Mermaid diagram rendering        | Node.js + mermaid-cli        | Mermaid diagram generation      |
| **mkdocs-dev-oci**    | MkDocs development server        | Python + mkdocs              | Live documentation preview      |
| **mkdocs-render-oci** | MkDocs static site generation    | Python + mkdocs              | Documentation builds            |
| **nginx-oci**         | NGINX web server                 | nginx:alpine                 | Serving static sites            |
| **pdf-cli-oci**       | PDF generation CLI               | Node.js + Puppeteer          | PDF rendering from HTML         |
| **pdf-oci**           | PDF utilities                    | Alpine + poppler             | PDF manipulation                |

## Architecture

OCI tools are used by EAC commands and CI/CD workflows:

```text
┌─────────────────────────────────┐
│      EAC Commands               │
│  (build, test, update-docs)     │
└───────────┬─────────────────────┘
            │
┌───────────▼─────────────────────┐
│   Docker / Container Runtime    │
└───────────┬─────────────────────┘
            │
    ┌───────┼───────┐
    │       │       │
┌───▼──┐ ┌─▼───┐ ┌─▼──────┐
│go-oci│ │mkdocs│ │drawio  │  ...
│      │ │-oci  │ │-oci    │
└──────┘ └──────┘ └────────┘
```

### Key Design Principles

1. **Minimal Base Images**: Use Alpine Linux where possible for small image sizes
2. **Single Responsibility**: Each container provides one tool or toolchain
3. **Versioned Tags**: Images tagged with SHA for reproducibility
4. **Multi-Platform**: Support linux/amd64 (and linux/arm64 where applicable)

## Container Registry

All OCI tools are published to GitHub Container Registry:

```text
ghcr.io/ready-to-release/{tool-name}:{tag}
```

### Tags

- **latest**: Most recent build (floating tag)
- **sha-{short-sha}**: Specific commit SHA (immutable tag)

### Example Usage

```bash
# Pull a specific version
docker pull ghcr.io/ready-to-release/go-oci:sha-abc1234

# Run go build in container
docker run --rm -v $(pwd):/workspace -w /workspace \
  ghcr.io/ready-to-release/go-oci:latest \
  go build -o bin/app ./cmd/app
```

## Build and Push Workflows

OCI tool images are built and pushed via CI/CD:

```yaml
# .github/workflows/oci-tools.yml
name: OCI Tools
on:
  push:
    paths:
      - "containers/**"

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    steps:
      - name: Build and push
        run: |
          docker build -t ghcr.io/ready-to-release/go-oci:latest \
            containers/go-oci/
          docker push ghcr.io/ready-to-release/go-oci:latest
```

## Container Categories

### Language Toolchains

Containers for language-specific builds:

- **go-oci**: Go compilation and testing
- **cgo-oci**: C/Go cross-compilation
- **dotnet-oci**: .NET builds and tests

### Documentation Tools

Containers for documentation generation:

- **mkdocs-dev-oci**: Live documentation server
- **mkdocs-render-oci**: Static site generation
- **drawio-oci**: Diagram rendering (Draw.io XML → PNG/SVG)
- **mermaid-oci**: Diagram rendering (Mermaid → PNG/SVG)
- **pdf-cli-oci**: HTML → PDF conversion
- **pdf-oci**: PDF utilities (merge, split)

### Visualization Tools

Containers for repository visualization:

- **gource-oci**: Repository history animations

### Infrastructure Tools

Containers for supporting infrastructure:

- **git-oci**: Git operations
- **nginx-oci**: Static file serving

## Usage in EAC

EAC commands invoke OCI tools via the Docker adapter:

```go
// Example: Build Go module using go-oci
func buildGoModule(ctx context.Context, module string) error {
    return docker.Run(ctx, docker.RunConfig{
        Image: "ghcr.io/ready-to-release/go-oci:latest",
        WorkDir: "/workspace",
        Volumes: []string{
            fmt.Sprintf("%s:/workspace", workspace),
        },
        Command: []string{"go", "build", "-o", "bin/app", "./cmd/app"},
    })
}
```

## Dockerfile Structure

OCI tool Dockerfiles follow a consistent pattern:

```dockerfile
# Multi-stage build for minimal final image
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git make

FROM alpine:3.19
COPY --from=builder /usr/local/go /usr/local/go
ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /workspace
ENTRYPOINT ["/bin/sh", "-c"]
```

## Performance Optimizations

### Layer Caching

- Base images cached locally and in CI
- Multi-stage builds minimize final image size
- Shared base images across containers

### Image Size

| Image             | Size    | Optimization                   |
| ----------------- | ------- | ------------------------------ |
| go-oci            | ~350 MB | Alpine base, multi-stage build |
| mkdocs-render-oci | ~200 MB | Minimal Python dependencies    |
| git-oci           | ~50 MB  | Alpine + git only              |

## Security

### Non-Root User

Containers run as non-root user where possible:

```dockerfile
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

USER appuser
```

### Minimal Attack Surface

- Minimal base images (Alpine)
- Only necessary tools included
- Regular security updates via CI rebuilds

## CI/CD Integration

OCI tools are used in GitHub Actions workflows:

```yaml
# Example: Build documentation using mkdocs-render-oci
- name: Build docs
  run: |
    docker run --rm \
      -v ${{ github.workspace }}:/workspace \
      ghcr.io/ready-to-release/mkdocs-render-oci:latest \
      mkdocs build
```

## Adding New OCI Tools

To add a new OCI tool:

1. **Create Dockerfile**: `containers/{tool-name}-oci/Dockerfile`
2. **Test locally**: `docker build -t {tool-name}-oci containers/{tool-name}-oci/`
3. **Add CI workflow**: Update `.github/workflows/oci-tools.yml`
4. **Document**: Create `docs/reference/eac/modules/oci-tools/{tool-name}-oci.md`

## See Also

- [Adapters Module](adapters.md) - Container runtime and tool integration
- [Continuous Delivery](../../../explanation/continuous-delivery/index.md) - CI/CD workflows
- [EAC Architecture](../architecture/index.md) - Overall architecture
- [Modules Index](../index.md) - Complete module reference
