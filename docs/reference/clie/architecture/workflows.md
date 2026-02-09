# Workflows

Dynamic workflow diagrams showing the execution flow for key CLIE CLI operations.

## Run Workflow

The flow when executing `clie run <extension> <command>` or `clie <extension> <command>`.

<!-- structurizr:clie:RunWorkflow -->

**Workflow steps:**

1. Developer executes command (e.g., `clie eac build`)
2. Command layer parses arguments via EBNF parser
3. Configuration loaded from `.clie/clie.yml`
4. Extension system ensures image exists locally
5. GitHub registry checked for updates (if pull policy allows)
6. Image pulled if needed
7. Container created with volume mounts
8. Container started with I/O attached
9. Wait for container completion
10. Exit code returned to caller

---

## Install Workflow

The flow when installing extensions via `clie install <extension>`.

<!-- structurizr:clie:InstallWorkflow -->

**Workflow steps:**

1. Developer executes `clie install <extension>`
2. GitHub registry queried for latest SHA tag
3. Available tags fetched and filtered
4. Configuration updated with SHA-pinned image reference
5. Image pulled from registry
6. Success reported to user

**SHA Pinning:**

Extensions are pinned to specific SHA digests for reproducibility:

```yaml
extensions:
  - name: 'eac'
    image: 'ghcr.io/ready-to-release/eac-ext@sha256:abc123...'
```

This ensures:

- Reproducible builds across environments
- Protection against supply chain attacks
- Explicit control over extension versions

## See Also

- [System Context](./system-context.md) - High-level architecture
- [Subsystems](./subsystems.md) - Component details
- [Install Command](../commands/install.md) - Install command reference
