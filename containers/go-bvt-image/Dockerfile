# Go BVT (Build Verification Test) Image
# Optimized container for running CI workflows with pre-installed tools
#
# Usage in GitHub Actions:
#   container:
#     image: ghcr.io/ready-to-release/go-bvt-image:latest
#
# Build locally:
#   docker build -t go-bvt-image:local containers/go-bvt-image/

FROM golang:1.24-bookworm

LABEL org.opencontainers.image.source="https://github.com/ready-to-release/eac"
LABEL org.opencontainers.image.description="Go BVT container with Go toolchain, Docker, and CI tools"
LABEL org.opencontainers.image.licenses="MIT"

# Install common CI tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    jq \
    curl \
    ca-certificates \
    gnupg \
    lsb-release \
    && rm -rf /var/lib/apt/lists/*

# Install Docker CLI and Docker Buildx
RUN curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null \
    && apt-get update \
    && apt-get install -y --no-install-recommends \
    docker-ce-cli \
    docker-buildx-plugin \
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
RUN while IFS= read -r line; do \
      # Skip comments and empty lines
      case "$line" in \
        "#"*|"") continue ;; \
      esac; \
      echo "Downloading: $line"; \
      go get "$line" || true; \
    done < go-cache-warm.txt \
    && go install github.com/cucumber/godog/cmd/godog@latest \
    && rm -rf /tmp/cache-warm

# Set working directory for CI
WORKDIR /workspace

# Configure git to trust the workspace directory (fixes VCS status errors in mounted volumes)
RUN git config --global --add safe.directory /workspace

# Verify installations
RUN go version && git --version && jq --version && gh --version && docker --version
