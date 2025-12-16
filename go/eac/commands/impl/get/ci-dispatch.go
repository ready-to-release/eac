// Command: get ci-dispatch
// Short: Filter modules for CI dispatch based on existing valid CI
// Flags:
//   --directly-changed <modules>: Space-separated list of directly changed modules (always dispatched)
//   --invalidated <modules>: Space-separated list of invalidated modules (checked for valid CI)
//   --head-sha <sha>: Current HEAD SHA to check against (defaults to git HEAD)
//   --mock <json>: Use mock CI status instead of querying GitHub
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
// Long:
// Long: Filters modules for CI dispatch by checking if invalidated modules already have
// Long: valid CI at the current HEAD. Directly changed modules are always dispatched.
// Long:
// Long: Mock mode (--mock) accepts JSON mapping module names to CI validity:
// Long:   {"eac-commands": true, "docs": false}
// Long: where true = has valid CI (skip), false = needs CI (dispatch)
package get

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(GetCIDispatch)
}

// CIDispatchResult represents the output of the get ci-dispatch command
type CIDispatchResult struct {
	// Modules to dispatch CI for
	Dispatch []string `json:"dispatch" yaml:"dispatch" toml:"dispatch"`
	// Modules skipped (valid CI exists)
	Skipped []string `json:"skipped" yaml:"skipped" toml:"skipped"`
	// Per-module reasoning
	Reasons map[string]string `json:"reasons" yaml:"reasons" toml:"reasons"`
	// The HEAD SHA used for checking
	HeadSHA string `json:"head_sha" yaml:"head_sha" toml:"head_sha"`
}

func GetCIDispatch() int {
	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Parse flags
	directlyChanged := ""
	invalidated := ""
	headSHA := ""
	mockJSON := ""

	for i, arg := range os.Args {
		switch arg {
		case "--directly-changed":
			if i+1 < len(os.Args) {
				directlyChanged = os.Args[i+1]
			}
		case "--invalidated":
			if i+1 < len(os.Args) {
				invalidated = os.Args[i+1]
			}
		case "--head-sha":
			if i+1 < len(os.Args) {
				headSHA = os.Args[i+1]
			}
		case "--mock":
			if i+1 < len(os.Args) {
				mockJSON = os.Args[i+1]
			}
		case "--help", "-h":
			printCIDispatchUsage()
			return 0
		}
	}

	// Get HEAD SHA if not provided
	if headSHA == "" {
		sha, err := getCurrentSHA(workspaceRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to get current SHA: %v\n", err)
			return 1
		}
		headSHA = sha
	}

	// Parse mock data if provided
	var mockStatus map[string]bool
	if mockJSON != "" {
		if err := json.Unmarshal([]byte(mockJSON), &mockStatus); err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid mock JSON: %v\n", err)
			return 1
		}
	}

	return internal.ExecuteGetCommand(func() (interface{}, error) {
		return filterCIDispatch(directlyChanged, invalidated, headSHA, mockStatus, workspaceRoot)
	})
}

func filterCIDispatch(directlyChangedStr, invalidatedStr, headSHA string, mockStatus map[string]bool, workspaceRoot string) (*CIDispatchResult, error) {
	result := &CIDispatchResult{
		Dispatch: []string{},
		Skipped:  []string{},
		Reasons:  make(map[string]string),
		HeadSHA:  headSHA,
	}

	// Parse directly changed modules (always dispatch)
	directlyChanged := parseModuleList(directlyChangedStr)
	for _, module := range directlyChanged {
		result.Dispatch = append(result.Dispatch, module)
		result.Reasons[module] = "directly_changed"
	}

	// Parse invalidated modules (check for valid CI)
	invalidatedModules := parseModuleList(invalidatedStr)

	for _, module := range invalidatedModules {
		hasValidCI, reason, err := checkModuleCIValidity(module, headSHA, mockStatus, workspaceRoot)
		if err != nil {
			// On error, dispatch to be safe
			result.Dispatch = append(result.Dispatch, module)
			result.Reasons[module] = fmt.Sprintf("error_checking: %v", err)
			continue
		}

		if hasValidCI {
			result.Skipped = append(result.Skipped, module)
			result.Reasons[module] = reason
		} else {
			result.Dispatch = append(result.Dispatch, module)
			result.Reasons[module] = reason
		}
	}

	return result, nil
}

