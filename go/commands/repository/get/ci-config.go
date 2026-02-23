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
	"github.com/ready-to-release/eac/go/commands/repository/get/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/config"
)

type getCIConfigCommand struct{}

var _ core.SimpleCommandPort = (*getCIConfigCommand)(nil)

func (c *getCIConfigCommand) Name() string { return "get ci-config" }

func (c *getCIConfigCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-ci-config",
		Short:         "Derive CI configuration for a module from core config",
		Long: "Derives CI configuration for any module from the core config system.\n" +
			"All values are computed from repository.yml + blueprints.yml - no duplicate\n" +
			"configuration in workflow files needed.\n" +
			"\n" +
			"With --format shell, outputs shell variable assignments for eval:\n" +
			"  IS_CONTAINER=true|false\n" +
			"  IS_MULTI_CONTAINER=true|false\n" +
			"  CONTAINER_COMPONENTS=[{\"name\":\"...\"},...]\n" +
			"  CONTAINER_PUSH=true|false\n" +
			"  HAS_TESTS=true|false\n" +
			"  TEST_ON_WINDOWS=true|false\n" +
			"  TEST_ON_MACOS=true|false\n" +
			"  SCANS=sbom,vuln,secrets\n" +
			"  SCAN_FAIL_MODE=warn\n" +
			"  BUILD_EVIDENCE=true|false\n" +
			"  CROSS_COMPILE_WINDOWS=true|false\n" +
			"  BUILD_ARGS=--all\n" +
			"  DOWNLOAD_MODULES=clie\n" +
			"  CONTAINER_TEST_SCRIPT=path\n" +
			"  TEST_SUITES=unit,integration\n" +
			"  TEST_SUITES_FULL=unit,integration,acceptance\n" +
			"\n" +
			"With --format github-output, outputs KEY=value lines for $GITHUB_OUTPUT.",
		Flags: []core.FlagSpec{
			{Name: "module", Type: "string", Usage: "Module moniker to derive CI config for", Required: true, Completion: []string{"modules"}},
			{Name: "format", Type: "string", Usage: "Output format: shell (eval-friendly) or github-output (KEY=value for $GITHUB_OUTPUT)"},
		},
	}
}

func (c *getCIConfigCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetCIConfig()
}

// CIConfigResult represents the derived CI configuration for a module.
type CIConfigResult struct {
	Module              string `json:"module" yaml:"module"`
	IsContainer         bool   `json:"is_container" yaml:"is_container"`
	IsMultiContainer    bool   `json:"is_multi_container" yaml:"is_multi_container"`
	ContainerComponents string `json:"container_components,omitempty" yaml:"container_components,omitempty"`
	ContainerPush       bool   `json:"container_push" yaml:"container_push"`
	HasTests            bool   `json:"has_tests" yaml:"has_tests"`
	TestOnWindows       bool   `json:"test_on_windows" yaml:"test_on_windows"`
	TestOnMacos         bool   `json:"test_on_macos" yaml:"test_on_macos"`
	Scans               string `json:"scans" yaml:"scans"`
	ScanFailMode        string `json:"scan_fail_mode" yaml:"scan_fail_mode"`
	BuildEvidence       bool   `json:"build_evidence" yaml:"build_evidence"`
	CrossCompileWindows bool   `json:"cross_compile_windows" yaml:"cross_compile_windows"`
	BuildArgs           string `json:"build_args" yaml:"build_args"`
	DownloadModules     string `json:"download_modules" yaml:"download_modules"`
	ContainerTestScript string `json:"container_test_script" yaml:"container_test_script"`
	TestSuites          string `json:"test_suites" yaml:"test_suites"`
	TestSuitesFull      string `json:"test_suites_full" yaml:"test_suites_full"`
}

func GetCIConfig() int {
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	module := ""
	format := ""

	for i, arg := range os.Args {
		switch arg {
		case "--module":
			if i+1 < len(os.Args) {
				module = os.Args[i+1]
			}
		case "--format":
			if i+1 < len(os.Args) {
				format = os.Args[i+1]
			}
		case "--help", "-h":
			printCIConfigUsage()
			return 0
		}
	}

	if module == "" {
		fmt.Fprintf(os.Stderr, "Error: --module is required\n")
		return 1
	}

	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load configuration: %v\n", err)
		return 1
	}

	result, err := deriveCIConfig(cfg, module)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	switch format {
	case "shell":
		outputCIConfigShell(result)
		return 0
	case "github-output":
		outputCIConfigGitHub(result)
		return 0
	default:
		return internal.ExecuteGetCommand(func() (interface{}, error) {
			return result, nil
		})
	}
}

