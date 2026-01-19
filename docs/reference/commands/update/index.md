# update Commands

Update existing documentation, architecture diagrams, and run linters.

## Commands in this Category

| Command                                      | Purpose                                    |
| -------------------------------------------- | ------------------------------------------ |
| [update](./update.md)                        | Base update command                        |
| [update design](./design.md)                 | Update existing workspace.dsl for a module |
| [update lint](./lint/index.md)               | Run linters on modules                     |
| [update lint (Go)](./lint/go.md)             | Go linting with golangci-lint              |
| [update lint (Markdown)](./lint/markdown.md) | Markdown linting with markdownlint-cli2    |

## Quick Examples

```bash
# Update architecture documentation
r2r eac update design src-auth

# Lint all modules
r2r eac update lint

# Lint with auto-fix
r2r eac update lint --fix
```

## See Also

- [Category Overview](../categories/update.md)
- [create design](../create/design.md)
- [build Commands](../build/index.md)
