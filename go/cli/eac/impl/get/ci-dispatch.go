package get

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/cli/eac/impl/get/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/ghexec"
	"github.com/ready-to-release/eac/go/core/cache"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/github"
	"github.com/ready-to-release/eac/go/core/repository"
)

type getCIDispatchCommand struct{}

var _ core.SimpleCommandPort = (*getCIDispatchCommand)(nil)

func (c *getCIDispatchCommand) Name() string { return "get ci-dispatch" }

func (c *getCIDispatchCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-ci-dispatch",
		Short:         "Filter modules for CI dispatch based on existing valid CI",
		Long: "Filters modules for CI dispatch by checking if invalidated modules already have\n" +
			"valid CI at the current HEAD. Directly changed modules are always dispatched.\n" +
			"\n" +
			"Mock mode (--mock) accepts JSON mapping module names to CI validity:\n" +
			"  {\"eac-cli\": true, \"docs\": false}\n" +
			"where true = has valid CI (skip), false = needs CI (dispatch)\n" +
			"\n" +
			"With --format shell, outputs shell variable assignments:\n" +
			"  DISPATCH=\"mod1 mod2 mod3\"\n" +
			"  SKIPPED=\"mod4 mod5\"",
		Flags: []core.FlagSpec{
			{Name: "directly-changed", Type: "string", Usage: "Space-separated list of directly changed modules (always dispatched)"},
			{Name: "invalidated", Type: "string", Usage: "Space-separated list of invalidated modules (checked for valid CI)"},
			{Name: "head-sha", Type: "string", Usage: "Current HEAD SHA to check against (defaults to git HEAD)"},
			{Name: "mock", Type: "string", Usage: "Use mock CI status instead of querying GitHub (JSON format)"},
			{Name: "format", Type: "string", Usage: "Output format (shell outputs shell variables; otherwise uses standard get command formats)"},
		},
	}
}

func (c *getCIDispatchCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetCIDispatch()
}

// CIDispatchResult represents the output of the get ci-dispatch command.
// Combines filtering (dispatch vs skip) with dependency-aware ordering.
type CIDispatchResult struct {
	// Modules to dispatch CI for, in dependency order (topologically sorted by CI deps).
	// Modules with no CI deps come first, followed by modules that depend on them.
	Dispatch []string `json:"dispatch" yaml:"dispatch" toml:"dispatch"`
	// Modules skipped (valid CI exists at HEAD)
	Skipped []string `json:"skipped" yaml:"skipped" toml:"skipped"`
	// Per-module reasoning for dispatch/skip decision
	Reasons map[string]string `json:"reasons" yaml:"reasons" toml:"reasons"`
	// CI dependency graph (module -> modules it must wait for).
	// Only includes deps within the dispatch set.
	CIDependencies map[string][]string `json:"ci_dependencies" yaml:"ci_dependencies" toml:"ci_dependencies"`
	// The HEAD SHA used for CI validity checking
	HeadSHA string `json:"head_sha" yaml:"head_sha" toml:"head_sha"`
	// Total modules considered (dispatch + skipped)
	TotalModules int `json:"total_modules" yaml:"total_modules" toml:"total_modules"`
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

	// Build CICacheChecker: mock mode or real GitHub
	var checker *cache.CICacheChecker
	if mockJSON != "" {
		var mockStatus map[string]bool
		if err := json.Unmarshal([]byte(mockJSON), &mockStatus); err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid mock JSON: %v\n", err)
			return 1
		}
		checker = mockCheckerFromStatus(mockStatus, headSHA)
	} else {
		api := github.NewGHClient(ghexec.New(workspaceRoot), workspaceRoot)
		querier := NewGHCIRunQuerier(api)
		checker = cache.NewCICacheChecker(querier, nil)
	}

	// Build result
	result, err := FilterCIDispatch(directlyChanged, invalidated, headSHA, checker, workspaceRoot)
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
		fmt.Fprintf(os.Stderr, "\nFinal dispatch list (in dependency order): %s\n", strings.Join(result.Dispatch, " "))
		fmt.Fprintf(os.Stderr, "Skipped modules: %s\n", strings.Join(result.Skipped, " "))
		if len(result.CIDependencies) > 0 {
			fmt.Fprintf(os.Stderr, "CI dependencies: %v\n", result.CIDependencies)
		}
		fmt.Fprintf(os.Stderr, "\n")

		// Output variables to stdout for eval
		fmt.Printf("DISPATCH=\"%s\"\n", strings.Join(result.Dispatch, " "))
		fmt.Printf("SKIPPED=\"%s\"\n", strings.Join(result.Skipped, " "))
		fmt.Printf("DISPATCH_COUNT=%d\n", len(result.Dispatch))
		fmt.Printf("SKIPPED_COUNT=%d\n", len(result.Skipped))

		// Output CI dependencies as JSON for programmatic access
		if ciDepsJSON, err := json.Marshal(result.CIDependencies); err == nil {
			fmt.Printf("CI_DEPS_JSON='%s'\n", string(ciDepsJSON))
		}
		return 0
	}

	return internal.ExecuteGetCommand(func() (interface{}, error) {
		return result, nil
	})
}

