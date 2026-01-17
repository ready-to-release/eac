package reports

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/git"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/testing"
)

// SpecFile represents a specification file with its status and metadata.
type SpecFile struct {
	FilePath      string `json:"file_path" yaml:"file_path" toml:"file_path"`
	RelativePath  string `json:"relative_path" yaml:"relative_path" toml:"relative_path"`
	Module        string `json:"module" yaml:"module" toml:"module"`
	FeatureName   string `json:"feature_name" yaml:"feature_name" toml:"feature_name"`
	Title         string `json:"title" yaml:"title" toml:"title"`
	Status        string `json:"status" yaml:"status" toml:"status"` // Added, Modified, Deleted
	ScenarioCount int    `json:"scenario_count" yaml:"scenario_count" toml:"scenario_count"`
}

// SpecsReport represents specifications for a module and version.
type SpecsReport struct {
	Module         string     `json:"module" yaml:"module" toml:"module"`
	Version        string     `json:"version" yaml:"version" toml:"version"`
	SpecFiles      []SpecFile `json:"spec_files" yaml:"spec_files" toml:"spec_files"`
	AddedCount     int        `json:"added_count" yaml:"added_count" toml:"added_count"`
	ModifiedCount  int        `json:"modified_count" yaml:"modified_count" toml:"modified_count"`
	DeletedCount   int        `json:"deleted_count" yaml:"deleted_count" toml:"deleted_count"`
	TotalScenarios int        `json:"total_scenarios" yaml:"total_scenarios" toml:"total_scenarios"`
}

// gitRepo holds the git repository instance for testing (allows mock injection).
var gitRepo git.GitRepository

// getGitRepo returns the git repository, initializing it if needed.
// This pattern matches other commands (create/squash-message, create/spec, etc.)
func getGitRepo(workspaceRoot string) (git.GitRepository, error) {
	if gitRepo != nil {
		return gitRepo, nil
	}
	repo, err := git.NewManager(nil).Open(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}
	return repo, nil
}

// SetGitRepo allows tests to inject a mock repository.
// This enables proper unit testing without requiring a real git repository.
func SetGitRepo(repo git.GitRepository) {
	gitRepo = repo
}

// GetSpecs loads specifications for a module and version
// Version can be: empty (unreleased), "latest", "unreleased", or specific version number
// For bundle/container modules with dependencies, specs are aggregated from all dependent modules
// Branch specifies which branch to query (default "main"). Use "HEAD" or "current" for current branch.
func GetSpecs(workspaceRoot, module, version, branch string) (*SpecsReport, error) {
	// Load config to get module information
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	modContract, ok := cfg.Repository.GetModule(module)
	if !ok {
		return nil, fmt.Errorf("module not found: %s", module)
	}

	// Resolve version using common helper with tag validation
	versionInfo, err := ResolveVersionWithValidation(workspaceRoot, module, version)
	if err != nil {
		return nil, err
	}

	// Get git repository using the standard interface pattern
	repo, err := getGitRepo(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	// Get commits for the version range
	var commits []git.CommitInfo

	// Determine end reference based on branch parameter and version
	endRef := branch
	if endRef == "" {
		// Default to configured trunk branch (usually "main" or "master")
		endRef = cfg.Repository.Repository.TrunkBranch
		if endRef == "" {
			endRef = "main" // Fallback if not configured
		}
	}
	// Special case: "HEAD" or "current" means use current branch
	if endRef == "HEAD" || endRef == "current" {
		endRef = "HEAD"
	}
	// Try origin/branch if local branch doesn't exist
	// (resolveRef in git layer handles this automatically)

	// For specific versions (not unreleased), use the version's tag as end point
	if !versionInfo.IsUnreleased && versionInfo.GitTag != "" {
		endRef = versionInfo.GitTag
	}

	if versionInfo.PreviousGitTag != "" {
		commits, err = repo.CommitsBetween(versionInfo.PreviousGitTag, endRef)
	} else {
		// No previous release - get all commits up to the end point
		commits, err = repo.CommitsBetween("", endRef)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	// Determine which modules to query for specs
	// Bundle/container modules aggregate specs from all dependencies
	var modulesToQuery []string
	if len(modContract.DependsOn) > 0 {
		// Bundle module - include all dependencies
		modulesToQuery = make([]string, 0, len(modContract.DependsOn)+1)
		modulesToQuery = append(modulesToQuery, modContract.DependsOn...)
		// Also include the bundle's own specs if directory exists
		modulesToQuery = append(modulesToQuery, module)
	} else {
		// Regular module - just query itself
		modulesToQuery = []string{module}
	}

	// Build report aggregating specs from all modules
	report := &SpecsReport{
		Module:    versionInfo.Module,
		Version:   versionInfo.VersionNumber,
		SpecFiles: []SpecFile{},
	}

	// Aggregate specs from all modules
	for _, mod := range modulesToQuery {
		specFiles := extractSpecFiles(workspaceRoot, mod, commits)

		for filePath, status := range specFiles {
			spec := SpecFile{
				FilePath:     filePath,
				RelativePath: relativePath(filePath, workspaceRoot),
				Status:       status,
			}

			// Parse feature file if it exists (not deleted)
			if status != "Deleted" {
				feature, err := testing.ParseFeatureFile(filePath)
				if err == nil {
					spec.Module = feature.Module
					spec.FeatureName = feature.FeatureName
					spec.Title = feature.Title
					spec.ScenarioCount = len(feature.Scenarios)
					report.TotalScenarios += spec.ScenarioCount
				}
			}

			report.SpecFiles = append(report.SpecFiles, spec)

			// Update counts
			switch status {
			case "Added":
				report.AddedCount++
			case "Modified":
				report.ModifiedCount++
			case "Deleted":
				report.DeletedCount++
			}
		}
	}

	return report, nil
}

// extractSpecFiles extracts unique .feature files from commits
// Returns a map of absolute file path -> status (Added/Modified/Deleted).
func extractSpecFiles(workspaceRoot, module string, commits []git.CommitInfo) map[string]string {
	specFiles := make(map[string]string)
	moduleSpecsPrefix := filepath.Join(paths.SpecsDir, module)

	// Normalize to forward slashes for comparison
	moduleSpecsPrefixNorm := filepath.ToSlash(moduleSpecsPrefix)

	// Track file states - iterate from oldest to newest
	// Reverse the commits slice to process oldest first
	for i := len(commits) - 1; i >= 0; i-- {
		commit := commits[i]
		for _, file := range commit.Files {
			// Normalize to forward slashes for comparison
			normalizedFile := filepath.ToSlash(file)

			// Only process .feature files in the module's specs directory
			if !strings.HasPrefix(normalizedFile, moduleSpecsPrefixNorm+"/") {
				continue
			}
			if !strings.HasSuffix(normalizedFile, ".feature") {
				continue
			}

			// Build absolute path
			absPath := filepath.Join(workspaceRoot, file)

			// Determine status
			// If file doesn't exist now, it was deleted
			// If file exists and wasn't seen before, it was added
			// If file exists and was seen before, it was modified

			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				specFiles[absPath] = "Deleted"
			} else {
				// File exists
				if _, seen := specFiles[absPath]; !seen {
					// First time seeing this file - it was added
					specFiles[absPath] = "Added"
				} else {
					// File seen before and still exists - it was modified
					if specFiles[absPath] != "Deleted" {
						specFiles[absPath] = "Modified"
					}
				}
			}
		}
	}

	return specFiles
}

func relativePath(absPath, root string) string {
	rel, err := filepath.Rel(root, absPath)
	if err != nil {
		return absPath
	}
	return filepath.ToSlash(rel)
}
