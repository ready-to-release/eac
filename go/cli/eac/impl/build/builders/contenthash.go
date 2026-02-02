// contenthash.go - Content-based caching for incremental builds
package builders

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"

	// do not delete: log.Debugf is used to track cache usage
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
)

var log = logging.C()

// BookBuildState tracks the state of a book build for incremental builds.
type BookBuildState struct {
	BookName    string `json:"book_name"`
	ContentHash string `json:"content_hash"` // SHA256 of staging directory
	Theme       string `json:"theme"`        // dark, light, or all
	OutputPath  string `json:"output_path"`  // Path to generated PDF
}

// ComputeStagingHash computes a SHA256 hash of all files in the staging directory.
// This is used after preprocessing to detect if PDF generation can be skipped.
func ComputeStagingHash(stagingDir string) (string, error) {
	h := sha256.New()

	var files []string
	err := filepath.WalkDir(stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}
		if d.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return "", err
	}

	// Sort for deterministic order
	sort.Strings(files)

	for _, path := range files {
		// Hash relative path
		relPath, _ := filepath.Rel(stagingDir, path)
		h.Write([]byte(relPath))

		// Hash file content
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		_, _ = io.Copy(h, f)
		f.Close()
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// getBuildStateCacheDir returns the directory for build state cache files.
// Uses .cache/eac/build/state/ which persists across build retries.
func getBuildStateCacheDir(workspaceRoot string) string {
	return paths.BuildStateCachePath(workspaceRoot)
}

// LoadBookBuildState loads the build state for a book from the cache file.
func LoadBookBuildState(bookName, theme, workspaceRoot string) (*BookBuildState, error) {
	cacheDir := getBuildStateCacheDir(workspaceRoot)
	statePath := filepath.Join(cacheDir, bookName+"-"+theme+".json")
	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, err
	}

	var state BookBuildState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// SaveBookBuildState saves the build state for a book to the cache file.
func SaveBookBuildState(state *BookBuildState, workspaceRoot string) error {
	cacheDir := getBuildStateCacheDir(workspaceRoot)
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return err
	}

	statePath := filepath.Join(cacheDir, state.BookName+"-"+state.Theme+".json")
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statePath, data, 0o644)
}

// ShouldSkipPDFGeneration checks if PDF generation can be skipped after preprocessing.
// Called AFTER preprocessing completes, compares staging directory hash with previous build.
// Returns (canSkip, reason) where reason explains why the build is needed or skipped.
func ShouldSkipPDFGeneration(bookName, theme, stagingDir, workspaceRoot, outputDir string) (bool, string) {
	log.Debugf("cache: ShouldSkipPDFGeneration book=%s theme=%s staging=%s", bookName, theme, stagingDir)

	// Compute current staging hash
	currentHash, err := ComputeStagingHash(stagingDir)
	if err != nil {
		log.Debugf("cache: PDF skip check failed - cannot compute hash: %v", err)
		return false, "failed to compute staging hash"
	}
	log.Debugf("cache: PDF current staging hash=%s", currentHash[:8])

	// Load previous build state from cache (.cache/eac/build/state/)
	state, err := LoadBookBuildState(bookName, theme, workspaceRoot)
	if err != nil {
		log.Debugf("cache: PDF cache MISS - no previous state: %v", err)
		return false, "no previous build state"
	}
	log.Debugf("cache: PDF previous hash=%s", state.ContentHash[:8])

	// Check if hash matches
	if state.ContentHash != currentHash {
		log.Debugf("cache: PDF cache MISS - hash changed (prev=%s curr=%s)", state.ContentHash[:8], currentHash[:8])
		return false, "staging content changed"
	}

	// Check if output file still exists in build output
	expectedPDF := filepath.Join(outputDir, bookName+"-"+theme+".pdf")
	if _, err := os.Stat(expectedPDF); os.IsNotExist(err) {
		log.Debugf("cache: PDF cache MISS - output missing: %s", expectedPDF)
		return false, "output PDF missing"
	}

	log.Debugf("cache: PDF cache HIT - skipping build for %s-%s", bookName, theme)
	return true, "staging unchanged (hash: " + currentHash[:8] + "...)"
}

