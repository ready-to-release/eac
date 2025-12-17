// Command: get cli-release-notes
// Description: Generate release notes for CLI releases
// Short: Generate release notes for CLI releases
// Long: The get cli-release-notes command generates release notes markdown for CLI binary releases.
// Long:
// Long: It includes installation instructions, binary download links with sizes, and supply chain
// Long: security information about build attestations.
// Long:
// Long: Flag.module: type=string, default=r2r-cli, usage=Module name
// Long: Flag.binary-prefix: type=string, default=r2r, usage=Binary name prefix
// Long: Flag.version: type=string, usage=Version string (required)
// Long: Flag.tag: type=string, usage=Git tag name (required)
// Long: Flag.commit: type=string, usage=Git commit SHA (required)
// Long: Flag.repo: type=string, usage=GitHub repository (owner/repo)
// Long: Flag.run-id: type=string, usage=GitHub Actions run ID
// Long:
// Long: Example:
// Long:   get cli-release-notes --version 1.0.0 --tag r2r-cli/1.0.0 --commit abc123
package get

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(GetCLIReleaseNotes)
}

func GetCLIReleaseNotes() int {
	args := os.Args[3:] // Skip program name, "get", and "cli-release-notes"

	module := "r2r-cli"
	binaryPrefix := "r2r"
	version := ""
	tag := ""
	commit := ""
	repo := ""
	runID := ""

	// Parse flags
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case strings.HasPrefix(arg, "--module="):
			module = strings.TrimPrefix(arg, "--module=")
		case arg == "--module" && i+1 < len(args):
			module = args[i+1]
			i++
		case strings.HasPrefix(arg, "--binary-prefix="):
			binaryPrefix = strings.TrimPrefix(arg, "--binary-prefix=")
		case arg == "--binary-prefix" && i+1 < len(args):
			binaryPrefix = args[i+1]
			i++
		case strings.HasPrefix(arg, "--version="):
			version = strings.TrimPrefix(arg, "--version=")
		case arg == "--version" && i+1 < len(args):
			version = args[i+1]
			i++
		case strings.HasPrefix(arg, "--tag="):
			tag = strings.TrimPrefix(arg, "--tag=")
		case arg == "--tag" && i+1 < len(args):
			tag = args[i+1]
			i++
		case strings.HasPrefix(arg, "--commit="):
			commit = strings.TrimPrefix(arg, "--commit=")
		case arg == "--commit" && i+1 < len(args):
			commit = args[i+1]
			i++
		case strings.HasPrefix(arg, "--repo="):
			repo = strings.TrimPrefix(arg, "--repo=")
		case arg == "--repo" && i+1 < len(args):
			repo = args[i+1]
			i++
		case strings.HasPrefix(arg, "--run-id="):
			runID = strings.TrimPrefix(arg, "--run-id=")
		case arg == "--run-id" && i+1 < len(args):
			runID = args[i+1]
			i++
		}
	}

	if version == "" {
		log.Errorf("Error: --version is required")
		return 1
	}
	if tag == "" {
		log.Errorf("Error: --tag is required")
		return 1
	}
	if commit == "" {
		log.Errorf("Error: --commit is required")
		return 1
	}

	// Get workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to get repository root: %v", err)
		return 1
	}

	// Build output directory
	buildDir := filepath.Join(workspaceRoot, "out", "build", module)

	// Get binary sizes
	sizes := make(map[string]string)
	for _, variant := range standardBinaryVariants {
		filename := binaryPrefix + variant.Suffix
		fullPath := filepath.Join(buildDir, filename)

		if info, err := os.Stat(fullPath); err == nil {
			sizeMB := float64(info.Size()) / 1048576.0
			sizes[variant.Name] = fmt.Sprintf("%.1f", sizeMB)
		} else {
			sizes[variant.Name] = "?"
		}
	}

	// Generate release notes
	notes := generateCLIReleaseNotes(module, binaryPrefix, version, tag, commit, repo, runID, sizes)
	fmt.Print(notes)

	return 0
}

func generateCLIReleaseNotes(module, binaryPrefix, version, tag, commit, repo, runID string, sizes map[string]string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s v%s\n\n", module, version))

	sb.WriteString("## Installation\n\n")
	sb.WriteString("Download the appropriate binary for your platform:\n\n")

	// Standard binaries
	sb.WriteString("### Standard Binaries (Recommended)\n\n")
	sb.WriteString(fmt.Sprintf("- **Linux (AMD64)** - `%s-linux-amd64` (%s MB)\n", binaryPrefix, sizes["linux-amd64"]))
	sb.WriteString(fmt.Sprintf("- **Linux (ARM64)** - `%s-linux-arm64` (%s MB)\n", binaryPrefix, sizes["linux-arm64"]))
	sb.WriteString(fmt.Sprintf("- **macOS (Intel)** - `%s-darwin-amd64` (%s MB)\n", binaryPrefix, sizes["darwin-amd64"]))
	sb.WriteString(fmt.Sprintf("- **macOS (Apple Silicon)** - `%s-darwin-arm64` (%s MB)\n", binaryPrefix, sizes["darwin-arm64"]))
	sb.WriteString(fmt.Sprintf("- **Windows (AMD64)** - `%s-windows-amd64.exe` (%s MB)\n", binaryPrefix, sizes["windows-amd64"]))
	sb.WriteString("\n")

	// UPX binaries
	sb.WriteString("### UPX-Compressed Binaries (Minimal Size)\n\n")
	sb.WriteString(fmt.Sprintf("- **Linux (AMD64)** - `%s-linux-amd64-upx` (%s MB)\n", binaryPrefix, sizes["linux-amd64-upx"]))
	sb.WriteString(fmt.Sprintf("- **Windows (AMD64)** - `%s-windows-amd64-upx.exe` (%s MB)\n", binaryPrefix, sizes["windows-amd64-upx"]))
	sb.WriteString("\n")

	// Supply chain security
	sb.WriteString("## Supply Chain Security\n\n")
	sb.WriteString("All binaries include [build attestations](https://docs.github.com/en/actions/security-for-github-actions/using-artifact-attestations). Verify with:\n\n")
	sb.WriteString("```bash\n")
	if repo != "" {
		sb.WriteString(fmt.Sprintf("gh attestation verify <binary-file> --repo %s\n", repo))
	} else {
		sb.WriteString("gh attestation verify <binary-file> --repo <owner>/<repo>\n")
	}
	sb.WriteString("```\n\n")

	// Release information
	sb.WriteString("## Release Information\n\n")
	sb.WriteString(fmt.Sprintf("- **Version**: %s\n", version))
	sb.WriteString(fmt.Sprintf("- **Commit**: %s\n", commit))
	sb.WriteString(fmt.Sprintf("- **Tag**: %s\n", tag))
	sb.WriteString("\n")

	// Footer
	sb.WriteString("---\n")
	if repo != "" && runID != "" {
		sb.WriteString(fmt.Sprintf("Generated with [GitHub Actions](https://github.com/%s/actions/runs/%s)\n", repo, runID))
	} else {
		sb.WriteString("Generated with GitHub Actions\n")
	}

	return sb.String()
}
