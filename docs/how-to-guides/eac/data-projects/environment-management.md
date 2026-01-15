# Environment Management

## What You'll Learn

Set up and manage environments for Databricks data projects using Unity Catalog for isolation, Delta Lake for versioning, and PLTE for production-like testing.

## Prerequisites

- Databricks workspace with Unity Catalog enabled
- Admin access to create catalogs and schemas
- Terraform or Databricks CLI for infrastructure as code
- Understanding of CD model environment types

## Environment Strategy

| Environment | CD Stages | Purpose | Characteristics |
|------------|-----------|---------|-----------------|
| **DevBox** | Stage 1-2 | Local development | Databricks Connect, local Spark, small data samples |
| **Build Agents** | Stage 2-4 | CI validation | GitHub Actions runners, unit/component tests |
| **PLTE (Staging)** | Stage 5-6 | Acceptance testing | Production-like config, cloned data, ephemeral or persistent |
| **Demo** | Stage 7 | Stakeholder validation | Trunk-HEAD or release-HEAD, stable test data |
| **Production** | Stage 10-12 | Live operation | Real data, monitored, strict access controls |

## Unity Catalog Structure

### Catalog-Based Isolation

Use separate catalogs for each environment:

```sql
-- Development catalog
CREATE CATALOG IF NOT EXISTS dev
COMMENT 'Development environment - individual developers';

-- Staging catalog (PLTE)
CREATE CATALOG IF NOT EXISTS staging
COMMENT 'Staging environment - integration testing';

-- Production catalog
CREATE CATALOG IF NOT EXISTS production
COMMENT 'Production environment - live data';

-- Demo catalog
CREATE CATALOG IF NOT EXISTS demo
COMMENT 'Demo environment - stakeholder validation';
```

### Schema Organization

Organize by data layer (medallion architecture):

```sql
-- Development environment setup
CREATE SCHEMA IF NOT EXISTS dev.bronze
COMMENT 'Raw data ingestion layer';

CREATE SCHEMA IF NOT EXISTS dev.silver
COMMENT 'Cleaned and enriched data layer';

CREATE SCHEMA IF NOT EXISTS dev.gold
COMMENT 'Business-level aggregates and models';

CREATE SCHEMA IF NOT EXISTS dev.artifacts
COMMENT 'Compiled wheels, JARs, and assets';

-- Repeat for staging, production, demo
```

### Volume Structure for Artifacts

```sql
-- Create volumes for artifact storage
CREATE VOLUME IF NOT EXISTS staging.artifacts.wheels;
CREATE VOLUME IF NOT EXISTS staging.artifacts.configs;

CREATE VOLUME IF NOT EXISTS production.artifacts.wheels;
CREATE VOLUME IF NOT EXISTS production.artifacts.configs;
```

## Access Control

### Environment-Specific Permissions

```sql
-- Development: Read/write for developers
GRANT CREATE, USE CATALOG ON CATALOG dev TO `data-engineers`;
GRANT CREATE, USE SCHEMA ON SCHEMA dev.bronze TO `data-engineers`;
GRANT SELECT, INSERT, UPDATE, DELETE ON SCHEMA dev.bronze TO `data-engineers`;

-- Staging: CI/CD service principal + developers (read-only)
GRANT CREATE, USE CATALOG ON CATALOG staging TO `cicd-service-principal`;
GRANT SELECT ON CATALOG staging TO `data-engineers`;

-- Production: CI/CD (write), engineers (read-only), analysts (read)
GRANT CREATE, USE CATALOG ON CATALOG production TO `cicd-service-principal`;
GRANT SELECT ON CATALOG production TO `data-engineers`;
GRANT SELECT ON SCHEMA production.gold TO `business-analysts`;
```

### Service Principal Setup

```bash
# Create service principal for CI/CD
databricks service-principals create \
  --display-name "GitHub Actions CI/CD" \
  --active

# Generate token
databricks tokens create \
  --application-id <service-principal-id> \
  --comment "CI/CD pipeline token" \
  --lifetime-seconds 31536000  # 1 year
```