// FilterCIDispatch filters modules for CI dispatch and returns them in dependency order.
// This is the main entry point that loads CI deps from config.
func FilterCIDispatch(directlyChangedStr, invalidatedStr, headSHA string, checker *cache.CICacheChecker, workspaceRoot string) (*CIDispatchResult, error) {
	// Load CI artifact deps from config.
	var ciDeps map[string][]string
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err == nil && cfg.Repository != nil {
		ciDeps = buildCIArtifactDeps(cfg.Repository)
	}

	return filterCIDispatchWithDeps(directlyChangedStr, invalidatedStr, headSHA, checker, workspaceRoot, ciDeps)
}

// filterCIDispatchWithDeps filters modules for CI dispatch with explicit CI deps (for testing).
func filterCIDispatchWithDeps(directlyChangedStr, invalidatedStr, headSHA string, checker *cache.CICacheChecker, workspaceRoot string, ciDeps map[string][]string) (*CIDispatchResult, error) {
	result := &CIDispatchResult{
		Dispatch:       []string{},
		Skipped:        []string{},
		Reasons:        make(map[string]string),
		CIDependencies: make(map[string][]string),
		HeadSHA:        headSHA,
	}

	// Get valid CI workflow modules for validation (skip when checker uses mock)
	var validModules map[string]bool
	if workspaceRoot != "" {
		var err error
		validModules, err = getValidCIModules(workspaceRoot)
		if err != nil {
			// Non-fatal: skip validation if we can't read workflows
			validModules = nil
		}
	}

	// Parse directly changed modules (always dispatch)
	directlyChanged := parseModuleList(directlyChangedStr)
	for _, module := range directlyChanged {
		// Validate module exists
		if validModules != nil && !validModules[module] {
			return nil, fmt.Errorf("invalid module %q: no CI workflow ci-%s.yaml exists", module, module)
		}
		result.Dispatch = append(result.Dispatch, module)
		result.Reasons[module] = "directly_changed"
	}

	// Parse invalidated modules (check for valid CI)
	invalidatedModules := parseModuleList(invalidatedStr)
	for _, module := range invalidatedModules {
		// Validate module exists
		if validModules != nil && !validModules[module] {
			return nil, fmt.Errorf("invalid module %q: no CI workflow ci-%s.yaml exists", module, module)
		}
	}

	for _, module := range invalidatedModules {
		ciResult := checker.Check(module, headSHA)

		if ciResult.Cached {
			result.Skipped = append(result.Skipped, module)
			result.Reasons[module] = ciResult.Reason
		} else {
			result.Dispatch = append(result.Dispatch, module)
			result.Reasons[module] = ciResult.Reason
		}
	}

	// Sort dispatch list by CI dependencies (topological sort)
	result.Dispatch = computeCIDispatchOrder(result.Dispatch, ciDeps)

	// Build filtered CI dependencies (only deps within dispatch set)
	result.CIDependencies = filterDependenciesToSet(result.Dispatch, ciDeps)

	// Set total modules count
	result.TotalModules = len(result.Dispatch) + len(result.Skipped)

	return result, nil
}

