# CI/CD Pipeline for Databricks

## What You'll Learn

Implement automated CI/CD pipelines for Databricks data projects using GitHub Actions and Asset Bundles.

## Prerequisites

- GitHub repository with Databricks code
- Databricks workspaces (staging, production)
- GitHub Actions enabled
- Databricks CLI configured
- Unity Catalog enabled

## Pipeline Overview

```text
┌─────────────┐
│ Git Push    │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────┐
│ Stage 2-4: CI Pipeline          │
│                                 │
│ ├─ Lint & Format                │
│ ├─ Unit Tests (L0)              │
│ ├─ Component Tests (L1)         │
│ ├─ Bundle Validate              │
│ └─ Security Scan                │
└──────┬──────────────────────────┘
       │ (on main branch)
       ▼
┌─────────────────────────────────┐
│ Stage 5-6: Deploy to Staging    │
│                                 │
│ ├─ Build & Package Artifacts    │
│ ├─ Upload to Unity Catalog      │
│ ├─ Deploy Asset Bundle          │
│ ├─ Run Integration Tests (L2)   │
│ └─ Generate Evidence            │
└──────┬──────────────────────────┘
       │ (on release tag)
       ▼
┌─────────────────────────────────┐
│ Stage 9-10: Production Deploy   │
│                                 │
│ ├─ Approval Gate (RA)           │
│ ├─ Deploy to Production         │
│ ├─ Run Smoke Tests              │
│ └─ Monitor Initial Run          │
└─────────────────────────────────┘
```

## Authentication Setup

### Workload Identity Federation (Recommended)

Configure OIDC for GitHub Actions:

```yaml
# .github/workflows/databricks-ci.yml
permissions:
  id-token: write
  contents: read

jobs:
  deploy:
    steps:
      - name: Configure Databricks CLI
        uses: databricks/setup-cli@main
        with:
          host: https://staging.cloud.databricks.com

      - name: Authenticate with OIDC
        run: |
          databricks auth login \
            --host $<< secrets.DATABRICKS_HOST >> \
            --client-id $<< secrets.DATABRICKS_CLIENT_ID >> \
            --client-secret $<< secrets.DATABRICKS_CLIENT_SECRET >>
```

### Service Principal (Alternative)

```yaml
# Use OIDC (no secrets needed - recommended)
- name: Authenticate with Databricks
  uses: databricks/setup-cli@main
  with:
    oidc: true

# OR use OAuth with client credentials (if OIDC not available)
# GitHub Secrets needed:
# DATABRICKS_CLIENT_ID, DATABRICKS_CLIENT_SECRET
- name: Authenticate with OAuth
  uses: databricks/setup-cli@main
  with:
    oauth-client-id: $<< secrets.DATABRICKS_CLIENT_ID >>
    oauth-client-secret: $<< secrets.DATABRICKS_CLIENT_SECRET >>
```

## CI Pipeline (Stages 2-4)

### Workflow File: `.github/workflows/ci.yml`

```yaml
name: CI - Databricks Pipeline

on:
  pull_request:
    branches: [main]
  push:
    branches: [main]

env:
  PYTHON_VERSION: "3.10"
  DATABRICKS_HOST: $<< secrets.DATABRICKS_STAGING_HOST >>
  DATABRICKS_TOKEN: $<< secrets.DATABRICKS_TOKEN >>

jobs:
  lint-and-format:
    name: Lint and Format
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: $<< env.PYTHON_VERSION >>

      - name: Install dependencies
        run: |
          pip install black pylint mypy

      - name: Run black
        run: black --check src/ tests/

      - name: Run pylint
        run: pylint src/ --fail-under=8.0

      - name: Run mypy
        run: mypy src/ --ignore-missing-imports

  unit-tests:
    name: Unit Tests (L0)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: $<< env.PYTHON_VERSION >>

      - name: Install dependencies
        run: |
          pip install -r requirements.txt
          pip install pytest pytest-cov chispa

      - name: Run unit tests
        run: |
          pytest tests/unit/ \
            -v \
            --junitxml=test-results/unit.xml \
            --cov=src \
            --cov-report=xml:coverage.xml

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.xml
          flags: unit

  component-tests:
    name: Component Tests (L1)
    runs-on: ubuntu-latest
    needs: [unit-tests]
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4

      - name: Set up Databricks CLI
        uses: databricks/setup-cli@main

      - name: Configure authentication
        run: |
          databricks configure --token <<EOF
          $<< env.DATABRICKS_HOST >>
          $<< env.DATABRICKS_TOKEN >>
          EOF

      - name: Run component tests
        run: |
          pytest tests/component/ \
            -v \
            --junitxml=test-results/component.xml

  bundle-validate:
    name: Validate Asset Bundle
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Databricks CLI
        uses: databricks/setup-cli@main

      - name: Validate bundle
        run: |
          databricks bundle validate --target staging

  security-scan:
    name: Security Scan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run Trivy security scan
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          format: 'sarif'
          output: 'trivy-results.sarif'

      - name: Upload scan results
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: 'trivy-results.sarif'

      - name: Scan for secrets
        uses: trufflesecurity/trufflehog@main
        with:
          path: ./
          base: $<< github.event.repository.default_branch >>
          head: HEAD
```

