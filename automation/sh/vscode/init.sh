#!/bin/bash
# Initialize the VSCode extension and MCP servers

set -e  # Exit on error

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "=================================="
echo "Initializing CLI Project"
echo "=================================="
echo ""

# Check prerequisites
echo "Checking prerequisites..."
echo ""

# Check Go
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed"
    echo "   Please install Go 1.21 or higher from https://golang.org/dl/"
    exit 1
fi
echo "✓ Go $(go version | awk '{print $3}')"

# Check Node.js
if ! command -v node &> /dev/null; then
    echo "❌ Error: Node.js is not installed"
    echo "   Please install Node.js 18.x or higher from https://nodejs.org/"
    exit 1
fi
echo "✓ Node.js $(node --version)"

# Check npm
if ! command -v npm &> /dev/null; then
    echo "❌ Error: npm is not installed"
    echo "   npm should come with Node.js"
    exit 1
fi
echo "✓ npm $(npm --version)"

echo ""
echo "=================================="
echo "Installing VSCode Extension"
echo "=================================="
echo ""

cd "$PROJECT_ROOT/.vscode/extensions/claude-mcp-vscode"

echo "Installing dependencies..."
npm install

echo ""
echo "Compiling TypeScript..."
npm run compile

echo ""
echo "=================================="
echo "Verifying MCP Servers"
echo "=================================="
echo ""

# Check if Go modules are initialized
for server in pwsh docs github vscode; do
    echo "Checking $server server..."
    cd "$PROJECT_ROOT/src/mcp/$server"

    if [ ! -f "go.mod" ]; then
        echo "  ❌ go.mod not found"
        continue
    fi

    if [ ! -f "main.go" ]; then
        echo "  ❌ main.go not found"
        continue
    fi

    # Try to build (but don't keep the binary)
    if go build -o /tmp/mcp-test-$server 2>&1 | grep -q "error"; then
        echo "  ⚠️  Build warnings/errors (check manually)"
    else
        echo "  ✓ $server server ready"
        rm -f /tmp/mcp-test-$server
    fi
done

echo ""
echo "=================================="
echo "✓ Initialization Complete!"
echo "=================================="
echo ""
echo "Next steps:"
echo "1. Open this workspace in VSCode"
echo "2. Press F5 to launch the Extension Development Host"
echo "3. Open the Source Control view (Ctrl+Shift+G)"
echo "4. Click the robot icon button in the toolbar"
echo ""
