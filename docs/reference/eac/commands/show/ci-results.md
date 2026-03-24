# show ci-results

<!-- book:cmd show ci-results -->

## Output Sections

A table with columns:

| Column   | Description                                                     |
| -------- | --------------------------------------------------------------- |
| Workflow | Workflow name (orchestrator listed first)                       |
| Status   | Conclusion or in-progress status; failures show `FAILED: <job>` |
| Link     | URL to the run or the specific failed job                       |

Summary line shows the short SHA with pass/fail counts.

## See Also

- [get ci-results](../get/ci-results.md) - JSON output
- [show ci-summary](./ci-summary.md)
