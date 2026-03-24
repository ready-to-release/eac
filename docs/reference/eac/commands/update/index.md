# Update Commands

The **update** category contains commands for updating existing documentation and architecture diagrams.

**Key Features**:

- AI-powered documentation updates
- Preserves existing structure while enhancing content
- Integration with Structurizr workspace files

## Commands in this Category

| Command                                      | Purpose                                    |
| -------------------------------------------- | ------------------------------------------ |
| [update](./update.md)                        | Base update command                        |
| [update ai-summary](./ai-summary.md)            | Update AI summary                              |
| [update cache-clear](./cache-clear.md)           | Clear build and test caches                    |
| [update dependabot](./dependabot.md)             | Update Dependabot configuration                |
| [update design](./design.md)                     | Update existing workspace.dsl for a module     |
| [update docs](./docs.md)                         | Update documentation                           |
| [update docs-manifest](./docs-manifest.md)       | Update documentation assets manifest           |
| [update evidence](./evidence.md)                 | Update evidence artifacts                      |
| [update go-mod-sums](./go-mod-sums.md)           | Update go.sum files                            |
| [update go-tidy](./go-tidy.md)                   | Run go mod tidy across modules                 |
| [update pdf-screenshots](./pdf-screenshots.md)   | Update PDF screenshots                         |
| [update structurizr](./structurizr.md)           | Update Structurizr diagrams                    |

## Common Use Cases

### Update Architecture Documentation

```bash
eac update design src-auth
```

## See Also

- [create design](../create/design.md)
- [serve design](../serve/design.md)
- [validate design](../validate/design.md)
- [build Commands](../build/index.md)
