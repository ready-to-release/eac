// Command: get ci-dispatch
// Short: Filter modules for CI dispatch based on existing valid CI
// Flag.directly-changed: type=string, usage=Space-separated list of directly changed modules (always dispatched)
// Flag.invalidated: type=string, usage=Space-separated list of invalidated modules (checked for valid CI)
// Flag.head-sha: type=string, usage=Current HEAD SHA to check against (defaults to git HEAD)
// Flag.mock: type=string, usage=Use mock CI status instead of querying GitHub (JSON format)
// Flag.format: type=string, usage=Output format (shell outputs shell variables; otherwise uses standard get command formats)
// Long:
// Long: Filters modules for CI dispatch by checking if invalidated modules already have
// Long: valid CI at the current HEAD. Directly changed modules are always dispatched.
// Long:
// Long: Mock mode (--mock) accepts JSON mapping module names to CI validity:
// Long:   {"eac-commands": true, "docs": false}
// Long: where true = has valid CI (skip), false = needs CI (dispatch)
// Long:
// Long: With --format shell, outputs shell variable assignments:
// Long:   DISPATCH="mod1 mod2 mod3"
// Long:   SKIPPED="mod4 mod5"
package get

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/github"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(GetCIDispatch)
}

// ciDispatchFlags defines valid flags for the get ci-dispatch command

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
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

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
	format := ""

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
		case "--format":
			if i+1 < len(os.Args) {
				format = os.Args[i+1]
			}
		case "--help", "-h":
			printCIDispatchUsage()
			return 0
		}
	}

	// Detect HEAD SHA using shared logic (explicit > GITHUB_SHA > origin/main)
	shaResult, err := DetectCurrentSHA(workspaceRoot, headSHA)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to detect SHA: %v\n", err)
		return 1
	}
	headSHA = shaResult.SHA

	// Parse mock data if provided
	var mockStatus map[string]bool
	if mockJSON != "" {
		if err := json.Unmarshal([]byte(mockJSON), &mockStatus); err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid mock JSON: %v\n", err)
			return 1
		}
	}

	// Build result
	result, err := filterCIDispatch(directlyChanged, invalidated, headSHA, mockStatus, workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Handle shell format output
	if format == "shell" {
		// Output reasoning to stderr for logging (doesn't interfere with eval)
		fmt.Fprintf(os.Stderr, "Per-module decisions:\n")
		for _, module := range result.Dispatch {
			if reason, ok := result.Reasons[module]; ok {
				fmt.Fprintf(os.Stderr, "  %s: %s (DISPATCH)\n", module, reason)
			}
		}
		for _, module := range result.Skipped {
			if reason, ok := result.Reasons[module]; ok {
				fmt.Fprintf(os.Stderr, "  %s: %s (SKIP)\n", module, reason)
			}
		}
		fmt.Fprintf(os.Stderr, "\nFinal dispatch list: %s\n", strings.Join(result.Dispatch, " "))
		fmt.Fprintf(os.Stderr, "Skipped modules: %s\n\n", strings.Join(result.Skipped, " "))

		// Output variables to stdout for eval
		fmt.Printf("DISPATCH=\"%s\"\n", strings.Join(result.Dispatch, " "))
		fmt.Printf("SKIPPED=\"%s\"\n", strings.Join(result.Skipped, " "))
		return 0
	}

	return internal.ExecuteGetCommand(func() (interface{}, error) {
		return result, nil
	})
}

func filterCIDispatch(directlyChangedStr, invalidatedStr, headSHA string, mockStatus map[string]bool, workspaceRoot string) (*CIDispatchResult, error) {
	result := &CIDispatchResult{
		Dispatch: []string{},
		Skipped:  []string{},
		Reasons:  make(map[string]string),
		HeadSHA:  headSHA,
	}

	// Get valid CI workflow modules for validation (skip when using mock for tests)
	var validModules map[string]bool
	if mockStatus == nil {
		var err error
		validModules, err = getValidCIModules(workspaceRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to get valid CI modules: %w", err)
		}
	}

	// Parse directly changed modules (always dispatch)
	directlyChanged := parseModuleList(directlyChangedStr)
	for _, module := range directlyChanged {
		// Validate module exists (skip validation in mock mode)
		if validModules != nil && !validModules[module] {
			return nil, fmt.Errorf("invalid module %q: no CI workflow ci-%s.yaml exists", module, module)
		}
		result.Dispatch = append(result.Dispatch, module)
		result.Reasons[module] = "directly_changed"
	}

	// Parse invalidated modules (check for valid CI)
	invalidatedModules := parseModuleList(invalidatedStr)
	for _, module := range invalidatedModules {
		// Validate module exists (skip validation in mock mode)
		if validModules != nil && !validModules[module] {
			return nil, fmt.Errorf("invalid module %q: no CI workflow ci-%s.yaml exists", module, module)
		}
	}

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
	return checkModuleCIValidityWithAPI(module, headSHA, mockStatus, workspaceRoot, nil)
}

// checkModuleCIValidityWithAPI checks if a module has valid CI at the given HEAD SHA
// Accepts an optional github.API for testing. If nil, creates a new GHClient.
// Returns (hasValidCI, reason, error)
func checkModuleCIValidityWithAPI(module, headSHA string, mockStatus map[string]bool, workspaceRoot string, api github.API) (bool, string, error) {
	// Use mock status if provided (for backward compatibility with --mock flag)
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

	// Use provided API or create a new client
	if api == nil {
		api = github.NewGHClient(workspaceRoot)
	}

	// Query GitHub for last successful CI run
	workflowName := fmt.Sprintf("ci-%s.yaml", module)
	lastSuccessSHA, err := getLastSuccessfulWorkflowSHAWithAPI(workflowName, api)
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

// getLastSuccessfulWorkflowSHAWithAPI queries GitHub for the last successful run of a workflow using the API interface
func getLastSuccessfulWorkflowSHAWithAPI(workflowName string, api github.API) (string, error) {
	runs, err := api.ListRuns(workflowName, github.ListRunsOpts{
		Status: "success",
		Limit:  1,
	})
	if err != nil {
		return "", fmt.Errorf("ListRuns failed: %w", err)
	}

	if len(runs) == 0 {
		return "", nil
	}

	return runs[0].HeadSHA, nil
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

// getValidCIModules returns a set of valid CI workflow module names
func getValidCIModules(workspaceRoot string) (map[string]bool, error) {
	workflowsDir := workspaceRoot + "/.github/workflows"
	pattern := workflowsDir + "/ci-*.yaml"

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob CI workflows: %w", err)
	}

	modules := make(map[string]bool, len(matches))
	for _, match := range matches {
		base := filepath.Base(match)
		// ci-foo.yaml -> foo
		module := strings.TrimSuffix(strings.TrimPrefix(base, "ci-"), ".yaml")
		modules[module] = true
	}

	return modules, nil
}