func deriveCIConfig(cfg *config.EACConfig, moniker string) (*CIConfigResult, error) {
	if cfg.Repository == nil {
		return nil, fmt.Errorf("repository config not loaded")
	}

	mod := cfg.Repository.GetByMoniker(moniker)
	if mod == nil {
		return nil, fmt.Errorf("module %q not found", moniker)
	}

	result := &CIConfigResult{
		Module:       moniker,
		ScanFailMode: "warn",
		TestSuites:   "unit,integration",
		TestSuitesFull: "unit,integration,acceptance",
	}

	// IS_CONTAINER + CONTAINER_PUSH: Check for dockerfile component with docker_build
	dockerCfg := mod.GetDockerBuildConfig()
	if dockerCfg != nil {
		result.IsContainer = true
		// Default push=true for containers; override if explicitly set
		result.ContainerPush = true
		if dockerCfg.Push != nil {
			result.ContainerPush = *dockerCfg.Push
		}
	}

	// IS_MULTI_CONTAINER + CONTAINER_COMPONENTS: matrix for multi-container modules
	pushable := mod.GetPushableContainerComponents()
	if len(pushable) > 1 {
		result.IsMultiContainer = true
		type containerComponentCI struct {
			Name string `json:"name"`
		}
		entries := make([]containerComponentCI, len(pushable))
		for i, name := range pushable {
			entries[i] = containerComponentCI{Name: name}
		}
		jsonBytes, _ := json.Marshal(entries)
		result.ContainerComponents = string(jsonBytes)
	}

	// HAS_TESTS: Any component has testers defined in its component kind
	if cfg.ComponentKinds != nil {
		for compName, entry := range mod.Components {
			if entry == nil {
				continue
			}
			compType := compName
			if entry.Type != "" {
				compType = entry.Type
			}
			ct := cfg.ComponentKinds.Get(compType)
			if ct != nil && ct.IsTestable() {
				result.HasTests = true
				break
			}
		}
	}

	// TEST_ON_WINDOWS, TEST_ON_MACOS, CROSS_COMPILE_WINDOWS: from artifact_matrix on Go component
	_, goComp := mod.Components.GetFirstByType("go")
	matrixName := ""
	if goComp != nil && goComp.Build != nil {
		matrixName = goComp.Build.ArtifactMatrixRef
	}
	if matrixName != "" && cfg.Blueprints != nil {
		result.CrossCompileWindows = strings.Contains(matrixName, "cross-platform")
		// Check matrix entries for platform indicators
		if matrix, ok := cfg.Blueprints.ArtifactMatrices[matrixName]; ok && matrix != nil {
			for _, entry := range matrix.Entries {
				if strings.Contains(entry.ID, "windows") || strings.Contains(entry.Pattern, "windows") {
					result.TestOnWindows = true
					result.CrossCompileWindows = true
				}
				if strings.Contains(entry.ID, "darwin") || strings.Contains(entry.Pattern, "darwin") {
					result.TestOnMacos = true
				}
			}
		}
		// Also check parent matrices
		if result.CrossCompileWindows {
			result.TestOnWindows = true
		}
	}

	// SCANS: Union of scanners from all component kinds, deduplicated
	scanSet := make(map[string]bool)
	if cfg.ComponentKinds != nil {
		for compName, entry := range mod.Components {
			if entry == nil {
				continue
			}
			compType := compName
			if entry.Type != "" {
				compType = entry.Type
			}
			ct := cfg.ComponentKinds.Get(compType)
			if ct != nil {
				for _, s := range ct.GetScanners() {
					scanSet[s] = true
				}
			}
		}
	}
	if len(scanSet) > 0 {
		scanList := make([]string, 0, len(scanSet))
		for s := range scanSet {
			scanList = append(scanList, s)
		}
		sort.Strings(scanList)
		result.Scans = strings.Join(scanList, ",")
	}

	// BUILD_EVIDENCE: has evidence-book components
	result.BuildEvidence = len(mod.GetEvidenceBooks()) > 0

	// BUILD_ARGS: CI always builds all components
	result.BuildArgs = "--all"

	// DOWNLOAD_MODULES: from CIDeps (depends_on_ci)
	if len(mod.CIDeps) > 0 {
		result.DownloadModules = strings.Join(mod.CIDeps, " ")
	}

	// CONTAINER_TEST_SCRIPT: convention containers/<module>/ci-test.sh
	containerTestPath := filepath.Join("containers", moniker, "ci-test.sh")
	if cfg.RepoRoot != "" {
		fullPath := filepath.Join(cfg.RepoRoot, containerTestPath)
		if _, err := os.Stat(fullPath); err == nil {
			result.ContainerTestScript = containerTestPath
		}
	}

	return result, nil
}

