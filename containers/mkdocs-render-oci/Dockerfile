# MkDocs Site Container (HTML only)
# Lean image for building documentation sites
# ~150MB vs ~800MB for full PDF container

ARG PYTHON_VERSION=3.14
FROM python:${PYTHON_VERSION}-alpine

WORKDIR /docs

# Suppress mkdocs-material 9.7.2+ warning about MkDocs 2.0 (we pin mkdocs<2.0)
# See: https://github.com/squidfunk/mkdocs-material/issues/8567
ENV NO_MKDOCS_2_WARNING=1

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
