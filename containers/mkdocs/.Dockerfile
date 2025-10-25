# MkDocs Documentation Container
# Provides MkDocs with Material theme and all plugins

FROM python:3.11-slim

LABEL maintainer="CLI Project Team"
LABEL description="MkDocs container with Material theme and plugins"
LABEL version="1.0"

# Set working directory
WORKDIR /docs

# Environment variables
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1

# Install system dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        git \
        openssh-client \
        ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements file
COPY requirements.txt /tmp/requirements.txt

# Install Python dependencies
RUN pip install --no-cache-dir -r /tmp/requirements.txt && \
    rm /tmp/requirements.txt

# Create non-root user
RUN useradd -m -u 1000 -s /bin/bash mkdocs && \
    chown -R mkdocs:mkdocs /docs

# Switch to non-root user
USER mkdocs

# Expose MkDocs development server port
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import http.client; conn = http.client.HTTPConnection('localhost:8000'); conn.request('GET', '/'); r = conn.getresponse(); exit(0 if r.status == 200 else 1)"

# Default command: serve documentation
CMD ["mkdocs", "serve", "--dev-addr=0.0.0.0:8000"]
