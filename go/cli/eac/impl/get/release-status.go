// Command: get release-status
// Short: Get release status for modules
// Long: Checks GitHub releases for a list of modules and returns their status.
// Long:
// Long: This command queries GitHub releases API to determine which modules
// Long: have releases and their current versions. Used by release-clie-eac-bundle.
// Long:
// Long: SHA Detection:
// Long:   1. --sha flag (explicit)
// Long:   2. GITHUB_SHA env var (CI)
// Long:   3. origin/main (devbox)
// Long:
// Long: Output formats:
// Long:   --format json: {"module": {"tag": "...", "version": "...", "released": true}}
// Long:   --format shell: RELEASED="mod1 mod2" MISSING="mod3"
// Long:
// Long: Example:
// Long:   get release-status --modules "clie eac-ext docs"
// Long:   get release-status --modules "clie eac-ext" --format shell
// Flag.modules: type=string, usage=Space-separated list of modules to check
// Flag.format: type=string, usage=Output format (json, shell)
package get

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/ghexec"
	"github.com/ready-to-release/eac/go/core/repository"
)

// ModuleReleaseStatus represents the release status of a module.
type ModuleReleaseStatus struct {
	Tag      string `json:"tag"`
	Version  string `json:"version"`
	Released bool   `json:"released"`
}

// ReleaseStatusOutput is the output structure.
type ReleaseStatusOutput struct {
	Modules  map[string]ModuleReleaseStatus `json:"modules"`
	Released []string                       `json:"released"`
	Missing  []string                       `json:"missing"`
}

func GetReleaseStatus() int {
	// Parse flags
	modules := ""
	format := "json"

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--modules" && i+1 < len(os.Args):
			modules = os.Args[i+1]
			i++
		case arg == "--format" && i+1 < len(os.Args):
			format = os.Args[i+1]
			i++
		}
	}

	if modules == "" {
		fmt.Fprintln(os.Stderr, "Error: --modules is required")
		fmt.Fprintln(os.Stderr, "Usage: get release-status --modules \"mod1 mod2\" [--format json|shell]")
		return 1
	}

	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	moduleList := strings.Fields(modules)
	result := ReleaseStatusOutput{
		Modules:  make(map[string]ModuleReleaseStatus),
		Released: []string{},
		Missing:  []string{},
	}

	for _, mod := range moduleList {
		status := checkModuleRelease(mod, workspaceRoot)
		result.Modules[mod] = status
		if status.Released {
			result.Released = append(result.Released, mod)
		} else {
			result.Missing = append(result.Missing, mod)
		}
	}

	// Output based on format
	switch format {
	case "shell":
		fmt.Printf("RELEASED=\"%s\"\n", strings.Join(result.Released, " "))
		fmt.Printf("MISSING=\"%s\"\n", strings.Join(result.Missing, " "))
		fmt.Printf("ALL_RELEASED=\"%t\"\n", len(result.Missing) == 0)
	default:
		if output, err := json.MarshalIndent(result, "", "  "); err == nil {
			fmt.Println(string(output))
		}
	}

	return 0
}

// checkModuleRelease queries GitHub for the latest release of a module.
func checkModuleRelease(module, workspaceRoot string) ModuleReleaseStatus {
	// Query GitHub releases for this module's tag pattern
	// Tags are formatted as: module/version (e.g., clie/1.0.0)
	output, err := ghexec.Run(workspaceRoot, "release", "list",
		"--limit", "10",
		"--json", "tagName",
		"-q", fmt.Sprintf(".[] | select(.tagName | startswith(\"%s/\")) | .tagName", module),
	)
	if err != nil {
		return ModuleReleaseStatus{Released: false}
	}

	tags := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(tags) == 0 || tags[0] == "" {
		return ModuleReleaseStatus{Released: false}
	}

	// Get the latest tag (first in list)
	latestTag := tags[0]
	version := strings.TrimPrefix(latestTag, module+"/")

	return ModuleReleaseStatus{
		Tag:      latestTag,
		Version:  version,
		Released: true,
	}
}
