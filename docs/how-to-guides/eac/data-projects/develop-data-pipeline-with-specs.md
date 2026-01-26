# Develop Data Pipeline with Specifications

## What You'll Accomplish

Develop a complete data pipeline from specifications through production deployment, following the 12-stage CD model.

## Prerequisites

- Databricks workspace with Unity Catalog
- Git repository initialized
- Databricks CLI installed
- Local development environment (Python ≥ 3.9)
- Asset Bundles configured

## Example: Customer Segmentation Pipeline

This guide develops a pipeline that segments customers based on purchase behavior.

**Pipeline Overview:**

- **Input**: Raw customer events from cloud storage
- **Transform**: Aggregate features, apply business rules segmentation
- **Output**: Customer segments in Delta table
- **Schedule**: Daily at 2 AM UTC

## Stage 1: Authoring - Write Specifications

### Create Feature Specification

Create `specs/customer-segmentation_pipeline.feature`:

```gherkin
@L2 @ov @control:si-10
Feature: customer-segmentation_pipeline
  Segment customers by purchase behavior for targeted marketing campaigns

  As a marketing analyst
  I want to segment customers by purchase behavior
  So that I can target campaigns effectively

  Background:
    Given Unity Catalog is enabled
    And catalog "production" exists
    And schema "analytics" exists

  Rule: Pipeline loads and validates customer events from bronze layer

    @ov
    Scenario: Load customer events from bronze layer
      Given bronze.customer_events contains purchase data
      And events include customer_id, event_date, amount
      When I read events from last 90 days
      Then I have events for active customers only
      And events are deduplicated by (customer_id, event_date, event_id)

  Rule: Pipeline aggregates customer purchase behavior features

    @ov
    Scenario: Aggregate customer features
      Given customer events from bronze layer
      When I aggregate by customer_id
      Then silver.customer_features contains:
        | Column           | Type    | Description                  |
        | customer_id      | BIGINT  | Unique customer identifier   |
        | total_purchases  | DECIMAL | Sum of purchase amounts      |
        | purchase_count   | INT     | Number of purchases          |
        | avg_order_value  | DECIMAL | Average purchase amount      |
        | days_since_last  | INT     | Days since most recent order |
        | first_purchase   | DATE    | Date of first purchase       |
        | last_purchase    | DATE    | Date of most recent purchase |

  Rule: Pipeline assigns customers to segments using business rules

    @ov
    Scenario: Apply customer segmentation
      Given silver.customer_features table
      When I apply business rules to categorize customers
      Then gold.customer_segments contains:
        | Column        | Type    | Description                            |
        | customer_id   | BIGINT  | Unique customer identifier             |
        | segment       | STRING  | Segment: VIP, Active, At-Risk, Churned |
        | assigned_date | DATE    | Date segment was assigned              |
      And VIP customers have total_purchases > 10000
      And Active customers have days_since_last <= 30
      And At-Risk customers have days_since_last between 30 and 90
      And Churned customers have days_since_last > 90

  Rule: Pipeline enforces data quality requirements

    @ov
    Scenario: Data quality checks
      Given gold.customer_segments table
      Then table has no null customer_ids
      And table has no duplicate customer_ids
      And segment values are in allowed list
      And assigned_date is today's date
```

### Create Notebook Structure

Create notebooks following this structure:

```text
src/
├── bronze/
│   └── ingest_customer_events.py       # Raw data ingestion
├── silver/
│   └── aggregate_customer_features.py  # Feature engineering
└── gold/
    └── segment_customers.py            # Segmentation logic
```

### Develop Bronze Layer Notebook

`src/bronze/ingest_customer_events.py`:

