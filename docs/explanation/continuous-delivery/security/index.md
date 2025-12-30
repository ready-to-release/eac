# Security

Security integration throughout all stages of the CD Model using open-source tools.

## In This Section

| Topic                                | Description                                                   |
| ------------------------------------ | ------------------------------------------------------------- |
| [Shift-Left Security](shift-left.md) | Philosophy, defense in depth, and stage integration matrix    |
| [SAST](sast.md)                      | Static Application Security Testing with Trivy                |
| [DAST](dast.md)                      | Dynamic Application Security Testing with OWASP ZAP           |
| [Supply Chain](supply-chain.md)      | Dependency scanning, Dependabot, and container security       |
| [Remediation](remediation.md)        | Vulnerability workflow, blocking strategy, and best practices |

## Tools

| Tool                                                              | Purpose                        | Cost |
| ----------------------------------------------------------------- | ------------------------------ | ---- |
| [Trivy](https://aquasecurity.github.io/trivy/)                    | SAST, dependencies, containers | Free |
| [OWASP ZAP](https://www.zaproxy.org/)                             | DAST for web applications      | Free |
| [Dependabot](https://docs.github.com/en/code-security/dependabot) | Automated dependency updates   | Free |
