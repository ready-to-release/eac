package show

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/render"
)

type showDepsSetupSummaryCommand struct{}

var _ core.SimpleCommandPort = (*showDepsSetupSummaryCommand)(nil)

func (c *showDepsSetupSummaryCommand) Name() string { return "show deps-setup-summary" }

func (c *showDepsSetupSummaryCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "show-deps-setup-summary",
		Short:         "Generate dependencies setup summary",
		Long:          "The show deps-setup-summary command generates a formatted summary of dependency setup results.\nThis command is designed to be used in GitHub Actions workflows to create consistent setup summaries.\nThe output is formatted as Markdown and can be redirected to $GITHUB_STEP_SUMMARY.\n\nExpected Output:\n- Markdown-formatted setup summary with dependencies table\n- Shows which dependencies were installed or already available",
		Flags: []core.FlagSpec{
			{Name: "module", Type: "string", Usage: "Module name (required)"},
			{Name: "deps", Type: "string", Usage: "Comma-separated list of dependencies (go,node,docker,buildx,upx)"},
			{Name: "go-available", Type: "bool", Usage: "Whether Go was already available"},
			{Name: "node-available", Type: "bool", Usage: "Whether Node was already available"},
			{Name: "buildx-available", Type: "bool", Usage: "Whether Docker Buildx was already available"},
			{Name: "qemu-available", Type: "bool", Usage: "Whether QEMU was already available"},
			{Name: "upx-available", Type: "bool", Usage: "Whether UPX was already available"},
		},
	}
}

func (c *showDepsSetupSummaryCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ShowDepsSetupSummary()
}

// ShowDepsSetupSummary generates a dependencies setup summary.
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
		tb.AddRow("npm dependencies", "Installed")
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
