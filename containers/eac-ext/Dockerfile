# ============================================================
# RUNTIME STAGE - Final image with pre-built binary
# ============================================================
# The eac module cross-compiles the binary during its own build phase.
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

ENV TZ=UTC
ENV CLIE_CONTAINER_ROOT=/app

# Copy pre-built binary from eac build output.
# TARGETARCH is set automatically by docker buildx (amd64 or arm64).
# The binary was cross-compiled by the eac module's go builder.
# The Docker build context is scoped to out/build/eac/eac_go/.
ARG TARGETARCH
COPY eac-linux-${TARGETARCH} /app/out/tools/eac

# Add OCI labels
LABEL org.opencontainers.image.title="EAC Extension"
LABEL org.opencontainers.image.description="Everything as Code - Repository management tooling for CLIE CLI"
LABEL org.opencontainers.image.version="1.0.0"
LABEL org.opencontainers.image.source="https://github.com/ready-to-release/eac"
LABEL org.opencontainers.image.vendor="Ready to Release"
LABEL org.opencontainers.image.licenses="MIT"

# Add CLIE CLI extension labels
LABEL clie.extension.name="eac"
LABEL clie.extension.category="tools"
LABEL clie.extension.type="commands"
LABEL clie.extension.maturity="stable"
LABEL clie.version.min="1.0.0"
LABEL clie.capabilities="repository-management,module-discovery,dependency-analysis,testing,documentation,specifications,architecture-design,git-worktree,build-automation,pipeline-execution"
LABEL clie.runtime="go"
LABEL clie.os.support="linux"
LABEL clie.resources.cpu.min="0.1"
LABEL clie.resources.memory.min="256Mi"

# Multi-platform support labels
LABEL clie.multi-arch="true"
LABEL clie.platforms="linux/amd64,linux/arm64"
LABEL clie.multi-platform.support="true"

# Set working directory for command execution (clie CLI mounts repo here)
WORKDIR /var/task

# Use pre-built eac-commands binary as entrypoint
ENTRYPOINT ["/app/out/tools/eac"]
