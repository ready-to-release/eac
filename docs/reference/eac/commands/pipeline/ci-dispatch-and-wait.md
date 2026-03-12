# Pipeline ci dispatch-and-wait

<!-- book:cmd pipeline ci dispatch-and-wait -->

Dispatches a GitHub Actions workflow and waits for it to complete, or waits for an existing run by ID. Useful for CI orchestration when you need to trigger a workflow and block until it finishes.

If `--workflow` is provided, the command dispatches the workflow on the specified ref (or current branch), then polls until completion. If `--run-id` is provided, it skips dispatch and only waits.

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--workflow` | string | | Workflow file name to dispatch (e.g. `ci-clie.yaml`) |
| `--ref` | string | current branch | Git ref to run workflow on |
| `--run-id` | string | | Existing run ID to wait for (skips dispatch) |
| `--timeout` | int | `300` | Timeout in seconds |
| `--inputs` | string | | Workflow inputs as JSON object |

Either `--workflow` or `--run-id` is required.

## Output

- Dispatch confirmation and run ID
- Periodic status updates with elapsed time
- Exit code 0 on success, 1 on failure or timeout

## Examples

```bash
eac pipeline ci dispatch-and-wait --workflow ci-clie.yaml --ref main
eac pipeline ci dispatch-and-wait --run-id 12345678 --timeout 600
eac pipeline ci dispatch-and-wait --workflow ci-clie.yaml --inputs '{"version":"1.0.0"}'
```

## See Also

- [pipeline wait](./wait.md)
- [pipeline ci](./ci.md)
- [pipeline Commands](../categories/pipeline.md)
