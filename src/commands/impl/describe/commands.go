// Command: describe commands
// Description: Output structured command information for shell integration
// HasSideEffects: false
package describe

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/contracts/reports"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register(DescribeCommands)
}

// CommandInfo represents structured information about a command
type CommandInfo struct {
	Name        string   `json:"name"`        // Full command name: "show modules"
	Parts       []string `json:"parts"`       // Command parts: ["show", "modules"]
	Description string   `json:"description"` // Command description
	Parent      string   `json:"parent"`      // Parent command: "show" (empty for root)
	IsLeaf      bool     `json:"is_leaf"`     // True if this is an executable command
	Args        string   `json:"args,omitempty"` // Argument completion type: "modules", "files", etc.
}

// CommandTree represents the hierarchical structure
type CommandTree struct {
	Commands []CommandInfo       `json:"commands"` // All commands
	Tree     map[string][]string `json:"tree"`     // Parent -> children mapping
	Modules  []string            `json:"modules"`  // Available module monikers for completion
}

func DescribeCommands() int {
	tree := buildCommandTree()

	// Output as JSON
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(tree); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		return 1
	}

	return 0
}

func buildCommandTree() CommandTree {
	var infos []CommandInfo
	treeMap := make(map[string][]string)

	// Process all registered commands
	commandRegistry := registry.GetCommandRegistry()
	for _, reg := range commandRegistry {
		cmdName := reg.ActualCommand
		parts := strings.Fields(cmdName)

		info := CommandInfo{
			Name:        cmdName,
			Parts:       parts,
			Description: reg.Short,
			IsLeaf:      true,
			Args:        reg.Args,
		}

		// Determine parent
		if len(parts) > 1 {
			info.Parent = strings.Join(parts[:len(parts)-1], " ")

			// Add to tree mapping
			if _, exists := treeMap[info.Parent]; !exists {
				treeMap[info.Parent] = []string{}
			}
			treeMap[info.Parent] = append(treeMap[info.Parent], parts[len(parts)-1])
		} else {
			info.Parent = ""
			// Root command
			if _, exists := treeMap[""]; !exists {
				treeMap[""] = []string{}
			}
			treeMap[""] = append(treeMap[""], cmdName)
		}

		infos = append(infos, info)
	}

	// Load module monikers for completion
	modules := loadModuleMonikers()

	return CommandTree{
		Commands: infos,
		Tree:     treeMap,
		Modules:  modules,
	}
}

// loadModuleMonikers loads all module monikers from the contracts
func loadModuleMonikers() []string {
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return []string{}
	}

	moduleReport, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		return []string{}
	}

	var monikers []string
	for _, module := range moduleReport.Registry.All() {
		monikers = append(monikers, module.Moniker)
	}

	// Sort for consistent output
	sort.Strings(monikers)

	return monikers
}
