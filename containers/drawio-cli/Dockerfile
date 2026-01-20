# Drawio CLI Container
# Extracts embedded XML from drawio.png files and converts between formats
# Used by eac-commands for drawio diagram manipulation

FROM python:3.12-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy extraction scripts
COPY extract.py /app/extract.py
COPY drawio_cli.py /app/drawio_cli.py

# Make scripts executable
RUN chmod +x /app/extract.py /app/drawio_cli.py

# Set working directory for runtime (workspace gets mounted here)
WORKDIR /docs

# Add scripts to PATH
ENV PATH="/app:$PATH"

# Default command shows help
CMD ["python", "/app/drawio_cli.py", "--help"]
