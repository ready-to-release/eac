# ============================================================
# RUNTIME STAGE - Final image with pre-built binary
# ============================================================
# The eac-cli module cross-compiles the binary during its own build phase.
# We consume the pre-built binary instead of recompiling from source,
# eliminating ~60s of redundant Go compilation per architecture.
FROM golang:1.25-alpine

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

# Note: CLIE_DOCKER_MODE is set by clie CLI at runtime, not here
# This allows tests to find /app via .git while normal usage gets /var/task
ENV TZ=UTC
ENV CLIE_CONTAINER_ROOT=/app

# Copy source code to /app (needed for go test in acceptance tests)
COPY go/core/ ./go/core/
COPY go/clibase/ ./go/clibase/
COPY go/cli/eac/ ./go/cli/eac/
COPY go/adapters/ ./go/adapters/
COPY go/specs/ ./go/specs/
COPY contracts/ ./contracts/
COPY .clie/ ./.clie/
COPY specs/ ./specs/
COPY templates/ ./templates/

# Copy pre-built binary from eac-cli build output.
# TARGETARCH is set automatically by docker buildx (amd64 or arm64).
# The binary was cross-compiled by the eac-cli module's go builder.
ARG TARGETARCH
COPY out/build/eac-cli/go-go/eac-linux-${TARGETARCH} /app/out/tools/eac

# Create fake .git so GetRepositoryRoot() finds /app when running tests
# (tests run from /app, not /var/task which is for user repo mounts)
RUN mkdir -p /app/.git

# Copy test assets needed for godog tests (scripts for installer tests)
COPY scripts/ ./scripts/

# Add OCI labels
LABEL org.opencontainers.image.title="EAC Extension"
LABEL org.opencontainers.image.description="Everything as Code - Repository management tooling for CLIE CLI"
LABEL org.opencontainers.image.version="1.0.0"
LABEL org.opencontainers.image.source="https://github.com/ready-to-release/eac"
LABEL org.opencontainers.image.vendor="Ready to Release"
LABEL org.opencontainers.image.licenses="MIT"

# Add CLIE CLI extension labels
LABEL clie-cli.extension.name="eac"
LABEL clie-cli.extension.category="tools"
LABEL clie-cli.extension.type="commands"
LABEL clie-cli.extension.maturity="stable"
LABEL clie-cli.version.min="1.0.0"
LABEL clie-cli.capabilities="repository-management,module-discovery,dependency-analysis,testing,documentation,specifications,architecture-design,git-worktree,build-automation,pipeline-execution"
LABEL clie-cli.runtime="go"
LABEL clie-cli.os.support="linux"
LABEL clie-cli.resources.cpu.min="0.1"
LABEL clie-cli.resources.memory.min="256Mi"

# Multi-platform support labels
LABEL clie-cli.multi-arch="true"
LABEL clie-cli.platforms="linux/amd64,linux/arm64"
LABEL clie-cli.multi-platform.support="true"

# Set working directory for command execution (clie CLI mounts repo here)
WORKDIR /var/task

# Use pre-built eac-commands binary as entrypoint
ENTRYPOINT ["/app/out/tools/eac"]