```python
# Databricks notebook source
# MAGIC %md
# MAGIC # Ingest Customer Events (Incremental)
# MAGIC Load new customer events from cloud storage into bronze layer using watermark-based incremental processing

# COMMAND ----------
from pyspark.sql import DataFrame
from pyspark.sql.functions import col, to_date, input_file_name, lit, current_timestamp
from datetime import datetime, timedelta

# COMMAND ----------
def get_last_watermark(catalog: str, table_name: str) -> str:
    """Get the last processed event_date from the target table."""
    try:
        max_date = spark.sql(f"""
            SELECT COALESCE(MAX(event_date), '1970-01-01') as max_date
            FROM {catalog}.bronze.{table_name}
        """).first().max_date
        return max_date.strftime("%Y-%m-%d") if max_date else "1970-01-01"
    except Exception:
        # Table doesn't exist yet (first run)
        return "1970-01-01"

def load_customer_events_incremental(
    source_path: str,
    last_watermark: str,
    catalog: str
) -> DataFrame:
    """Load only NEW customer events since last watermark."""
    try:
        df = (spark.read
              .format("parquet")
              .load(source_path)
              .filter(col("event_date") > last_watermark)  # INCREMENTAL: Only new data
              .filter(col("event_type") == "purchase")
              .filter(col("amount") > 0)  # Data quality: reject negative amounts
              .select(
                  col("customer_id").cast("bigint"),
                  to_date(col("event_timestamp")).alias("event_date"),
                  col("amount").cast("decimal(10,2)"),
                  col("event_id"),
                  input_file_name().alias("source_file"),
                  current_timestamp().alias("ingestion_timestamp")
              )
              .dropDuplicates(["customer_id", "event_date", "event_id"])
        )

        # Validate data
        row_count = df.count()
        if row_count == 0:
            print(f"⚠️ No new events found since {last_watermark}")
        else:
            print(f"✅ Loaded {row_count} new events since {last_watermark}")

        return df

    except Exception as e:
        print(f"🚨 Error loading events: {str(e)}")
        # Write to dead letter queue
        spark.createDataFrame([(str(e), source_path, last_watermark)],
                             ["error", "source_path", "watermark"]) \
             .write.mode("append") \
             .saveAsTable(f"{catalog}.bronze.ingestion_errors")
        raise

# COMMAND ----------
# Configuration
catalog = spark.conf.get("catalog", "production")
source_path = spark.conf.get("source_path", "s3://customer-data/events/")

# Get last processed date (watermark)
last_watermark = get_last_watermark(catalog, "customer_events")
print(f"Last watermark: {last_watermark}")

# Load only NEW events (incremental)
events_df = load_customer_events_incremental(source_path, last_watermark, catalog)

# Write incrementally (append new data)
if events_df.count() > 0:
    events_df.write.mode("append").saveAsTable(f"{catalog}.bronze.customer_events")
    print(f"✅ Appended {events_df.count()} new events")
else:
    print("ℹ️ No new data to process")

# Display summary
events_df.groupBy("event_date").count().orderBy("event_date").show()
```

### Develop Silver Layer Notebook

`src/silver/aggregate_customer_features.py`:

