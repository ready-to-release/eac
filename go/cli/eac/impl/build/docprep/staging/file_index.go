package staging

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// FileIndex provides a pre-built index of files in the staging directory.
// Built once at preprocessing start, shared across all steps to avoid repeated WalkDir calls.
type FileIndex struct {
	StagingDir    string                // Root staging directory
	MarkdownFiles []string             // All .md files (absolute paths)
	AllFiles      []string             // All files (absolute paths)
	FileInfo      map[string]os.FileInfo // Path -> FileInfo for quick lookups
	mu            sync.RWMutex         // Protects the index during updates
}

// NewFileIndex creates a new FileIndex by scanning the staging directory.
func NewFileIndex(stagingDir string) (*FileIndex, error) {
	idx := &FileIndex{
		StagingDir: stagingDir,
		FileInfo:   make(map[string]os.FileInfo),
	}

	err := filepath.WalkDir(stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil // Skip if we can't get info
		}

		idx.AllFiles = append(idx.AllFiles, path)
		idx.FileInfo[path] = info

		if strings.HasSuffix(strings.ToLower(path), ".md") {
			idx.MarkdownFiles = append(idx.MarkdownFiles, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return idx, nil
}

// Refresh rebuilds the file index after files have been modified.
func (idx *FileIndex) Refresh() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.MarkdownFiles = idx.MarkdownFiles[:0]
	idx.AllFiles = idx.AllFiles[:0]
	idx.FileInfo = make(map[string]os.FileInfo)

	err := filepath.WalkDir(idx.StagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		idx.AllFiles = append(idx.AllFiles, path)
		idx.FileInfo[path] = info

		if strings.HasSuffix(strings.ToLower(path), ".md") {
			idx.MarkdownFiles = append(idx.MarkdownFiles, path)
		}

		return nil
	})

	return err
}

// GetMarkdownFiles returns a copy of the markdown files slice.
func (idx *FileIndex) GetMarkdownFiles() []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	result := make([]string, len(idx.MarkdownFiles))
	copy(result, idx.MarkdownFiles)
	return result
}

// GetAllFiles returns a copy of all files slice.
func (idx *FileIndex) GetAllFiles() []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	result := make([]string, len(idx.AllFiles))
	copy(result, idx.AllFiles)
	return result
}

// FileCount returns the total number of files indexed.
func (idx *FileIndex) FileCount() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.AllFiles)
}

// MarkdownCount returns the number of markdown files indexed.
func (idx *FileIndex) MarkdownCount() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.MarkdownFiles)
}

// RemoveFile removes a file from the index (called after deletion).
func (idx *FileIndex) RemoveFile(path string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	delete(idx.FileInfo, path)

	for i, f := range idx.AllFiles {
		if f == path {
			idx.AllFiles = append(idx.AllFiles[:i], idx.AllFiles[i+1:]...)
			break
		}
	}

	if strings.HasSuffix(strings.ToLower(path), ".md") {
		for i, f := range idx.MarkdownFiles {
			if f == path {
				idx.MarkdownFiles = append(idx.MarkdownFiles[:i], idx.MarkdownFiles[i+1:]...)
				break
			}
		}
	}
}
