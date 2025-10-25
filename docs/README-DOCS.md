# Documentation Structure

This directory is the root of the MkDocs documentation site.

## Structure

```text
docs/
├── index.md                    # Homepage
├── assets/                     # Binary files ONLY (GIFs, images, PDFs)
│   └── .gitkeep
├── guide/                      # User guides
│   └── vscode-extension/
│       ├── index.md            # Extension usage and development
│       └── recording-demos.md  # How to create demo videos
├── tutorials/                  # Step-by-step tutorials
├── reference/                  # API and technical reference
└── development/                # Contributing and development docs
```

## Rules

### 1. docs/assets/ is for Binary Files ONLY

**DO:**

- ✅ Place `.gif`, `.png`, `.jpg`, `.svg` files here
- ✅ Place `.pdf`, `.zip` files here
- ✅ Keep this folder for visual assets only

**DON'T:**

- ❌ Put `.md` files in assets/
- ❌ Put text files in assets/
- ❌ Put documentation in assets/

### 2. Markdown Files Go in Proper Directories

| Content Type | Directory |
|--------------|-----------|
| User guides | `docs/guide/` |
| Tutorials | `docs/tutorials/` |
| API reference | `docs/reference/` |
| Contributing | `docs/development/` |

### 3. Navigation Structure

All navigation is defined in `mkdocs.yml` at the project root.

To add a new page:

1. Create the markdown file in the appropriate directory
2. Add it to `mkdocs.yml` nav section:

```yaml
nav:
  - Guides:
    - Your New Guide: guide/your-new-guide.md
```

## Building the Docs

### Install MkDocs

```bash
# Install MkDocs with Material theme
pip install mkdocs-material

# Optional plugins
pip install mkdocs-git-revision-date-localized-plugin
```

### Serve Locally

```bash
# From project root
mkdocs serve

# Open: http://127.0.0.1:8000
```

### Build Static Site

```bash
mkdocs build

# Output: site/ directory
```

## Adding Content

### New Guide

```bash
# Create file
touch docs/guide/my-new-guide.md

# Edit content
# Add to mkdocs.yml nav
```

### New Tutorial

```bash
# Create file
touch docs/tutorials/my-tutorial.md

# Edit content
# Add to mkdocs.yml nav
```

### Adding Images

```bash
# 1. Place image in assets/
cp myimage.png docs/assets/

# 2. Reference in markdown:
![Description](../assets/myimage.png)

# Or with HTML for sizing:
<img src="../assets/myimage.png" width="600">
```

## Content Guidelines

### File Names

- Use lowercase
- Use hyphens, not underscores: `my-guide.md` not `my_guide.md`
- Be descriptive: `vscode-extension-setup.md` not `setup.md`

### Markdown Style

```markdown
# Page Title (H1) - Only one per page

## Section (H2)

### Subsection (H3)

**Bold** for emphasis
`code` for inline code
[Link text](url)

\`\`\`bash
# Code blocks with language
echo "hello"
\`\`\`
```

### Code Blocks

Always specify the language:

```bash
./script.sh
```

```typescript
function hello() {
  console.log('world');
}
```

```go
func main() {
  fmt.Println("hello")
}
```

### Admonitions

Use MkDocs admonitions for notes/warnings:

```markdown
!!! note
    This is a note

!!! warning
    This is a warning

!!! tip
    This is a helpful tip
```

## Current Pages

### Existing

- ✅ `docs/index.md` - Homepage
- ✅ `docs/guide/vscode-extension/index.md` - Extension guide
- ✅ `docs/guide/vscode-extension/recording-demos.md` - Recording guide

### To Create

- ⏳ `docs/installation.md` - Installation instructions
- ⏳ `docs/guide/mcp-servers.md` - MCP server guide
- ⏳ `docs/guide/automation.md` - Automation scripts guide
- ⏳ `docs/reference/structure.md` - Project structure
- ⏳ `docs/reference/configuration.md` - Configuration reference
- ⏳ `docs/reference/mcp-api.md` - MCP API reference
- ⏳ `docs/reference/extension-api.md` - Extension API
- ⏳ `docs/development/contributing.md` - Contributing guide
- ⏳ `docs/development/architecture.md` - Architecture overview
- ⏳ `docs/development/testing.md` - Testing guide

## Assets to Create

Binary files needed in `docs/assets/`:

- ⏳ `quick-start-guide.gif` (~3MB) - Quick start demo
- ⏳ `git-commit-demo.gif` (~2MB) - Git commit demo
- ⏳ `mcp-server-test.gif` (~2MB) - MCP server test demo

See `docs/guide/vscode-extension/recording-demos.md` for recording instructions.

## MkDocs Configuration

Main configuration in `mkdocs.yml`:

```yaml
site_name: CLI Project Documentation
theme:
  name: material
  features:
    - navigation.tabs
    - navigation.sections
    - search.suggest
nav:
  - Home: index.md
  - Guides: ...
```

## Deployment

### GitHub Pages

```bash
# Deploy to gh-pages branch
mkdocs gh-deploy
```

### Manual Deployment

```bash
# Build site
mkdocs build

# Upload site/ directory to your hosting
```

## Testing

Before committing:

```bash
# Check for broken links
mkdocs build --strict

# Preview locally
mkdocs serve

# Check navigation works
# Check images load
# Check code blocks render
```

## Questions?

- MkDocs docs: <https://www.mkdocs.org/>
- Material theme: <https://squidfunk.github.io/mkdocs-material/>
- Markdown guide: <https://www.markdownguide.org/>
