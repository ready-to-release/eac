# Go BVT (Build Verification Test) Image
# Lightweight container for running Go CI workflows
#
# Usage in GitHub Actions:
#   container:
#     image: ghcr.io/ready-to-release/go-bvt-image:latest
#
# Build locally:
#   docker build -t go-bvt-image:local containers/go-bvt-image/

FROM golang:1.24-alpine

LABEL org.opencontainers.image.source="https://github.com/ready-to-release/eac"
LABEL org.opencontainers.image.description="Lightweight Go CI container with Go toolchain and essential tools"
LABEL org.opencontainers.image.licenses="MIT"

# Install essential CI tools
# - git: version control
# - jq: JSON processing
# - build-base: gcc, make, etc. for cgo
RUN apk add --no-cache \
    git \
    jq \
    build-base

# Set working directory for CI
WORKDIR /workspace

# Configure git to trust the workspace directory (fixes VCS status errors in mounted volumes)
RUN git config --global --add safe.directory /workspace

# Verify installations
RUN go version && git --version && jq --version
