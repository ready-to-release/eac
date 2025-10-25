# MkDocs MCP Server

A Model Context Protocol server for accessing, searching, and managing MkDocs documentation with Docker container integration.

## Features

- **Search documentation content** - Full-text search across all docs
- **Navigate documentation structure** - List and access pages
- **Retrieve specific pages** - Get content of any documentation page
- **Build documentation** - Build static site using Docker container
- **Serve documentation** - Start development server with live reload
- **Manage documentation server** - Start/stop the MkDocs container

## Configuration

The server is configured in `.mcp.json` as:

```json
{
  "docs": {
    "command": "bash",
    "args": ["src/mcp/docs/run.sh"],
    "env": {
      "DOCS_PATH": "${DOCS_PATH:-.}"
    }
  }
}
```

## Environment Variables

- `DOCS_PATH` - Path to MkDocs root directory (defaults to current directory)

## Tools Provided

### Documentation Access

#### `search-docs`
Full-text search across all documentation files.

**Arguments:**
- `query` (string, required) - Search query

**Example:**
```json
{
  "name": "search-docs",
  "arguments": {
    "query": "authentication"
  }
}
```

**Returns:** List of matching files with context snippets.

#### `get-doc-page`
Retrieve the content of a specific documentation page.

**Arguments:**
- `path` (string, required) - Relative path to documentation page

**Example:**
```json
{
  "name": "get-doc-page",
  "arguments": {
    "path": "guide/vscode-ext-claude-commit/index.md"
  }
}
```

**Returns:** Full content of the requested page.

#### `list-docs`
List all available documentation pages.

**Arguments:** None

**Example:**
```json
{
  "name": "list-docs",
  "arguments": {}
}
```

**Returns:** Array of all documentation file paths.

### Docker Container Integration

#### `build-docs`
Build the static documentation site using the MkDocs Docker container.

**Arguments:** None

**Example:**
```json
{
  "name": "build-docs",
  "arguments": {}
}
```

**Returns:** Build output and success message. Creates static site in `site/` directory.

**What it does:**
1. Uses Docker Compose to run MkDocs container
2. Executes `mkdocs build --clean --strict`
3. Generates static HTML in `site/` directory
4. Returns build output

#### `serve-docs`
Start the MkDocs development server with live reload using Docker container.

**Arguments:**
- `detached` (boolean, optional) - Run in background (default: true)

**Example:**
```json
{
  "name": "serve-docs",
  "arguments": {
    "detached": true
  }
}
```

**Returns:** Server start message with URL.

**What it does:**
1. Starts Docker container with MkDocs
2. Launches development server on http://localhost:8000
3. Enables live reload for documentation changes
4. Runs in background by default

**Access:** http://localhost:8000

#### `stop-docs`
Stop the MkDocs development server.

**Arguments:** None

**Example:**
```json
{
  "name": "stop-docs",
  "arguments": {}
}
```

**Returns:** Stop confirmation message.

**What it does:**
1. Stops the Docker container
2. Removes the container
3. Frees up port 8000

## Usage

### Via Claude Code

**Search documentation:**
```
Search the docs for information about VSCode extension
```

**Get specific page:**
```
Show me the VSCode extension guide
```

**Build documentation:**
```
Build the MkDocs documentation
```

**Start documentation server:**
```
Start the documentation server
```

**Stop documentation server:**
```
Stop the documentation server
```

### Via JSON-RPC

**Test search:**
```bash
cd src/mcp/docs
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"search-docs","arguments":{"query":"extension"}}}' | go run .
```

**Test build:**
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"build-docs","arguments":{}}}' | go run .
```

**Test serve:**
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"serve-docs","arguments":{"detached":true}}}' | go run .
```

## Docker Container Requirements

The Docker container tools require:

- **Docker** 20.10+
- **Docker Compose** 2.0+
- **Container location**: `containers/mkdocs/`

The MkDocs container provides:
- MkDocs 1.5.3 with Material theme
- All required plugins
- Live reload development server
- Static site builder

See [containers/mkdocs/README.md](../../../containers/mkdocs/README.md) for container details.