// RecordPDFBuildComplete records that a PDF build completed successfully.
// Called AFTER PDF generation, saves staging hash to cache.
func RecordPDFBuildComplete(bookName, theme, stagingDir, workspaceRoot, outputDir string) error {
	hash, err := ComputeStagingHash(stagingDir)
	if err != nil {
		return err
	}

	state := &BookBuildState{
		BookName:    bookName,
		ContentHash: hash,
		Theme:       theme,
		OutputPath:  filepath.Join(outputDir, bookName+"-"+theme+".pdf"),
	}

	return SaveBookBuildState(state, workspaceRoot)
}

// SiteBuildState tracks the state of a site build for incremental builds.
type SiteBuildState struct {
	ModuleMoniker string `json:"module_moniker"`
	ContentHash   string `json:"content_hash"` // SHA256 of staging directory
	OutputPath    string `json:"output_path"`  // Path to generated site
}

// LoadSiteBuildState loads the site build state from the cache file.
func LoadSiteBuildState(moduleMoniker, workspaceRoot string) (*SiteBuildState, error) {
	cacheDir := getBuildStateCacheDir(workspaceRoot)
	statePath := filepath.Join(cacheDir, "site-"+moduleMoniker+".json")
	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, err
	}

	var state SiteBuildState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// SaveSiteBuildState saves the site build state to the cache file.
func SaveSiteBuildState(state *SiteBuildState, workspaceRoot string) error {
	cacheDir := getBuildStateCacheDir(workspaceRoot)
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return err
	}

	statePath := filepath.Join(cacheDir, "site-"+state.ModuleMoniker+".json")
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statePath, data, 0o644)
}

// ShouldSkipSiteBuild checks if site generation can be skipped after preprocessing.
// Called AFTER preprocessing completes, compares staging directory hash with previous build.
// When reproducible is true (CI mode), always returns false to force rebuild.
// Returns (canSkip, reason) where reason explains why the build is needed or skipped.
func ShouldSkipSiteBuild(moduleMoniker, stagingDir, workspaceRoot, outputDir string, reproducible bool) (bool, string) {
	log.Debugf("cache: ShouldSkipSiteBuild module=%s staging=%s", moduleMoniker, stagingDir)

	// In reproducible mode (CI), always rebuild to ensure consistent output
	if reproducible {
		log.Debugf("cache: site cache SKIP - reproducible mode enabled")
		return false, "reproducible mode enabled"
	}

	// Compute current staging hash
	currentHash, err := ComputeStagingHash(stagingDir)
	if err != nil {
		log.Debugf("cache: site skip check failed - cannot compute hash: %v", err)
		return false, "failed to compute staging hash"
	}
	log.Debugf("cache: site current staging hash=%s", currentHash[:8])

	// Load previous build state from cache
	state, err := LoadSiteBuildState(moduleMoniker, workspaceRoot)
	if err != nil {
		log.Debugf("cache: site cache MISS - no previous state: %v", err)
		return false, "no previous build state"
	}
	log.Debugf("cache: site previous hash=%s", state.ContentHash[:8])

	// Check if hash matches
	if state.ContentHash != currentHash {
		log.Debugf("cache: site cache MISS - hash changed (prev=%s curr=%s)", state.ContentHash[:8], currentHash[:8])
		return false, "staging content changed"
	}

	// Check if output site directory still exists
	siteDir := filepath.Join(outputDir, "site")
	if _, err := os.Stat(siteDir); os.IsNotExist(err) {
		log.Debugf("cache: site cache MISS - output missing: %s", siteDir)
		return false, "output site missing"
	}

	// Check if site has content (index.html)
	indexPath := filepath.Join(siteDir, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		log.Debugf("cache: site cache MISS - index.html missing: %s", indexPath)
		return false, "output site incomplete"
	}

	log.Debugf("cache: site cache HIT - skipping build for %s", moduleMoniker)
	return true, "staging unchanged (hash: " + currentHash[:8] + "...)"
}

// RecordSiteBuildComplete records that a site build completed successfully.
// Called AFTER site generation, saves staging hash to cache.
func RecordSiteBuildComplete(moduleMoniker, stagingDir, workspaceRoot, outputDir string) error {
	hash, err := ComputeStagingHash(stagingDir)
	if err != nil {
		return err
	}

	state := &SiteBuildState{
		ModuleMoniker: moduleMoniker,
		ContentHash:   hash,
		OutputPath:    filepath.Join(outputDir, "site"),
	}

	return SaveSiteBuildState(state, workspaceRoot)
}
