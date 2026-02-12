package pdfscreenshots

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// scanForPDFs finds all PDF files in the build directory.
func scanForPDFs(buildDir, moduleFilter string) ([]PDFInfo, error) {
	var pdfs []PDFInfo

	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		return pdfs, nil
	}

	err := filepath.WalkDir(buildDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".pdf") {
			return nil
		}

		relPath, relErr := filepath.Rel(buildDir, path)
		if relErr != nil {
			relPath = path // Fallback to absolute path
		}
		parts := strings.Split(filepath.ToSlash(relPath), "/")
		module := ""
		if len(parts) > 0 {
			module = parts[0]
		}

		// Apply module filter if specified
		if moduleFilter != "" && module != moduleFilter {
			return nil
		}

		// Extract book name from filename (without .pdf extension)
		baseName := filepath.Base(path)
		bookName := strings.TrimSuffix(baseName, filepath.Ext(baseName))

		pdfs = append(pdfs, PDFInfo{
			Path:     path,
			RelPath:  relPath,
			Module:   module,
			BookName: bookName,
		})

		return nil
	})

	return pdfs, err
}

// hashFile returns the SHA256 hash of a file (first 12 characters).
func hashFile(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return "unknown"
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "unknown"
	}

	return fmt.Sprintf("%x", h.Sum(nil))[:12]
}

// cacheValidWithHash checks if cache dir exists with correct hash marker.
func cacheValidWithHash(cacheDir, hash string) bool {
	// Check for hash marker file: <hash>.cache
	hashMarker := filepath.Join(cacheDir, hash+".cache")
	if _, err := os.Stat(hashMarker); err != nil {
		return false
	}

	// Also verify at least one PNG exists
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".png") {
			return true
		}
	}
	return false
}

// countPages counts the number of PNG files in a directory.
func countPages(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}

	count := 0
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".png") {
			count++
		}
	}
	return count
}
