# Drawio CLI Container
# Extracts embedded XML from drawio.png files and converts between formats
# Used by eac-commands for drawio diagram manipulation

FROM python:3.12-slim

WORKDIR /app

# Install system dependencies for cairosvg
RUN apt-get update && apt-get install -y --no-install-recommends \
    libcairo2 \
    libpango-1.0-0 \
    libpangocairo-1.0-0 \
    libgdk-pixbuf-2.0-0 \
    libffi-dev \
    shared-mime-info \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy all Python scripts
COPY extract.py /app/extract.py
COPY codec.py /app/codec.py
COPY embed.py /app/embed.py
COPY render.py /app/render.py
COPY drawio_cli.py /app/drawio_cli.py

# Make scripts executable
RUN chmod +x /app/extract.py /app/codec.py /app/embed.py /app/render.py /app/drawio_cli.py

# Set working directory for runtime (workspace gets mounted here)
WORKDIR /docs

# Add scripts to PATH
ENV PATH="/app:$PATH"

# Default command shows help
CMD ["python", "/app/drawio_cli.py", "--help"]
