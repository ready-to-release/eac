# MkDocs Mermaid Container
# Lightweight container for mermaid-cli (mmdc) diagram rendering
# Used by eac-commands for pre-rendering mermaid diagrams to SVG

ARG NODE_VERSION=25
FROM node:${NODE_VERSION}-slim

# Install mermaid-cli to /opt/mermaid (NOT /docs which gets volume-mounted at runtime)
WORKDIR /opt/mermaid

# Install Chromium and dependencies for mermaid-cli
RUN apt-get update && apt-get install -y --no-install-recommends \
    chromium \
    # Fonts for proper text rendering
    fonts-liberation \
    fonts-dejavu-core \
    # Required libraries for Chromium
    libnss3 \
    libnspr4 \
    libatk1.0-0 \
    libatk-bridge2.0-0 \
    libcups2 \
    libdrm2 \
    libxkbcommon0 \
    libxcomposite1 \
    libxdamage1 \
    libxfixes3 \
    libxrandr2 \
    libgbm1 \
    libasound2 \
    && rm -rf /var/lib/apt/lists/*

# Install mermaid-cli from package.json
COPY package.json .
RUN npm install && npm cache clean --force

# Add mermaid-cli binaries to PATH (in /opt/mermaid, not /docs which gets mounted)
ENV PATH="/opt/mermaid/node_modules/.bin:$PATH"

# Set working directory for runtime (this is where workspace gets mounted)
WORKDIR /docs

# Chromium wrapper with container-safe flags
# NOTE: Do not redirect stderr - Puppeteer needs it to get the WebSocket debugging port
RUN printf '#!/bin/sh\nexec /usr/bin/chromium \\\n  --no-sandbox \\\n  --disable-setuid-sandbox \\\n  --disable-dev-shm-usage \\\n  --disable-gpu \\\n  --disable-software-rasterizer \\\n  --disable-extensions \\\n  --disable-features=TranslateUI \\\n  --js-flags="--max-old-space-size=2048" \\\n  "$@"\n' > /usr/local/bin/chromium-wrapper \
    && chmod +x /usr/local/bin/chromium-wrapper

# Copy configuration files
COPY puppeteer-config.json /etc/mermaid/puppeteer-config.json
COPY mermaid-config.json /etc/mermaid/mermaid-config.json

# Copy batch rendering script
COPY batch-render.js /app/batch-render.js

# Set environment for puppeteer to use our wrapper
ENV PUPPETEER_EXECUTABLE_PATH=/usr/local/bin/chromium-wrapper
ENV PUPPETEER_SKIP_DOWNLOAD=true

# Default command shows version info
CMD ["mmdc", "--version"]
