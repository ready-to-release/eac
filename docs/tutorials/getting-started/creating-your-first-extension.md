# Creating Your First Extension

Learn how to build a custom clie extension - a containerized command that extends the clie CLI with your own functionality.

**Prerequisites:** [Quick Start Guide](./quick-start.md), Docker installed, basic Go knowledge

## What You'll Learn

By the end of this tutorial, you'll be able to:

- Understand what clie extensions are and how they work
- Set up a proper Go project structure for extensions
- Implement the required extension metadata
- Build and test your extension locally
- Use your extension via the clie CLI

## What is an Extension?

An **clie extension** is a Docker container that adds custom commands to the clie CLI.

Extensions run in isolation and receive arguments from the CLI.

**Key characteristics:**

- **Containerized** - Packaged as Docker images
- **CLI-integrated** - Invoked via `clie <extension-name> <command>`
- **Self-describing** - Provide metadata about capabilities and requirements
- **Portable** - Run consistently across platforms

**Example:** The `ext-env-check` extension validates environment variables:

```bash
clie env-check HOME USER SHELL
```

!!! tip "Complete Guide Available"

    This tutorial provides a quick introduction. For comprehensive details, see:
    - **[Creating Extensions Guide](../../how-to-guides/clie/creating-extensions.md)** - Complete reference with production patterns
    - **[ext-env-check Repository](https://github.com/ready-to-release/ext-env-check)** - Working example to study

## Step 1: Create Project Structure

Create a new directory for your extension:

```bash
mkdir my-extension
cd my-extension
```

Set up the recommended structure:

```text
my-extension/
├── go/
│   └── my-extension/
│       ├── cmd/
│       │   └── my-extension/
│       │       └── main.go
│       ├── internal/
│       │   ├── metadata/
│       │   │   └── metadata.go
│       │   └── core/
│       │       └── logic.go
│       ├── go.mod
│       └── go.sum
└── containers/
    └── my-extension/
        └── Dockerfile
```

**Why this structure?**

- `cmd/` - Entry point separate from logic
- `internal/` - Encapsulated packages
- `internal/metadata/` - Extension metadata implementation
- `internal/core/` - Business logic
- `containers/` - Docker configuration

## Step 2: Implement Extension Metadata

Every extension **must** implement the `extension-meta` command.

Create `go/my-extension/internal/metadata/metadata.go`:

```go
package metadata

import (
    "encoding/json"
    "os"
)

type Metadata struct {
    Name          string       `json:"name"`
    Version       string       `json:"version"`
    Description   string       `json:"description"`
    SchemaVersion string       `json:"schema-version"`
    Capabilities  []string     `json:"capabilities"`
    Requirements  Requirements `json:"requirements"`
}

type Requirements struct {
    CLIEVersion       string `json:"clie-version"`
    ContainerRuntime string `json:"container-runtime"`
    MinimumMemory    string `json:"minimum-memory"`
    MinimumCPU       string `json:"minimum-cpu"`
}

func Print(version string) {
    meta := Metadata{
        Name:          "my-extension",
        Version:       version,
        Description:   "My custom extension",
        SchemaVersion: "1.0",
        Capabilities:  []string{"custom-task"},
        Requirements: Requirements{
            CLIEVersion:       ">=0.1.0",
            ContainerRuntime: "docker",
            MinimumMemory:    "64MB",
            MinimumCPU:       "0.1",
        },
    }

    encoder := json.NewEncoder(os.Stdout)
    encoder.SetIndent("", "  ")
    encoder.Encode(meta)
}
```

## Step 3: Create the Main Entry Point

Create `go/my-extension/cmd/my-extension/main.go`:

```go
package main

import (
    "fmt"
    "os"

    "github.com/your-org/my-extension/internal/metadata"
)

var Version = "dev"

func main() {
    // Handle special commands FIRST
    if len(os.Args) > 1 {
        switch os.Args[1] {
        case "extension-meta":
            metadata.Print(Version)
            os.Exit(0)
        case "version":
            fmt.Printf("my-extension version %s\n", Version)
            os.Exit(0)
        case "help", "--help", "-h":
            printHelp()
            os.Exit(0)
        }
    }

    // Your command logic here
    if len(os.Args) < 2 {
        fmt.Println("Usage: my-extension <command>")
        os.Exit(1)
    }

    // Handle commands
    switch os.Args[1] {
    case "greet":
        fmt.Println("Hello from my-extension!")
    default:
        fmt.Printf("Unknown command: %s\n", os.Args[1])
        os.Exit(1)
    }
}

func printHelp() {
    fmt.Println("my-extension - My custom extension")
    fmt.Println()
    fmt.Println("Commands:")
    fmt.Println("  greet          Say hello")
    fmt.Println("  version        Show version")
    fmt.Println("  extension-meta Show metadata")
    fmt.Println("  help           Show this help")
}
```

## Step 4: Initialize Go Module

```bash
cd go/my-extension
go mod init github.com/your-org/my-extension
```

## Step 5: Create Dockerfile

Create `containers/my-extension/Dockerfile`:

```dockerfile
# Multi-stage build for minimal image size
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy dependency files first (caching)
COPY go/my-extension/go.mod go/my-extension/go.sum ./
RUN go mod download

# Copy source code
COPY go/my-extension/ ./

# Build with version injection
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X main.Version=${VERSION}" \
    -o my-extension \
    ./cmd/my-extension

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /workspace

COPY --from=builder /build/my-extension /usr/local/bin/my-extension

ENTRYPOINT ["my-extension"]
CMD ["--help"]
```

## Step 6: Build and Test

Build the Docker image:

```bash
cd ../..  # Back to project root
docker build -f containers/my-extension/Dockerfile -t my-extension:dev .
```

Test the container:

```bash
# Test metadata
docker run --rm my-extension:dev extension-meta

# Test help
docker run --rm my-extension:dev --help

# Test your command
docker run --rm my-extension:dev greet
```

## Step 7: Configure for clie

Create `.clie/clie-cli.local.yml`:

```yaml
version: "1.0"
extensions:
  - name: my-ext
    image: my-extension:dev
    load_local: true
    image_pull_policy: Never
```

## Step 8: Use via clie CLI

Test your extension through clie:

```bash
clie my-ext greet
clie my-ext --help
clie my-ext extension-meta
```

## What You Learned

Congratulations! You've successfully:

- ✅ Created a proper Go project structure for extensions
- ✅ Implemented required extension metadata
- ✅ Built a multi-stage Docker image
- ✅ Tested the container locally
- ✅ Configured clie to use your extension
- ✅ Invoked your extension via clie CLI

## Key Concepts Covered

- **Extension metadata** - Required `extension-meta` command for discovery
- **Project structure** - `cmd/` and `internal/` organization
- **Docker multi-stage builds** - Optimized container images
- **Local development** - Building and testing without publishing
- **clie integration** - Configuration and invocation

## Next Steps

### Enhance Your Extension

Now that you have a working extension, you can:

- **Add more commands** - Extend the switch statement in main.go
- **Add business logic** - Implement in `internal/core/`
- **Add output formatting** - Support text, JSON, etc.
- **Add tests** - Write table-driven unit tests
- **Add specifications** - Write Gherkin feature files

### Learn Production Patterns

For production-ready extensions, see:

- **[Creating Extensions Guide](../../how-to-guides/clie/creating-extensions.md)** - Complete patterns:
  - Testing strategies (unit, behavior, container)
  - EAC integration for automated builds
  - Publishing to registries
  - Multi-platform builds
  - Release workflows

- **[ext-env-check Example](https://github.com/ready-to-release/ext-env-check)** - Study a real extension:
  - Production project structure
  - Comprehensive testing
  - CI/CD workflows
  - Changelog-based releases

### Related Guides

- **[Local Development Workflows](../../how-to-guides/local-setup/local-dev-workflows.md)** - Iterate faster
- **[Testing in External Repos](../../how-to-guides/clie/testing-in-external-repos.md)** - Validate across projects
