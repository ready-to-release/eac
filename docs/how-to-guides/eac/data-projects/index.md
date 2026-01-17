# Data Projects

Learn how to apply the 12-stage CD model to Databricks data projects including pipelines, dashboards, and business logic.

## In This Section

### Understanding the CD Model

| Guide                                                         | What You'll Learn                                                                  |
| ------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| [CD Model for Data Projects](./cd-model-for-data-projects.md) | How each of the 12 stages applies to data pipelines, notebooks, and business logic |

### Development Workflow

| Guide                                                                     | What You'll Learn                                                      |
| ------------------------------------------------------------------------- | ---------------------------------------------------------------------- |
| [Develop Data Pipeline with Specs](./develop-data-pipeline-with-specs.md) | End-to-end pipeline development from specifications through deployment |
| [Databricks Asset Bundles](./databricks-asset-bundles.md)                 | Structure, configuration, and deployment of Asset Bundles              |

### Testing and Quality

| Guide                                                 | What You'll Learn                                                     |
| ----------------------------------------------------- | --------------------------------------------------------------------- |
| [Testing Data Pipelines](./testing-data-pipelines.md) | Unit tests, integration tests, and data quality validation strategies |

### CI/CD and Operations

| Guide                                                          | What You'll Learn                                                   |
| -------------------------------------------------------------- | ------------------------------------------------------------------- |
| [CI/CD Pipeline for Databricks](./cicd-pipeline-databricks.md) | GitHub Actions workflows and automated deployment                   |
| [Environment Management](./environment-management.md)          | Unity Catalog isolation, Delta Lake versioning, and PLTE setup      |
| [Monitoring Data Quality](./monitoring-data-quality.md)        | Pipeline observability, data quality metrics, and incident response |

## Why Apply CD Model to Data Projects?

Data projects face unique challenges:

- **Long-running pipelines** require different validation strategies
- **Data quality** issues emerge in production, not just code bugs
- **Stateful systems** make rollback more complex
- **Multiple artifacts** (notebooks, jobs, models, dashboards) must deploy together

The 12-stage CD model provides:

- **Shift-left validation** - Catch data quality issues early
- **Environment isolation** - Test with production-like data safely
- **Automated testing** - Validate transformations, schemas, and quality
- **Traceability** - Link data changes to requirements and tests
- **Controlled deployment** - RA or CDe patterns based on risk

## Quick Reference: CD Stages for Data

| Stage               | Data Activities                           | Key Tools                      |
| ------------------- | ----------------------------------------- | ------------------------------ |
| 1. Authoring        | Write specs, develop notebooks locally    | Databricks Connect, VS Code    |
| 2. Pre-commit       | Unit tests, linting, bundle validate      | pytest, chispa, databricks CLI |
| 3. Merge Request    | Code review, CI validation                | GitHub Actions, Asset Bundles  |
| 4. Commit           | Build artifacts, version Delta tables     | GitHub Actions, Unity Catalog  |
| 5. Acceptance Test  | Pipeline testing with realistic data      | PLTE, Delta Lake clones        |
| 6. Extended Test    | Performance, security, data quality       | DAST, pytest assertions        |
| 7. Exploration      | Stakeholder validation, dashboard review  | Demo workspace                 |
| 8. Start Release    | Tag version, generate release notes       | Git tags, changelog            |
| 9. Release Approval | Approve for production (RA) or auto (CDe) | Quality gates, approvals       |
| 10. Prod Deploy     | Deploy bundles, update schemas            | Asset Bundles, Unity Catalog   |
| 11. Live            | Monitor pipelines, data quality, costs    | Databricks monitoring, alerts  |
| 12. Toggling        | Feature flags for transformations         | Custom feature flag system     |

## Prerequisites

Before following these guides, ensure you have:

- Databricks workspace access (AWS, Azure, or GCP)
- Unity Catalog enabled
- Git repository for version control
- CI/CD platform (GitHub Actions recommended)
- Local development environment:
  - Python ≥ 3.9
  - Databricks CLI
  - Databricks Connect (optional for local execution)
- Basic familiarity with:
  - Databricks notebooks and jobs
  - Delta Lake tables
  - CI/CD concepts
  - Git workflows

## Example Used Throughout

These guides use a consistent **Customer Segmentation Pipeline** example:

- **Input**: Raw customer events from cloud storage
- **Transformation**: Feature aggregation and business rule segmentation
- **Output**: Customer segments in Delta table (VIP, Active, At-Risk, Churned)
- **Dashboard**: Segment distribution visualization

This provides continuity across articles and shows how pieces fit together.

## Next Steps

Start with [CD Model for Data Projects](./cd-model-for-data-projects.md) to understand how the 12 stages apply to your work, then follow the development workflow guides.
