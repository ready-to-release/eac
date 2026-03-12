# Pipeline ci summary-link

<!-- book:cmd pipeline ci summary-link -->

Generates a markdown code block with `gh` CLI diagnostic commands for CI failure investigation. The output is designed to be piped directly into `$GITHUB_STEP_SUMMARY` in workflows.

The `--type` flag controls which diagnostic commands are generated:

- **build** (default) -- failed step logs and full log download
- **test** -- failed logs plus test artifact download
- **container** -- failed logs plus docker pull/test commands
- **release** -- failed logs, CI status check, and release/tag listing
- **docs** -- failed logs, artifact download, and GitHub Pages status
- **deviation** -- full rebuild comparison with incremental CI runs

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--type` | string | `build` | Failure type: build, test, container, release, docs, deviation |
| `--artifact` | string | | Artifact name to include in download command |
| `--image` | string | | Container image for container-type diagnostics |
| `--workflow` | string | | CI workflow name for release-type diagnostics |
| `--commit` | string | | Commit SHA for release-type diagnostics |

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<run-id>` | Yes | GitHub Actions run ID |

## Output

Markdown with a clickable run link and a fenced code block containing `gh` commands.

## Examples

```bash
eac pipeline ci summary-link 12345678
eac pipeline ci summary-link 12345678 --type test --artifact results
eac pipeline ci summary-link 12345678 --type container --image ghcr.io/org/img:latest

# In a GitHub Actions workflow
go run ./go/cli/eac pipeline ci summary-link ${{ github.run_id }} >> $GITHUB_STEP_SUMMARY
```

## See Also

- [pipeline ci](./ci.md)
- [pipeline status](./status.md)
- [pipeline Commands](../categories/pipeline.md)
