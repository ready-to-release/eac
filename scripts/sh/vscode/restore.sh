#!/bin/bash
# Restore dependencies for the VSCode extension

set -e  # Exit on error

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "=================================="
echo "Restoring VSCode Extension"
echo "=================================="
echo ""

cd "$PROJECT_ROOT/.vscode/extensions/vscode-ext-commit"

# Check if node_modules exists
if [ -d "node_modules" ]; then
    echo "Removing existing node_modules..."
    rm -rf node_modules
fi

# Check if package-lock.json exists
if [ -f "package-lock.json" ]; then
    echo "Removing existing package-lock.json..."
    rm -f package-lock.json
fi

# Check if out directory exists
if [ -d "out" ]; then
    echo "Removing existing build artifacts..."
    rm -rf out
fi

echo ""
echo "Installing fresh dependencies..."
npm install

echo ""
echo "Compiling TypeScript..."
npm run compile

echo ""
echo "=================================="
echo "✓ Restore Complete!"
echo "=================================="
echo ""
echo "The extension has been restored with fresh dependencies."
echo ""
