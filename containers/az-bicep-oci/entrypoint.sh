#!/bin/bash
#
# Thin az CLI wrapper. Auth via env vars, everything else via args.
# The Go deploy command reads environments.yml + component config,
# resolves all values, and passes the full az CLI args.
#
# Env vars (auth only, forwarded by host-env):
#   AZURE_TENANT_ID
#   AZURE_CLIENT_ID
#   AZURE_CLIENT_SECRET
#
# Usage: entrypoint.sh <az cli args...>
# Example: entrypoint.sh deployment group create \
#            --resource-group rg-myapp --template-file /app/main.bicep \
#            --parameters @/app/configuration/development.json \
#            --name eac-20260302-143025 --output json

set -euo pipefail

# Login with service principal credentials
az login --service-principal \
    --tenant "${AZURE_TENANT_ID:?AZURE_TENANT_ID is required}" \
    --username "${AZURE_CLIENT_ID:?AZURE_CLIENT_ID is required}" \
    --password "${AZURE_CLIENT_SECRET:?AZURE_CLIENT_SECRET is required}" \
    --output none

# Pass all args straight through to az
exec az "$@"
