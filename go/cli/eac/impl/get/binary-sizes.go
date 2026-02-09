// Command: get binary-sizes
// Short: Get binary file sizes for a module
// Flag.module: type=string, usage=Module to get binary sizes for (required)
// Flag.binary-prefix: type=string, usage=Binary name prefix (e.g., clie for clie-linux-amd64)
// Flag.format: type=string, default=shell, usage=Output format (shell, json, yaml, markdown)
// Long: The get binary-sizes command calculates file sizes for binaries produced by a module build.
// Long:
// Long: It scans the build output directory for the specified module and reports sizes for all
// Long: binary files found. This is useful for release notes and tracking binary size trends.
// Long:
// Long: Expected Output (--format shell):
// Long:   SIZE_LINUX_AMD64="12.3"
// Long:   SIZE_LINUX_AMD64_UPX="4.5"
// Long:   SIZE_DARWIN_ARM64="13.1"
// Long:   ...
// Long:
// Long: Example:
// Long:   get binary-sizes --module clie --binary-prefix clie
// Long:   eval $(get binary-sizes --module clie --binary-prefix clie --format shell)
package get

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/core/repository"
	"gopkg.in/yaml.v3"
)

// BinarySize represents a single binary file's size.
type BinarySize struct {
	Name      string  `json:"name" yaml:"name"`
	Path      string  `json:"path" yaml:"path"`
	SizeBytes int64   `json:"size_bytes" yaml:"size_bytes"`
	SizeMB    float64 `json:"size_mb" yaml:"size_mb"`
	Exists    bool    `json:"exists" yaml:"exists"`
}

// BinarySizesOutput contains all binary sizes.
type BinarySizesOutput struct {
	Module   string                `json:"module" yaml:"module"`
	Binaries map[string]BinarySize `json:"binaries" yaml:"binaries"`
}

// Standard binary variants for CLI builds.
var standardBinaryVariants = []struct {
	Name   string
	Suffix string
	EnvVar string
}{
	{"linux-amd64", "-linux-amd64", "SIZE_LINUX_AMD64"},
	{"linux-amd64-upx", "-linux-amd64-upx", "SIZE_LINUX_AMD64_UPX"},
	{"linux-arm64", "-linux-arm64", "SIZE_LINUX_ARM64"},
	{"darwin-amd64", "-darwin-amd64", "SIZE_DARWIN_AMD64"},
	{"darwin-arm64", "-darwin-arm64", "SIZE_DARWIN_ARM64"},
	{"windows-amd64", "-windows-amd64.exe", "SIZE_WINDOWS_AMD64"},
	{"windows-amd64-upx", "-windows-amd64-upx.exe", "SIZE_WINDOWS_AMD64_UPX"},
}

func GetBinarySizes() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "get", and "binary-sizes"

	module := ""
	binaryPrefix := ""
	format := "shell"

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
		case strings.HasPrefix(arg, "--format="):
			format = strings.TrimPrefix(arg, "--format=")
		case arg == "--format" && i+1 < len(args):
			format = args[i+1]
			i++
		}
	}

	if module == "" {
		log.Errorf("Error: --module is required")
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

	// Calculate sizes
	output := BinarySizesOutput{
		Module:   module,
		Binaries: make(map[string]BinarySize),
	}

	for _, variant := range standardBinaryVariants {
		filename := binaryPrefix + variant.Suffix
		fullPath := filepath.Join(buildDir, filename)

		size := BinarySize{
			Name: variant.Name,
			Path: fullPath,
		}

		if info, err := os.Stat(fullPath); err == nil {
			size.Exists = true
			size.SizeBytes = info.Size()
			size.SizeMB = float64(info.Size()) / 1048576.0
		} else {
			size.Exists = false
			size.SizeBytes = 0
			size.SizeMB = 0
		}

		output.Binaries[variant.Name] = size
	}

	// Output in requested format
	switch format {
	case "shell":
		for _, variant := range standardBinaryVariants {
			size := output.Binaries[variant.Name]
			if size.Exists {
				fmt.Printf("%s=%q\n", variant.EnvVar, fmt.Sprintf("%.1f", size.SizeMB))
			} else {
				fmt.Printf("%s=%q\n", variant.EnvVar, "?")
			}
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
	case "markdown":
		fmt.Println(formatBinarySizesMarkdown(output, binaryPrefix))
	default:
		// Simple list
		for _, variant := range standardBinaryVariants {
			size := output.Binaries[variant.Name]
			if size.Exists {
				fmt.Printf("%s: %.1f MB\n", variant.Name, size.SizeMB)
			}
		}
	}

	return 0
}

func formatBinarySizesMarkdown(output BinarySizesOutput, binaryPrefix string) string {
	var sb strings.Builder

	sb.WriteString("### Standard Binaries (Recommended)\n\n")

	tb := render.NewTableBuilder().
		WithHeaders("Platform", "Binary", "Size")

	// Standard binaries first
	standardVariants := []string{"linux-amd64", "linux-arm64", "darwin-amd64", "darwin-arm64", "windows-amd64"}
	for _, name := range standardVariants {
		if size, ok := output.Binaries[name]; ok && size.Exists {
			platform := formatPlatformName(name)
			filename := binaryPrefix + getBinarySuffix(name)
			tb.AddRow(platform, fmt.Sprintf("`%s`", filename), fmt.Sprintf("%.1f MB", size.SizeMB))
		}
	}
	sb.WriteString(tb.Build())

	// UPX binaries
	sb.WriteString("\n### UPX-Compressed Binaries (Minimal Size)\n\n")
	tbUpx := render.NewTableBuilder().
		WithHeaders("Platform", "Binary", "Size")

	upxVariants := []string{"linux-amd64-upx", "windows-amd64-upx"}
	for _, name := range upxVariants {
		if size, ok := output.Binaries[name]; ok && size.Exists {
			platform := formatPlatformName(strings.TrimSuffix(name, "-upx"))
			filename := binaryPrefix + getBinarySuffix(name)
			tbUpx.AddRow(platform, fmt.Sprintf("`%s`", filename), fmt.Sprintf("%.1f MB", size.SizeMB))
		}
	}
	sb.WriteString(tbUpx.Build())

	return sb.String()
}

func formatPlatformName(variant string) string {
	switch variant {
	case "linux-amd64":
		return "Linux (AMD64)"
	case "linux-arm64":
		return "Linux (ARM64)"
	case "darwin-amd64":
		return "macOS (Intel)"
	case "darwin-arm64":
		return "macOS (Apple Silicon)"
	case "windows-amd64":
		return "Windows (AMD64)"
	default:
		return variant
	}
}

func getBinarySuffix(variant string) string {
	for _, v := range standardBinaryVariants {
		if v.Name == variant {
			return v.Suffix
		}
	}
	return ""
}
