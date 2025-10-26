# Documentation Structure

This document explains the organization of documentation in this project.

## Overview

Documentation is organized using **MkDocs** with the Material theme. The structure follows best practices with `/docs` as the documentation root.

## Directory Structure

```text
cli/
├── mkdocs.yml                           # MkDocs configuration
│
├── docs/                                # Documentation root (MkDocs)
│   ├── index.md                         # Documentation homepage
│   ├── README-DOCS.md                   # Documentation guide (this structure)
│   │
│   ├── assets/                          # Binary assets ONLY
│   │   ├── .gitkeep                     # Git tracking
│   │   └── (GIF files to be created)
│   │
│   ├── guide/                           # User guides
│   │   └── vscode-extension/
│   │       ├── index.md                 # Extension guide
│   │       └── recording-demos.md       # Recording instructions
│   │
│   ├── tutorials/                       # Step-by-step tutorials
│   ├── reference/                       # Technical reference
│   └── development/                     # Contributing docs
│
├── QUICKSTART.md                        # Quick start (root level)
├── USAGE.md                             # Usage guide (root level)
├── README.md                            # Project README (root level)
└── CLAUDE.md                            # Claude Code instructions

automation/sh-vscode/
└── README.md                            # Automation scripts docs
```

## Key Principles

### 1. Binary Files Go in docs/assets/

