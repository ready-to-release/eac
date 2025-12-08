# Provisioning PLTE Environments

How to create and manage Production-Like Test Environments.

## Ephemeral PLTE Lifecycle

### Step 1: Trigger

PLTE provisioning is triggered by:

- Merge to main branch
- Release candidate creation
- Manual request for feature branch testing

### Step 2: Provision (5-10 minutes)

Create infrastructure from IaC templates:

```bash
# Terraform example
terraform init
terraform plan -var="environment=plte-${BUILD_ID}"
terraform apply -auto-approve
```

### Step 3: Deploy

Install application and seed test data:

```bash
# Deploy application
kubectl apply -f k8s/deployment.yaml

# Seed test data
./scripts/seed-test-data.sh
```

### Step 4: Test (1-4 hours)

Run acceptance and extended tests:

```bash
# Acceptance tests (IV, OV, PV)
eac test acceptance

# Extended tests (performance, security)
eac test extended
```

### Step 5: Destroy

Tear down infrastructure after testing:

```bash
terraform destroy -auto-approve
```

## Infrastructure as Code Template

```hcl
# plte/main.tf
variable "environment" {
  description = "PLTE environment identifier"
}

resource "azurerm_resource_group" "plte" {
  name     = "rg-plte-${var.environment}"
  location = "westeurope"

  tags = {
    environment = "plte"
    ephemeral   = "true"
    build_id    = var.environment
  }
}

# Add app services, databases, etc.
```

## Cost Management

- **Short-lived**: Hours, not days
- **Automated cleanup**: Destroy after testing
- **Resource limits**: Prevent overprovisioning
- **Scheduled cleanup**: Nightly job to catch orphans

```bash
# Cleanup orphaned PLTEs older than 24 hours
./scripts/cleanup-old-plte.sh --max-age 24h
```

## PLTE Characteristics

- Production-like infrastructure
- Production-like configuration (without production credentials)
- Realistic test data (anonymized)
- Isolated per feature/release
- Network isolation for security

## Related

- [Environments](./environments.md)
- [Deployment Strategies](../deployment/deployment-strategies.md)

---

_[Tutorials](../../../tutorials/) | [How-to Guides](../../../how-to-guides/) | **Explanation** | [Reference](../../../reference/)_

**You are here:** Explanation — understanding-oriented discussion that clarifies concepts.
