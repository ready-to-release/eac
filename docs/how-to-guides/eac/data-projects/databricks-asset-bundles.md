# Databricks Asset Bundles

## What You'll Learn

Master Databricks Asset Bundles for unified deployment of notebooks, jobs, pipelines, and infrastructure across environments.

## Prerequisites

- Databricks CLI installed (`curl -fsSL https://raw.githubusercontent.com/databricks/setup-cli/main/install.sh | sh`)
- Git repository
- Databricks workspace access
- Understanding of YAML syntax

## Asset Bundle Overview

Asset Bundles package and deploy:

- Notebooks and Python wheels
- Jobs and workflows
- Delta Live Tables pipelines
- ML models
- Cluster configurations
- Permissions and access controls

**Benefits:**

- **Single source of truth** - All resources defined as code
- **Environment consistency** - Same bundle, different targets
- **Atomic deployment** - All resources deploy together
- **Version control** - Track infrastructure changes in Git
- **Validation** - Catch errors before deployment

## Repository Structure: Monorepo with Modules

Organize data projects as **modules within a monorepo**, following the module contract pattern. Each data pipeline is a module with its own versioning, dependencies, CI workflow, and release cycle.

**Recommended Structure:**

```text
data-platform/                    # Repository root
├── .r2r/eac/
│   ├── repository.yml           # Module contracts (see below)
│   └── module-types.yml
├── pipelines/
│   ├── customer-segmentation/   # Module: customer-segmentation
│   │   ├── databricks.yml      # Bundle configuration
│   │   ├── specs/              # Gherkin specifications
│   │   │   └── customer-segmentation_pipeline.feature
│   │   ├── src/                # Source code
│   │   │   ├── bronze/
│   │   │   │   └── ingest_events.py
│   │   │   ├── silver/
│   │   │   │   └── aggregate_features.py
│   │   │   └── gold/
│   │   │       └── segment_customers.py
│   │   ├── tests/              # Unit/integration tests
│   │   │   └── test_features.py
│   │   ├── requirements.txt
│   │   ├── CHANGELOG.md        # Module-specific changelog
│   │   └── README.md
│   │
│   └── revenue-forecasting/    # Module: revenue-forecasting
│       ├── databricks.yml
│       ├── specs/
│       ├── src/
│       └── tests/
├── shared/
│   └── data-quality-lib/       # Module: shared libraries
│       ├── src/
│       └── tests/
├── .github/workflows/
│   ├── ci-customer-segmentation.yaml
│   └── ci-revenue-forecasting.yaml
└── README.md
```

### Module Contract Configuration

Define each data pipeline as a module in `.r2r/eac/repository.yml`:

```yaml
repository:
  type: mono                     # Monorepo

modules:
  # Shared library (Layer 0)
  - moniker: data-quality-lib
    name: Data Quality Library
    type: python
    description: Shared data quality utilities
    versioning:
      scheme: SemVer
    files:
      root: shared/data-quality-lib
      source:
        - "**/*.py"
      tests:
        - "**/*.py"
      config:
        - requirements.txt
        - setup.py
      workflows:
        ci: .github/workflows/ci-data-quality-lib.yaml

  # Customer segmentation pipeline (Layer 1 - depends on shared lib)
  - moniker: customer-segmentation
    name: Customer Segmentation Pipeline
    type: databricks-bundle
    description: Daily customer segmentation using business rules
    versioning:
      scheme: CalVer             # Use CalVer for data pipelines (YYYY.MM.DD)
    depends_on:
      - data-quality-lib
    files:
      root: pipelines/customer-segmentation
      source:
        - "src/**/*.py"
      tests:
        - "tests/**/*.py"
      config:
        - databricks.yml
        - requirements.txt
      workflows:
        ci: .github/workflows/ci-customer-segmentation.yaml
        release: .github/workflows/release-customer-segmentation.yaml
      repo:
        specs:
          - specs/**/*.feature
    databricks_bundle:
      target_environments:
        - dev
        - staging
        - production

  # Revenue forecasting pipeline (Layer 1)
  - moniker: revenue-forecasting
    name: Revenue Forecasting Pipeline
    type: databricks-bundle
    description: ML-based revenue forecasting with Prophet
    versioning:
      scheme: CalVer
    depends_on:
      - data-quality-lib
      - customer-segmentation      # Depends on segmentation output
    files:
      root: pipelines/revenue-forecasting
      workflows:
        ci: .github/workflows/ci-revenue-forecasting.yaml
        release: .github/workflows/release-revenue-forecasting.yaml
```

