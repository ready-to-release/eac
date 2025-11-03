# Setup Gauge

Install and configure Gauge for ATDD acceptance testing.

---

## Prerequisites

- Go 1.21 or later installed
- Command-line access (terminal/PowerShell)

---

## Step 1: Install Gauge

### On macOS (Homebrew)

```bash
brew install gauge
```

### On Linux

```bash
curl -SsL https://downloads.gauge.org/stable | sh
```

### On Windows (Chocolatey)

```powershell
choco install gauge
```

### Alternative: Install via Go

```bash
go install github.com/getgauge/gauge@latest
```

---

## Step 2: Verify Gauge Installation

Check that Gauge is installed correctly:

```bash
gauge version
```

**Expected output**:

```text
Gauge version: <version>
```

---

## Step 3: Install Gauge Go Plugin

Gauge needs a language plugin to execute tests. Install the Go plugin:

```bash
gauge install go
```

**Expected output**:

```text
Successfully installed plugin 'go'.
```

---

## Step 4: Verify Go Plugin Installation

List installed plugins:

```bash
gauge list
```

**Expected output should include**:

```text
go (<version>)
```

---

## Step 5: Configure Project for Gauge

### Create Gauge Configuration Directory

In your project root:

```bash
mkdir -p .gauge
```

### Create gauge.properties

Create `.gauge/gauge.properties` with default settings:

```bash
cat > .gauge/gauge.properties << 'EOF'
# Default gauge properties
screenshot_on_failure = true
enable_multithreading = false
gauge_reports_dir = test-results/gauge
EOF
```

### Create requirements Directory

Create the directory structure for test specifications:

```bash
mkdir -p requirements
```

---

## Step 6: Verify Setup

### Create Test Specification

Create a simple test spec to verify everything works:

```bash
mkdir -p requirements/test
cat > requirements/test/acceptance.spec << 'EOF'
# Test Specification

## Verification Test

* This is a test step
EOF
```

### Create Test Implementation

Create the step implementation:

```bash
cat > requirements/test/acceptance_test.go << 'EOF'
package test

import (
    "github.com/getgauge-contrib/gauge-go/gauge"
)

func init() {
    gauge.Step("This is a test step", testStep)
}

func testStep() {
    // Test implementation
}
EOF
```

### Initialize Go Module (if needed)

If you don't have a go.mod file in requirements/test/:

```bash
cd requirements/test
go mod init test
go get github.com/getgauge-contrib/gauge-go
cd ../..
```

### Run Verification Test

```bash
gauge run requirements/test/
```

**Expected output**:

```text
# Test Specification

## Verification Test
  * This is a test step       ✓

Successfully generated html-report to => test-results/gauge/html-report/index.html
Specifications: 1 executed     1 passed     0 failed     0 skipped
Scenarios:      1 executed     1 passed     0 failed     0 skipped
```

---

## Step 7: Clean Up Test Files

Remove the test files after verification:

```bash
rm -rf requirements/test
```

---

## Troubleshooting

### Gauge command not found

**Problem**: `gauge: command not found`

**Solution**:

- Ensure Gauge is in your PATH
- Restart terminal after installation
- Try installing via different method

### Go plugin not installed

**Problem**: `Failed to start gauge runner`

**Solution**:

```bash
gauge install go
```

### Permission errors on Linux/macOS

**Problem**: Permission denied during installation

**Solution**:

```bash
sudo curl -SsL https://downloads.gauge.org/stable | sh
```

---

## Next Steps

- ✅ Gauge is now installed and configured
- **Next**: [Setup Godog](./setup-godog.md) for BDD testing
- **Then**: [Create Feature Spec](./create-feature-spec.md) to start testing

---

## Related Documentation

- [Gauge Commands](../../reference/testing/gauge-commands.md) - Command reference
- [ATDD Format](../../reference/testing/atdd-format.md) - Specification format
- [ATDD Concepts](../../explanation/testing/atdd-concepts.md) - Understanding ATDD
- [Official Gauge Docs](https://docs.gauge.org/) - Gauge documentation
