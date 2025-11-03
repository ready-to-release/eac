#!/bin/bash
# Clean all build artifacts and dependencies

set -e  # Exit on error

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "=================================="
echo "Cleaning Build Artifacts"
echo "=================================="
echo ""

# Clean VSCode extension
echo "Cleaning VSCode extension..."
cd "$PROJECT_ROOT/.vscode/extensions/vscode-ext-commit"

CLEANED=false

if [ -d "node_modules" ]; then
    echo "  Removing node_modules..."
    rm -rf node_modules
    CLEANED=true
fi

if [ -d "out" ]; then
    echo "  Removing out directory..."
    rm -rf out
    CLEANED=true
fi

if [ -f "package-lock.json" ]; then
    echo "  Removing package-lock.json..."
    rm -f package-lock.json
    CLEANED=true
fi

if [ -d ".vscode-test" ]; then
    echo "  Removing .vscode-test..."
    rm -rf .vscode-test
    CLEANED=true
fi

if [ -f "*.vsix" ]; then
    echo "  Removing packaged extensions..."
    rm -f *.vsix
    CLEANED=true
fi

if [ -f "*.tsbuildinfo" ]; then
    echo "  Removing TypeScript build info..."
    rm -f *.tsbuildinfo
    CLEANED=true
fi

if [ "$CLEANED" = false ]; then
    echo "  Already clean!"
fi

echo ""
echo "Cleaning MCP server build artifacts..."
cd "$PROJECT_ROOT"

MCP_CLEANED=false

for server in pwsh docs github vscode; do
    SERVER_DIR="$PROJECT_ROOT/src/mcp/$server"
    if [ -d "$SERVER_DIR" ]; then
        cd "$SERVER_DIR"

        # Remove compiled binaries
        if ls *.exe >/dev/null 2>&1 || ls mcp-server-* >/dev/null 2>&1; then
            echo "  Removing $server binaries..."
            rm -f *.exe mcp-server-* 2>/dev/null || true
            MCP_CLEANED=true
        fi

        # Remove test binaries
        if ls *.test >/dev/null 2>&1; then
            echo "  Removing $server test binaries..."
            rm -f *.test 2>/dev/null || true
            MCP_CLEANED=true
        fi
    fi
done

if [ "$MCP_CLEANED" = false ]; then
    echo "  Already clean!"
fi

echo ""
echo "=================================="
echo "✓ Clean Complete!"
echo "=================================="
echo ""
echo "All build artifacts and dependencies have been removed."
echo "Run './automation/sh/vscode/init.sh' to restore and rebuild."
echo ""
