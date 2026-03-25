# Configure Azure MCP Server

Set up the Azure MCP server so Claude Code can create and manage Azure resources
directly from the chat interface.

## The Windows Credential Problem

The Azure MCP server runs as a Docker Linux container. On Windows, `az login`
encrypts credentials using DPAPI — a Windows-only mechanism. The Linux container
cannot decrypt them even when the `.azure` directory is volume-mounted.

There are two solutions:

| Option                                             | How it authenticates                                    | Prerequisite                                 |
| -------------------------------------------------- | ------------------------------------------------------- | -------------------------------------------- |
| **A — Service Principal** (recommended on Windows) | Four env vars passed to the container                   | An Azure service principal                   |
| **B — WSL credential store**                       | Unencrypted credential files mounted into the container | Ubuntu/Debian WSL distro with `az` installed |

> **Note**: Docker Desktop ships with a minimal `docker-desktop` WSL distro that
> has no bash and no `az`. Option B requires a **separate** Ubuntu or Debian WSL
> distro (`wsl --install -d Ubuntu`). If you only have `docker-desktop`, use
> Option A.

---

## Option A — Service Principal (recommended on Windows)

### How it works

A Service Principal (SP) is an Azure app identity — like a service account.
You create one with scoped permissions, then authenticate by setting four
environment variables on your Windows machine. Docker inherits them from the
host and passes them to the container. The Azure SDK reads them automatically
via `DefaultAzureCredential`.

No credential files, no WSL, no encryption issues.

### Step 1 — Create a Service Principal

Run this from your Windows terminal (requires an active `az login` session):

```powershell
$sub = az account show --query id -o tsv

az ad sp create-for-rbac `
  --name eac-mcp-local `
  --role Contributor `
  --scopes /subscriptions/$sub `
  --output json
```

Save the output — you will not see the secret again:

```json
{
  "appId":       "<AZURE_CLIENT_ID>",
  "password":    "<AZURE_CLIENT_SECRET>",
  "tenant":      "<AZURE_TENANT_ID>"
}
```

> **Scope**: Using subscription-level Contributor is convenient for local dev.
> For tighter security, scope to a specific resource group instead:
> `--scopes /subscriptions/$sub/resourceGroups/rg-eac-test`

### Step 2 — Set environment variables on Windows

Set the four variables as **permanent user environment variables** so they
survive reboots and are inherited by Docker:

```powershell
$sub = az account show --query id -o tsv

[System.Environment]::SetEnvironmentVariable("AZURE_CLIENT_ID",       "<appId>",    "User")
[System.Environment]::SetEnvironmentVariable("AZURE_CLIENT_SECRET",   "<password>", "User")
[System.Environment]::SetEnvironmentVariable("AZURE_TENANT_ID",       "<tenant>",   "User")
[System.Environment]::SetEnvironmentVariable("AZURE_SUBSCRIPTION_ID", $sub,         "User")
```

Verify they are set:

```powershell
[System.Environment]::GetEnvironmentVariable("AZURE_CLIENT_ID", "User")
```

> **Security**: User-scoped env vars are visible only to your Windows account.
> Never commit these values to source control.

### Step 3 — Check `.mcp.json`

The repository `.mcp.json` is already configured to pass the env vars to the
container:

```json
"azure": {
  "type": "stdio",
  "command": "docker",
  "args": [
    "run", "-i", "--rm",
    "-e", "AZURE_CLIENT_ID",
    "-e", "AZURE_CLIENT_SECRET",
    "-e", "AZURE_TENANT_ID",
    "-e", "AZURE_SUBSCRIPTION_ID",
    "mcr.microsoft.com/azure-sdk/azure-mcp:latest"
  ],
  "env": {}
}
```

The `-e VARNAME` syntax (no `=value`) tells Docker to inherit the value from the
host environment at container startup.

### Step 4 — Restart Claude Code

Close and reopen Claude Code. The `mcp__azure__*` tools will be available.

Verify with `/go:status` — you should see the `azure` MCP server listed as
connected.

### Rotating credentials

