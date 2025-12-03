# MkDocs Site Container (HTML only)
# Lean image for building documentation sites
# ~150MB vs ~800MB for full PDF container

FROM python:3.12-alpine

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

EXPOSE 8000

CMD ["mkdocs", "serve", "--dev-addr=0.0.0.0:8000"]
