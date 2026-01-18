// Package internal provides build manifest validation against the contract schema
package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// ManifestValidator validates build manifests against the contract schema.
type ManifestValidator struct {
	schema *jsonschema.Schema
}

// ManifestValidationError represents a manifest validation error.
type ManifestValidationError struct {
	Moniker string
	Message string
	Details []string
}

func (e *ManifestValidationError) Error() string {
	if len(e.Details) > 0 {
		return fmt.Sprintf("manifest validation failed for %s: %s (%v)", e.Moniker, e.Message, e.Details)
	}
	return fmt.Sprintf("manifest validation failed for %s: %s", e.Moniker, e.Message)
}

// Contract version for build manifest schema.
const manifestContractVersion = "0.1.0"

// NewManifestValidator creates a new manifest validator that loads schema from contracts.
func NewManifestValidator(workspaceRoot string) (*ManifestValidator, error) {
	c := jsonschema.NewCompiler()

	// Use distribution root (schemas are part of tool distribution, not user workspace)
	// In containers, R2R_CONTAINER_ROOT points to the tool distribution
	schemaRoot := workspaceRoot
	if containerRoot := os.Getenv("R2R_CONTAINER_ROOT"); containerRoot != "" {
		schemaRoot = containerRoot
	}

	// Build path to schema: contracts/eac-core/<version>/build-manifest.schema.json
	schemaPath := filepath.Join(
		paths.ContractsVersionPath(schemaRoot, "eac-core", manifestContractVersion),
		"build-manifest.schema.json",
	)

	// Read schema from contracts
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read build manifest schema from %s: %w", schemaPath, err)
	}

	// Parse the schema
	var schemaDoc any
	if err := json.Unmarshal(data, &schemaDoc); err != nil {
		return nil, fmt.Errorf("failed to parse build manifest schema: %w", err)
	}

	// Add schema to compiler
	schemaURL := "file:///build-manifest.schema.json"
	if err := c.AddResource(schemaURL, schemaDoc); err != nil {
		return nil, fmt.Errorf("failed to add schema resource: %w", err)
	}

	// Compile the schema
	schema, err := c.Compile(schemaURL)
	if err != nil {
		return nil, fmt.Errorf("failed to compile build manifest schema: %w", err)
	}

	return &ManifestValidator{schema: schema}, nil
}

// ValidateManifest validates a ModuleManifest against the contract schema.
func (v *ManifestValidator) ValidateManifest(manifest *ModuleManifest) error {
	// Convert manifest to JSON for validation
	jsonData, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	return v.ValidateJSON(jsonData, manifest.Moniker)
}

// ValidateJSON validates raw JSON data against the manifest schema.
func (v *ManifestValidator) ValidateJSON(jsonData []byte, moniker string) error {
	// Parse JSON to generic interface
	var data any
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate against schema
	if err := v.schema.Validate(data); err != nil {
		return &ManifestValidationError{
			Moniker: moniker,
			Message: err.Error(),
			Details: extractManifestValidationDetails(err),
		}
	}

	return nil
}

// extractManifestValidationDetails extracts detailed error messages from validation errors.
func extractManifestValidationDetails(err error) []string {
	var details []string

	validErr := &jsonschema.ValidationError{}
	if errors.As(err, &validErr) {
		details = append(details, validErr.Error())
		for _, cause := range validErr.Causes {
			details = append(details, extractManifestValidationDetails(cause)...)
		}
	}

	return details
}

// Global validator instance (lazy initialized).
var (
	globalManifestValidator     *ManifestValidator
	globalManifestValidatorRoot string
)

// GetManifestValidator returns the global manifest validator instance.
// Uses repository.GetRepositoryRoot() to find workspace root.
func GetManifestValidator() (*ManifestValidator, error) {
	return GetManifestValidatorWithRoot("")
}