## Implementation

✅ **GO IMPLEMENTATION** - Working MCP server written in Go.

### Architecture

```
MCP Server (Go)
├── Documentation Access Tools
│   ├── search-docs        → Search markdown files
│   ├── get-doc-page       → Read markdown content
│   └── list-docs          → List all pages
└── Docker Integration Tools
    ├── build-docs         → docker-compose run mkdocs build
    ├── serve-docs         → docker-compose up -d
    └── stop-docs          → docker-compose down
```

### How Docker Integration Works

1. **Build Tool:**
   ```bash
   cd containers/mkdocs
   docker-compose run --rm mkdocs mkdocs build --clean --strict
   ```

2. **Serve Tool:**
   ```bash
   cd containers/mkdocs
   docker-compose up -d
   ```

3. **Stop Tool:**
   ```bash
   cd containers/mkdocs
   docker-compose down
   ```

All Docker commands are executed from the project root using relative paths.

## Development

### Run Locally

```bash
cd src/mcp/docs
go run .
```

### Test Initialization

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | go run .
```

### Test Tools List

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}
{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | go run .
```

### Test Search

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"search-docs","arguments":{"query":"docker"}}}' | go run .
```

### Test Build

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"build-docs","arguments":{}}}' | go run .
```

### Test Serve

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"serve-docs","arguments":{"detached":true}}}' | go run .

# Then access: http://localhost:8000
```

### Test Stop

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"stop-docs","arguments":{}}}' | go run .
```

## Use Cases

### Documentation Workflow

1. **Write documentation:**
   - Edit markdown files in `docs/`

2. **Preview locally:**
   - Use `serve-docs` tool
   - Access http://localhost:8000
   - See changes live

3. **Build for production:**
   - Use `build-docs` tool
   - Static site created in `site/`

4. **Deploy:**
   - Upload `site/` directory
   - Or use `mkdocs gh-deploy`

### Search and Discovery

1. **Find information:**
   - Use `search-docs` with query
   - Get relevant files and snippets

2. **Read specific pages:**
   - Use `get-doc-page` with path
   - Get full markdown content

3. **Navigate structure:**
   - Use `list-docs`
   - See all available pages

## Troubleshooting

### Docker Container Not Found

**Error:** `Error building docs: docker-compose not found`

**Solution:**
```bash
# Check Docker is installed
docker --version
docker-compose --version

# Check container exists
ls containers/mkdocs/
```

### Port Already in Use

**Error:** `port 8000 already in use`

**Solution:**
```bash
# Stop existing server
# Use stop-docs tool or:
cd containers/mkdocs
docker-compose down

# Or use different port in docker-compose.yml
```

### Build Fails

**Error:** `Error building docs: mkdocs build failed`

**Solution:**
```bash
# Check mkdocs.yml syntax
cd containers/mkdocs
docker-compose run --rm mkdocs mkdocs build --verbose

# Check for broken links
docker-compose run --rm mkdocs mkdocs build --strict --verbose
```

### Permission Errors

**Error:** `permission denied`

**Solution:**
```bash
# Check file permissions
ls -la docs/

# Fix permissions
chmod -R 755 docs/
```

## Performance

- **Search:** Fast full-text search across all markdown files
- **Build:** Typically 5-30 seconds depending on documentation size
- **Serve:** Instant startup with live reload
- **Docker:** ~200MB container with all dependencies

## Security

- Runs as non-root user in container (UID 1000)
- No cache poisoning (PIP_NO_CACHE_DIR=1)
- Read-only access to documentation files
- Container isolation for safe execution

## Future Enhancements

Potential additions:
- [ ] Deploy to GitHub Pages
- [ ] Generate PDF documentation
- [ ] Validate links automatically
- [ ] Track documentation metrics
- [ ] Multi-version documentation
- [ ] Search result ranking
- [ ] Code example testing

## Resources

- [MkDocs Documentation](https://www.mkdocs.org/)
- [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/)
- [Model Context Protocol](https://modelcontextprotocol.io/)
- [Container Setup](../../../containers/mkdocs/README.md)
