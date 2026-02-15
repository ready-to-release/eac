package release

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"gopkg.in/yaml.v3"
)

type releaseExtractVersionCommand struct{}

var _ core.SimpleCommandPort = (*releaseExtractVersionCommand)(nil)

func (c *releaseExtractVersionCommand) Name() string { return "release extract-version" }

func (c *releaseExtractVersionCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "release-extract-version",
		Short:         "Extract and validate release version from tag or input",
		Long:          "The release extract-version command extracts version information from a git tag ref\nor explicit version input, validates the format (semver or calver), and outputs\nstructured data for use in release workflows.\n\nThis command replaces the extract-release-version GitHub Action with pure Go logic,\nmaking it testable and usable locally.\n\nExpected Output (--format shell):\n  VERSION=\"1.0.0\"\n  TAG_NAME=\"clie/1.0.0\"\n  IS_VALID=\"true\"\n\nExample:\n  release extract-version --module clie --ref refs/tags/clie/1.0.0\n  release extract-version --module docs --type calver --version \"\"\n  eval $(release extract-version --module clie --ref \"$GITHUB_REF\" --format shell)",
		Flags: []core.FlagSpec{
			{Name: "module", Type: "string", Usage: "Module prefix (e.g., clie, docs)"},
			{Name: "type", Type: "string", DefaultValue: "semver", Usage: "Version type (semver or calver)"},
			{Name: "ref", Type: "string", Usage: "Git ref (e.g., refs/tags/clie/1.0.0)"},
			{Name: "version", Type: "string", Usage: "Explicit version (for workflow_dispatch)"},
			{Name: "format", Type: "string", DefaultValue: "shell", Usage: "Output format (shell, json, yaml)"},
		},
	}
}

func (c *releaseExtractVersionCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ExtractVersion()
}

var (
	reCalver       = regexp.MustCompile(`^\d{4}\.\d{4}\.\d{4}$`)
	reSemverNoV    = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`)
)

// ExtractVersionOutput contains the extracted version information.
type ExtractVersionOutput struct {
	Version string `json:"version" yaml:"version"`
	TagName string `json:"tag_name" yaml:"tag_name"`
	IsValid bool   `json:"is_valid" yaml:"is_valid"`
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}

// ExtractVersion extracts and validates a release version.
func ExtractVersion() int {
	s, exitCode := newReleaseScaffold()
	if s == nil {
		return exitCode
	}

	args := os.Args[3:] // Skip program name, "release", and "extract-version"

	module := ""
	versionType := "semver"
	ref := ""
	version := ""
	format := "shell"

	// Parse flags
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case strings.HasPrefix(arg, "--module="):
			module = strings.TrimPrefix(arg, "--module=")
		case strings.HasPrefix(arg, "--type="):
			versionType = strings.TrimPrefix(arg, "--type=")
		case strings.HasPrefix(arg, "--ref="):
			ref = strings.TrimPrefix(arg, "--ref=")
		case strings.HasPrefix(arg, "--version="):
			version = strings.TrimPrefix(arg, "--version=")
		case strings.HasPrefix(arg, "--format="):
			format = strings.TrimPrefix(arg, "--format=")
		}
	}

	if module == "" {
		fmt.Fprintln(os.Stderr, "Error: --module is required")
		return 1
	}

	output := extractVersionLogic(module, versionType, ref, version)

	// Output in requested format
	switch format {
	case "shell":
		fmt.Printf("VERSION=%q\n", output.Version)
		fmt.Printf("TAG_NAME=%q\n", output.TagName)
		fmt.Printf("IS_VALID=%q\n", fmt.Sprintf("%t", output.IsValid))
		if output.Message != "" {
			fmt.Printf("MESSAGE=%q\n", output.Message)
		}
	case "json":
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to marshal JSON: %v\n", err)
			return 1
		}
		fmt.Println(string(data))
	case "yaml":
		data, err := yaml.Marshal(output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to marshal YAML: %v\n", err)
			return 1
		}
		fmt.Print(string(data))
	default:
		// Simple output for direct use
		fmt.Println(output.Version)
	}

	return 0
}

func extractVersionLogic(module, versionType, ref, explicitVersion string) ExtractVersionOutput {
	output := ExtractVersionOutput{}

	// Determine version based on inputs
	if ref != "" && strings.HasPrefix(ref, "refs/tags/") {
		// Extract from tag ref
		tagName := strings.TrimPrefix(ref, "refs/tags/")
		output.TagName = tagName
		output.Version = strings.TrimPrefix(tagName, module+"/")
	} else if explicitVersion != "" {
		// Use explicit version
		output.Version = explicitVersion
		output.TagName = module + "/" + explicitVersion
	} else if versionType == "calver" {
		// Auto-generate calver
		output.Version = generateCalver()
		output.TagName = module + "/" + output.Version
		output.Message = "Auto-generated calver"
	} else {
		// No version available
		output.IsValid = false
		output.Message = "No version provided (--ref or --version required)"
		return output
	}

	// Validate version format
	if versionType == "calver" {
		output.IsValid = isValidCalver(output.Version)
		if !output.IsValid {
			output.Message = fmt.Sprintf("Invalid calver format: %s (expected YYYY.MMDD.HHMM)", output.Version)
		}
	} else {
		output.IsValid = isValidSemver(output.Version)
		if !output.IsValid {
			output.Message = fmt.Sprintf("Invalid semver format: %s (expected x.y.z)", output.Version)
		}
	}

	return output
}

// generateCalver generates a calver timestamp in YYYY.MMDD.HHMM format.
func generateCalver() string {
	now := time.Now().UTC()
	return fmt.Sprintf("%d.%02d%02d.%02d%02d",
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute())
}

// isValidCalver validates a calver string (YYYY.MMDD.HHMM).
func isValidCalver(version string) bool {
	return reCalver.MatchString(version)
}

// isValidSemver validates a semver string (x.y.z without v prefix).
func isValidSemver(version string) bool {
	// No v prefix allowed
	if strings.HasPrefix(version, "v") || strings.HasPrefix(version, "V") {
		return false
	}
	return reSemverNoV.MatchString(version)
}
