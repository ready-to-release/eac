package books

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var manifestLog = logging.C()

// StagingManifest tracks the state and configuration of a staging directory
type StagingManifest struct {
	Version       string       `json:"version"`
	BookName      string       `json:"book_name"`
	ConfigHash    string       `json:"config_hash"`
	CreatedAt     time.Time    `json:"created_at"`
	LastValidated time.Time    `json:"last_validated"`
	PDFMode       bool         `json:"pdf_mode"`
	Sources       []SourceInfo `json:"sources"`
	ExpectedFiles []string     `json:"expected_files"`
	BuildStatus   string       `json:"build_status"` // "success", "failed", "incomplete"
	BuildDuration int64        `json:"build_duration_ms"`
}

// SourceInfo captures metadata about a book source
type SourceInfo struct {
	Type      string `json:"type"`       // "copy", "command", "inline"
	From      string `json:"from,omitempty"`
	To        string `json:"to,omitempty"`
	Command   string `json:"command,omitempty"`
	Target    string `json:"target,omitempty"`
	FileCount int    `json:"file_count,omitempty"`
	ExitCode  int    `json:"exit_code,omitempty"`
}

// calculateBookConfigHash generates a hash of the book configuration
// Any change to sources, commands, or PDF mode will change this hash
func calculateBookConfigHash(book *config.Book, pdfMode bool) string {
	h := sha256.New()

	// Hash book metadata
	fmt.Fprintf(h, "name:%s\n", book.Name)
	fmt.Fprintf(h, "description:%s\n", book.Description)
	fmt.Fprintf(h, "pdf_mode:%v\n", pdfMode)
	fmt.Fprintf(h, "site_url:%s\n", book.SiteURL)

	// Hash all sources in order
	for i, src := range book.Sources {
		fmt.Fprintf(h, "source[%d]:type=%s\n", i, src.Type)
		fmt.Fprintf(h, "source[%d]:from=%s\n", i, src.From)
		fmt.Fprintf(h, "source[%d]:to=%s\n", i, src.Target)
		fmt.Fprintf(h, "source[%d]:command=%s\n", i, src.Command)
		fmt.Fprintf(h, "source[%d]:order=%d\n", i, src.Order)

		// Hash frontmatter keys (sorted for deterministic hashing)
		if len(src.Frontmatter) > 0 {
			// Sort keys to ensure deterministic hash
			keys := make([]string, 0, len(src.Frontmatter))
			for k := range src.Frontmatter {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				fmt.Fprintf(h, "source[%d]:fm[%s]=%v\n", i, k, src.Frontmatter[k])
			}
		}
	}

	return fmt.Sprintf("sha256:%x", h.Sum(nil)[:16]) // First 16 bytes = 32 hex chars
}

// loadManifest reads and parses the staging manifest file
func loadManifest(stagingDir string) (*StagingManifest, error) {
	manifestPath := filepath.Join(stagingDir, ".staging-manifest.json")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	var manifest StagingManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("invalid manifest format: %w", err)
	}

	if manifest.Version != "1.0" {
		return nil, fmt.Errorf("unsupported manifest version: %s", manifest.Version)
	}

	return &manifest, nil
}

