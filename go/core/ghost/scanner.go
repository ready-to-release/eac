package ghost

import (
	"path"
	"strings"

	"github.com/ready-to-release/eac/go/core/domain/modules"
)

// FileSource provides access to tracked files.
// This interface allows the scanner to work with FileCache or mocks.
type FileSource interface {
	TrackedFiles() ([]string, error)
}

// ScanOptions configures the ghost scanner.
type ScanOptions struct {
	// Alias is the ghost prefix (default: "ghost")
	Alias string
}

// DefaultScanOptions returns sensible defaults.
func DefaultScanOptions() ScanOptions {
	return ScanOptions{
		Alias: "ghost",
	}
}

// Scanner discovers ghost entities using a FileSource.
// It does NOT walk the filesystem - all discovery uses the provided file list.
type Scanner struct {
	source   FileSource
	registry *modules.Registry
	opts     ScanOptions
}

// NewScanner creates a ghost scanner.
// The FileSource must be provided - scanner does not create its own.
// Registry is optional and used for module ownership resolution.
func NewScanner(source FileSource, registry *modules.Registry, opts ScanOptions) *Scanner {
	if opts.Alias == "" {
		opts.Alias = "ghost"
	}
	return &Scanner{
		source:   source,
		registry: registry,
		opts:     opts,
	}
}

// Scan discovers all ghost entities from the FileSource.
// Returns both ghost files and ghost directories (derived from file paths).
func (s *Scanner) Scan() ([]Ghost, error) {
	// Get all tracked files from source (no filesystem walk)
	files, err := s.source.TrackedFiles()
	if err != nil {
		return nil, err
	}

	var ghosts []Ghost
	seenDirs := make(map[string]bool) // Track ghost directories to avoid duplicates

	for _, filePath := range files {
		// Check if file itself is a ghost
		fileName := path.Base(filePath)
		if s.isGhostName(fileName) {
			ghost := s.createGhost(filePath, fileName, GhostTypeFile)
			ghosts = append(ghosts, ghost)
		}

		// Check path components for ghost directories
		// e.g., "foo/ghost-feature/bar.go" -> ghost-feature is a ghost directory
		dir := path.Dir(filePath)
		for dir != "." && dir != "" {
			dirName := path.Base(dir)
			if s.isGhostName(dirName) && !seenDirs[dir] {
				seenDirs[dir] = true
				ghost := s.createGhost(dir, dirName, GhostTypeDirectory)
				ghosts = append(ghosts, ghost)
			}
			dir = path.Dir(dir)
		}
	}

	return ghosts, nil
}

// isGhostName checks if a name matches ghost patterns.
func (s *Scanner) isGhostName(name string) bool {
	alias := s.opts.Alias
	return strings.HasPrefix(name, alias+"-") || // ghost-feature.go
		strings.HasPrefix(name, alias+".") || // ghost.config
		name == alias // exact match
}

// createGhost builds a Ghost struct with module resolution.
func (s *Scanner) createGhost(filePath, name string, ghostType GhostType) Ghost {
	ghost := Ghost{
		Path:      filePath,
		Name:      name,
		Type:      ghostType,
		GhostName: s.extractGhostName(name),
	}

	// Resolve module ownership
	if s.registry != nil {
		matchingModules := s.registry.FindModulesForFile(filePath)
		if len(matchingModules) > 0 {
			// Use first non-repository-root module
			for _, m := range matchingModules {
				if m.Moniker != "repository" {
					ghost.Module = m.Moniker
					break
				}
			}
			// Fallback to repository if that's all we have
			if ghost.Module == "" && len(matchingModules) > 0 {
				ghost.Module = matchingModules[0].Moniker
			}
		}
	}

	return ghost
}

// extractGhostName removes the ghost prefix from a name.
func (s *Scanner) extractGhostName(name string) string {
	alias := s.opts.Alias
	if strings.HasPrefix(name, alias+"-") {
		return strings.TrimPrefix(name, alias+"-")
	}
	if strings.HasPrefix(name, alias+".") {
		return strings.TrimPrefix(name, alias+".")
	}
	if name == alias {
		return ""
	}
	return name
}
