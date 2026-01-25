# Release Workflows

Reference for release workflows that publish deployable artifacts.

## Overview

Release workflows build and publish deployable artifacts to production environments. Each deployable module has a dedicated release workflow following a standard pattern with CI verification, version validation, artifact building, and release creation.

**Location:** `.github/workflows/release-{moniker}.yaml`

**Release Type Requirement:** Only modules with `release_type: published` or `release_type: bundle` can have release workflows. Internal modules and modules with `release_type: none` do not trigger releases. See [Understanding Release Types](../release-types.md) for details.

## Release Workflow Pattern

All release workflows follow a consistent structure:

### Standard Structure

1. **CI Verification** - Requires successful CI run for the commit being released
2. **Version Extraction** - Extracts and validates version from git tag or input
3. **Existing Release Check** - Prevents duplicate releases
4. **Build from Source** - Builds release artifacts from source code
5. **Attestations** - Generates build provenance for supply chain security
6. **Release Creation** - Creates GitHub release with artifacts
7. **Verification** - Verifies release assets uploaded successfully

### Trigger Conditions

**Tag-based triggers:**

```yaml
on:
  push:
    tags:
      - '{moniker}/*'
```

**Manual triggers:**

```yaml
on:
  workflow_dispatch:
    inputs:
      version:
        description: 'Version to release (semver format: x.y.z)'
        required: true
        type: string
```

### Permissions

```yaml
permissions:
  contents: write        # Create releases and upload assets
  id-token: write        # Sign attestations (Sigstore)
  attestations: write    # Persist attestations
```

## Release Workflows Inventory

| Workflow               | Module          | Release Type | Artifact Type           | Destination     | Versioning |
| ---------------------- | --------------- | ------------ | ----------------------- | --------------- | ---------- |
| `release-r2r-cli.yaml` | r2r-cli         | published    | Cross-platform binaries | GitHub Releases | SemVer     |
| `release-ext-eac.yaml` | ext-eac         | published    | Docker extension        | Docker Hub      | SemVer     |
| `release-docs.yaml`    | docs            | published    | Static site             | GitHub Pages    | CalVer     |
| `release-books.yaml`   | books           | published    | PDF documents           | GitHub Releases | CalVer     |
| `release-bundle.yaml`  | r2r-eac-bundle  | bundle       | Meta-release bundle     | GitHub Releases | SemVer     |

**Note**: Internal modules (eac-commands, eac-mcp-commands, r2r-installer, vscode-ext-commit) do not have release workflows because they are not released independently.

## Example: release-r2r-cli.yaml

Complete specification for the R2R CLI binary release workflow.

**File:** `.github/workflows/release-r2r-cli.yaml`

**Module:** r2r-cli

**Artifact:** Cross-platform binaries (Linux, macOS, Windows)

### Triggers

```yaml
on:
  push:
    tags:
      - 'r2r-cli/*'
      - 'src-cli/*'  # Legacy tag format for backwards compatibility
  workflow_dispatch:
    inputs:
      version:
        description: 'Version to release (semver format: x.y.z)'
        required: true
        type: string
```

**Tag Format:**

- Current: `r2r-cli/1.0.0`
- Legacy: `src-cli/1.0.0` (supported for backwards compatibility)

### Permissions

```yaml
permissions:
  contents: write        # Create releases and upload assets
  id-token: write        # Required for Sigstore signing
  attestations: write    # Required to persist attestations
```

### Job: release

Runs on `ubuntu-latest` with all release steps.

#### Step 1: Checkout Repository

```yaml
- name: Checkout repository
  uses: actions/checkout@v6
  with:
    fetch-depth: 0  # Full history for version extraction
```

#### Step 2: Setup Commands Binary

```yaml
- name: Setup Commands Binary
  id: commands
  uses: ./.github/actions/setup-commands
```

#### Step 3: Check CI Status

Verifies that CI has run successfully on the commit being released.

```yaml
- name: Check CI status
  run: |
    commands release check-ci \
      --workflow ci-r2r-cli.yaml \
      --commit ⟪ github.sha ⟫ \
      --timeout 300
  env:
    GH_TOKEN: ⟪ github.token ⟫
```

**Behavior:**

- Queries GitHub API for successful `ci-r2r-cli.yaml` run on the commit
- Waits up to 300 seconds for CI to complete if still running
- Fails if no successful CI run found

**Purpose:** Ensures code is built and tested before release

#### Step 4: Extract and Validate Version

```yaml
- name: Extract and validate version
  id: extract_version
  uses: ./.github/actions/extract-release-version
  with:
    module-prefix: r2r-cli
    legacy-prefixes: src-cli
    commands-path: ⟪ steps.commands.outputs.commands-path ⟫
```