## DevBox Environment

### Local Development Setup

**Option 1: Databricks Connect**

```python
# Configure Databricks Connect for local execution
from databricks.connect import DatabricksSession

spark = DatabricksSession.builder \
    .remote(
        host="https://staging.cloud.databricks.com",
        token="dapi...",
        cluster_id="1234-567890-abc123"
    ) \
    .getOrCreate()

# Use dev catalog for local work
spark.sql("USE CATALOG dev")

# Run local notebook against remote cluster
df = spark.table("dev.bronze.customer_events")
df.show()
```

**Option 2: Local Spark**

```python
# For unit tests, use local Spark
from pyspark.sql import SparkSession

spark = SparkSession.builder \
    .master("local[*]") \
    .appName("local-dev") \
    .config("spark.sql.extensions", "io.delta.sql.DeltaSparkSessionExtension") \
    .config("spark.sql.catalog.spark_catalog", "org.apache.spark.sql.delta.catalog.DeltaCatalog") \
    .getOrCreate()
```

### Test Data Samples

```python
# scripts/create_dev_samples.py
def create_dev_test_data():
    """Create small sample datasets for local development."""
    # Sample 1000 rows from production for realistic development
    sample_df = (spark.table("production.bronze.customer_events")
        .sample(fraction=0.01)
        .limit(1000)
    )

    # Write to dev catalog
    sample_df.write.mode("overwrite").saveAsTable("dev.bronze.customer_events")

    print(f"Created dev sample: {sample_df.count()} rows")
```

## PLTE (Production-Like Test Environment)

### Characteristics

- **Same infrastructure as production**: Cluster configs, Unity Catalog setup
- **Production-like data**: Cloned or masked from production
- **Isolated execution**: No impact on production
- **Ephemeral or persistent**: Choose based on cost vs. setup time

### Strategy 1: Ephemeral PLTE (Per-PR)

Create temporary environment for each pull request:

```python
# scripts/create_ephemeral_plte.py
import os

def create_ephemeral_plte(pr_number: int):
    """Create ephemeral test environment for PR."""
    catalog_name = f"plte_pr_{pr_number}"

    # Create catalog
    spark.sql(f"CREATE CATALOG IF NOT EXISTS {catalog_name}")
    spark.sql(f"CREATE SCHEMA {catalog_name}.bronze")
    spark.sql(f"CREATE SCHEMA {catalog_name}.silver")
    spark.sql(f"CREATE SCHEMA {catalog_name}.gold")

    # Clone production data (shallow clone = metadata only)
    spark.sql(f"""
        CREATE TABLE {catalog_name}.bronze.customer_events
        SHALLOW CLONE production.bronze.customer_events
    """)

    print(f"✅ Created ephemeral PLTE: {catalog_name}")
    return catalog_name

def cleanup_ephemeral_plte(pr_number: int):
    """Cleanup after PR merged/closed."""
    catalog_name = f"plte_pr_{pr_number}"
    spark.sql(f"DROP CATALOG IF EXISTS {catalog_name} CASCADE")
    print(f"🗑️ Cleaned up PLTE: {catalog_name}")
```

**GitHub Actions Integration:**

```yaml
- name: Create ephemeral PLTE
  id: plte
  run: |
    CATALOG=$(python scripts/create_ephemeral_plte.py ${{ github.event.number }})
    echo "catalog=$CATALOG" >> $GITHUB_OUTPUT

- name: Run tests in PLTE
  run: |
    pytest tests/integration/ \
      --catalog ${{ steps.plte.outputs.catalog }}

- name: Cleanup PLTE
  if: always()
  run: |
    python scripts/cleanup_ephemeral_plte.py ${{ github.event.number }}
```

### Strategy 2: Persistent Staging Environment

Single staging environment shared across the team:

```python
# Staging catalog already exists, just refresh test data
def refresh_staging_data():
    """Refresh staging environment with latest production data."""
    # Option A: Shallow clone (fast, shares storage)
    spark.sql("""
        CREATE OR REPLACE TABLE staging.bronze.customer_events
        SHALLOW CLONE production.bronze.customer_events
    """)

    # Option B: Deep clone (independent, can modify)
    spark.sql("""
        CREATE OR REPLACE TABLE staging.bronze.customer_events
        DEEP CLONE production.bronze.customer_events
    """)

    # Option C: Sample for performance (10% of data)
    spark.sql("""
        CREATE OR REPLACE TABLE staging.bronze.customer_events AS
        SELECT * FROM production.bronze.customer_events
        TABLESAMPLE (10 PERCENT)
    """)
```

## Delta Lake Versioning

### Time Travel for Testing

```sql
-- View table history
DESCRIBE HISTORY staging.gold.customer_segments;

-- Query specific version
SELECT * FROM staging.gold.customer_segments VERSION AS OF 42;

-- Query as of specific timestamp
SELECT * FROM staging.gold.customer_segments
TIMESTAMP AS OF '2024-01-15 10:00:00';

-- Restore to previous version (rollback)
RESTORE TABLE staging.gold.customer_segments TO VERSION AS OF 41;
```

### Clone Types

**Shallow Clone** (metadata only, fast):
```sql
-- Shares storage with source, can't modify independently
CREATE TABLE staging.bronze.events
SHALLOW CLONE production.bronze.events;
```

**Deep Clone** (full copy):
```sql
-- Independent copy, can modify without affecting source
CREATE TABLE staging.bronze.events
DEEP CLONE production.bronze.events;
```

### Schema Evolution Testing

```python
def test_schema_evolution():
    """Test adding column with backward compatibility."""
    # Stage 1: Add column with default value
    spark.sql("""
        ALTER TABLE staging.gold.customer_segments
        ADD COLUMN risk_score DECIMAL(3,2) DEFAULT 0.5
    """)

    # Stage 2: Run pipeline to populate new column
    run_pipeline(catalog="staging")

    # Stage 3: Verify backward compatibility
    assert spark.table("staging.gold.customer_segments").schema.fieldNames() == [
        "customer_id", "segment", "segment_score", "assigned_date", "risk_score"
    ]

    # Stage 4: If tests pass, apply to production
    # If tests fail, restore previous schema
    spark.sql("RESTORE TABLE staging.gold.customer_segments TO VERSION AS OF 1")
```

## Configuration Management

### Environment-Specific Configs

**Asset Bundle Variables:**

```yaml
# databricks.yml
variables:
  catalog:
    description: Target catalog for data
    default: production

  source_path:
    description: Input data location
    default: s3://prod-data/events/

  cluster_size:
    description: Cluster configuration
    default: medium

targets:
  dev:
    variables:
      catalog: dev
      source_path: s3://dev-data/events/
      cluster_size: small

  staging:
    variables:
      catalog: staging
      source_path: s3://staging-data/events/
      cluster_size: medium

  production:
    variables:
      catalog: production
      source_path: s3://prod-data/events/
      cluster_size: large
```

### Secret Management

```python
# Reference secrets in notebooks
catalog = spark.conf.get("catalog", "dev")
api_key = dbutils.secrets.get(scope="prod-keys", key="external-api-key")

# Create secret scopes (once per environment)
databricks secrets create-scope --scope staging-keys
databricks secrets put --scope staging-keys --key api-key

databricks secrets create-scope --scope prod-keys
databricks secrets put --scope prod-keys --key api-key
```

## Data Masking for Testing

### Production Data Masking