## Staging Deployment (Stages 5-6)

### Workflow File: `.github/workflows/deploy-staging.yml`

```yaml
name: Deploy to Staging

on:
  push:
    branches: [main]

env:
  DATABRICKS_HOST: $<< secrets.DATABRICKS_STAGING_HOST >>
  DATABRICKS_TOKEN: $<< secrets.DATABRICKS_TOKEN >>

jobs:
  build-and-upload:
    name: Build & Upload Artifacts
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: "3.10"

      - name: Build wheel package
        run: |
          pip install build
          python -m build --wheel

      - name: Set up Databricks CLI
        uses: databricks/setup-cli@main

      - name: Upload artifacts to Unity Catalog
        run: |
          ARTIFACT_PATH="dist/customer_segmentation-0.1.0-py3-none-any.whl"
          UC_VOLUME="staging.artifacts.wheels"

          databricks fs cp \
            $ARTIFACT_PATH \
            "dbfs:/Volumes/$UC_VOLUME/customer_segmentation-$<< github.sha >>.whl"

  deploy-staging:
    name: Deploy Asset Bundle
    runs-on: ubuntu-latest
    needs: [build-and-upload]
    environment: staging
    steps:
      - uses: actions/checkout@v4

      - name: Set up Databricks CLI
        uses: databricks/setup-cli@main

      - name: Deploy bundle
        run: |
          databricks bundle deploy --target staging

      - name: Trigger pipeline run
        id: trigger_run
        run: |
          RUN_ID=$(databricks jobs run-now \
            --job-id $<< secrets.STAGING_JOB_ID >> \
            --output json | jq -r '.run_id')
          echo "run_id=$RUN_ID" >> $GITHUB_OUTPUT

      - name: Wait for completion
        run: |
          databricks runs wait $<< steps.trigger_run.outputs.run_id >>

  integration-tests:
    name: Integration Tests (L2)
    runs-on: ubuntu-latest
    needs: [deploy-staging]
    steps:
      - uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: "3.10"

      - name: Install test dependencies
        run: pip install pytest databricks-connect

      - name: Run integration tests
        env:
          DATABRICKS_HOST: $<< env.DATABRICKS_HOST >>
          DATABRICKS_TOKEN: $<< env.DATABRICKS_TOKEN >>
        run: |
          pytest tests/integration/ \
            -v \
            --junitxml=test-results/integration.xml

      - name: Verify data quality
        run: |
          python scripts/verify_staging_quality.py

  generate-evidence:
    name: Generate Compliance Evidence
    runs-on: ubuntu-latest
    needs: [integration-tests]
    steps:
      - uses: actions/checkout@v4

      - name: Collect test results
        run: |
          mkdir -p evidence
          cp test-results/*.xml evidence/

      - name: Generate evidence report
        run: |
          python scripts/generate_evidence.py \
            --commit-sha $<< github.sha >> \
            --output evidence/report.json

      - name: Upload evidence
        uses: actions/upload-artifact@v3
        with:
          name: compliance-evidence
          path: evidence/
```

## Production Deployment (Stages 9-10)

### Workflow File: `.github/workflows/deploy-production.yml`

```yaml
name: Deploy to Production

on:
  push:
    tags:
      - 'v*'

env:
  DATABRICKS_HOST: $<< secrets.DATABRICKS_PROD_HOST >>
  DATABRICKS_TOKEN: $<< secrets.DATABRICKS_PROD_TOKEN >>

jobs:
  approval-gate:
    name: Release Approval (RA Pattern)
    runs-on: ubuntu-latest
    environment:
      name: production
      url: https://prod.cloud.databricks.com
    steps:
      - name: Wait for approval
        run: echo "Manual approval required via GitHub environment protection"

  deploy-production:
    name: Deploy to Production
    runs-on: ubuntu-latest
    needs: [approval-gate]
    steps:
      - uses: actions/checkout@v4
        with:
          ref: $<< github.ref >>

      - name: Set up Databricks CLI
        uses: databricks/setup-cli@main

      - name: Extract version
        id: version
        run: |
          VERSION=${GITHUB_REF#refs/tags/v}
          echo "version=$VERSION" >> $GITHUB_OUTPUT

      - name: Upload production artifact
        run: |
          ARTIFACT_PATH="dist/customer_segmentation-$<< steps.version.outputs.version >>-py3-none-any.whl"
          UC_VOLUME="production.artifacts.wheels"

          databricks fs cp \
            $ARTIFACT_PATH \
            "dbfs:/Volumes/$UC_VOLUME/customer_segmentation-$<< steps.version.outputs.version >>.whl"

      - name: Deploy bundle
        run: |
          databricks bundle deploy --target production

      - name: Update job schedule
        run: |
          databricks jobs update $<< secrets.PROD_JOB_ID >> \
            --json @production-job-config.json

  smoke-tests:
    name: Production Smoke Tests
    runs-on: ubuntu-latest
    needs: [deploy-production]
    steps:
      - uses: actions/checkout@v4

      - name: Trigger production run
        id: trigger_run
        run: |
          RUN_ID=$(databricks jobs run-now \
            --job-id $<< secrets.PROD_JOB_ID >> \
            --output json | jq -r '.run_id')
          echo "run_id=$RUN_ID" >> $GITHUB_OUTPUT

      - name: Monitor first run
        run: |
          databricks runs wait $<< steps.trigger_run.outputs.run_id >> \
            --timeout 1800

      - name: Verify output quality
        run: |
          python scripts/verify_production_quality.py

      - name: Send notification
        if: success()
        uses: 8398a7/action-slack@v3
        with:
          status: custom
          text: |
            ✅ Production deployment successful
            Version: $<< steps.version.outputs.version >>
            Run ID: $<< steps.trigger_run.outputs.run_id >>
          webhook_url: $<< secrets.SLACK_WEBHOOK >>

      - name: Rollback on failure
        if: failure()
        run: |
          echo "Smoke tests failed, initiating rollback"
          python scripts/rollback_production.py
```