// filterDependenciesToSet filters a dependency map to only include modules in the given set.
func filterDependenciesToSet(modules []string, allDeps map[string][]string) map[string][]string {
	moduleSet := make(map[string]bool, len(modules))
	for _, m := range modules {
		moduleSet[m] = true
	}

	filtered := make(map[string][]string)
	for _, module := range modules {
		if deps, ok := allDeps[module]; ok {
			filteredDeps := make([]string, 0, len(deps))
			for _, dep := range deps {
				if moduleSet[dep] {
					filteredDeps = append(filteredDeps, dep)
				}
			}
			if len(filteredDeps) > 0 {
				filtered[module] = filteredDeps
			}
		}
	}
	return filtered
}

// computeCIDispatchOrder topologically sorts modules by CI artifact dependencies.
// Modules with no CI deps come first, followed by modules that depend on them.
// Uses Kahn's algorithm with pre-computed reverse dependencies for efficiency.
func computeCIDispatchOrder(modules []string, ciArtifactDeps map[string][]string) []string {
	if len(modules) == 0 {
		return []string{}
	}

	// Build module set for filtering
	moduleSet := make(map[string]bool, len(modules))
	for _, m := range modules {
		moduleSet[m] = true
	}

	// Build reverse dependency map (who depends on each module) and count in-degrees
	reverseDeps := make(map[string][]string)
	inDegree := make(map[string]int, len(modules))
	for _, m := range modules {
		inDegree[m] = 0
	}

	for _, module := range modules {
		if deps, ok := ciArtifactDeps[module]; ok {
			for _, dep := range deps {
				if moduleSet[dep] {
					inDegree[module]++
					reverseDeps[dep] = append(reverseDeps[dep], module)
				}
			}
		}
	}

	// Kahn's algorithm: start with modules that have no deps
	result := make([]string, 0, len(modules))
	queue := make([]string, 0, len(modules))

	for _, m := range modules {
		if inDegree[m] == 0 {
			queue = append(queue, m)
		}
	}
	sort.Strings(queue) // Deterministic order within same level

	for len(queue) > 0 {
		// Pop first element
		module := queue[0]
		queue = queue[1:]
		result = append(result, module)

		// Process only actual dependents using reverse deps map
		for _, dependent := range reverseDeps[module] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
		sort.Strings(queue) // Keep sorted for determinism
	}

	return result
}

// buildCIArtifactDeps builds a map of module -> CI artifact dependencies from config.
// Uses both CIDeps (from depends_on_ci) AND DependsOn (from depends_on) for modules
// that have CI workflows. This enables full dependency-aware wave computation.
func buildCIArtifactDeps(cfg *config.RepositoryConfig) map[string][]string {
	// Build a set of modules that have CI workflows
	ciModules := make(map[string]bool)
	for i := range cfg.Modules {
		m := &cfg.Modules[i]
		// Modules with versioning (non-Implicit) or explicit CIDeps likely have CI
		if m.Versioning != nil && m.Versioning.Scheme != "Implicit" {
			ciModules[m.Moniker] = true
		}
		if len(m.CIDeps) > 0 {
			ciModules[m.Moniker] = true
		}
	}

	deps := make(map[string][]string)
	for i := range cfg.Modules {
		module := &cfg.Modules[i]
		seen := make(map[string]bool)
		var allDeps []string

		// Add CIDeps first (explicit CI artifact deps)
		for _, d := range module.CIDeps {
			if !seen[d] {
				seen[d] = true
				allDeps = append(allDeps, d)
			}
		}

		// Add DependsOn entries that are CI modules
		for _, d := range module.DependsOn {
			if ciModules[d] && !seen[d] {
				seen[d] = true
				allDeps = append(allDeps, d)
			}
		}

		if len(allDeps) > 0 {
			deps[module.Moniker] = allDeps
		}
	}
	return deps
}

func parseModuleList(input string) []string {
	if input == "" {
		return []string{}
	}
	return strings.Fields(input)
}

