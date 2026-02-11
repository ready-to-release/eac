package get

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/cli/eac/impl/get/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/config"
)

type getReleaseConfigCommand struct{}

var _ core.SimpleCommandPort = (*getReleaseConfigCommand)(nil)

func (c *getReleaseConfigCommand) Name() string { return "get release-config" }

func (c *getReleaseConfigCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-release-config",
		Short:         "Derive release configuration for a module from core config",
		Long: "Derives release configuration for any module from the core config system.\n" +
			"All values are computed from repository.yml + blueprints.yml.\n" +
			"\n" +
			"With --format shell, outputs shell variable assignments for eval:\n" +
			"  RELEASE_TYPE=cli-binary|container|docs-site|bundle|none\n" +
			"  VERSION_TYPE=semver|calver\n" +
			"  HAS_EVIDENCE=true|false\n" +
			"  AWAIT_MODULE_RELEASES=true|false\n" +
			"\n" +
			"Release type resolution:\n" +
			"  - bundle -> bundle\n" +
			"  - published + dockerfile with push -> container\n" +
			"  - published + docs-site component -> docs-site\n" +
			"  - published + go component with binary_name -> cli-binary\n" +
			"  - internal / none -> none",
		Flags: []core.FlagSpec{
			{Name: "module", Type: "string", Usage: "Module moniker to derive release config for", Required: true, Completion: []string{"modules"}},
			{Name: "format", Type: "string", Usage: "Output format: shell (eval-friendly) or github-output (KEY=value for $GITHUB_OUTPUT)"},
		},
	}
}

func (c *getReleaseConfigCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetReleaseConfig()
}

// ReleaseConfigResult represents the derived release configuration for a module.
type ReleaseConfigResult struct {
	Module              string `json:"module" yaml:"module"`
	ReleaseType         string `json:"release_type" yaml:"release_type"`
	VersionType         string `json:"version_type" yaml:"version_type"`
	HasEvidence         bool   `json:"has_evidence" yaml:"has_evidence"`
	AwaitModuleReleases bool   `json:"await_module_releases" yaml:"await_module_releases"`
}

func GetReleaseConfig() int {
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	module := ""
	format := ""

	for i, arg := range os.Args {
		switch arg {
		case "--module":
			if i+1 < len(os.Args) {
				module = os.Args[i+1]
			}
		case "--format":
			if i+1 < len(os.Args) {
				format = os.Args[i+1]
			}
		case "--help", "-h":
			printReleaseConfigUsage()
			return 0
		}
	}

	if module == "" {
		fmt.Fprintf(os.Stderr, "Error: --module is required\n")
		return 1
	}

	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load configuration: %v\n", err)
		return 1
	}

	result, err := deriveReleaseConfig(cfg, module)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	switch format {
	case "shell":
		outputReleaseConfigShell(result)
		return 0
	case "github-output":
		outputReleaseConfigGitHub(result)
		return 0
	default:
		return internal.ExecuteGetCommand(func() (interface{}, error) {
			return result, nil
		})
	}
}

func deriveReleaseConfig(cfg *config.EACConfig, moniker string) (*ReleaseConfigResult, error) {
	if cfg.Repository == nil {
		return nil, fmt.Errorf("repository config not loaded")
	}

	mod := cfg.Repository.GetByMoniker(moniker)
	if mod == nil {
		return nil, fmt.Errorf("module %q not found", moniker)
	}

	result := &ReleaseConfigResult{
		Module: moniker,
	}

	// VERSION_TYPE: from versioning.scheme
	if mod.Versioning != nil {
		switch strings.ToLower(mod.Versioning.Scheme) {
		case "semver":
			result.VersionType = "semver"
		case "calver":
			result.VersionType = "calver"
		default:
			result.VersionType = "none"
		}
	} else {
		result.VersionType = "none"
	}

	// HAS_EVIDENCE: evidence_books non-empty
	result.HasEvidence = len(mod.EvidenceBooks) > 0

	// RELEASE_TYPE: derived from versioning.release_type + components
	result.ReleaseType = resolveReleaseType(mod)

	// AWAIT_MODULE_RELEASES: true for bundle release type
	result.AwaitModuleReleases = result.ReleaseType == "bundle"

	return result, nil
}

// resolveReleaseType determines the release type from versioning config and components.
func resolveReleaseType(mod *config.Module) string {
	if mod.Versioning == nil {
		return "none"
	}

	releaseType := mod.Versioning.ReleaseType

	switch releaseType {
	case "bundle":
		return "bundle"

	case "published":
		// Check for container (dockerfile with push)
		dockerCfg := mod.GetDockerBuildConfig()
		if dockerCfg != nil {
			shouldPush := true
			if dockerCfg.Push != nil {
				shouldPush = *dockerCfg.Push
			}
			if shouldPush {
				return "container"
			}
		}

		// Check for docs-site component
		if mod.Components.HasComponentType("docs-site") || mod.Components.HasComponentType("site-render") {
			return "docs-site"
		}

		// Check for CLI binary (go component with binary_name)
		for _, entry := range mod.Components {
			if entry == nil || entry.Build == nil {
				continue
			}
			if entry.Build.BinaryName != "" {
				return "cli-binary"
			}
		}

		// Published but no recognized artifact type - default to cli-binary
		return "cli-binary"

	case "internal", "none":
		return "none"

	default:
		return "none"
	}
}

func outputReleaseConfigShell(r *ReleaseConfigResult) {
	fmt.Printf("RELEASE_TYPE=\"%s\"\n", r.ReleaseType)
	fmt.Printf("VERSION_TYPE=\"%s\"\n", r.VersionType)
	fmt.Printf("HAS_EVIDENCE=%s\n", boolToStr(r.HasEvidence))
	fmt.Printf("AWAIT_MODULE_RELEASES=%s\n", boolToStr(r.AwaitModuleReleases))
}

func outputReleaseConfigGitHub(r *ReleaseConfigResult) {
	fmt.Printf("release-type=%s\n", r.ReleaseType)
	fmt.Printf("version-type=%s\n", r.VersionType)
	fmt.Printf("has-evidence=%s\n", boolToStr(r.HasEvidence))
	fmt.Printf("await-module-releases=%s\n", boolToStr(r.AwaitModuleReleases))
}

func printReleaseConfigUsage() {
	fmt.Println("Derive release configuration for a module from core config")
	fmt.Println("")
	fmt.Println("Usage: eac get release-config --module <moniker> [--format shell|github-output]")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  --module <moniker>    Module moniker (required)")
	fmt.Println("  --format <format>     Output format: shell, github-output, yaml (default), json")
	fmt.Println("")
	fmt.Println("Output variables:")
	fmt.Println("  RELEASE_TYPE             cli-binary, container, docs-site, bundle, or none")
	fmt.Println("  VERSION_TYPE             semver or calver")
	fmt.Println("  HAS_EVIDENCE             evidence_books is non-empty")
	fmt.Println("  AWAIT_MODULE_RELEASES    true for bundle releases (awaits dependency releases)")
	fmt.Println("")
	fmt.Println("Release type resolution:")
	fmt.Println("  bundle                 → bundle")
	fmt.Println("  published + dockerfile → container")
	fmt.Println("  published + docs-site  → docs-site")
	fmt.Println("  published + binary     → cli-binary")
	fmt.Println("  internal / none        → none")
}