func outputCIConfigShell(r *CIConfigResult) {
	fmt.Printf("IS_CONTAINER=%s\n", boolToStr(r.IsContainer))
	fmt.Printf("IS_MULTI_CONTAINER=%s\n", boolToStr(r.IsMultiContainer))
	if r.ContainerComponents != "" {
		fmt.Printf("CONTAINER_COMPONENTS='%s'\n", r.ContainerComponents)
	}
	fmt.Printf("CONTAINER_PUSH=%s\n", boolToStr(r.ContainerPush))
	fmt.Printf("HAS_TESTS=%s\n", boolToStr(r.HasTests))
	fmt.Printf("TEST_ON_WINDOWS=%s\n", boolToStr(r.TestOnWindows))
	fmt.Printf("TEST_ON_MACOS=%s\n", boolToStr(r.TestOnMacos))
	fmt.Printf("SCANS=\"%s\"\n", r.Scans)
	fmt.Printf("SCAN_FAIL_MODE=\"%s\"\n", r.ScanFailMode)
	fmt.Printf("BUILD_EVIDENCE=%s\n", boolToStr(r.BuildEvidence))
	fmt.Printf("CROSS_COMPILE_WINDOWS=%s\n", boolToStr(r.CrossCompileWindows))
	fmt.Printf("BUILD_ARGS=\"%s\"\n", r.BuildArgs)
	fmt.Printf("DOWNLOAD_MODULES=\"%s\"\n", r.DownloadModules)
	fmt.Printf("CONTAINER_TEST_SCRIPT=\"%s\"\n", r.ContainerTestScript)
	fmt.Printf("TEST_SUITES=\"%s\"\n", r.TestSuites)
	fmt.Printf("TEST_SUITES_FULL=\"%s\"\n", r.TestSuitesFull)
}

func outputCIConfigGitHub(r *CIConfigResult) {
	fmt.Printf("is-container=%s\n", boolToStr(r.IsContainer))
	fmt.Printf("is-multi-container=%s\n", boolToStr(r.IsMultiContainer))
	if r.ContainerComponents != "" {
		fmt.Printf("container-components=%s\n", r.ContainerComponents)
	}
	fmt.Printf("container-push=%s\n", boolToStr(r.ContainerPush))
	fmt.Printf("has-tests=%s\n", boolToStr(r.HasTests))
	fmt.Printf("test-on-windows=%s\n", boolToStr(r.TestOnWindows))
	fmt.Printf("test-on-macos=%s\n", boolToStr(r.TestOnMacos))
	fmt.Printf("scans=%s\n", r.Scans)
	fmt.Printf("scan-fail-mode=%s\n", r.ScanFailMode)
	fmt.Printf("build-evidence=%s\n", boolToStr(r.BuildEvidence))
	fmt.Printf("cross-compile-windows=%s\n", boolToStr(r.CrossCompileWindows))
	fmt.Printf("build-args=%s\n", r.BuildArgs)
	fmt.Printf("download-modules=%s\n", r.DownloadModules)
	fmt.Printf("container-test-script=%s\n", r.ContainerTestScript)
	fmt.Printf("test-suites=%s\n", r.TestSuites)
	fmt.Printf("test-suites-full=%s\n", r.TestSuitesFull)
}

func boolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func printCIConfigUsage() {
	fmt.Println("Derive CI configuration for a module from core config")
	fmt.Println("")
	fmt.Println("Usage: eac get ci-config --module <moniker> [--format shell|github-output]")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  --module <moniker>    Module moniker (required)")
	fmt.Println("  --format <format>     Output format: shell, github-output, yaml (default), json")
	fmt.Println("")
	fmt.Println("Output variables:")
	fmt.Println("  IS_CONTAINER           Has dockerfile component with push=true")
	fmt.Println("  CONTAINER_PUSH         docker_build.push value")
	fmt.Println("  HAS_TESTS              Has components with testers defined")
	fmt.Println("  TEST_ON_WINDOWS        artifact_matrix includes windows")
	fmt.Println("  TEST_ON_MACOS          artifact_matrix includes macos")
	fmt.Println("  SCANS                  Aggregated scanners from component kinds")
	fmt.Println("  SCAN_FAIL_MODE         Default scan failure mode")
	fmt.Println("  BUILD_EVIDENCE         has evidence-book components")
	fmt.Println("  CROSS_COMPILE_WINDOWS  artifact_matrix is cross-platform")
	fmt.Println("  BUILD_ARGS             CI always builds all components")
	fmt.Println("  DOWNLOAD_MODULES       From depends_on_ci, space-separated")
	fmt.Println("  CONTAINER_TEST_SCRIPT  Convention: containers/<module>/ci-test.sh")
	fmt.Println("  TEST_SUITES            PR test suites")
	fmt.Println("  TEST_SUITES_FULL       Full CI test suites")
}