**Outputs:**

- `version` - Extracted version (e.g., `1.0.0`)
- `tag_name` - Full tag name (e.g., `r2r-cli/1.0.0`)
- `is_valid` - Boolean indicating valid semver format

**Validation:**

```yaml
- name: Verify version is valid
  if: steps.extract_version.outputs.is_valid != 'true'
  run: |
    echo "Error: Invalid version format: ⟪ steps.extract_version.outputs.version ⟫"
    echo "Expected format: x.y.z (e.g., 1.0.0)"
    exit 1
```

#### Step 5: Check for Existing Release

```yaml
- name: Check for existing release
  run: |
    TAG_NAME="⟪ steps.extract_version.outputs.tag_name ⟫"

    if gh release view "$TAG_NAME" 2>/dev/null; then
      echo "Error: Release already exists for tag $TAG_NAME"
      exit 1
    fi
```

**Purpose:** Prevents duplicate releases and tag conflicts

#### Step 6: Install Build Dependencies

```yaml
- name: Install UPX
  run: |
    sudo apt-get update && sudo apt-get install -y upx-ucl
    upx --version
```

**UPX:** Universal Packer for eXecutables - compresses binaries to reduce size

#### Step 7: Build Binaries

```yaml
- name: Build binaries for multiple platforms
  env:
    VERSION: ⟪ steps.extract_version.outputs.version ⟫
  run: |
    commands build r2r-cli --all --version "$VERSION" --no-tidy
```

**Build Targets:**

- Linux AMD64 (standard and UPX-compressed)
- Linux ARM64
- macOS Intel (AMD64)
- macOS Apple Silicon (ARM64)
- Windows AMD64 (standard and UPX-compressed)

**Total binaries:** 7 files

**Build Flags:**

- `--all` - Build all platforms and UPX variants
- `--version` - Embed version in binary
- `--no-tidy` - Skip `go mod tidy` (already validated in CI)

#### Step 8: Verify Binaries

```yaml
- name: Verify binaries
  run: |
    ls -lh out/build/r2r-cli/

    REQUIRED_FILES=(
      "out/build/r2r-cli/r2r-linux-amd64"
      "out/build/r2r-cli/r2r-linux-amd64-upx"
      "out/build/r2r-cli/r2r-linux-arm64"
      "out/build/r2r-cli/r2r-darwin-amd64"
      "out/build/r2r-cli/r2r-darwin-arm64"
      "out/build/r2r-cli/r2r-windows-amd64.exe"
      "out/build/r2r-cli/r2r-windows-amd64-upx.exe"
    )

    for file in "${REQUIRED_FILES[@]}"; do
      if [ ! -f "$file" ]; then
        echo "Error: Required file missing: $file"
        exit 1
      fi
    done

    # Test binary execution
    chmod +x out/build/r2r-cli/r2r-linux-amd64
    ./out/build/r2r-cli/r2r-linux-amd64 version
```

#### Step 9: Generate Build Attestations

```yaml
- name: Generate build attestations
  uses: actions/attest-build-provenance@v3
  with:
    subject-path: |
      out/build/r2r-cli/r2r-linux-amd64
      out/build/r2r-cli/r2r-linux-amd64-upx
      out/build/r2r-cli/r2r-linux-arm64
      out/build/r2r-cli/r2r-darwin-amd64
      out/build/r2r-cli/r2r-darwin-arm64
      out/build/r2r-cli/r2r-windows-amd64.exe
      out/build/r2r-cli/r2r-windows-amd64-upx.exe
```

**Purpose:** Supply chain security via Sigstore attestations

**Verification:**

```bash
gh attestation verify r2r-linux-amd64 --repo <owner>/<repo>
```

#### Step 10: Create Release Notes

```yaml
- name: Create release notes
  id: release_notes
  run: |
    cat << EOF > release_notes.md
    # r2r CLI ⟪ steps.extract_version.outputs.version ⟫

    ## What's New

    Release of r2r CLI version ⟪ steps.extract_version.outputs.version ⟫.

    ## Installation

    [Binary download table with sizes...]

    ## Supply Chain Security

    All binaries include build attestations...
    EOF
```

**Content:**

- Version and release information
- Installation instructions
- Binary download table with sizes
- Supply chain security verification instructions
- Platform-specific notes (UPX compression)

#### Step 11: Create GitHub Release

