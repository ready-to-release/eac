# MkDocs Docker Container
# Environment: mkdocs-docker
# Purpose: Serve project documentation using MkDocs Material theme

FROM python:3.14-slim

# Set working directory
WORKDIR /docs

# Copy requirements file
COPY requirements.txt /tmp/requirements.txt

# Install MkDocs and dependencies from requirements.txt
RUN pip install --no-cache-dir -r /tmp/requirements.txt && \
    rm /tmp/requirements.txt

# Expose port for MkDocs server
EXPOSE 8000

# Note: User is specified at runtime via --user flag for Docker-in-Docker compatibility
# The /docs directory is a volume mount, so permissions come from the host

# Default command: serve documentation
CMD ["mkdocs", "serve", "--dev-addr=0.0.0.0:8000"]