```python
# Databricks notebook source
# MAGIC %md
# MAGIC # Aggregate Customer Features (Deterministic)
# MAGIC Calculate purchase behavior metrics for each customer using MERGE for incremental updates

# COMMAND ----------
from pyspark.sql import DataFrame
from pyspark.sql.functions import col, sum, count, avg, max, min, datediff, lit, current_date
from delta.tables import DeltaTable

# COMMAND ----------
def aggregate_customer_features(events_df: DataFrame, as_of_date: str) -> DataFrame:
    """Aggregate customer purchase behavior into features.

    Args:
        events_df: Customer events DataFrame
        as_of_date: Reference date for calculating days_since_last (DETERMINISTIC)

    Note: Using as_of_date parameter instead of current_date() makes this reproducible
    """
    features_df = (events_df
        .groupBy("customer_id")
        .agg(
            sum("amount").alias("total_purchases"),
            count("*").alias("purchase_count"),
            avg("amount").alias("avg_order_value"),
            min("event_date").alias("first_purchase"),
            max("event_date").alias("last_purchase")
        )
        .withColumn("days_since_last", datediff(lit(as_of_date), col("last_purchase")))
        .withColumn("feature_date", lit(as_of_date))
    )

    return features_df

def merge_features_incremental(catalog: str, new_features_df: DataFrame):
    """Merge new features into target table (upsert pattern)."""
    target_table = f"{catalog}.silver.customer_features"

    try:
        # Check if target table exists
        if spark.catalog.tableExists(target_table):
            # MERGE (upsert) - update existing, insert new
            target = DeltaTable.forName(spark, target_table)

            target.alias("target").merge(
                new_features_df.alias("source"),
                "target.customer_id = source.customer_id"
            ).whenMatchedUpdateAll() \
             .whenNotMatchedInsertAll() \
             .execute()

            print(f"✅ Merged {new_features_df.count()} customer features")
        else:
            # First run - create table
            new_features_df.write.mode("overwrite").saveAsTable(target_table)
            print(f"✅ Created table with {new_features_df.count()} customer features")

    except Exception as e:
        print(f"🚨 Error merging features: {str(e)}")
        # Log error
        spark.createDataFrame([(str(e), target_table)], ["error", "table"]) \
             .write.mode("append") \
             .saveAsTable(f"{catalog}.silver.transformation_errors")
        raise

# COMMAND ----------
# Configuration
catalog = spark.conf.get("catalog", "production")
as_of_date = spark.conf.get("as_of_date", current_date().strftime("%Y-%m-%d"))  # Default to today

print(f"Computing features as of: {as_of_date}")

try:
    # Load bronze data
    events_df = spark.table(f"{catalog}.bronze.customer_events")

    # Validate input
    if events_df.count() == 0:
        print("⚠️ No events in bronze layer, skipping feature aggregation")
        dbutils.notebook.exit("NO_DATA")

    # Transform with deterministic as_of_date
    features_df = aggregate_customer_features(events_df, as_of_date)

    # Merge incrementally (upsert)
    merge_features_incremental(catalog, features_df)

    # Display summary
    print(f"✅ Features computed for {features_df.count()} customers")
    features_df.select("customer_id", "total_purchases", "purchase_count", "days_since_last").show(10)

except Exception as e:
    print(f"🚨 Pipeline failed: {str(e)}")
    dbutils.notebook.exit(f"FAILED: {str(e)}")
```

### Develop Gold Layer Notebook

`src/gold/segment_customers.py`:

```python
# Databricks notebook source
# MAGIC %md
# MAGIC # Customer Segmentation (Business Rules)
# MAGIC Apply deterministic business rules to categorize customers into segments

# COMMAND ----------
from pyspark.sql import DataFrame
from pyspark.sql.functions import col, when, lit
from delta.tables import DeltaTable

# COMMAND ----------
def segment_customers(features_df: DataFrame) -> DataFrame:
    """Apply business rules to segment customers.

    Segmentation Logic:
    - VIP: Total purchases > $10,000
    - Active: Purchased in last 30 days
    - At-Risk: Last purchase 30-90 days ago
    - Churned: No purchase in 90+ days

    Args:
        features_df: Customer features DataFrame

    Returns:
        DataFrame with customer_id, segment, assigned_date
    """
    segments_df = (features_df
        .withColumn("segment",
            when(col("total_purchases") > 10000, "VIP")
            .when(col("days_since_last") <= 30, "Active")
            .when(col("days_since_last") <= 90, "At-Risk")
            .otherwise("Churned")
        )
        .withColumn("assigned_date", col("feature_date"))
        .select("customer_id", "segment", "assigned_date")
    )

    return segments_df

# COMMAND ----------
# Configuration
catalog = spark.conf.get("catalog", "production")

try:
    # Load silver data
    features_df = spark.table(f"{catalog}.silver.customer_features")

    # Validate input
    if features_df.count() == 0:
        print("⚠️ No features available, skipping segmentation")
        dbutils.notebook.exit("NO_DATA")

    # Apply business rules
    segments_df = segment_customers(features_df)

    # Write with MERGE for idempotency
    target_table = f"{catalog}.gold.customer_segments"

    if spark.catalog.tableExists(target_table):
        target = DeltaTable.forName(spark, target_table)
        target.alias("target").merge(
            segments_df.alias("source"),
            "target.customer_id = source.customer_id AND target.assigned_date = source.assigned_date"
        ).whenMatchedUpdateAll() \
         .whenNotMatchedInsertAll() \
         .execute()
        print(f"✅ Merged {segments_df.count()} customer segments")
    else:
        segments_df.write.mode("overwrite").saveAsTable(target_table)
        print(f"✅ Created table with {segments_df.count()} segments")

    # Display summary
    segments_df.groupBy("segment").count().orderBy("segment").show()
    print(f"✅ Segmentation complete: {segments_df.count()} customers")

except Exception as e:
    print(f"🚨 Segmentation failed: {str(e)}")
    # Log error
    spark.createDataFrame([(str(e), catalog)], ["error", "catalog"]) \
         .write.mode("append") \
         .saveAsTable(f"{catalog}.gold.segmentation_errors")
    dbutils.notebook.exit(f"FAILED: {str(e)}")
```

