# ============================================================
# DEPENDENCIES STAGE - Download Go dependencies (runs on build host)
# ============================================================
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS deps

# Install git for go mod download
RUN apk add --no-cache git

ENV GO111MODULE=on
ENV GOWORK=off
ENV CGO_ENABLED=0

WORKDIR /app

# Copy go.mod and go.sum files for dependency caching
COPY go/eac/core/go.mod go/eac/core/go.sum ./go/eac/core/
COPY go/eac/commands/go.mod go/eac/commands/go.sum ./go/eac/commands/
COPY go/eac/godog/go.mod go/eac/godog/go.sum ./go/eac/godog/
COPY contracts/eac-core/0.1.0/interfaces/go.mod ./contracts/eac-core/0.1.0/interfaces/
COPY contracts/eac-adapters/ai/go.mod ./contracts/eac-adapters/ai/
COPY contracts/eac-adapters/docker/go.mod contracts/eac-adapters/docker/go.sum ./contracts/eac-adapters/docker/
COPY contracts/eac-adapters/tui/go.mod ./contracts/eac-adapters/tui/
COPY go/eac/adapters/ai/go.mod go/eac/adapters/ai/go.sum ./go/eac/adapters/ai/
COPY go/eac/adapters/docker/go.mod go/eac/adapters/docker/go.sum ./go/eac/adapters/docker/
COPY go/eac/adapters/tui/go.mod go/eac/adapters/tui/go.sum ./go/eac/adapters/tui/

# Download dependencies with BuildKit cache mount for faster rebuilds
RUN --mount=type=cache,target=/go/pkg/mod \
    cd go/eac/commands && go mod download

# ============================================================
# BUILD STAGE - Cross-compile eac-commands binary (runs on build host)
# ============================================================
FROM deps AS builder

# Target architecture for cross-compilation
ARG TARGETARCH

# Copy source code
COPY go/eac/core/ ./go/eac/core/
COPY go/eac/commands/ ./go/eac/commands/
COPY go/eac/adapters/ ./go/eac/adapters/
COPY contracts/eac-core/0.1.0/interfaces/ ./contracts/eac-core/0.1.0/interfaces/
COPY contracts/eac-adapters/ ./contracts/eac-adapters/

# Cross-compile for target architecture (fast - no QEMU needed)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOARCH=${TARGETARCH} cd go/eac/commands && go build -o /app/out/tools/commands .

# ============================================================
# RUNTIME STAGE - Final image (runs on target platform)
# ============================================================
FROM golang:1.25-alpine AS runtime

# Install runtime dependencies
# Note: github-cli requires community repo. We detect Alpine version dynamically.
RUN apk add --no-cache \
    ca-certificates \
    git \
    tzdata \
    bash \
    nodejs \
    npm \
    docker-cli \
    docker-cli-buildx \
    curl \
    && ALPINE_VERSION=$(cat /etc/alpine-release | cut -d. -f1,2) \
    && apk add --no-cache \
        --repository=https://dl-cdn.alpinelinux.org/alpine/v${ALPINE_VERSION}/community \
        github-cli

ENV GO111MODULE=on
ENV GOWORK=off
ENV CGO_ENABLED=0

WORKDIR /app

# Note: R2R_DOCKER_MODE is set by r2r CLI at runtime, not here
# This allows tests to find /app via .git while normal usage gets /var/task
ENV TZ=UTC
ENV R2R_CONTAINER_ROOT=/app

# Copy source code to /app (needed for go test in acceptance tests)
COPY go/eac/core/ ./go/eac/core/
COPY go/eac/commands/ ./go/eac/commands/
COPY go/eac/adapters/ ./go/eac/adapters/
COPY go/eac/specs/ ./go/eac/specs/
COPY contracts/ ./contracts/
COPY .r2r/ ./.r2r/
COPY specs/ ./specs/
COPY templates/ ./templates/

# Copy cross-compiled binary from builder to tools directory
# (out/tools is for CI tools, out/build is for module build outputs)
COPY --from=builder /app/out/tools/commands /app/out/tools/commands

# Create fake .git so GetRepositoryRoot() finds /app when running tests
# (tests run from /app, not /var/task which is for user repo mounts)
RUN mkdir -p /app/.git

# Copy test assets needed for godog tests (scripts for installer tests)
COPY scripts/ ./scripts/

# Add OCI labels
LABEL org.opencontainers.image.title="EAC Extension"
LABEL org.opencontainers.image.description="Everything as Code - Repository management tooling for R2R CLI"
LABEL org.opencontainers.image.version="1.0.0"
LABEL org.opencontainers.image.source="https://github.com/ready-to-release/eac"
LABEL org.opencontainers.image.vendor="Ready to Release"
LABEL org.opencontainers.image.licenses="MIT"

# Add R2R CLI extension labels
LABEL r2r-cli.extension.name="eac"
LABEL r2r-cli.extension.category="tools"
LABEL r2r-cli.extension.type="commands"
LABEL r2r-cli.extension.maturity="stable"
LABEL r2r-cli.version.min="1.0.0"
LABEL r2r-cli.capabilities="repository-management,module-discovery,dependency-analysis,testing,documentation,specifications,architecture-design,git-worktree,build-automation,pipeline-execution"
LABEL r2r-cli.runtime="go"
LABEL r2r-cli.os.support="linux"
LABEL r2r-cli.resources.cpu.min="0.1"
LABEL r2r-cli.resources.memory.min="256Mi"

# Multi-platform support labels
LABEL r2r-cli.multi-arch="true"
LABEL r2r-cli.platforms="linux/amd64,linux/arm64"
LABEL r2r-cli.multi-platform.support="true"

# Set working directory for command execution (r2r CLI mounts repo here)
WORKDIR /var/task

# Use pre-built eac-commands binary as entrypoint
ENTRYPOINT ["/app/out/tools/commands"]
