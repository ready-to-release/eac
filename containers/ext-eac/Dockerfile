# ============================================================
# RUNTIME STAGE - Single stage with Go runtime
# ============================================================
FROM golang:1.24-alpine

# Build arguments for multi-platform support
ARG TARGETARCH

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    git \
    tzdata \
    github-cli \
    && rm -rf /var/cache/apk/*

ENV DOCKER_R2R_MODE=true
ENV TZ=UTC
ENV GO111MODULE=on
ENV GOWORK=off

WORKDIR /app

# Copy source code for all dependencies
COPY src/core/ ./src/core/
COPY src/commands/ ./src/commands/
COPY contracts/ ./contracts/

# Download dependencies - cached unless go.mod/go.sum changes
RUN --mount=type=cache,target=/go/pkg/mod \
    cd src/commands/ext && go mod download

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

# Run Go source directly instead of compiled binary
# Use shell form to support changing directory first
ENTRYPOINT ["sh", "-c", "cd /app/src/commands/ext && exec go run . \"$@\"", "--"]
