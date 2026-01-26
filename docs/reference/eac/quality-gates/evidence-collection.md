# Evidence Collection

Test execution and security scan evidence formats for quality gate validation.

---

## Test Execution Evidence

### Required Evidence (Stages 2-6)

| Evidence Type            | Format                    | Source                      |
| ------------------------ | ------------------------- | --------------------------- |
| Unit test results        | JUnit XML                 | `go test`, `pytest`, `jest` |
| Integration test results | JUnit XML                 | CI pipeline                 |
| Acceptance test results  | JUnit XML + Cucumber JSON | BDD test runner             |
| Code coverage            | Cobertura XML, HTML       | Coverage tools              |

### JUnit XML Format

Standard test result format:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="42" failures="0" errors="0" time="12.345">
  <testsuite name="TestUserService" tests="10" failures="0" time="1.234">
    <testcase name="TestCreateUser" classname="user_service" time="0.123"/>
    <testcase name="TestDeleteUser" classname="user_service" time="0.089"/>
  </testsuite>
</testsuites>
```

### Generate Test Evidence

```bash
# Go with JUnit output
go test ./... -v 2>&1 | go-junit-report > test-results.xml

# Python with pytest
pytest --junitxml=test-results.xml

# JavaScript with Jest
jest --reporters=jest-junit
```

---

## Security Scan Evidence

### Required Evidence (Stages 2, 3, 6)

| Scan Type           | Tool            | Format      |
| ------------------- | --------------- | ----------- |
| SAST                | Semgrep, Gosec  | SARIF, JSON |
| Dependency scanning | Trivy           | SARIF, JSON |
| Container scanning  | Trivy           | SARIF, JSON |
| Secret detection    | Trivy, Gitleaks | JSON        |
| DAST                | OWASP ZAP       | JSON, HTML  |

### SARIF Format

Static Analysis Results Interchange Format:

```json
{
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "version": "2.1.0",
  "runs": [{
    "tool": {
      "driver": {
        "name": "Semgrep",
        "version": "1.0.0"
      }
    },
    "results": [{
      "ruleId": "go.lang.security.injection.sql-injection",
      "level": "error",
      "message": { "text": "SQL injection vulnerability" },
      "locations": [{
        "physicalLocation": {
          "artifactLocation": { "uri": "src/db/query.go" },
          "region": { "startLine": 42 }
        }
      }]
    }]
  }]
}
```

### Generate Security Evidence

```bash
# Trivy (SARIF output)
trivy fs --format sarif --output trivy-results.sarif .

# Semgrep (SARIF output)
semgrep --sarif --output semgrep-results.sarif .

# Gosec (SARIF output)
gosec -fmt sarif -out gosec-results.sarif ./...

# Gitleaks (JSON output)
gitleaks detect --report-format json --report-path gitleaks-results.json
```

---

## Performance Evidence

### Required Evidence (Stage 6)

| Metric                        | Tool                | Format    |
| ----------------------------- | ------------------- | --------- |
| Response time (P50, P95, P99) | k6, JMeter, Gatling | JSON, XML |
| Throughput (req/sec)          | k6, JMeter          | JSON, XML |
| Resource utilization          | Prometheus, Grafana | JSON      |
| Regression analysis           | Custom              | JSON      |

### k6 Output Format

```bash
# Run load test with JSON output
k6 run --out json=results.json load-test.js
```

```json
{
  "metrics": {
    "http_req_duration": {
      "type": "trend",
      "contains": "time",
      "values": {
        "avg": 45.23,
        "min": 12.1,
        "med": 38.5,
        "max": 234.7,
        "p(90)": 89.2,
        "p(95)": 112.4
      }
    }
  }
}
```

---

## Evidence Directory Structure

```text
out/
├── <module>/
│   ├── test/
│   │   ├── unit-results.xml          # JUnit XML
│   │   ├── integration-results.xml   # JUnit XML
│   │   ├── acceptance-results.xml    # JUnit XML
│   │   ├── coverage.xml              # Cobertura
│   │   └── coverage.html             # HTML report
│   ├── scan/
│   │   ├── sast-results.sarif        # SARIF
│   │   ├── vuln-results.sarif        # SARIF
│   │   ├── secrets-results.json      # JSON
│   │   └── container-results.sarif   # SARIF
│   └── perf/
│       ├── load-test-results.json    # k6 JSON
│       └── regression-analysis.json  # Custom
└── risk/
    └── assessment-results.json       # OSCAL
```

---

## Evidence Commands

```bash
# Build test evidence
r2r eac update evidence <module>

# Show evidence summary
r2r eac show test-summary <module>
r2r eac show scan-summary <module>

# Generate OSCAL assessment
r2r eac create risk-assess --profile specs/.risk-controls/risk-profile.json
```

---

## Related Documentation

- [Test Suites](../testing/test-suites.md) - Test execution commands
- [Security Scanning](../security/) - Security scan commands
- [Risk Controls](../compliance/risk-controls.md) - OSCAL assessment generation
