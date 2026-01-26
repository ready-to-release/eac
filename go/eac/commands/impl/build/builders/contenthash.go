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
)

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
// Uses out/cache/build-state/ which persists across build retries.
func getBuildStateCacheDir(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, "out", "cache", "build-state")
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
	// Compute current staging hash
	currentHash, err := ComputeStagingHash(stagingDir)
	if err != nil {
		return false, "failed to compute staging hash"
	}

	// Load previous build state from cache (out/cache/build-state/)
	state, err := LoadBookBuildState(bookName, theme, workspaceRoot)
	if err != nil {
		return false, "no previous build state"
	}

	// Check if hash matches
	if state.ContentHash != currentHash {
		return false, "staging content changed"
	}

	// Check if output file still exists in build output
	expectedPDF := filepath.Join(outputDir, bookName+"-"+theme+".pdf")
	if _, err := os.Stat(expectedPDF); os.IsNotExist(err) {
		return false, "output PDF missing"
	}

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
