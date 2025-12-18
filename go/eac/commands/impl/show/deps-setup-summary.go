// Command: show deps-setup-summary
// Description: Generate dependencies setup summary
// Short: Generate dependencies setup summary
// Flag.module: type=string, usage=Module name (required)
// Flag.deps: type=string, usage=Comma-separated list of dependencies (go,node,docker,buildx,upx)
// Flag.go-available: type=bool, usage=Whether Go was already available
// Flag.node-available: type=bool, usage=Whether Node was already available
// Flag.buildx-available: type=bool, usage=Whether Docker Buildx was already available
// Flag.qemu-available: type=bool, usage=Whether QEMU was already available
// Flag.upx-available: type=bool, usage=Whether UPX was already available
// Long: The show deps-setup-summary command generates a formatted summary of dependency setup results.
// Long: This command is designed to be used in GitHub Actions workflows to create consistent setup summaries.
// Long: The output is formatted as Markdown and can be redirected to $GITHUB_STEP_SUMMARY.
// Long:
// Long: Expected Output:
// Long: - Markdown-formatted setup summary with dependencies table
// Long: - Shows which dependencies were installed or already available

package show

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(ShowDepsSetupSummary)
}

// ShowDepsSetupSummary generates a dependencies setup summary
func ShowDepsSetupSummary() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "show", and "deps-setup-summary"

	module := ""
	deps := ""
	goAvailable := false
	nodeAvailable := false
	buildxAvailable := false
	qemuAvailable := false
	upxAvailable := false

	// Parse flags
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case strings.HasPrefix(arg, "--module="):
			module = strings.TrimPrefix(arg, "--module=")
		case strings.HasPrefix(arg, "--deps="):
			deps = strings.TrimPrefix(arg, "--deps=")
		case arg == "--go-available" || arg == "--go-available=true":
			goAvailable = true
		case arg == "--node-available" || arg == "--node-available=true":
			nodeAvailable = true
		case arg == "--buildx-available" || arg == "--buildx-available=true":
			buildxAvailable = true
		case arg == "--qemu-available" || arg == "--qemu-available=true":
			qemuAvailable = true
		case arg == "--upx-available" || arg == "--upx-available=true":
			upxAvailable = true
		}
	}

	if module == "" {
		log.Errorf("Usage: show deps-setup-summary --module=<name> --deps=<list> [--go-available] [--node-available] ...")
		return 1
	}

	return generateDepsSetupSummary(module, deps, goAvailable, nodeAvailable, buildxAvailable, qemuAvailable, upxAvailable)
}

func generateDepsSetupSummary(module, deps string, goAvailable, nodeAvailable, buildxAvailable, qemuAvailable, upxAvailable bool) int {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("### Dependencies Setup for `%s`\n\n", module))

	// Build table
	tb := render.NewTableBuilder().
		WithHeaders("Dependency", "Status")

	// Check each dependency type
	if strings.Contains(deps, "go") {
		if goAvailable {
			tb.AddRow("Go", "Already available")
		} else {
			tb.AddRow("Go", "Installed")
		}
	}

	if strings.Contains(deps, "upx") {
		if upxAvailable {
			tb.AddRow("UPX", "Already available")
		} else {
			tb.AddRow("UPX", "Installed")
		}
	}

	if strings.Contains(deps, "node") || strings.Contains(deps, "npm") {
		if nodeAvailable {
			tb.AddRow("Node.js", "Already available")
		} else {
			tb.AddRow("Node.js", "Installed")
		}
		tb.AddRow("npm packages", "Installed")
	}

	if strings.Contains(deps, "buildx") {
		if buildxAvailable {
			tb.AddRow("Docker Buildx", "Already configured")
		} else {
			tb.AddRow("Docker Buildx", "Configured")
		}
		if qemuAvailable {
			tb.AddRow("QEMU", "Already available")
		} else {
			tb.AddRow("QEMU", "Configured")
		}
	} else if strings.Contains(deps, "docker") {
		if buildxAvailable {
			tb.AddRow("Docker Buildx", "Already configured")
		} else {
			tb.AddRow("Docker Buildx", "Configured")
		}
	}

	sb.WriteString(tb.Build())

	fmt.Print(sb.String())
	return 0
}
