#!/bin/bash
# MkDocs Development Server
# MkDocs 1.2.2+ uses polling observer by default, which works with Docker volumes.
# We just need to start MkDocs and let it handle file watching internally.

set -e

DOCS_DIR="${DOCS_DIR:-/workspace/docs}"
CONFIG_FILE="${CONFIG_FILE:-/docs/mkdocs.yml}"
DEV_ADDR="${DEV_ADDR:-0.0.0.0:8000}"

echo "=== MkDocs Dev Server ==="
echo "Docs directory: $DOCS_DIR"
echo "Config file: $CONFIG_FILE"
echo ""

# Validate docs directory exists
if [ ! -d "$DOCS_DIR" ]; then
    echo "ERROR: Docs directory not found: $DOCS_DIR"
    exit 1
fi

# Validate config file exists
if [ ! -f "$CONFIG_FILE" ]; then
    echo "ERROR: MkDocs config not found: $CONFIG_FILE"
    exit 1
fi

# Create wrapper config if using repo config
EFFECTIVE_CONFIG="$CONFIG_FILE"

if [[ "$CONFIG_FILE" == /workspace/* ]]; then
    echo "Using repository mkdocs.yml with container path overrides..."
    cat > /tmp/mkdocs-dev.yml << EOF
INHERIT: $CONFIG_FILE
docs_dir: $DOCS_DIR
dev_addr: '0.0.0.0:8000'
EOF
    EFFECTIVE_CONFIG="/tmp/mkdocs-dev.yml"
    echo "Created wrapper config at $EFFECTIVE_CONFIG"
fi

echo ""
echo "Starting MkDocs with built-in polling file watcher..."
echo "Changes will be detected automatically (no restart needed)"
echo ""

# Run MkDocs serve - it handles file watching via polling internally
# The --dirty flag enables incremental rebuilds (only changed files)
exec mkdocs serve -f "$EFFECTIVE_CONFIG" --dev-addr="$DEV_ADDR" --dirty
