# Go Build Container with CGO Support
# Enables CGO_ENABLED=1 and race detection for Go builds/tests

ARG GO_VERSION=1.24
FROM golang:${GO_VERSION}-alpine

WORKDIR /app

# Install CGO dependencies (gcc, musl-dev for static linking)
RUN apk add --no-cache \
    gcc \
    musl-dev \
    git \
    && rm -rf /var/cache/apk/*

# Enable CGO by default
ENV CGO_ENABLED=1

# Pre-create cache directories
RUN mkdir -p /go/pkg /root/.cache/go-build

CMD ["go", "build", "./..."]
