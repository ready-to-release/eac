# get ci-results

<!-- book:cmd get ci-results -->

## Output Structure

```yaml
head_sha: abc1234...
orchestrator:
  module: ci-orchestrator
  workflow: change-trigger.yaml
  run_id: 12345678
  status: completed
  conclusion: success
runs:
  - module: core
    workflow: ci-core.yaml
    run_id: 12345679
    status: completed
    conclusion: success
    jobs:
      - name: build
        status: completed
        conclusion: success
        duration: 2m30s
    artifacts:
      - name: core-evidence
        size_bytes: 1048576
        expired: false
    links:
      web_url: https://github.com/org/repo/actions/runs/12345679
      view_logs: gh run view 12345679 --repo org/repo --log
      view_failed_logs: gh run view 12345679 --repo org/repo --log-failed
      download_all: gh run download 12345679 --repo org/repo
total_runs: 5
passed: 4
failed: 1
```
