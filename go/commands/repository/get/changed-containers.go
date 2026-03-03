package get

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/commands/repository/get/internal"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/tool"
)

const defaultRepositoryOwner = "ready-to-release"

type getChangedContainersCommand struct{}

var _ core.SimpleCommandPort = (*getChangedContainersCommand)(nil)

func (c *getChangedContainersCommand) Name() string { return "get changed-containers" }

func (c *getChangedContainersCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-changed-containers",
		Short:         "Get container components requiring rebuild based on per-component file changes",
		Long: "Determines which container components in a multi-container module need rebuilding.\n" +
			"\n" +
			"For each component, queries the container registry (GHCR) for the last-build SHA tag,\n" +
			"then checks if any of the component's files changed since that SHA.\n" +
			"\n" +
			"Fail-open design: any error (registry query, git diff) triggers a rebuild for safety.\n" +
			"\n" +
			"Output formats:\n" +
			"  --format github-output: KEY=value lines for $GITHUB_OUTPUT\n" +
			"  --format shell: shell variable assignments\n" +
			"  default: YAML/JSON/TOML via standard get command",
		Flags: []core.FlagSpec{
			{Name: "module", Type: "string", Usage: "Module moniker", Required: true, Completion: []string{"modules"}},
			{Name: "components", Type: "string", Usage: "JSON array of component objects, e.g. [{\"name\":\"drawio-oci\"}]"},
			{Name: "head-sha", Type: "string", Usage: "HEAD commit SHA to compare against"},
			{Name: "format", Type: "string", Usage: "Output format: github-output, shell, or standard get formats"},
			{Name: "force-all", Type: "bool", Usage: "Skip detection and return all components as changed"},
		},
	}
}

func (c *getChangedContainersCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetChangedContainers()
}

// ContainerChangeResult represents per-component change detection output.
type ContainerChangeResult struct {
	Module            string                     `json:"module" yaml:"module"`
	HeadSHA           string                     `json:"head_sha" yaml:"head_sha"`
	AllComponents     []string                   `json:"all_components" yaml:"all_components"`
	ChangedComponents []ContainerComponentStatus `json:"changed_components" yaml:"changed_components"`
	SkippedComponents []ContainerComponentStatus `json:"skipped_components" yaml:"skipped_components"`
	FilteredMatrix    string                     `json:"filtered_matrix,omitempty" yaml:"filtered_matrix,omitempty"`
}

// ContainerComponentStatus tracks change status for a single container component.
type ContainerComponentStatus struct {
	Name         string `json:"name" yaml:"name"`
	LastBuildSHA string `json:"last_build_sha,omitempty" yaml:"last_build_sha,omitempty"`
	Reason       string `json:"reason" yaml:"reason"`
	FilesChanged int    `json:"files_changed,omitempty" yaml:"files_changed,omitempty"`
}

// containerComponentInput represents a component entry from the --components JSON array.
type containerComponentInput struct {
	Name string `json:"name"`
}

func GetChangedContainers() int {
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	module := ""
	componentsJSON := ""
	headSHA := ""
	format := ""
	forceAll := false

	for i, arg := range os.Args {
		switch arg {
		case "--module":
			if i+1 < len(os.Args) {
				module = os.Args[i+1]
			}
		case "--components":
			if i+1 < len(os.Args) {
				componentsJSON = os.Args[i+1]
			}
		case "--head-sha":
			if i+1 < len(os.Args) {
				headSHA = os.Args[i+1]
			}
		case "--format":
			if i+1 < len(os.Args) {
				format = os.Args[i+1]
			}
		case "--force-all":
			forceAll = true
		}
	}

	if module == "" {
		fmt.Fprintf(os.Stderr, "Error: --module is required\n")
		return 1
	}

	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Parse component names from JSON or discover from config
	componentNames, err := resolveComponentNames(componentsJSON, module)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	if len(componentNames) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no container components found for module %q\n", module)
		return 1
	}

	// Resolve HEAD SHA
	if headSHA == "" {
		headSHA, err = getCurrentSHA(workspaceRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: getting current SHA: %v\n", err)
			return 1
		}
	}

	// Build result
	var result *ContainerChangeResult
	if forceAll {
		result = buildForceAllResult(module, headSHA, componentNames)
	} else {
		// Create registry querier
		owner := os.Getenv("GITHUB_REPOSITORY_OWNER")
		if owner == "" {
			owner = defaultRepositoryOwner
		}

		// Check for mock support
		querier := loadMockedContainerRegistry()
		if querier == nil {
			querier = NewGHCRQuerier(tool.GlobalToolSystem(), owner, workspaceRoot)
		}

		result, err = detectChangedContainers(module, componentNames, headSHA, workspaceRoot, querier)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
	}

	// Output
	switch format {
	case "github-output":
		outputChangedContainersGitHub(result)
		return 0
	case "shell":
		outputChangedContainersShell(result)
		return 0
	default:
		return internal.ExecuteGetCommand(func() (interface{}, error) {
			return result, nil
		})
	}
}

