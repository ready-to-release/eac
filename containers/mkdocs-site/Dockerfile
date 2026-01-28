# MkDocs Site Container (HTML only)
# Lean image for building documentation sites
# ~150MB vs ~800MB for full PDF container

ARG PYTHON_VERSION=3.14
FROM python:${PYTHON_VERSION}-alpine

WORKDIR /docs

# Minimal dependencies for mkdocs + material theme
RUN apk add --no-cache \
    git \
    && rm -rf /var/cache/apk/*

# Install Python packages
COPY requirements.txt /tmp/requirements.txt
RUN pip install --no-cache-dir -r /tmp/requirements.txt \
    && rm /tmp/requirements.txt \
    && rm -rf ~/.cache/pip

# Copy mkdocs config and macros
COPY mkdocs.yml /docs/mkdocs.yml
COPY mkdocs_macros.py /docs/main.py

EXPOSE 8000

CMD ["mkdocs", "serve", "--dev-addr=0.0.0.0:8000"]
