#!/bin/bash
# Serve MkDocs documentation with live reload

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "================================================"
echo "Starting MkDocs Development Server"
echo "================================================"
echo ""
echo "Documentation will be available at:"
echo "  http://localhost:8000"
echo ""
echo "Press Ctrl+C to stop"
echo ""

cd "$SCRIPT_DIR"
docker-compose up