// writeManifest creates a manifest file for the staging directory
func writeManifest(stagingDir string, book *config.Book, pdfMode bool, expectedFiles []string, buildDuration time.Duration) error {
	manifest := StagingManifest{
		Version:       "1.0",
		BookName:      book.Name,
		ConfigHash:    calculateBookConfigHash(book, pdfMode),
		CreatedAt:     time.Now(),
		LastValidated: time.Now(),
		PDFMode:       pdfMode,
		ExpectedFiles: expectedFiles,
		BuildStatus:   "success",
		BuildDuration: buildDuration.Milliseconds(),
	}

	// Capture source metadata
	for _, src := range book.Sources {
		info := SourceInfo{
			Type:    src.Type,
			From:    src.From,
			To:      src.Target,
			Command: src.Command,
			Target:  src.Target,
		}
		manifest.Sources = append(manifest.Sources, info)
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	manifestPath := filepath.Join(stagingDir, ".staging-manifest.json")
	return os.WriteFile(manifestPath, data, 0644)
}

// validateStagingIntegrity checks if the staging directory matches the manifest
func validateStagingIntegrity(stagingDir string, manifest *StagingManifest) error {
	// Collect all actual files in staging (excluding metadata)
	actualFiles := make(map[string]bool)
	err := filepath.WalkDir(stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(stagingDir, path)
		if err != nil {
			return err
		}

		// Normalize to forward slashes for comparison
		relPath = filepath.ToSlash(relPath)

		// Skip metadata files
		if relPath == ".staging-manifest.json" || relPath == ".staging-lock" ||
		   relPath == ".gitignore" || strings.HasSuffix(relPath, ".nav.yml") {
			return nil
		}

		actualFiles[relPath] = true
		return nil
	})
	if err != nil {
		return fmt.Errorf("walking staging directory: %w", err)
	}

	// Check all expected files exist
	expectedSet := make(map[string]bool)
	for _, file := range manifest.ExpectedFiles {
		expectedSet[file] = true
		if !actualFiles[file] {
			return fmt.Errorf("expected file missing: %s", file)
		}
	}

	// Check for unexpected files (potential contamination)
	for file := range actualFiles {
		if !expectedSet[file] {
			// Allow assets directory and generated nav files
			if strings.HasPrefix(file, "assets/") || strings.HasSuffix(file, ".nav.yml") {
				continue
			}
			return fmt.Errorf("unexpected file found: %s (possible stale content)", file)
		}
	}

	return nil
}

// CheckStagingCache validates existing staging or determines if rebuild is needed
// Returns: (stagingDir, useCache, error)
func CheckStagingCache(book *config.Book, workspaceRoot string, pdfMode bool) (string, bool, error) {
	stagingDir := filepath.Join(workspaceRoot, "out", "staging", book.Name)
	currentHash := calculateBookConfigHash(book, pdfMode)

	// Try to load existing manifest
	manifest, err := loadManifest(stagingDir)
	if err != nil {
		if os.IsNotExist(err) {
			manifestLog.Infof("📝 No cache manifest for %s - will rebuild", book.Name)
			return stagingDir, false, nil
		}
		manifestLog.Warnf("⚠️  Invalid manifest for %s: %v - will rebuild", book.Name, err)
		return stagingDir, false, nil
	}

	// Check if configuration changed
	if manifest.ConfigHash != currentHash {
		manifestLog.Infof("📝 Config changed for %s (was: %s, now: %s) - invalidating cache",
			book.Name, manifest.ConfigHash[:12], currentHash[:12])
		return stagingDir, false, nil
	}

	// Check build status
	if manifest.BuildStatus != "success" {
		manifestLog.Warnf("⚠️  Previous build failed for %s - will rebuild", book.Name)
		return stagingDir, false, nil
	}

	// Validate integrity
	if err := validateStagingIntegrity(stagingDir, manifest); err != nil {
		manifestLog.Warnf("⚠️  Staging corrupted for %s: %v - will rebuild", book.Name, err)
		return stagingDir, false, nil
	}

	// Cache is valid!
	manifestLog.Infof("✅ Using cached staging for %s (built: %s)",
		book.Name, manifest.CreatedAt.Format("2006-01-02 15:04:05"))
	return stagingDir, true, nil
}

// collectStagingFiles returns a list of all files in staging (for manifest)
func collectStagingFiles(stagingDir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(stagingDir, path)
		if err != nil {
			return err
		}

		// Normalize to forward slashes
		relPath = filepath.ToSlash(relPath)

		// Skip metadata files
		if relPath == ".staging-manifest.json" || relPath == ".staging-lock" ||
		   relPath == ".gitignore" || strings.HasSuffix(relPath, ".nav.yml") {
			return nil
		}

		files = append(files, relPath)
		return nil
	})

	return files, err
}
