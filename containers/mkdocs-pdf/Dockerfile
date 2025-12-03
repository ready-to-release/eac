# MkDocs PDF Container
# Full image with PDF generation support via WeasyPrint
# Includes Chromium for mermaid-to-svg pre-rendering

FROM python:3.12-slim

WORKDIR /docs

# System dependencies for PDF generation
RUN apt-get update && apt-get install -y --no-install-recommends \
    # WeasyPrint dependencies (Pango, Cairo, GDK-PixBuf)
    libpango-1.0-0 \
    libpangoft2-1.0-0 \
    libpangocairo-1.0-0 \
    libgdk-pixbuf-2.0-0 \
    libcairo2 \
    libffi-dev \
    shared-mime-info \
    # Fonts
    fontconfig \
    fonts-liberation \
    fonts-dejavu-core \
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
    && rm -rf /var/lib/apt/lists/*

# Install mermaid-cli for SVG pre-rendering
RUN npm install -g @mermaid-js/mermaid-cli \
    && npm cache clean --force

# Chromium wrapper with container-safe flags
RUN printf '#!/bin/sh\n/usr/bin/chromium --no-sandbox --disable-setuid-sandbox --disable-dev-shm-usage --disable-gpu "$@" 2>/dev/null\n' > /usr/local/bin/chromium-wrapper \
    && chmod +x /usr/local/bin/chromium-wrapper

ENV PUPPETEER_EXECUTABLE_PATH=/usr/local/bin/chromium-wrapper
ENV PUPPETEER_SKIP_DOWNLOAD=true

# Install Python packages
COPY requirements.txt /tmp/requirements.txt
RUN pip install --no-cache-dir -r /tmp/requirements.txt \
    && rm /tmp/requirements.txt

EXPOSE 8000

CMD ["mkdocs", "serve", "--dev-addr=0.0.0.0:8000"]
