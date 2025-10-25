#!/bin/bash
# Build static MkDocs site

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "================================================"
echo "Building MkDocs Static Site"
echo "================================================"
echo ""

cd "$SCRIPT_DIR"

# Build with strict mode
docker-compose run --rm mkdocs mkdocs build --clean --strict

echo ""
echo "================================================"
echo "✓ Build Complete!"
echo "================================================"
echo ""
echo "Static site created at:"
echo "  $PROJECT_ROOT/site/"
echo ""
echo "To preview:"
echo "  cd site && python -m http.server 8000"
echo ""