// resolveComponentNames parses component names from JSON or discovers them from config.
func resolveComponentNames(componentsJSON, module string) ([]string, error) {
	if componentsJSON != "" {
		var inputs []containerComponentInput
		if err := json.Unmarshal([]byte(componentsJSON), &inputs); err != nil {
			return nil, fmt.Errorf("parsing --components JSON: %w", err)
		}
		names := make([]string, len(inputs))
		for i, input := range inputs {
			names[i] = input.Name
		}
		return names, nil
	}

	// Discover from config (uses config.Module which has GetPushableContainerComponents)
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	if cfg.Repository == nil {
		return nil, fmt.Errorf("repository config not loaded")
	}
	mod := cfg.Repository.GetByMoniker(module)
	if mod == nil {
		return nil, fmt.Errorf("module %q not found", module)
	}
	return mod.GetPushableContainerComponents(), nil
}

func buildForceAllResult(module, headSHA string, components []string) *ContainerChangeResult {
	result := &ContainerChangeResult{
		Module:        module,
		HeadSHA:       headSHA,
		AllComponents: components,
	}
	for _, name := range components {
		result.ChangedComponents = append(result.ChangedComponents, ContainerComponentStatus{
			Name:   name,
			Reason: "force_all",
		})
	}
	result.FilteredMatrix = buildFilteredMatrix(result.ChangedComponents)
	return result
}

func detectChangedContainers(
	module string,
	components []string,
	headSHA string,
	workspaceRoot string,
	registryQuerier ContainerRegistryQuerier,
) (*ContainerChangeResult, error) {
	// Load module registry for file ownership
	registry, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("loading module registry: %w", err)
	}

	mod, ok := registry.Get(module)
	if !ok {
		return nil, fmt.Errorf("module %q not found", module)
	}

	result := &ContainerChangeResult{
		Module:        module,
		HeadSHA:       headSHA,
		AllComponents: components,
	}

	shortHead := headSHA
	if len(shortHead) > 7 {
		shortHead = shortHead[:7]
	}

	// Track changed files per base SHA to avoid redundant git diffs
	diffCache := make(map[string][]string)

	for _, compName := range components {
		ctx := context.Background()

		// 1. Query registry for last-build SHA
		lastBuildSHA, err := registryQuerier.LastBuildSHA(ctx, compName)
		if err != nil {
			// Registry query failed -> rebuild to be safe
			result.ChangedComponents = append(result.ChangedComponents, ContainerComponentStatus{
				Name:   compName,
				Reason: fmt.Sprintf("registry_query_failed: %v", err),
			})
			continue
		}

		// 2. No previous build -> must build
		if lastBuildSHA == "" {
			result.ChangedComponents = append(result.ChangedComponents, ContainerComponentStatus{
				Name:   compName,
				Reason: "no_previous_build",
			})
			continue
		}

		// 3. Already built at HEAD -> skip
		if lastBuildSHA == shortHead {
			result.SkippedComponents = append(result.SkippedComponents, ContainerComponentStatus{
				Name:         compName,
				LastBuildSHA: lastBuildSHA,
				Reason:       "already_built_at_head",
			})
			continue
		}

		// 4. Different SHA -> check if component's files changed
		changedFiles, ok := diffCache[lastBuildSHA]
		if !ok {
			changedFiles, err = getChangedFilesBetweenSHAs(lastBuildSHA, headSHA, workspaceRoot)
			if err != nil {
				// git diff failed -> rebuild
				result.ChangedComponents = append(result.ChangedComponents, ContainerComponentStatus{
					Name:         compName,
					LastBuildSHA: lastBuildSHA,
					Reason:       fmt.Sprintf("diff_failed: %v", err),
				})
				continue
			}
			diffCache[lastBuildSHA] = changedFiles
		}

		// 5. Check for shared files (module-owned but not component-owned)
		sharedFiles := detectSharedFileChanges(changedFiles, mod, components)

		if len(sharedFiles) > 0 {
			result.ChangedComponents = append(result.ChangedComponents, ContainerComponentStatus{
				Name:         compName,
				LastBuildSHA: lastBuildSHA,
				Reason:       fmt.Sprintf("shared_files_changed_since_%s", lastBuildSHA),
				FilesChanged: len(sharedFiles),
			})
			continue
		}

		// 6. Filter to files owned by THIS component
		componentFiles := filterFilesForComponentOwnership(changedFiles, mod, compName)

		if len(componentFiles) == 0 {
			result.SkippedComponents = append(result.SkippedComponents, ContainerComponentStatus{
				Name:         compName,
				LastBuildSHA: lastBuildSHA,
				Reason:       "no_component_file_changes",
			})
		} else {
			result.ChangedComponents = append(result.ChangedComponents, ContainerComponentStatus{
				Name:         compName,
				LastBuildSHA: lastBuildSHA,
				Reason:       fmt.Sprintf("files_changed_since_%s", lastBuildSHA),
				FilesChanged: len(componentFiles),
			})
		}
	}

	// Build filtered matrix for GitHub Actions
	result.FilteredMatrix = buildFilteredMatrix(result.ChangedComponents)

	return result, nil
}

