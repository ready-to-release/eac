#!/usr/bin/env bash
# Container integration tests for eac-ext
# Env: IMAGE_NAME, IMAGE_TAG, MODULE, GITHUB_WORKSPACE
set -euo pipefail

# Setup clie for integration testing
cp out/build/clie/clie_go/clie-linux-amd64 ./clie 2>/dev/null || true
chmod +x ./clie 2>/dev/null || true

# Test via clie CLI
mkdir -p /tmp/eac-ext-test
cd /tmp/eac-ext-test
mkdir -p .git
cp -r "$GITHUB_WORKSPACE/contracts" ./contracts

mkdir -p .clie
cat > .clie/clie.yml << CLIEOF
extensions:
  - name: 'eac'
    image: 'eac-ext:ci-test'
    image_pull_policy: 'Never'
CLIEOF

cp "$GITHUB_WORKSPACE/clie" ./clie
chmod +x ./clie

echo "Testing 'show modules'..."
timeout 120 ./clie run eac show modules

echo "Testing 'extension-meta'..."
timeout 30 ./clie run eac extension-meta

# Test direct Docker
cd "$GITHUB_WORKSPACE"
echo "Test: help command"
docker run --rm -v "$(pwd):/var/task" "$IMAGE_NAME:$IMAGE_TAG" help 2>&1 | head -20 || true
echo "Test: extension-meta"
docker run --rm "$IMAGE_NAME:$IMAGE_TAG" extension-meta
