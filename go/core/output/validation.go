package output

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// HashFile computes the SHA256 hash of a file and returns its size.
// The hash is returned with a "sha256:" prefix.
func HashFile(path string) (size int64, hash string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, "", fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	// Check if it's a directory
	info, err := f.Stat()
	if err != nil {
		return 0, "", fmt.Errorf("stat file: %w", err)
	}
	if info.IsDir() {
		return 0, "", fmt.Errorf("path is a directory, not a file")
	}

	hasher := sha256.New()
	n, err := io.Copy(hasher, f)
	if err != nil {
		return 0, "", fmt.Errorf("hash file: %w", err)
	}

	hashBytes := hasher.Sum(nil)
	hashStr := "sha256:" + hex.EncodeToString(hashBytes)

	return n, hashStr, nil
}

// ValidateArtifacts checks that all artifacts exist and their hashes match.
// baseDir is joined with each artifact's path to form the full path.
// If baseDir is empty, artifact paths are used as-is (assumed absolute or relative to cwd).
func ValidateArtifacts(baseDir string, artifacts []Artifact) ValidationResult {
	result := ValidationResult{
		Valid:            true,
		ArtifactsValid:   true,
		MissingArtifacts: []string{},
		CorruptArtifacts: []string{},
	}

	// Empty or nil artifact list is valid
	if len(artifacts) == 0 {
		return result
	}

	for _, artifact := range artifacts {
		// Compute full path
		fullPath := artifact.Path
		if baseDir != "" && !filepath.IsAbs(artifact.Path) {
			fullPath = filepath.Join(baseDir, artifact.Path)
		}

		// Check if file exists
		_, err := os.Stat(fullPath)
		if os.IsNotExist(err) {
			result.Valid = false
			result.ArtifactsValid = false
			result.MissingArtifacts = append(result.MissingArtifacts, artifact.Path)
			continue
		}
		if err != nil {
			result.Valid = false
			result.ArtifactsValid = false
			result.Error = err
			continue
		}

		// Compute hash and compare
		_, actualHash, err := HashFile(fullPath)
		if err != nil {
			result.Valid = false
			result.ArtifactsValid = false
			result.Error = err
			continue
		}

		if actualHash != artifact.SHA256 {
			result.Valid = false
			result.ArtifactsValid = false
			result.CorruptArtifacts = append(result.CorruptArtifacts, artifact.Path)
		}
	}

	return result
}