### Advantages of Monorepo with Modules

**Atomic Changes:**

- Code, specs, tests, and bundle config change together in one PR
- Cross-pipeline changes (e.g., shared library update) are atomic

**Independent Versioning:**

- Each pipeline has its own version (CalVer: 2024.01.15)
- Shared libraries use SemVer (1.2.3)
- Independent release cycles per module

**Dependency Management:**

- Explicit `depends_on` relationships between modules
- CI automatically builds dependencies first
- Detect which pipelines need rebuild when shared code changes

**Selective CI:**

- Only run CI for changed modules (and their dependents)
- Faster feedback loops
- Resource-efficient CI

**Unified Workflows:**

- All pipelines use consistent CI/CD patterns
- Shared GitHub Actions across modules
- Centralized monitoring and governance

**Traceability:**

- Single source of truth for all data pipelines
- Unified change history across pipelines
- Easier compliance and audit

## Bundle Structure

### Basic Bundle Configuration

`databricks.yml`:

```yaml
bundle:
  name: customer-segmentation

include:
  - "resources/*.yml"  # Include additional resource files

variables:
  catalog:
    description: Target Unity Catalog
    default: production

workspace:
  host: https://prod.cloud.databricks.com

artifacts:
  default:
    type: whl
    build: poetry build
    path: .

resources:
  jobs:
    customer_segmentation_pipeline:
      name: "Customer Segmentation - ${bundle.target}"
      description: "Daily customer segmentation pipeline"

      job_clusters:
        - job_cluster_key: main_cluster
          new_cluster:
            spark_version: 13.3.x-scala2.12
            node_type_id: i3.xlarge
            num_workers: 2
            spark_conf:
              spark.databricks.delta.preview.enabled: "true"

      tasks:
        - task_key: ingest_events
          job_cluster_key: main_cluster
          notebook_task:
            notebook_path: ./src/bronze/ingest_events
            base_parameters:
              catalog: ${var.catalog}
              source_path: ${var.source_path}

        - task_key: aggregate_features
          depends_on:
            - task_key: ingest_events
          job_cluster_key: main_cluster
          notebook_task:
            notebook_path: ./src/silver/aggregate_features
            base_parameters:
              catalog: ${var.catalog}

        - task_key: segment_customers
          depends_on:
            - task_key: aggregate_features
          job_cluster_key: main_cluster
          notebook_task:
            notebook_path: ./src/gold/segment_customers
            base_parameters:
              catalog: ${var.catalog}

      schedule:
        quartz_cron_expression: "0 0 2 * * ?"  # Daily at 2 AM UTC
        timezone_id: "UTC"

      email_notifications:
        on_failure:
          - data-team@company.com

targets:
  dev:
    mode: development
    workspace:
      host: https://dev.cloud.databricks.com
    variables:
      catalog: dev
      source_path: s3://dev-data/events/

  staging:
    mode: development
    workspace:
      host: https://staging.cloud.databricks.com
    variables:
      catalog: staging
      source_path: s3://staging-data/events/

  production:
    mode: production
    workspace:
      host: https://prod.cloud.databricks.com
    variables:
      catalog: production
      source_path: s3://prod-data/events/
```

### Multi-File Configuration

Split large bundles into multiple files:

`databricks.yml`:

```yaml
bundle:
  name: customer-segmentation

include:
  - "resources/jobs/*.yml"
  - "resources/pipelines/*.yml"
  - "resources/clusters/*.yml"

variables:
  catalog:
    default: production
  environment:
    default: prod
```

`resources/jobs/segmentation.yml`:

```yaml
resources:
  jobs:
    customer_segmentation:
      name: "Customer Segmentation - ${var.environment}"
      tasks:
        - task_key: segment
          notebook_task:
            notebook_path: ./src/gold/segment_customers
```

`resources/pipelines/dlt.yml`:

