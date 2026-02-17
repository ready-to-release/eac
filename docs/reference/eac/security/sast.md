# Static Application Security Testing (SAST)

Static code analysis for security vulnerabilities using Semgrep and Trivy.

---

## Scanners

```bash
# Static code analysis
eac scan --scanner sast        # Semgrep code analysis

# Secret detection
eac scan --scanner secrets     # Trivy secret scanning

# Infrastructure as Code
eac scan --scanner iac         # Trivy IaC scanning

# All static scanners
eac scan --scanner sast,secrets,iac
```

**Output**: `out/scan/<module>/<scanner>/`

---

## SAST (Semgrep)

Detects security vulnerabilities in source code:

- **Injection flaws**: SQL injection, command injection, XSS
- **Authentication**: Weak auth, session management issues
- **Cryptography**: Weak algorithms, insecure random, hardcoded keys
- **Input validation**: Unvalidated input, path traversal
- **Error handling**: Information disclosure, stack traces

**Languages**: Go, JavaScript/TypeScript, Python, Java, Ruby, C#, PHP

**Rulesets**: OWASP Top 10, CWE Top 25, language-specific best practices

---

## Secret Detection (Trivy)

Finds hardcoded secrets:

- **API keys**: AWS, GCP, Azure credentials
- **Passwords**: Database passwords, service accounts
- **Tokens**: OAuth tokens, JWT secrets, SSH keys
- **Certificates**: Private keys, PEM files

**Scan locations**: Source code, config files, commit history, container images

---

## Infrastructure as Code (Trivy)

Scans IaC configurations for misconfigurations:

- **Terraform**: AWS, GCP, Azure resource misconfigurations
- **Kubernetes**: Pod security, RBAC, network policies
- **Docker**: Dockerfile best practices, image security
- **CloudFormation**: AWS resource security settings

**Checks**: CIS Benchmarks, compliance frameworks

---

## Related Documentation

- **[Security Index](./index.md)** - Security scanning overview
- **[Scan Commands](../commands/scan/index.md)** - Full scan command reference