// MockCheckerFromJSON parses a JSON mock string (map[string]bool) and returns a
// CICacheChecker backed by a MockCIRunQuerier. Returns nil if JSON is invalid.
func MockCheckerFromJSON(mockJSON, headSHA string) *cache.CICacheChecker {
	var mockStatus map[string]bool
	if err := json.Unmarshal([]byte(mockJSON), &mockStatus); err != nil {
		return nil
	}
	return mockCheckerFromStatus(mockStatus, headSHA)
}

// mockCheckerFromStatus converts the legacy --mock JSON format (map[string]bool)
// into a cache.CICacheChecker backed by a MockCIRunQuerier.
// Modules mapped to true get headSHA as their last successful run; false/missing get "".
func mockCheckerFromStatus(mockStatus map[string]bool, headSHA string) *cache.CICacheChecker {
	runs := make(map[string]string, len(mockStatus))
	for module, valid := range mockStatus {
		workflow := fmt.Sprintf("ci-%s.yaml", module)
		if valid {
			runs[workflow] = headSHA
		}
		// If !valid, we leave it absent → querier returns "" → checker says uncached
	}
	querier := cache.NewMockCIRunQuerier(runs)
	return cache.NewCICacheChecker(querier, nil)
}

func printCIDispatchUsage() {
	fmt.Println("Filter modules for CI dispatch based on existing valid CI")
	fmt.Println("")
	fmt.Println("Usage: clie get ci-dispatch [flags]")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  --directly-changed <modules>  Space-separated list of directly changed modules")
	fmt.Println("  --invalidated <modules>       Space-separated list of invalidated modules")
	fmt.Println("  --head-sha <sha>              HEAD SHA to check against (default: git HEAD)")
	fmt.Println("  --mock <json>                 Mock CI status (for testing)")
	fmt.Println("  --format shell                Output as shell variable assignments for eval")
	fmt.Println("  --as-yaml                     Output as YAML (default)")
	fmt.Println("  --as-json                     Output as JSON")
	fmt.Println("")
	fmt.Println("Logic:")
	fmt.Println("  - Directly changed modules are ALWAYS dispatched")
	fmt.Println("  - Invalidated modules are checked for existing valid CI:")
	fmt.Println("    - If successful CI exists at HEAD SHA: SKIP")
	fmt.Println("    - Otherwise: DISPATCH")
	fmt.Println("  - Dispatch list is topologically sorted by CI dependencies")
	fmt.Println("    (modules with no CI deps come first)")
	fmt.Println("")
	fmt.Println("Shell Format Output:")
	fmt.Println("  DISPATCH=\"mod1 mod2 mod3\"    # Modules in dependency order")
	fmt.Println("  SKIPPED=\"mod4\"               # Modules with valid CI at HEAD")
	fmt.Println("  DISPATCH_COUNT=3             # Number of modules to dispatch")
	fmt.Println("  SKIPPED_COUNT=1              # Number of skipped modules")
	fmt.Println("  CI_DEPS_JSON='{...}'         # CI dependency map as JSON")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  # Normal usage in CI")
	fmt.Println("  clie get ci-dispatch --directly-changed \"core\" --invalidated \"eac-cli docs\"")
	fmt.Println("")
	fmt.Println("  # Shell format for workflow scripting")
	fmt.Println("  eval $(clie get ci-dispatch --format shell --invalidated \"clie eac-ext\")")
	fmt.Println("  for module in $DISPATCH; do")
	fmt.Println("    gh workflow run \"ci-${module}.yaml\" --ref main")
	fmt.Println("  done")
	fmt.Println("")
	fmt.Println("  # Local testing with mock data")
	fmt.Println("  clie get ci-dispatch --directly-changed \"core\" --invalidated \"eac-cli docs\" \\")
	fmt.Println("    --mock '{\"eac-cli\": true, \"docs\": false}'")
	fmt.Println("")
	fmt.Println("  # Output as JSON")
	fmt.Println("  clie get ci-dispatch --as-json --directly-changed \"core\" --invalidated \"docs\"")
}

// getValidCIModules returns a set of valid CI workflow module names.
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