**Note**: This uses simple, deterministic business rules. For ML-based segmentation, see the ML Model Lifecycle guide (not included in this basic example).

## Stage 2: Pre-commit - Unit Tests

### Create Unit Tests

`tests/test_customer_features.py`:

```python
import pytest
from pyspark.sql import SparkSession
from datetime import datetime, timedelta
import sys
sys.path.append("src/silver")
from aggregate_customer_features import aggregate_customer_features

@pytest.fixture
def spark():
    return SparkSession.builder.master("local[1]").getOrCreate()

def test_aggregate_customer_features(spark):
    # Create test data
    test_data = [
        (1, "2024-01-01", 100.0, "evt1"),
        (1, "2024-01-05", 150.0, "evt2"),
        (2, "2024-01-02", 200.0, "evt3"),
    ]

    events_df = spark.createDataFrame(test_data, ["customer_id", "event_date", "amount", "event_id"])

    # Run transformation
    result = aggregate_customer_features(events_df)

    # Assertions
    assert result.count() == 2, "Should have 2 customers"

    customer_1 = result.filter("customer_id = 1").first()
    assert customer_1.total_purchases == 250.0
    assert customer_1.purchase_count == 2
    assert customer_1.avg_order_value == 125.0

def test_empty_input_handling(spark):
    empty_df = spark.createDataFrame([], ["customer_id", "event_date", "amount", "event_id"])
    result = aggregate_customer_features(empty_df)
    assert result.count() == 0
```

### Run Pre-commit Validation

```bash
# Lint code
black src/ tests/
pylint src/

# Run unit tests
pytest tests/ -v

# Validate Asset Bundle
databricks bundle validate
```

## Stage 3-4: Merge Request & Commit

### Create Pull Request

```bash
git checkout -b feature/customer-segmentation
git add specs/ src/ tests/
git commit -m "feat: add customer segmentation pipeline

- Ingest customer events from S3
- Aggregate purchase behavior features
- Apply business rules for segmentation
- Create gold.customer_segments table"

git push origin feature/customer-segmentation
```

### CI Pipeline Validates

GitHub Actions automatically runs:

- Unit tests
- Linting
- Bundle validation
- Security scans

### Merge to Main

After approval, squash merge triggers full pipeline validation.

## Stage 5: Acceptance Testing - PLTE Deployment

### Configure Asset Bundle for Staging

`databricks.yml`:

```yaml
bundle:
  name: customer-segmentation

variables:
  catalog:
    default: production
    lookup:
      development: dev
      staging: staging
      production: production

resources:
  jobs:
    customer_segmentation_pipeline:
      name: Customer Segmentation - ${bundle.target}
      tasks:
        - task_key: ingest_events
          notebook_task:
            notebook_path: ./src/bronze/ingest_customer_events.py
          new_cluster:
            spark_version: 13.3.x-scala2.12
            node_type_id: i3.xlarge
            num_workers: 2

        - task_key: aggregate_features
          depends_on:
            - task_key: ingest_events
          notebook_task:
            notebook_path: ./src/silver/aggregate_customer_features.py

        - task_key: segment_customers
          depends_on:
            - task_key: aggregate_features
          notebook_task:
            notebook_path: ./src/gold/segment_customers.py

targets:
  staging:
    mode: development
    workspace:
      host: https://staging.cloud.databricks.com
    variables:
      catalog: staging

  production:
    mode: production
    workspace:
      host: https://prod.cloud.databricks.com
    variables:
      catalog: production
```

### Deploy to Staging

```bash
# Deploy bundle
databricks bundle deploy --target staging

# Run pipeline
databricks bundle run customer_segmentation_pipeline --target staging

# Monitor run
databricks jobs runs get-output --run-id <run-id>
```

### Verify Output

```sql
-- Check segments created
SELECT segment, COUNT(*) as customer_count, AVG(segment_score) as avg_score
FROM staging.gold.customer_segments
GROUP BY segment
ORDER BY customer_count DESC;

-- Verify data quality
SELECT
  COUNT(*) as total_customers,
  COUNT(DISTINCT customer_id) as unique_customers,
  COUNT(*) - COUNT(DISTINCT customer_id) as duplicates,
  SUM(CASE WHEN segment IS NULL THEN 1 ELSE 0 END) as null_segments
FROM staging.gold.customer_segments;
```

## Stage 8-10: Release and Production Deployment

### Tag Release

```bash
git checkout main
git tag -a v2024.01.15 -m "Release: Customer Segmentation Pipeline

Features:
- Customer segmentation with 4 segments (VIP, Active, At-Risk, Churned)
- Daily automated pipeline
- Data quality validations

Schema Changes:
- Created gold.customer_segments table
"

git push origin v2024.01.15
```

### Deploy to Production

```bash
# Deploy production bundle
databricks bundle deploy --target production

# Schedule job for daily execution
databricks jobs update <job-id> --json '{
  "schedule": {
    "quartz_cron_expression": "0 0 2 * * ?",
    "timezone_id": "UTC"
  }
}'

# Trigger initial run
databricks bundle run customer_segmentation_pipeline --target production
```

## Stage 11: Monitor Production

### Set Up Alerts

```python
# Create monitoring notebook
from databricks.sdk.runtime import dbutils

# Check data freshness
latest_date = spark.sql("SELECT MAX(assigned_date) FROM production.gold.customer_segments").first()[0]
delay_hours = (datetime.now().date() - latest_date).total_seconds() / 3600

if delay_hours > 6:
    dbutils.notebook.exit(json.dumps({
        "status": "ALERT",
        "message": f"Data is {delay_hours:.1f} hours old"
    }))

# Check segment distribution
distribution = spark.sql("""
    SELECT segment, COUNT(*) * 100.0 / SUM(COUNT(*)) OVER() as pct
    FROM production.gold.customer_segments
    GROUP BY segment
""").collect()

for row in distribution:
    if row.pct < 10 or row.pct > 50:
        dbutils.notebook.exit(json.dumps({
            "status": "WARNING",
            "message": f"Segment {row.segment} has unusual distribution: {row.pct:.1f}%"
        }))
```

## Next Steps

- [Testing Data Pipelines](./testing-data-pipelines.md) - Deep dive on testing strategies
- [CI/CD Pipeline for Databricks](./cicd-pipeline-databricks.md) - Automate the complete workflow
- [Databricks Asset Bundles](./databricks-asset-bundles.md) - Master bundle configuration
