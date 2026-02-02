// Command: show release-summary
// Short: Generate release summary from layers JSON
// Long: Parses release layers JSON and generates a markdown summary.
// Long:
// Long: This command replaces the bash/jq loop that iterates over LAYERS_JSON
// Long: to generate the release summary in GITHUB_STEP_SUMMARY.
// Long:
// Long: Input: --layers JSON array of release layers
// Long:
// Long: Output: Markdown formatted release summary
// Long:
// Long: Example:
// Long:   show release-summary --layers '[[{"module":"docs","version":"2025.0116","type":"calver"}]]'
// Long:   show release-summary --layers "$LAYERS_JSON" >> $GITHUB_STEP_SUMMARY
// Flag.layers: type=string, usage=JSON array of release layers (required)
package show

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/clibase/registry"
)

func init() {
	registry.Register(ShowReleaseSummary)
}

// ReleaseModule represents a module in a release layer.
type ReleaseModule struct {
	Module  string `json:"module"`
	Version string `json:"version"`
	Tag     string `json:"tag"`
	Type    string `json:"type"`
}

func ShowReleaseSummary() int {
	// Parse flags
	layersJSON := ""

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--layers" && i+1 < len(os.Args):
			layersJSON = os.Args[i+1]
			i++
		case strings.HasPrefix(arg, "--layers="):
			layersJSON = strings.TrimPrefix(arg, "--layers=")
		}
	}

	if layersJSON == "" {
		fmt.Fprintln(os.Stderr, "Error: --layers is required")
		fmt.Fprintln(os.Stderr, "Usage: show release-summary --layers <json>")
		return 1
	}

	// Parse layers
	var layers [][]ReleaseModule
	if err := json.Unmarshal([]byte(layersJSON), &layers); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse layers JSON: %v\n", err)
		return 1
	}

	if len(layers) == 0 {
		fmt.Println("No releases to summarize.")
		return 0
	}

	// Generate markdown summary
	var sb strings.Builder

	for layerIdx, layer := range layers {
		if len(layer) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("**Layer %d:**\n", layerIdx+1))

		// Check if single module layer
		if len(layer) == 1 {
			mod := layer[0]
			sb.WriteString(fmt.Sprintf("- `%s` %s (%s)\n", mod.Module, mod.Version, mod.Type))
		} else {
			// Multiple modules in layer - show as list
			for _, mod := range layer {
				sb.WriteString(fmt.Sprintf("- `%s` %s (%s)\n", mod.Module, mod.Version, mod.Type))
			}
		}
		sb.WriteString("\n")
	}

	// Add tag summary table using proper markdown renderer
	tb := render.NewTableBuilder().
		WithHeaders("Module", "Version", "Tag", "Type")
	for _, layer := range layers {
		for _, mod := range layer {
			tb.AddRow(mod.Module, mod.Version, fmt.Sprintf("`%s`", mod.Tag), mod.Type)
		}
	}
	sb.WriteString(tb.Build())

	fmt.Print(sb.String())
	return 0
}
