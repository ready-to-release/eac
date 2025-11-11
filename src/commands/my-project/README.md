# Project Templates

Templates for the **Everything as Code** (EAC) paradigm. These templates help you initialize projects following continuous delivery, compliance automation, and executable specification best practices.

---

## Quick Start

### 1. Discover Required Variables

```bash
# See what placeholders this template repository uses
# and which files use each placeholder
r2r templates list
```

**Example output:**

```text
Template Placeholders in 'https://github.com/ready-to-release/eac':
----------------------------
  my-awesome-project
    - design/dr.md
    - implementation-report.md
    - README.md
  <no value>
    - design/dr.md
    - README.md
  <no value>
    - design/dr.md
    - README.md

Total: 3 placeholders
```

### 2. Create Values File

```bash
# Create your values file
# Only ProjectName is required - all other values are optional
cat > values.json <<EOF
{
  "ProjectName": "my-awesome-project"
}
EOF
```

**Note**: Only `ProjectName` is required. All other template variables are optional and will be replaced with empty strings if not provided. These can be filled in later by automated delivery processes.

### 3. Install Templates

```bash
# Install templates to your project
r2r templates install --values values.json --location .
```

---

## What's Included

### Design Templates

**`design/dr.md`** - Decision Record Template

- Document architectural decisions
- Track context, decision, and consequences
- Includes status tracking (Proposed/Accepted/Rejected/etc.)

**Required Variables**: `ProjectName`

### Specification Templates

**`specs/specification.feature`** - Gherkin Feature Template

- Executable specifications using Gherkin
- Follows ATDD/BDD three-layer approach
- Template-based structure for features and scenarios

**`specs/risk-controls/risk-control.feature`** - Risk Control Template

- Risk control requirements from assessments
- Links regulatory requirements to implementations
- Includes examples for authentication, data protection, audit, access control, privacy, and AI/ML controls

**Required Variables**: `ProjectName`

### Documentation Templates

**`implementation-report.md`** - Implementation Report Template

- Compliance-ready implementation documentation
- Requirements traceability
- Test results and verification (IV/OV/PV)
- Most values filled automatically by delivery processes

**Required Variables**: `ProjectName`
**Optional Variables**: See `values.example.json` for CI/CD and automated delivery values

---

## Installation Methods

### Method 1: Install from Default Repository (Recommended)

```bash
# Uses this repository by default
r2r templates install --values values.json --location ./my-project
```

### Method 2: Install from Specific Repository

```bash
# Use a custom template repository
r2r templates install \
  --template https://github.com/myorg/custom-templates \
  --values values.json \
  --location ./my-project
```

### Method 3: Install to Existing Project

```bash
# Add templates to existing project
cd my-existing-project
r2r templates install --values values.json --location .

# Review changes
git diff

# Keep what you want, discard what you don't
git add -p
git commit -m "Add EAC templates"
```

---

## Example: Initialize New Project

### Minimal Setup (Recommended)

```bash
# Create minimal values file - only ProjectName is required
cat > values.json <<EOF
{
  "ProjectName": "payment-service"
}
EOF

# Install templates
r2r templates install --values values.json --location ./payment-service

# Result:
# payment-service/
# ├── design/
# │   └── dr.md                           # Project: payment-service
# ├── specs/
# │   ├── specification.feature          # Template ready for use
# │   └── risk-controls/
# │       └── risk-control.feature       # Template with examples
# └── implementation-report.md            # Project: payment-service (automation fields empty)
```

### With Optional Values

```bash
# Provide additional values if needed
cat > values.json <<EOF
{
  "ProjectName": "payment-service",
  "Author": "Development Team",
  "Date": "2025-01-15",
  "ModuleName": "payments",
  "Context": "Payment processing for e-commerce platform"
}
EOF

# Install templates
r2r templates install --values values.json --location ./payment-service
```

**Note**: Most template variables (like `changed_requirements`, `feature_test_results`, etc.) are intended to be filled by automated delivery processes during CI/CD pipelines.

---

## Updating Templates

```bash
# Get latest template versions
r2r templates install --values values.json --location .

# Review what changed
git diff

# Accept or reject changes
git add -p
git commit -m "Update templates to latest version"
```

---

## Available Variables

### Required Variable

- **`ProjectName`** - Project name used across all templates (Required)

### Optional Variables

**User-provided** (optional, can be filled manually):

- `Author` - Document author name
- `Date` - Current date
- `Context` - Project or decision context
- `ModuleName` - Module or component name

**CI/CD Pipeline** (typically filled by automation):

- `ChangeType`, `PipelineID`, `Repository`, `Branch`, `BuildDate`, `TriggeredBy`

**Automated Delivery** (filled by delivery automation in implementation reports):

- `changed_requirements`, `req_approval_comments`, `release_notes`, `requirements`
- `feature_test_results`, `iv_test_traceability_report`, `ov_test_traceability_report`
- `pv_test_traceability_report`, `specs_and_test_results`

See [`values.example.json`](values.example.json) for example values and detailed categorization.

### Discover Variables

To see all variables used in a template repository:

```bash
# For default EAC repository (shows file locations)
r2r templates list

# For custom repository
r2r templates list --template https://github.com/myorg/templates

# For local templates
r2r templates list --template ./my-templates
```

The `templates list` command shows:

- ✅ All placeholder variables in the templates
- ✅ Which template files use each variable
- ✅ Example values.json structure

---

## Customization

### Override Default Values

Templates include guidance notes (in `!!! note` blocks) to help you replace content:

```markdown
## Context

<no value>

!!! note

    Describe the context and problem statement motivating this decision
```

### Remove Guidance Notes

After filling in your content, remove the `!!! note` blocks if desired:

```bash
# Remove all guidance notes
find . -type f -name "*.md" -exec sed -i '/^!!! note/,/^$/d' {} \;
```

---

## Template Development

### Creating Custom Templates

1. **Clone this repository**:

   ```bash
   git clone https://github.com/ready-to-release/eac my-templates
   cd my-templates/templates
   ```

2. **Add placeholders** using `<no value>` syntax:

   ```markdown
   # my-awesome-project Documentation

   Author: <no value>
   Date: <no value>
   ```

3. **Test locally**:

   ```bash
   r2r templates list --template .
   r2r templates install --template file://$(pwd) --values test-values.json --location /tmp/test
   ```

4. **Push to repository**:

   ```bash
   git add .
   git commit -m "Update templates"
   git push
   ```

### Placeholder Syntax

- **In content**: `<no value>`
- **In filenames**: `<no value>.md`
- **In directory names**: `specs/<no value>/`

**Rules**:

- Variable names must start with a letter
- Can contain letters, numbers, underscores
- Case-sensitive

---

## Support

- **Documentation**: [EAC Documentation](https://ready-to-release.github.io/eac/)
- **Issues**: [GitHub Issues](https://github.com/ready-to-release/eac/issues)

---

## License

This template repository is part of the Everything as Code project:

- **Software**: MIT License
- **Documentation**: CC BY-SA 4.0
