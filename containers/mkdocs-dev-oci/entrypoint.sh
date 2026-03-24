#!/bin/bash
# MkDocs Development Server
# Serves documentation with live-reload via polling-based file watching.
# The staging directory (mounted at /workspace) contains pre-processed docs
# with expanded command markers. Source repo is mounted at /source for config inheritance.

set -e

CONFIG_FILE="${CONFIG_FILE:-/docs/mkdocs.yml}"
DEV_ADDR="${DEV_ADDR:-0.0.0.0:8000}"

echo "=== MkDocs Dev Server ==="
echo "Config: $CONFIG_FILE"
echo ""

# Validate config file exists
if [ ! -f "$CONFIG_FILE" ]; then
    echo "ERROR: MkDocs config not found: $CONFIG_FILE"
    exit 1
fi

echo "Starting MkDocs with live-reload..."
echo ""

# --dirty: incremental rebuilds (only changed files)
# --livereload: explicit flag required due to click 8.3.x bug (mkdocs/mkdocs#4032)
exec mkdocs serve -f "$CONFIG_FILE" --dev-addr="$DEV_ADDR" --dirty --livereload