```yaml
resources:
  pipelines:
    customer_events_pipeline:
      name: "Customer Events DLT - ${var.environment}"
      catalog: ${var.catalog}
      target: gold
      configuration:
        pipeline.catalog: ${var.catalog}
```

## Deployment Workflow

### Local Development

```bash
# Validate bundle
databricks bundle validate --target dev

# Deploy to dev
databricks bundle deploy --target dev

# Run job
databricks bundle run customer_segmentation_pipeline --target dev

# Watch logs
databricks bundle run customer_segmentation_pipeline --target dev --watch

# Destroy (cleanup)
databricks bundle destroy --target dev
```

### Watch Mode for Development

```bash
# Auto-deploy on file changes
databricks bundle deploy --target dev --watch

# Now edit notebooks, save, and they auto-deploy
```

### CI/CD Integration

```yaml
# .github/workflows/deploy.yml
- name: Validate bundle
  run: databricks bundle validate --target staging

- name: Deploy to staging
  run: databricks bundle deploy --target staging --auto-approve

- name: Run bundle
  run: databricks bundle run customer_segmentation_pipeline --target staging
```

## Advanced Patterns

### Using Wheels for Code Packaging

`databricks.yml`:

```yaml
artifacts:
  customer_segmentation_lib:
    type: whl
    build: pip wheel . --no-deps --wheel-dir dist
    path: dist/*.whl

resources:
  jobs:
    segmentation:
      tasks:
        - task_key: run_segmentation
          libraries:
            - whl: ./dist/*.whl
          python_wheel_task:
            package_name: customer_segmentation
            entry_point: main
            parameters:
              - "--catalog"
              - "${var.catalog}"
```

### Delta Live Tables Pipeline

`databricks.yml`:

```yaml
resources:
  pipelines:
    customer_bronze_silver_gold:
      name: "Customer Pipeline - ${bundle.target}"
      catalog: ${var.catalog}
      target: gold

      libraries:
        - notebook:
            path: ./src/dlt/bronze_layer.py
        - notebook:
            path: ./src/dlt/silver_layer.py
        - notebook:
            path: ./src/dlt/gold_layer.py

      clusters:
        - label: default
          num_workers: 2
          node_type_id: i3.xlarge

      development: ${ bundle.target == 'dev' }
      continuous: false

      configuration:
        catalog: ${var.catalog}
        source_path: ${var.source_path}
```

### Permissions and Access Control

```yaml
resources:
  jobs:
    customer_segmentation:
      name: "Customer Segmentation"

      # Job-level permissions
      permissions:
        - level: CAN_VIEW
          group_name: "data-engineers"
        - level: CAN_MANAGE_RUN
          group_name: "data-team-leads"
        - level: IS_OWNER
          service_principal_name: "cicd-service-principal"

  pipelines:
    events_pipeline:
      name: "Events Pipeline"

      permissions:
        - level: CAN_VIEW
          group_name: "all-users"
        - level: CAN_RUN
          group_name: "data-engineers"
```

### Cluster Policies

```yaml
resources:
  cluster_policies:
    data_engineering_policy:
      name: "Data Engineering Standard - ${bundle.target}"
      definition:
        "spark_version": {
          "type": "fixed",
          "value": "13.3.x-scala2.12"
        }
        "node_type_id": {
          "type": "allowlist",
          "values": ["i3.xlarge", "i3.2xlarge"]
        }
        "num_workers": {
          "type": "range",
          "minValue": 1,
          "maxValue": 10
        }

  jobs:
    my_job:
      job_clusters:
        - job_cluster_key: main
          new_cluster:
            policy_id: ${resources.cluster_policies.data_engineering_policy.id}
```

## Environment-Specific Overrides

### Variable Lookup Tables

```yaml
variables:
  catalog:
    description: Target catalog
    lookup:
      dev: dev_catalog
      staging: staging_catalog
      production: prod_catalog

  cluster_size:
    lookup:
      dev:
        num_workers: 1
        node_type_id: i3.xlarge
      production:
        num_workers: 10
        node_type_id: i3.2xlarge

targets:
  dev:
    variables:
      environment: dev

  production:
    variables:
      environment: production

resources:
  jobs:
    pipeline:
      job_clusters:
        - job_cluster_key: main
          new_cluster:
            num_workers: ${var.cluster_size.num_workers}
            node_type_id: ${var.cluster_size.node_type_id}
```

