# ============================================================
# DEPENDENCIES STAGE - Download Go dependencies
# ============================================================
FROM golang:1.24-alpine AS deps

# Build arguments for multi-platform support
ARG TARGETARCH

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    git \
    tzdata \
    github-cli \
    bash \
    npm \
    docker-cli \
    && rm -rf /var/cache/apk/*

ENV GO111MODULE=on
ENV GOWORK=off
# Disable CGO for cross-compilation (Windows/macOS builds from Linux)
ENV CGO_ENABLED=0

WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker layer caching
# The ext module has replace directives pointing to parent modules,
# so we need all related go.mod files for dependency resolution
COPY src/core/go.mod src/core/go.sum ./src/core/
COPY src/core/ai/go.mod src/core/ai/go.sum ./src/core/ai/
COPY src/core/markdown/go.mod src/core/markdown/go.sum ./src/core/markdown/
COPY src/core/module-deps/go.mod src/core/module-deps/go.sum ./src/core/module-deps/
COPY src/commands/go.mod src/commands/go.sum ./src/commands/
COPY src/commands/ext/go.mod src/commands/ext/go.sum ./src/commands/ext/

# Download dependencies (this layer will be cached)
RUN cd src/commands/ext && go mod download -x

# ============================================================
# RUNTIME STAGE - Final image with source code
# ============================================================
FROM deps AS runtime

ENV DOCKER_R2R_MODE=true
ENV TZ=UTC

# Copy source code (won't invalidate the dependency layer above)
COPY src/core/ ./src/core/
COPY src/commands/ ./src/commands/
COPY contracts/ ./contracts/

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

# Run Go source directly to allow runtime metadata reading
ENTRYPOINT ["sh", "-c", "cd /app/src/commands/ext && exec go run . \"$@\"", "--"]
