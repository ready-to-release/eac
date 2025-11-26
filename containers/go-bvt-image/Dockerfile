# EAC CI Container
# Optimized container for running CI workflows with pre-installed tools
#
# Usage in GitHub Actions:
#   container:
#     image: ghcr.io/ready-to-release/eac-ci:latest
#
# Build locally:
#   docker build -t eac-ci:local containers/eac-ci/

FROM golang:1.24-bookworm

LABEL org.opencontainers.image.source="https://github.com/ready-to-release/eac"
LABEL org.opencontainers.image.description="CI container for EAC repository with Go toolchain and common tools"
LABEL org.opencontainers.image.licenses="MIT"

# Install common CI tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    jq \
    curl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install GitHub CLI
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg \
    && chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
    && apt-get update \
    && apt-get install -y gh \
    && rm -rf /var/lib/apt/lists/*

# Pre-warm Go module cache with common dependencies
# This speeds up first build in CI by having base modules cached
WORKDIR /tmp/cache-warm
COPY go-cache-warm.txt .
RUN go install github.com/cucumber/godog/cmd/godog@latest \
    && rm -rf /tmp/cache-warm

# Set working directory for CI
WORKDIR /workspace

# Verify installations
RUN go version && git --version && jq --version && gh --version