```python
# scripts/mask_production_data.py
from pyspark.sql.functions import lit, when, col, expr, monotonically_increasing_id
import uuid

def mask_sensitive_data(source_catalog: str, dest_catalog: str):
    """Create masked copy of production data for testing.

    SECURITY: Uses non-deterministic masking to prevent re-identification attacks.
    Each masked value is unique and cannot be reverse-engineered from the original.
    """
    # Read production data
    prod_df = spark.table(f"{source_catalog}.bronze.customers")

    # Mask PII fields with NON-DETERMINISTIC values
    # ⚠️ SECURITY: Do NOT use sha2() or other deterministic hashing
    # Deterministic: same email → same hash (allows re-identification)
    # Non-deterministic: same email → different random value each run
    masked_df = (prod_df
        # Option 1: Use UUID for unique, non-deterministic values
        .withColumn("email", expr("concat('test_user_', uuid(), '@test.example.com')"))
        .withColumn("phone", expr("concat('555-', lpad(cast(rand() * 10000 as int), 4, '0'))"))
        .withColumn("ssn", lit("XXX-XX-XXXX"))
        .withColumn("credit_card", lit("0000-0000-0000-0000"))

        # Option 2: Generalize location data (keep city for analytics, remove street)
        .withColumn("address",
            when(col("city").isNotNull(), concat(lit("123 Test St, "), col("city")))
            .otherwise(lit("Unknown"))
        )

        # Option 3: For customer_id, use sequential IDs (preserves cardinality)
        .withColumn("original_customer_id", col("customer_id"))  # Keep for joins
        .withColumn("customer_id", monotonically_increasing_id())
    )

    # Sample data for cost efficiency (optional)
    # Take 10% sample or max 100K records
    sample_size = min(100000, int(masked_df.count() * 0.1))
    masked_df = masked_df.limit(sample_size)

    # Write to staging
    masked_df.write.mode("overwrite").saveAsTable(f"{dest_catalog}.bronze.customers_masked")

    print(f"✅ Masked {masked_df.count()} records from {source_catalog} to {dest_catalog}")
    print(f"⚠️ NOTE: PII is non-deterministicly masked and cannot be reversed")
```

## Infrastructure as Code

### Terraform Setup

```hcl
# terraform/environments/staging/main.tf
resource "databricks_catalog" "staging" {
  name    = "staging"
  comment = "Staging environment for integration testing"
}

resource "databricks_schema" "staging_bronze" {
  catalog_name = databricks_catalog.staging.name
  name         = "bronze"
  comment      = "Raw data layer"
}

resource "databricks_schema" "staging_silver" {
  catalog_name = databricks_catalog.staging.name
  name         = "silver"
  comment      = "Cleaned data layer"
}

resource "databricks_schema" "staging_gold" {
  catalog_name = databricks_catalog.staging.name
  name         = "gold"
  comment      = "Business aggregates"
}

resource "databricks_grants" "staging_grants" {
  catalog = databricks_catalog.staging.name

  grant {
    principal  = "cicd-service-principal"
    privileges = ["USE_CATALOG", "USE_SCHEMA", "CREATE_TABLE", "MODIFY"]
  }

  grant {
    principal  = "data-engineers"
    privileges = ["USE_CATALOG", "USE_SCHEMA", "SELECT"]
  }
}
```

## Environment Promotion

### Data Promotion Strategy

```python
# scripts/promote_to_production.py
def promote_staging_to_production():
    """Promote validated staging data to production."""
    # Verify staging quality
    quality_check = run_quality_checks("staging")
    if not quality_check.passed:
        raise Exception(f"Quality checks failed: {quality_check.failures}")

    # Create production table from staging
    spark.sql("""
        CREATE OR REPLACE TABLE production.gold.customer_segments AS
        SELECT * FROM staging.gold.customer_segments
    """)

    # Tag version
    spark.sql("""
        ALTER TABLE production.gold.customer_segments
        SET TBLPROPERTIES (
          'promoted_from' = 'staging',
          'promoted_at' = current_timestamp(),
          'promoted_by' = current_user()
        )
    """)

    print("✅ Promoted staging to production")
```

## Next Steps

- [Databricks Asset Bundles](./databricks-asset-bundles.md) - Configure bundles for each environment
- [CI/CD Pipeline for Databricks](./cicd-pipeline-databricks.md) - Automate environment deployments
- [Monitoring Data Quality](./monitoring-data-quality.md) - Monitor production environment health
