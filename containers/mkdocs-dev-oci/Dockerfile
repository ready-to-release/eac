# MkDocs Development Server with Polling Support
# This container provides live-reload that works on Windows/Docker Desktop
# by using polling-based file watching instead of inotify

ARG PYTHON_VERSION=3.12
FROM python:${PYTHON_VERSION}-alpine

WORKDIR /docs

# Install dependencies
RUN apk add --no-cache \
    git \
    bash \
    && rm -rf /var/cache/apk/*

# Install Python packages (same as mkdocs-render-oci + polling support)
COPY requirements.txt /tmp/requirements.txt
RUN pip install --no-cache-dir -r /tmp/requirements.txt \
    && rm /tmp/requirements.txt \
    && rm -rf ~/.cache/pip

# Copy mkdocs config and macros
COPY mkdocs.yml /docs/mkdocs.yml
COPY mkdocs_macros.py /docs/main.py

# Copy the polling entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 8000

ENTRYPOINT ["/entrypoint.sh"]