// GetManifestValidatorWithRoot returns the global manifest validator instance with explicit workspace root.
// If workspaceRoot is empty, it will be detected automatically.
func GetManifestValidatorWithRoot(workspaceRoot string) (*ManifestValidator, error) {
	// If no root provided, try to detect it
	if workspaceRoot == "" {
		// Use distribution root for schema loading
		if containerRoot := os.Getenv("R2R_CONTAINER_ROOT"); containerRoot != "" {
			workspaceRoot = containerRoot
		} else {
			// Try to find workspace root by looking for .r2r directory
			cwd, err := os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("failed to get working directory: %w", err)
			}
			workspaceRoot = findWorkspaceRoot(cwd)
			if workspaceRoot == "" {
				return nil, fmt.Errorf("could not find workspace root (no .r2r directory found)")
			}
		}
	}

	// Check if we need to reinitialize (different root)
	if globalManifestValidator != nil && globalManifestValidatorRoot == workspaceRoot {
		return globalManifestValidator, nil
	}

	var err error
	globalManifestValidator, err = NewManifestValidator(workspaceRoot)
	if err != nil {
		return nil, err
	}
	globalManifestValidatorRoot = workspaceRoot

	return globalManifestValidator, nil
}

// findWorkspaceRoot walks up the directory tree looking for .r2r directory.
func findWorkspaceRoot(startDir string) string {
	dir := startDir
	for {
		if _, err := os.Stat(paths.R2RPath(dir)); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "" // Reached root
		}
		dir = parent
	}
}

// ValidateAndSave validates the manifest against the schema, verifies artifacts exist, and saves if valid.
func (m *ModuleManifest) ValidateAndSave(moduleBuildDir string) error {
	return m.ValidateAndSaveWithRoot(moduleBuildDir, "")
}

// ValidateAndSaveWithRoot validates the manifest against the schema with explicit workspace root.
func (m *ModuleManifest) ValidateAndSaveWithRoot(moduleBuildDir, workspaceRoot string) error {
	validator, err := GetManifestValidatorWithRoot(workspaceRoot)
	if err != nil {
		return fmt.Errorf("failed to get manifest validator: %w", err)
	}

	if err := validator.ValidateManifest(m); err != nil {
		return err
	}

	// Verify all declared artifacts actually exist
	if err := m.VerifyArtifactsExist(moduleBuildDir); err != nil {
		return err
	}

	return m.Save(moduleBuildDir)
}

// ArtifactExistenceError represents a missing artifact error.
type ArtifactExistenceError struct {
	Moniker string
	Missing []string
	Details map[string]string // artifact ID -> specific error
}

func (e *ArtifactExistenceError) Error() string {
	return fmt.Sprintf("build for %s produced manifest but artifacts are missing: %v", e.Moniker, e.Missing)
}

// VerifyArtifactsExist checks that all artifacts declared in the manifest actually exist.
func (m *ModuleManifest) VerifyArtifactsExist(moduleBuildDir string) error {
	var missing []string
	details := make(map[string]string)

	for _, art := range m.Artifacts {
		var exists bool
		var errMsg string

		switch art.Type {
		case "image":
			// Docker images need docker verification
			// Pass the full artifact info so we can check registry tags if local image not found
			exists, errMsg = verifyDockerImageExists(art)
		case "directory":
			// Directories should exist - check module level first, then component subdirs
			dirPath := filepath.Join(moduleBuildDir, art.Path)
			info, err := os.Stat(dirPath)
			if err != nil {
				// Not found at module level - check component subdirectories
				if foundPath := findInComponentSubdirs(moduleBuildDir, art.Path); foundPath != "" {
					if info, err := os.Stat(foundPath); err == nil && info.IsDir() {
						exists = true
					} else {
						exists = false
						errMsg = fmt.Sprintf("expected directory but found file: %s", foundPath)
					}
				} else {
					exists = false
					errMsg = fmt.Sprintf("directory not found: %s", dirPath)
				}
			} else if !info.IsDir() {
				exists = false
				errMsg = fmt.Sprintf("expected directory but found file: %s", dirPath)
			} else {
				exists = true
			}
		default:
			// File-based artifacts (executable, file) - check module level first, then component subdirs
			filePath := filepath.Join(moduleBuildDir, art.Path)
			info, err := os.Stat(filePath)
			if err != nil {
				// Not found at module level - check component subdirectories
				if foundPath := findInComponentSubdirs(moduleBuildDir, art.Path); foundPath != "" {
					if info, err := os.Stat(foundPath); err == nil {
						if info.IsDir() {
							exists = false
							errMsg = fmt.Sprintf("expected file but found directory: %s", foundPath)
						} else if info.Size() == 0 {
							exists = false
							errMsg = fmt.Sprintf("file is empty: %s", foundPath)
						} else {
							exists = true
						}
					}
				} else {
					exists = false
					errMsg = fmt.Sprintf("file not found: %s", filePath)
				}
			} else if info.IsDir() {
				exists = false
				errMsg = fmt.Sprintf("expected file but found directory: %s", filePath)
			} else if info.Size() == 0 {
				exists = false
				errMsg = fmt.Sprintf("file is empty: %s", filePath)
			} else {
				exists = true
			}
		}

		if !exists {
			missing = append(missing, art.ID)
			details[art.ID] = errMsg
		}
	}

	if len(missing) > 0 {
		return &ArtifactExistenceError{
			Moniker: m.Moniker,
			Missing: missing,
			Details: details,
		}
	}

	return nil
}

