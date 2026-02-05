# Go Toolchain Container
# General-purpose container for Go builds, tests, and toolchain operations.
# Used by the tool system as an alternative to native 'go' binary.
#
# Features:
# - CGO support (gcc, musl-dev)
# - Cross-compilation support (CGO_ENABLED=0 by default for portability)
# - Module caching (mount /go/pkg and /root/.cache/go-build)
# - Race detection (requires CGO_ENABLED=1)
#
# Usage via tool system:
#   ArgsOverrides: ["build", "-o", "output", "./..."]
#   ArgsOverrides: ["test", "-race", "./..."]
#   EnvOverrides: {"GOOS": "linux", "GOARCH": "amd64"}

ARG GO_VERSION=1.24
FROM golang:${GO_VERSION}-alpine

WORKDIR /app

# Install CGO dependencies and cross-compilation support
RUN apk add --no-cache \
    gcc \
    musl-dev \
    git \
    ca-certificates \
    && rm -rf /var/cache/apk/*

# Default: CGO disabled for cross-platform compatibility
# Enable via EnvOverrides when race detection or CGO is needed
ENV CGO_ENABLED=0
ENV GOPROXY=https://proxy.golang.org,direct

# Pre-create cache directories (will be mounted from host)
RUN mkdir -p /go/pkg /root/.cache/go-build

# Entrypoint is 'go' - all subcommands come from ArgsOverrides
# Examples:
#   ArgsOverrides: ["mod", "tidy"]
#   ArgsOverrides: ["build", "-o", "/output/bin", "./..."]
#   ArgsOverrides: ["test", "-json", "./..."]
ENTRYPOINT ["go"]

# Default command if no ArgsOverrides provided
CMD ["version"]
