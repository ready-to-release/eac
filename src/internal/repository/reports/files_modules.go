package reports

import (
	"github.com/ready-to-release/eac/src/internal/repository"
)

// FilesModulesReport contains statistics about file-module relationships
type FilesModulesReport struct {
	TotalFiles     int
	SingleOwner    int
	MultiOwner     int
	Orphan         int
	FilesByModule  map[string][]string
	MultiOwnership []repository.RepositoryFileWithModule
	OrphanFiles    []repository.RepositoryFileWithModule
	AllFiles       []repository.RepositoryFileWithModule
}

// GetFilesModulesReport generates a report of file-module ownership
//
// Parameters:
//   - trackedOnly: if true, only include files tracked by Git
//   - includeIgnored: if true, include files ignored by .gitignore
//   - stagedOnly: if true, only return files currently staged in Git index
//   - rootPath: repository root (if empty, will be detected automatically)
//   - version: module contract version (e.g., "0.1.0")
//
// Returns:
//   - FilesModulesReport containing all statistics and data
//   - Error if repository operations or module loading fails
func GetFilesModulesReport(trackedOnly bool, includeIgnored bool, stagedOnly bool, rootPath string, version string) (*FilesModulesReport, error) {
	// Get all files with module ownership
	files, err := repository.GetRepositoryFilesWithModules(trackedOnly, includeIgnored, stagedOnly, rootPath, version)
	if err != nil {
		return nil, err
	}

	// Calculate statistics
	report := &FilesModulesReport{
		TotalFiles: len(files),
		AllFiles:   files,
	}

	for _, file := range files {
		switch len(file.Modules) {
		case 0:
			report.Orphan++
		case 1:
			report.SingleOwner++
		default:
			report.MultiOwner++
		}
	}

	// Group by module
	report.FilesByModule = repository.GetFilesByModule(files)

	// Get multi-ownership and orphan files
	report.MultiOwnership = repository.GetMultiOwnershipFiles(files)
	report.OrphanFiles = repository.GetOrphanFiles(files)

	return report, nil
}

// GetCoveragePercentage returns the percentage of files covered by at least one module
func (r *FilesModulesReport) GetCoveragePercentage() float64 {
	if r.TotalFiles == 0 {
		return 0
	}
	covered := r.TotalFiles - r.Orphan
	return (float64(covered) / float64(r.TotalFiles)) * 100
}

// GetFileModules returns the modules that own a specific file
func (r *FilesModulesReport) GetFileModules(filePath string) []string {
	for _, file := range r.AllFiles {
		if file.Name == filePath {
			return file.Modules
		}
	}
	return nil
}

// GetModuleStats returns statistics for a specific module
func (r *FilesModulesReport) GetModuleStats(moniker string) (fileCount int, files []string) {
	files, exists := r.FilesByModule[moniker]
	if !exists {
		return 0, []string{}
	}
	return len(files), files
}