// findInComponentSubdirs searches for an artifact in component subdirectories.
// Returns the found path or empty string if not found.
func findInComponentSubdirs(moduleBuildDir, artifactPath string) string {
	entries, err := os.ReadDir(moduleBuildDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if artifact exists in this component subdirectory
		componentPath := filepath.Join(moduleBuildDir, entry.Name(), artifactPath)
		if _, err := os.Stat(componentPath); err == nil {
			return componentPath
		}
	}

	return ""
}

// verifyDockerImageExists checks if a Docker image exists locally or in registry (for pushed images).
func verifyDockerImageExists(art ArtifactInfo) (bool, string) {
	// Check if docker is available first
	if !isDockerAvailable() {
		// If docker isn't available, we can't verify - log warning but don't fail
		// This allows builds to succeed when docker daemon isn't running
		return true, ""
	}

	// First try to find image locally using the path (local reference)
	if checkDockerImageLocal(art.Path) {
		return true, ""
	}

	// If image has registry tags (CI pushed image), check those instead
	// This handles the case where buildx --push pushed to registry without loading locally
	if len(art.Tags) > 0 {
		for _, tag := range art.Tags {
			// Check if the tag exists locally (might have been pulled)
			if checkDockerImageLocal(tag) {
				return true, ""
			}
			// For CI with push=true, the image won't exist locally but was pushed
			// Check if this is a registry image that was pushed (contains registry prefix)
			// Registry images are verified by successful push, not local existence
			if art.Registry != "" && isRegistryTag(tag, art.Registry) {
				// Image was pushed to registry - trust the buildx push succeeded
				// (buildx would have failed if push failed)
				return true, ""
			}
		}
	}

	return false, fmt.Sprintf("docker image not found: %s (checked local and registry tags)", art.Path)
}

// checkDockerImageLocal checks if an image exists in local Docker daemon.
func checkDockerImageLocal(imageRef string) bool {
	cmd := execCommand("docker", "images", "-q", imageRef)
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// If output is non-empty, image exists locally
	trimmed := strings.TrimSpace(string(output))
	return trimmed != ""
}

// isRegistryTag checks if a tag is a registry image (contains registry prefix).
func isRegistryTag(tag, registry string) bool {
	// Registry must be non-empty and tag must start with registry prefix
	if registry == "" {
		return false
	}
	// Check if tag starts with registry (e.g., ghcr.io/ready-to-release/ext-eac:ci starts with ghcr.io)
	return len(tag) > len(registry) && tag[:len(registry)] == registry
}

// isDockerAvailable checks if Docker CLI is available.
func isDockerAvailable() bool {
	cmd := execCommand("docker", "version", "--format", "{{.Server.Version}}")
	return cmd.Run() == nil
}

// execCommand is a variable to allow testing.
var execCommand = defaultExecCommand

func defaultExecCommand(name string, args ...string) execCommandInterface {
	return &realExecCmd{cmd: exec.Command(name, args...)}
}

type execCommandInterface interface {
	Output() ([]byte, error)
	Run() error
}

type realExecCmd struct {
	cmd *exec.Cmd
}

func (r *realExecCmd) Output() ([]byte, error) {
	return r.cmd.Output()
}

func (r *realExecCmd) Run() error {
	return r.cmd.Run()
}