```yaml
- name: Create GitHub Release with binaries
  run: |
    TAG_NAME="⟪ steps.extract_version.outputs.tag_name ⟫"
    VERSION="⟪ steps.extract_version.outputs.version ⟫"

    if [ "⟪ github.event_name ⟫" = "workflow_dispatch" ]; then
      # Manual trigger: create tag and release together
      gh release create "$TAG_NAME" \
        --title "r2r CLI v$VERSION" \
        --notes-file release_notes.md \
        --target "⟪ github.sha ⟫" \
        out/build/r2r-cli/*
    else
      # Tag trigger: use existing tag
      gh release create "$TAG_NAME" \
        --title "r2r CLI v$VERSION" \
        --notes-file release_notes.md \
        --verify-tag \
        out/build/r2r-cli/*
    fi
  env:
    GH_TOKEN: ⟪ github.token ⟫
```

**Behavior:**

- **Manual trigger:** Creates tag atomically with release using `--target`
- **Tag trigger:** Uses existing tag with `--verify-tag`
- Uploads all 7 binaries as release assets

#### Step 12: Verify Release Assets

```yaml
- name: Verify release assets
  run: |
    gh release view ⟪ steps.extract_version.outputs.tag_name ⟫ \
      --json assets --jq '.assets[] | .name'
  env:
    GH_TOKEN: ⟪ github.token ⟫
```

#### Step 13: Generate Summaries

**Failure Summary:**

```yaml
- name: Summary (failure)
  if: failure()
  run: |
    echo "## Release Failed" >> $GITHUB_STEP_SUMMARY
    # Diagnostic table with common failure causes
```

**Success Summary:**

```yaml
- name: Summary (success)
  if: success()
  run: |
    echo "## r2r CLI Released" >> $GITHUB_STEP_SUMMARY
    # Release details and download link
```

## Supply Chain Security

All release workflows generate build attestations using Sigstore.

### Attestation Generation

```yaml
- name: Generate build attestations
  uses: actions/attest-build-provenance@v3
  with:
    subject-path: <artifact-paths>
```

### Attestation Verification

Users can verify artifact authenticity:

```bash
# Download binary and attestation
gh release download r2r-cli/1.0.0 -p 'r2r-linux-amd64'

# Verify attestation
gh attestation verify r2r-linux-amd64 --repo <owner>/<repo>
```

**Output:**

- Verified build provenance
- Workflow information
- Commit SHA
- Build environment details

## Release Naming Conventions

### GitHub Releases

**Title Format:** `{Module Name} v{version}`

**Examples:**

- `r2r CLI v1.0.0`
- `EAC Extension v0.1.0`
- `Documentation Site (2025.12.01)`

### Git Tags

**Format:** `{moniker}/{version}`

**Examples:**

- `r2r-cli/1.0.0`
- `ext-eac/0.1.0`
- `docs/2025.12.01`
- `books/2025.12.01`

### Release Assets

**Naming Pattern:** Varies by module type

**Binary releases:**

- `{name}-{os}-{arch}[.exe]`
- `{name}-{os}-{arch}-upx[.exe]`

**Examples:**

- `r2r-linux-amd64`
- `r2r-windows-amd64.exe`
- `r2r-linux-amd64-upx`

## Manual Release Workflow

### Triggering a Release

```bash
# Create and push tag (triggers release automatically)
git tag r2r-cli/1.0.0
git push origin r2r-cli/1.0.0

# Or trigger manually via workflow dispatch
gh workflow run release-r2r-cli.yaml -f version=1.0.0
```

### Release Checklist

1. **Update CHANGELOG.md** - Document changes for the version
2. **Run CI** - Ensure CI passes on the commit to be released
3. **Create tag** - Tag the commit with version
4. **Push tag** - Push tag to trigger release workflow
5. **Monitor workflow** - Watch release workflow execution
6. **Verify release** - Check GitHub releases page
7. **Test artifacts** - Download and test release artifacts
8. **Verify attestations** - Verify supply chain attestations

## Debugging Release Workflows

### View Release Workflow Runs

```bash
# List recent release runs
gh run list --workflow release-r2r-cli.yaml --limit 10

# View specific run
gh run view <run-id>

# View logs
gh run view <run-id> --log
```

### Check CI Status

```bash
# Check if CI passed for a commit
r2r eac release check-ci \
  --workflow ci-r2r-cli.yaml \
  --commit <sha>
```

### Validate Version Format

```bash
# Validate semver format
r2r eac validate release-version 1.0.0
```

### Test Release Locally

```bash
# Build release binaries locally
r2r eac build r2r-cli --all --version 1.0.0

# Verify binaries
ls -lh out/build/r2r-cli/
./out/build/r2r-cli/r2r-linux-amd64 version
```

## References

- [CI Workflows](./ci-workflows.md) - Module CI workflows that run before release
- [Versioning](../changelog/versioning.md) - Semantic versioning rules
- [Repository Layout](../../architecture/repository-layout.md) - Module structure
- Release workflow files: `.github/workflows/release-*.yaml`