## CDe Pattern (Automated Approval)

For continuous deployment without manual gates:

```yaml
name: Continuous Deployment (CDe)

on:
  push:
    branches: [main]

jobs:
  quality-gates:
    name: Automated Quality Gates
    runs-on: ubuntu-latest
    steps:
      - name: Check all tests passed
        run: |
          # Verify unit, component, integration tests all passed
          # Verify security scans passed
          # Verify performance benchmarks met

      - name: Auto-approve for production
        if: success()
        run: echo "Quality gates passed, proceeding to production"

  deploy-production:
    name: Deploy to Production (Auto)
    needs: [quality-gates]
    # No environment protection, deploys automatically
    steps:
      - name: Deploy
        run: databricks bundle deploy --target production
```

## Parallel Execution for Multiple Modules

```yaml
jobs:
  detect-changes:
    runs-on: ubuntu-latest
    outputs:
      modules: $<< steps.changed.outputs.modules >>
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Detect changed modules
        id: changed
        run: |
          # Detect which pipeline modules changed
          MODULES=$(python scripts/detect_changed_modules.py)
          echo "modules=$MODULES" >> $GITHUB_OUTPUT

  test-modules:
    needs: [detect-changes]
    runs-on: ubuntu-latest
    strategy:
      matrix:
        module: $<< fromJson(needs.detect-changes.outputs.modules) >>
    steps:
      - name: Test $<< matrix.module >>
        run: pytest tests/$<< matrix.module >>/
```

## Monitoring and Alerts

### Post-Deployment Monitoring

```yaml
  monitor-deployment:
    name: Monitor Deployment Health
    runs-on: ubuntu-latest
    needs: [deploy-production]
    steps:
      - name: Monitor for 1 hour
        run: |
          python scripts/monitor_deployment.py \
            --duration 3600 \
            --alert-threshold 0.05

      - name: Check data quality metrics
        run: |
          python scripts/check_quality_metrics.py \
            --catalog production \
            --table gold.customer_segments
```

### Rollback Script

```python
# scripts/rollback_production.py
import sys
from databricks.sdk import WorkspaceClient

def rollback_to_previous_version():
    """Rollback to previous working version."""
    w = WorkspaceClient()

    # Get current job configuration
    job_id = int(os.getenv("PROD_JOB_ID"))
    job = w.jobs.get(job_id)

    # Get previous version from Git
    previous_tag = subprocess.check_output(
        ["git", "describe", "--abbrev=0", "--tags", "HEAD~1"]
    ).decode().strip()

    print(f"Rolling back to {previous_tag}")

    # Deploy previous bundle version
    subprocess.run(
        ["git", "checkout", previous_tag],
        check=True
    )

    subprocess.run(
        ["databricks", "bundle", "deploy", "--target", "production"],
        check=True
    )

    # Revert Delta table if needed
    spark.sql("""
        RESTORE TABLE production.gold.customer_segments
        TO VERSION AS OF yesterday()
    """)

    print(f"✅ Rolled back to {previous_tag}")

if __name__ == "__main__":
    rollback_to_previous_version()
```

## Repository Structure

```text
.
├── .github/
│   └── workflows/
│       ├── ci.yml                      # Stage 2-4: CI pipeline
│       ├── deploy-staging.yml          # Stage 5-6: Staging deploy
│       └── deploy-production.yml       # Stage 9-10: Prod deploy
├── src/
│   ├── bronze/
│   ├── silver/
│   └── gold/
├── tests/
│   ├── unit/
│   ├── component/
│   ├── integration/
│   └── system/
├── scripts/
│   ├── verify_staging_quality.py
│   ├── verify_production_quality.py
│   ├── rollback_production.py
│   └── monitor_deployment.py
├── specs/
│   └── customer_segmentation.feature
├── databricks.yml                      # Asset Bundle config
├── requirements.txt
└── pyproject.toml
```

## Next Steps

- [Databricks Asset Bundles](./databricks-asset-bundles.md) - Deep dive on bundle configuration
- [Environment Management](./environment-management.md) - Setup staging and production environments
- [Monitoring Data Quality](./monitoring-data-quality.md) - Production monitoring strategies