// filterFilesForComponentOwnership returns files that are owned by a specific component.
func filterFilesForComponentOwnership(files []string, mod *modules.ModuleContract, componentName string) []string {
	var matched []string
	for _, f := range files {
		if mod.MatchesFileForComponent(f, componentName) {
			matched = append(matched, f)
		}
	}
	return matched
}

// detectSharedFileChanges finds files that belong to the module but not to any specific component.
func detectSharedFileChanges(changedFiles []string, mod *modules.ModuleContract, components []string) []string {
	var shared []string
	for _, f := range changedFiles {
		if !mod.MatchesFile(f) {
			continue
		}
		ownedByComponent := false
		for _, comp := range components {
			if mod.MatchesFileForComponent(f, comp) {
				ownedByComponent = true
				break
			}
		}
		if !ownedByComponent {
			shared = append(shared, f)
		}
	}
	return shared
}

// buildFilteredMatrix creates a JSON array of changed components for GitHub Actions matrix.
func buildFilteredMatrix(changed []ContainerComponentStatus) string {
	type matrixEntry struct {
		Name string `json:"name"`
	}
	entries := make([]matrixEntry, len(changed))
	for i, c := range changed {
		entries[i] = matrixEntry{Name: c.Name}
	}
	data, err := json.Marshal(entries)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func extractComponentNames(statuses []ContainerComponentStatus) []string {
	names := make([]string, len(statuses))
	for i, s := range statuses {
		names[i] = s.Name
	}
	return names
}

func outputChangedContainersGitHub(r *ContainerChangeResult) {
	fmt.Printf("filtered-components=%s\n", r.FilteredMatrix)
	fmt.Printf("has-changes=%s\n", boolToStr(len(r.ChangedComponents) > 0))
	fmt.Printf("skipped-components=%s\n", strings.Join(extractComponentNames(r.SkippedComponents), ","))
}

func outputChangedContainersShell(r *ContainerChangeResult) {
	fmt.Printf("CHANGED_COMPONENTS=\"%s\"\n", strings.Join(extractComponentNames(r.ChangedComponents), " "))
	fmt.Printf("SKIPPED_COMPONENTS=\"%s\"\n", strings.Join(extractComponentNames(r.SkippedComponents), " "))
	fmt.Printf("FILTERED_MATRIX='%s'\n", r.FilteredMatrix)
	fmt.Printf("HAS_CHANGES=\"%s\"\n", boolToStr(len(r.ChangedComponents) > 0))
}

// loadMockedContainerRegistry loads mocked container registry data from
// EAC_MOCK_CONTAINER_REGISTRY environment variable (JSON file path).
func loadMockedContainerRegistry() ContainerRegistryQuerier {
	mockPath := os.Getenv(environments.EnvEACMockContainerRegistry)
	if mockPath == "" {
		return nil
	}

	data, err := os.ReadFile(mockPath)
	if err != nil {
		return nil
	}

	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}

	results := make(map[string]string)
	errors := make(map[string]error)
	for k, v := range raw {
		if v == "__ERROR__" {
			errors[k] = fmt.Errorf("mock error for %s", k)
		} else {
			results[k] = v
		}
	}

	return &mockContainerRegistryQuerier{
		results: results,
		errors:  errors,
	}
}
