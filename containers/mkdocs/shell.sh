#!/bin/bash
# Open interactive shell in MkDocs container

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "================================================"
echo "Opening MkDocs Container Shell"
echo "================================================"
echo ""
echo "Available commands:"
echo "  mkdocs serve    - Start development server"
echo "  mkdocs build    - Build static site"
echo "  mkdocs --help   - Show help"
echo ""
echo "Type 'exit' to leave the container"
echo ""

cd "$SCRIPT_DIR"
docker-compose run --rm --entrypoint /bin/bash mkdocs