func parseModuleList(input string) []string {
	if input == "" {
		return []string{}
	}

	parts := strings.Fields(input)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// checkModuleCIValidity checks if a module has valid CI at the given HEAD SHA
// Returns (hasValidCI, reason, error)
func checkModuleCIValidity(module, headSHA string, mockStatus map[string]bool, workspaceRoot string) (bool, string, error) {
	// Use mock status if provided
	if mockStatus != nil {
		if valid, exists := mockStatus[module]; exists {
			if valid {
				return true, "mock:valid_ci_at_head", nil
			}
			return false, "mock:no_valid_ci", nil
		}
		// Module not in mock data - treat as no CI
		return false, "mock:not_specified", nil
	}

	// Query GitHub for last successful CI run
	workflowName := fmt.Sprintf("ci-%s.yaml", module)
	lastSuccessSHA, err := getLastSuccessfulWorkflowSHA(workflowName, workspaceRoot)
	if err != nil {
		return false, fmt.Sprintf("query_failed: %v", err), nil // Non-fatal, dispatch to be safe
	}

	if lastSuccessSHA == "" {
		return false, "no_ci_run", nil
	}

	if lastSuccessSHA == headSHA {
		return true, "valid_ci_at_head", nil
	}

	return false, fmt.Sprintf("ci_at_different_sha:%s", lastSuccessSHA[:min(7, len(lastSuccessSHA))]), nil
}

// getLastSuccessfulWorkflowSHA queries GitHub for the last successful run of a workflow
func getLastSuccessfulWorkflowSHA(workflowName, workspaceRoot string) (string, error) {
	cmd := exec.Command("gh", "run", "list",
		"-w", workflowName,
		"-s", "success",
		"-L", "1",
		"--json", "headSha",
		"-q", ".[0].headSha",
	)
	cmd.Dir = workspaceRoot

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("gh command failed: %w", err)
	}

	sha := strings.TrimSpace(string(output))
	return sha, nil
}

func printCIDispatchUsage() {
	fmt.Println("Filter modules for CI dispatch based on existing valid CI")
	fmt.Println("")
	fmt.Println("Usage: r2r get ci-dispatch [flags]")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  --directly-changed <modules>  Space-separated list of directly changed modules")
	fmt.Println("  --invalidated <modules>       Space-separated list of invalidated modules")
	fmt.Println("  --head-sha <sha>              HEAD SHA to check against (default: git HEAD)")
	fmt.Println("  --mock <json>                 Mock CI status (for testing)")
	fmt.Println("  --as-yaml                     Output as YAML (default)")
	fmt.Println("  --as-json                     Output as JSON")
	fmt.Println("")
	fmt.Println("Logic:")
	fmt.Println("  - Directly changed modules are ALWAYS dispatched")
	fmt.Println("  - Invalidated modules are checked for existing valid CI:")
	fmt.Println("    - If successful CI exists at HEAD SHA: SKIP")
	fmt.Println("    - Otherwise: DISPATCH")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  # Normal usage in CI")
	fmt.Println("  r2r get ci-dispatch --directly-changed \"eac-core\" --invalidated \"eac-commands docs\"")
	fmt.Println("")
	fmt.Println("  # Local testing with mock data")
	fmt.Println("  r2r get ci-dispatch --directly-changed \"eac-core\" --invalidated \"eac-commands docs\" \\")
	fmt.Println("    --mock '{\"eac-commands\": true, \"docs\": false}'")
	fmt.Println("")
	fmt.Println("  # Output as JSON")
	fmt.Println("  r2r get ci-dispatch --as-json --directly-changed \"eac-core\" --invalidated \"docs\"")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