**docs/assets/** is for binary files ONLY:

- ✅ Images (.png, .jpg, .gif, .svg)
- ✅ PDFs
- ✅ Videos (if small)
- ❌ NO .md files
- ❌ NO .txt files
- ❌ NO documentation

### 2. Markdown Files Go in Proper Directories

| Type          | Directory         | Example                                |
| ------------- | ----------------- | -------------------------------------- |
| User guides   | docs/guide/       | `docs/guide/vscode-extension/index.md`       |
| Tutorials     | docs/tutorials/   | `docs/tutorials/building-mcp-tools.md` |
| API reference | docs/reference/   | `docs/reference/mcp-api.md`            |
| Contributing  | docs/development/ | `docs/development/contributing.md`     |

### 3. Root-Level Docs for Quick Access

Important docs at root for easy access:

- `README.md` - Project overview (GitHub front page)
- `QUICKSTART.md` - 5-minute quick start
- `USAGE.md` - Complete usage examples
- `CLAUDE.md` - Claude Code instructions

## MkDocs Configuration

### mkdocs.yml

Main configuration file at project root:

```yaml
site_name: CLI Project Documentation
site_url: https://example.com

theme:
  name: material
  palette:
    - scheme: default  # Light mode
    - scheme: slate    # Dark mode
  features:
    - navigation.tabs
    - navigation.sections
    - navigation.expand
    - search.suggest
    - content.code.copy

nav:
  - Home: index.md
  - Getting Started:
    - Quick Start: ../QUICKSTART.md
    - Installation: installation.md
  - Guides:
    - VSCode Extension:
      - Overview: guide/vscode-extension/index.md
      - Recording Demos: guide/vscode-extension/recording-demos.md
```

## Content Organization

### VSCode Extension Documentation

Location: `docs/guide/vscode-extension/`

**index.md** - Main guide:

- Overview and quick start
- Architecture and how it works
- Extension development
- Customization
- Troubleshooting
- Testing and packaging

**recording-demos.md** - Recording guide:

- Tool recommendations
- Environment setup
- Scene-by-scene storyboards
- Recording process
- Optimization tips
- Troubleshooting

### Assets

Location: `docs/assets/`

**To be created:**

- `quick-start-guide.gif` (~3MB) - Complete setup demo
- `git-commit-demo.gif` (~2MB) - Button click workflow
- `mcp-server-test.gif` (~2MB) - Terminal MCP test

**Recording instructions:** See `docs/guide/vscode-extension/recording-demos.md`

## Building Documentation

### Install Dependencies

```bash
pip install mkdocs-material
pip install mkdocs-git-revision-date-localized-plugin
```

### Local Development

```bash
# Serve with live reload
mkdocs serve

# Open: http://127.0.0.1:8000
```

### Build Static Site

```bash
# Build to site/
mkdocs build

# Build with strict mode (fail on warnings)
mkdocs build --strict
```

### Deploy to GitHub Pages

```bash
mkdocs gh-deploy
```

## Navigation Structure

The navigation is hierarchical:

```text
Home
├── Getting Started
│   ├── Quick Start
│   ├── Installation
│   └── Usage Guide
│
├── Guides
│   ├── VSCode Extension
│   │   ├── Overview
│   │   └── Recording Demos
│   ├── MCP Servers
│   └── Automation Scripts
│
├── Tutorials
│   ├── Building MCP Tools
│   ├── Custom VSCode Commands
│   └── Extending Automation
│
├── Reference
│   ├── Project Structure
│   ├── Configuration
│   ├── MCP Server API
│   └── Extension API
│
└── Development
    ├── Contributing
    ├── Architecture
    ├── Testing
    └── Release Process
```

## Adding New Content

### New Guide Page

```bash
# 1. Create file
touch docs/guide/my-new-guide.md

# 2. Add content with front matter
cat > docs/guide/my-new-guide.md <<'EOF'
# My New Guide

Introduction to the guide...

## Section 1
...
EOF

# 3. Add to mkdocs.yml nav:
#    - Guides:
#      - My New Guide: guide/my-new-guide.md

# 4. Test
mkdocs serve
```

### New Tutorial

```bash
# 1. Create file
touch docs/tutorials/my-tutorial.md

# 2. Add content
# 3. Add to mkdocs.yml
# 4. Test with mkdocs serve
```

### Adding Images

```bash
# 1. Place in assets/
cp myimage.png docs/assets/

# 2. Reference in markdown
![Alt text](../assets/myimage.png)

# Or with HTML for sizing
<img src="../assets/myimage.png" width="600" alt="Description">
```

## Content Guidelines

### File Naming

- Lowercase: `my-guide.md` ✅
- Hyphens not underscores: `my-guide.md` not `my_guide.md` ✅
- Descriptive: `vscode-extension-setup.md` not `setup.md` ✅

### Markdown Formatting

```markdown
# Page Title (H1) - One per page

## Major Section (H2)

### Subsection (H3)

#### Minor Subsection (H4)

**Bold** for emphasis
*Italic* for terminology
`code` for inline code
```

### Code Blocks

Always specify language:

```bash
./script.sh
```

```typescript
function hello() {
  console.log('world');
}
```

### Admonitions

```markdown
!!! note
    Important information

!!! warning
    Be careful here

!!! tip
    Helpful hint

!!! danger
    Critical warning
```

## Current Status

### ✅ Complete

- MkDocs configuration (mkdocs.yml)
- Documentation structure (/docs)
- Homepage (docs/index.md)
- VSCode extension guide (docs/guide/vscode-extension/index.md)
- Recording guide (docs/guide/vscode-extension/recording-demos.md)
- Assets directory structure
- Documentation guide (docs/README-DOCS.md)

### ⏳ To Create

- Installation guide (docs/installation.md)
- MCP servers guide (docs/guide/mcp-servers.md)
- Automation guide (docs/guide/automation.md)
- Tutorial pages (docs/tutorials/)
- Reference pages (docs/reference/)
- Development pages (docs/development/)
- Demo GIF files (docs/assets/\*.gif)

## Migration Notes

### What Changed

**Old Structure:**

```text
docs/
├── assets/
│   ├── README.md                   # ❌ Removed
│   ├── RECORDING-GUIDE.md          # ❌ Removed
│   ├── quick-start-guide.md        # ❌ Removed
│   └── (many other .md files)      # ❌ Removed
└── README.md                       # ❌ Removed
```

**New Structure:**

```text
docs/
├── index.md                        # ✅ New homepage
├── README-DOCS.md                  # ✅ Structure guide
├── assets/                         # ✅ Binary only
│   └── .gitkeep
└── guide/
    └── vscode-extension/
        ├── index.md                # ✅ Main guide
        └── recording-demos.md      # ✅ Recording instructions
```

### Content Consolidation

All recording-related content was consolidated into:

- `docs/guide/vscode-extension/recording-demos.md`

All extension usage and development content was consolidated into:

- `docs/guide/vscode-extension/index.md`

## Resources

**MkDocs:**

- Main site: <https://www.mkdocs.org/>
- Material theme: <https://squidfunk.github.io/mkdocs-material/>
- User guide: <https://www.mkdocs.org/user-guide/>

**Writing Documentation:**

- Markdown guide: <https://www.markdownguide.org/>
- Material reference: <https://squidfunk.github.io/mkdocs-material/reference/>

## Questions?

See `docs/README-DOCS.md` for more details on contributing to documentation.
