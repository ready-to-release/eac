# CLI Documentation

Welcome to the CLI documentation.

This is an extensible CLI-in-a-box for managing software delivery flows with integrated MCP servers and VSCode extension.

---

## Diataxis Navigation

Documentation is organized using the [Diataxis framework](https://diataxis.fr/):

- **[Tutorials](tutorials/index.md)** - Learning-oriented: Step-by-step lessons for newcomers
- **[How-to Guides](how-to-guides/index.md)** - Problem-oriented: Recipes for specific tasks
- **[Reference](reference/index.md)** - Information-oriented: Technical descriptions and specifications
- **[Explanation](explanation/index.md)** - Understanding-oriented: Conceptual discussions and design rationale

**Choose based on what you need:**

- "I'm new and want to learn" → [Tutorials](tutorials/index.md)
- "I need to accomplish a task" → [How-to Guides](how-to-guides/index.md)
- "I need technical details" → [Reference](reference/index.md)
- "I want to understand why" → [Explanation](explanation/index.md)

---

## Working with Documentation

### Directory Structure

```text
docs/
├── index.md                    # This file
├── assets/                     # Binary files ONLY (.gif, .png, .pdf)
├── tutorials/                  # Learning-oriented guides
├── how-to-guides/              # Task-oriented recipes
├── reference/                  # Technical specifications
└── explanation/                # Conceptual discussions
```

### Building with MkDocs

This documentation uses [MkDocs](https://www.mkdocs.org/) with [Material theme](https://squidfunk.github.io/mkdocs-material/).

**Install:**

```bash
pip install mkdocs-material
pip install mkdocs-git-revision-date-localized-plugin
```

**Serve locally:**

```bash
mkdocs serve
# Open: http://127.0.0.1:8000
```

**Build static site:**

```bash
mkdocs build       # Output: site/
```

### Adding Content

**1. Create markdown file in appropriate directory:**

```bash
# Tutorial
touch docs/tutorials/my-tutorial.md

# How-to guide
touch docs/how-to-guides/my-guide.md

# Reference
touch docs/reference/my-reference.md

# Explanation
touch docs/explanation/my-explanation.md
```

**2. Add to navigation in `mkdocs.yml`:**

```yaml
nav:
  - Tutorials:
    - My Tutorial: tutorials/my-tutorial.md
```

**3. Preview with `mkdocs serve`**

### Adding Images

**Place binary files in `docs/assets/`:**

```bash
cp myimage.png docs/assets/
```

**Reference in markdown:**

```markdown
![Description](../assets/myimage.png)

<!-- Or with sizing -->
<img src="../assets/myimage.png" width="600">
```

**Important:** `docs/assets/` is for **binary files ONLY** (images, PDFs, videos). Never put `.md` files there.

### File Naming

- Lowercase: `my-guide.md` ✅
- Hyphens not underscores: `my-guide.md` not `my_guide.md` ✅
- Descriptive: `vscode-extension-setup.md` not `setup.md` ✅

### Code Blocks

Always specify language:

````markdown
```bash
./script.sh
```

```go
func main() {
  fmt.Println("hello")
}
```
````

### Admonitions

```markdown
!!! note
    This is a note

!!! warning
    This is a warning

!!! tip
    This is a helpful tip
```

---

## Resources

- **Diataxis**: <https://diataxis.fr/>
- **MkDocs**: <https://www.mkdocs.org/>
- **Material theme**: <https://squidfunk.github.io/mkdocs-material/>
- **Markdown guide**: <https://www.markdownguide.org/>
