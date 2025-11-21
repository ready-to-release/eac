# MkDocs Docker Container
# Environment: mkdocs-docker
# Purpose: Serve project documentation using MkDocs Material theme

FROM python:3.11-slim

# Set working directory
WORKDIR /docs

# Copy requirements file
COPY requirements.txt /tmp/requirements.txt

# Install MkDocs and dependencies from requirements.txt
RUN pip install --no-cache-dir -r /tmp/requirements.txt && \
    rm /tmp/requirements.txt

# Expose port for MkDocs server
EXPOSE 8000

# Default command: serve documentation
CMD ["mkdocs", "serve", "--dev-addr=0.0.0.0:8000"]
