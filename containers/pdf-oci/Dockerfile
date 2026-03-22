# MkDocs PDF Container
# Full image with PDF generation support via Playwright (mkdocs-exporter)
# Includes Chromium for mermaid-to-svg pre-rendering

ARG PYTHON_VERSION=3.14
FROM python:${PYTHON_VERSION}-slim

WORKDIR /docs

# System dependencies for PDF generation
RUN apt-get update && apt-get install -y --no-install-recommends \
    # Fonts - Noto family only (sans + mono + emoji)
    fontconfig \
    fonts-noto-core \
    fonts-noto-mono \
    fonts-noto-color-emoji \
    # Build tools
    gcc \
    python3-dev \
    # Image handling
    libjpeg62-turbo \
    libopenjp2-7 \
    # Node.js + Chromium for mermaid-cli
    nodejs \
    npm \
    chromium \
    # Playwright dependencies
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

# Install mermaid-cli for SVG pre-rendering
RUN npm install -g @mermaid-js/mermaid-cli@11.4.2 \
    && npm cache clean --force

# Chromium wrapper with container-safe flags (for mermaid-cli)
# NOTE: Do not redirect stderr - Puppeteer needs it to get the WebSocket debugging port
# NOTE: --single-process removed - causes instability under parallel load
# Performance flags added for PDF generation optimization
# Chromium uses /dev/shm for shared memory - we provide 1GB via --shm-size in tool-config.yml
# Removed --disable-dev-shm-usage to use the ramdisk properly
RUN printf '#!/bin/sh\nexec /usr/bin/chromium \\\n  --no-sandbox \\\n  --disable-setuid-sandbox \\\n  --disable-gpu \\\n  --disable-software-rasterizer \\\n  --disable-extensions \\\n  --disable-features=TranslateUI \\\n  --disable-background-timer-throttling \\\n  --disable-backgrounding-occluded-windows \\\n  --disable-renderer-backgrounding \\\n  --disable-hang-monitor \\\n  --disable-ipc-flooding-protection \\\n  --disable-popup-blocking \\\n  --disable-prompt-on-repost \\\n  --disable-sync \\\n  --metrics-recording-only \\\n  --mute-audio \\\n  --no-first-run \\\n  --safebrowsing-disable-auto-update \\\n  --js-flags="--max-old-space-size=4096" \\\n  "$@"\n' > /usr/local/bin/chromium-wrapper \
    && chmod +x /usr/local/bin/chromium-wrapper

ENV PUPPETEER_EXECUTABLE_PATH=/usr/local/bin/chromium-wrapper
ENV PUPPETEER_SKIP_DOWNLOAD=true
# Suppress mkdocs-material 9.7.2+ warning about MkDocs 2.0 (we pin mkdocs<2.0)
# See: https://github.com/squidfunk/mkdocs-material/issues/8567
ENV NO_MKDOCS_2_WARNING=1

# Install Python packages in stages to manage build layer size
# First: Install playwright separately (large package, 46MB)
RUN pip install --no-cache-dir --root-user-action=ignore playwright==1.58.0

# Then: Install remaining packages (excluding mkdocs-exporter - using fork)
COPY requirements.txt /tmp/requirements.txt
RUN pip install --no-cache-dir --root-user-action=ignore -r /tmp/requirements.txt \
    && rm /tmp/requirements.txt

# Install mkdocs-exporter fork with parallel merge optimization
COPY mkdocs-exporter-fork /tmp/mkdocs-exporter-fork
RUN pip install --no-cache-dir --root-user-action=ignore /tmp/mkdocs-exporter-fork \
    && rm -rf /tmp/mkdocs-exporter-fork

# Copy Material theme overrides for mkdocs-exporter compatibility
# The override sets body[mkdocs-exporter="true"] to signal page ready state
# Store in /opt so it doesn't conflict with mounted /docs volume
COPY overrides /opt/mkdocs-overrides

# Point Playwright at the apt-installed Chromium (already present above).
# This avoids a redundant 100-400 MB chromium-headless-shell download.
# PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH overrides which binary Playwright uses.
ENV PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH=/usr/local/bin/chromium-wrapper

# Use ramdisk for temp files at RUNTIME (PDF aggregation I/O optimization)
# Note: This is set after installs to avoid breaking pip/playwright downloads during build
ENV TMPDIR=/dev/shm

EXPOSE 8000

CMD ["mkdocs", "serve", "--dev-addr=0.0.0.0:8000"]