### Conditional Resources

**Option 1: Target-Specific Bundle Files**:

Create separate resource files per environment:

`resources/dev_jobs.yml`:

```yaml
resources:
  jobs:
    debug_job:
      name: "Debug Job - ${bundle.target}"
      tasks:
        - task_key: debug
          notebook_task:
            notebook_path: ./notebooks/debug
```

`resources/prod_jobs.yml`:

```yaml
resources:
  jobs:
    production_monitoring:
      name: "Production Monitoring"
      schedule:
        quartz_cron_expression: "0 */15 * * * ?"
      tasks:
        - task_key: monitor
          notebook_task:
            notebook_path: ./notebooks/monitor
```

`databricks.yml`:

```yaml
bundle:
  name: customer-segmentation

# Conditionally include files based on target
include:
  - "resources/common/*.yml"
  - "resources/${bundle.target}_jobs.yml"  # Loads dev_jobs.yml or prod_jobs.yml

targets:
  dev:
    # Includes resources/dev_jobs.yml
  production:
    # Includes resources/prod_jobs.yml
```

**Option 2: Disabled Jobs**:

Use `paused` parameter to disable jobs in certain environments:

```yaml
resources:
  jobs:
    debug_job:
      name: "Debug Job"
      # Disable in production
      pause_status: ${var.is_production ? "PAUSED" : "UNPAUSED"}
      tasks:
        - task_key: debug
          notebook_task:
            notebook_path: ./notebooks/debug

variables:
  is_production:
    lookup:
      dev: false
      staging: false
      production: true
```

**Option 3: Override Per Target**:

Define job in common file, override schedule per target:

`databricks.yml`:

```yaml
resources:
  jobs:
    monitoring_job:
      name: "Monitoring"
      schedule:
        quartz_cron_expression: "0 0 * * * ?"  # Default: daily

targets:
  production:
    resources:
      jobs:
        monitoring_job:
          schedule:
            quartz_cron_expression: "0 */15 * * * ?"  # Override: every 15 min
```

## Validation and Testing

### Bundle Validation

```bash
# Validate syntax and references
databricks bundle validate --target staging

# Show what would be deployed
databricks bundle deploy --target staging --dry-run

# Show JSON representation
databricks bundle deploy --target staging --dry-run --json
```

### Pre-deployment Checks

```python
# scripts/validate_bundle.py
import yaml
import sys

def validate_bundle_config():
    """Validate bundle configuration before deployment."""
    with open("databricks.yml") as f:
        config = yaml.safe_load(f)

    # Check required fields
    assert "bundle" in config, "Missing bundle section"
    assert "name" in config["bundle"], "Missing bundle name"

    # Validate targets
    assert "targets" in config, "No targets defined"
    required_targets = ["dev", "staging", "production"]
    for target in required_targets:
        assert target in config["targets"], f"Missing target: {target}"

    # Validate jobs
    if "resources" in config and "jobs" in config["resources"]:
        for job_name, job_config in config["resources"]["jobs"].items():
            assert "tasks" in job_config, f"Job {job_name} has no tasks"
            # Validate task dependencies
            validate_task_dependencies(job_config["tasks"])

    print("✅ Bundle configuration valid")

if __name__ == "__main__":
    validate_bundle_config()
```

## Troubleshooting

### View Deployed Resources

```bash
# List deployed jobs
databricks jobs list --output json | jq '.jobs[] | select(.settings.name | contains("Customer Segmentation"))'

# View job configuration
databricks jobs get --job-id 12345

# View pipeline configuration
databricks pipelines get --pipeline-id <pipeline-id>
```

### Debug Deployment Issues

```bash
# Enable verbose logging
databricks bundle deploy --target staging --debug

# Check bundle state
databricks bundle sync --target staging

# Force redeploy
databricks bundle deploy --target staging --force
```

## Next Steps

- [CI/CD Pipeline for Databricks](./cicd-pipeline-databricks.md) - Automate bundle deployments
- [Environment Management](./environment-management.md) - Configure Unity Catalog for each target
- [Develop Data Pipeline with Specs](./develop-data-pipeline-with-specs.md) - Build pipeline to deploy with bundles