Service principal secrets expire based on your policy (default 1 year). To
rotate:

```powershell
az ad sp credential reset --name eac-mcp-local --output json
# Update AZURE_CLIENT_SECRET env var with the new password
[System.Environment]::SetEnvironmentVariable("AZURE_CLIENT_SECRET", "<new-password>", "User")
```

Restart Claude Code after updating the variable.

---

## Option B — WSL Credential Store

Use this if you have Ubuntu or Debian WSL installed and prefer not to create a
service principal.

### Prerequisites

- Ubuntu or Debian WSL distro installed (`wsl --install -d Ubuntu`)
- Azure CLI installed in that distro

### Step 1 — Create a WSL credential store

Open an **Ubuntu WSL terminal** and run:

```bash
mkdir /mnt/c/users/<username>/.azure-wsl
AZURE_CONFIG_DIR=/mnt/c/users/<username>/.azure-wsl az login
```

This stores credentials as plain JSON files readable by a Linux container.
Your Windows `az` session is unaffected.

### Step 2 — Update `.mcp.json`

Change the azure entry to mount the WSL credential directory:

```json
"azure": {
  "type": "stdio",
  "command": "docker",
  "args": [
    "run", "-i", "--rm",
    "--volume", "C:/Users/<username>/.azure-wsl:/root/.azure",
    "mcr.microsoft.com/azure-sdk/azure-mcp:latest"
  ],
  "env": {}
}
```

### Step 3 — Restart Claude Code

Tokens expire like any `az` session. Re-run Step 1 to refresh — no Claude Code
restart needed.

---

## How Authentication Works

The container uses `DefaultAzureCredential`, which checks sources in this order:

1. `AZURE_CLIENT_ID` + `AZURE_CLIENT_SECRET` + `AZURE_TENANT_ID` env vars → **Option A**
2. Mounted `.azure` directory with valid token cache → **Option B**
3. Managed Identity (when running on Azure infrastructure)

| Scenario                      | Recommended option                               |
| ----------------------------- | ------------------------------------------------ |
| Windows, Docker Desktop only  | Option A (Service Principal)                     |
| Windows, Ubuntu WSL installed | Option A or B                                    |
| Linux / macOS                 | Option B (plain `~/.azure` mount, no WSL needed) |
| CI / GitHub Actions           | Option A (SP credentials as secrets)             |

---

## Troubleshooting

### Tools not appearing after restart

The env vars must be set **before** starting Claude Code. Confirm they are
present in a new terminal:

```powershell
echo $env:AZURE_CLIENT_ID
```

If empty, the `SetEnvironmentVariable` call may have used the wrong scope, or
the terminal was opened before the variables were set. Open a fresh terminal and
check again.

### Container exits immediately with no output

Normal behaviour when tested interactively — the MCP server exits when stdin
closes. Claude Code holds stdin open during a session.

Test that the container starts and authenticates:

```powershell
docker run --rm `
  -e AZURE_CLIENT_ID `
  -e AZURE_CLIENT_SECRET `
  -e AZURE_TENANT_ID `
  -e AZURE_SUBSCRIPTION_ID `
  mcr.microsoft.com/azure-sdk/azure-mcp:latest --help
```

### `Unauthorized` or `Authentication failed` errors

The service principal credentials are invalid or expired. Verify:

```powershell
az login --service-principal `
  -u $env:AZURE_CLIENT_ID `
  -p $env:AZURE_CLIENT_SECRET `
  --tenant $env:AZURE_TENANT_ID
az account show
```

If this fails, rotate the credentials (see "Rotating credentials" above).

---

## Related

- [Platform Troubleshooting](./platform-troubleshooting.md) — Docker Desktop and WSL issues
- [Configure Claude Code](./configure-claude-code.md) — MCP server overview
- [Azure MCP Server docs](https://github.com/microsoft/mcp/tree/main/servers/Azure.Mcp.Server)
- [Azure SDK issue #19167](https://github.com/Azure/azure-sdk-for-net/issues/19167) — Windows credential encryption root cause
