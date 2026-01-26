# PDF Tools Container
# Lightweight container for PDF manipulation (page extraction to images)
ARG ALPINE_VERSION=3.20
FROM alpine:${ALPINE_VERSION}

LABEL org.opencontainers.image.source="https://github.com/ready-to-release/eac"
LABEL org.opencontainers.image.description="PDF tools (poppler-utils) for page extraction"

RUN apk add --no-cache \
    poppler-utils \
    libpng \
    libjpeg-turbo-utils

WORKDIR /workspace
