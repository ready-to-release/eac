package reports

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/src/repository"
)

// FilesModulesData contains structured file-module relationship data
type FilesModulesData struct {
	FileMap   map[string][]string // Key: file path, Value: list of module monikers
	ModuleMap map[string][]string // Key: module moniker, Value: list of file paths
}

// FilesModulesReport contains statistics and data about file-module relationships
type FilesModulesReport struct {
	TotalFiles      int
	SingleOwner     int
	MultiOwner      int
	Orphan          int
	FilesByModule   map[string][]string
	MultiOwnership  []repository.RepositoryFileWithModule
	OrphanFiles     []repository.RepositoryFileWithModule
	AllFiles        []repository.RepositoryFileWithModule
}

// GetFilesModulesReport generates a comprehensive report of file-module ownership
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
//
// Example:
//
//	report, err := reports.GetFilesModulesReport(true, false, false, "", "0.1.0")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Total files: %d\n", report.TotalFiles)
//	fmt.Printf("Orphan files: %d\n", report.Orphan)
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

// FormatReport returns a formatted string representation of the report
func (r *FilesModulesReport) FormatReport() string {
	var sb strings.Builder

	sb.WriteString("=== Repository Files with Module Ownership ===\n\n")
	sb.WriteString(fmt.Sprintf("✅ Found %d tracked files\n\n", r.TotalFiles))

	// Statistics
	sb.WriteString("=== Statistics ===\n")
	sb.WriteString(fmt.Sprintf("Single owner:     %d files\n", r.SingleOwner))
	sb.WriteString(fmt.Sprintf("Multiple owners:  %d files\n", r.MultiOwner))
	sb.WriteString(fmt.Sprintf("Orphan (no owner): %d files\n", r.Orphan))
	sb.WriteString("\n")

	// Modules Affected
	if len(r.FilesByModule) > 0 {
		sb.WriteString("=== Modules Affected ===\n")
		moduleNum := 1
		for module, paths := range r.FilesByModule {
			sb.WriteString(fmt.Sprintf("%d. %s (%d files)\n", moduleNum, module, len(paths)))
			moduleNum++
		}
		sb.WriteString("\n")
	}

	// Files by module
	sb.WriteString("=== Files by Module ===\n")
	for module, paths := range r.FilesByModule {
		sb.WriteString(fmt.Sprintf("\n📦 %s: %d files\n", module, len(paths)))

		// Show all files
		for _, path := range paths {
			sb.WriteString(fmt.Sprintf("   - %s\n", path))
		}
	}

	// Multi-ownership
	if len(r.MultiOwnership) > 0 {
		sb.WriteString("\n=== Multi-Ownership Files ===\n")
		sb.WriteString(fmt.Sprintf("Found %d files owned by multiple modules:\n\n", len(r.MultiOwnership)))

		// Show as markdown table
		sb.WriteString("| File | Owner Count | Owners |\n")
		sb.WriteString("|------|-------------|--------|\n")
		for _, file := range r.MultiOwnership {
			ownerCount := len(file.Modules)
			owners := strings.Join(file.Modules, ", ")
			sb.WriteString(fmt.Sprintf("| %s | %d | %s |\n", file.Name, ownerCount, owners))
		}
	}

	// Orphan files
	if len(r.OrphanFiles) > 0 {
		sb.WriteString("\n=== Orphan Files (No Module Owner) ===\n")
		sb.WriteString(fmt.Sprintf("Found %d files without module ownership:\n\n", len(r.OrphanFiles)))

		// Show all orphan files
		for _, file := range r.OrphanFiles {
			sb.WriteString(fmt.Sprintf("❓ %s\n", file.Name))
		}
	}

	sb.WriteString("\n✅ Module ownership analysis complete!\n")

	return sb.String()
}

// GetModuleStats returns statistics for a specific module
func (r *FilesModulesReport) GetModuleStats(moniker string) (fileCount int, files []string) {
	files, exists := r.FilesByModule[moniker]
	if !exists {
		return 0, []string{}
	}
	return len(files), files
}

// HasMultiOwnership checks if a specific file has multiple owners
func (r *FilesModulesReport) HasMultiOwnership(filePath string) (bool, []string) {
	for _, file := range r.MultiOwnership {
		if file.Name == filePath {
			return true, file.Modules
		}
	}
	return false, nil
}

// IsOrphan checks if a specific file is an orphan (no module owner)
func (r *FilesModulesReport) IsOrphan(filePath string) bool {
	for _, file := range r.OrphanFiles {
		if file.Name == filePath {
			return true
		}
	}
	return false
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

// GetCoveragePercentage returns the percentage of files covered by at least one module
func (r *FilesModulesReport) GetCoveragePercentage() float64 {
	if r.TotalFiles == 0 {
		return 0
	}
	covered := r.TotalFiles - r.Orphan
	return (float64(covered) / float64(r.TotalFiles)) * 100
}

// PrintSummary prints a concise summary of the report
func (r *FilesModulesReport) PrintSummary() {
	fmt.Println("=== Module Coverage Summary ===")
	fmt.Printf("Total files:      %d\n", r.TotalFiles)
	fmt.Printf("Covered:          %d (%.1f%%)\n", r.TotalFiles-r.Orphan, r.GetCoveragePercentage())
	fmt.Printf("Orphan:           %d\n", r.Orphan)
	fmt.Printf("Multi-ownership:  %d\n", r.MultiOwner)
	fmt.Printf("Modules:          %d\n", len(r.FilesByModule))
}

// GetFilesModules returns structured file-module relationship data
//
// Parameters:
//   - trackedOnly: if true, only include files tracked by Git
//   - includeIgnored: if true, include files ignored by .gitignore
//   - stagedOnly: if true, only return files currently staged in Git index
//   - rootPath: repository root (if empty, will be detected automatically)
//   - version: module contract version (e.g., "0.1.0")
//
// Returns:
//   - FilesModulesData with FileMap (file -> modules) and ModuleMap (module -> files)
//   - Error if repository operations or module loading fails
//
// Example:
//
//	data, err := reports.GetFilesModules(true, false, false, "", "0.1.0")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	modules := data.FileMap["src/main.go"]  // Get modules owning this file
//	files := data.ModuleMap["src-cli"]      // Get files owned by this module
func GetFilesModules(trackedOnly bool, includeIgnored bool, stagedOnly bool, rootPath string, version string) (*FilesModulesData, error) {
	// Get all files with module ownership
	files, err := repository.GetRepositoryFilesWithModules(trackedOnly, includeIgnored, stagedOnly, rootPath, version)
	if err != nil {
		return nil, err
	}

	data := &FilesModulesData{
		FileMap:   make(map[string][]string),
		ModuleMap: make(map[string][]string),
	}

	// Build FileMap: file -> modules
	for _, file := range files {
		data.FileMap[file.Name] = file.Modules
	}

	// Build ModuleMap: module -> files
	for _, file := range files {
		for _, module := range file.Modules {
			data.ModuleMap[module] = append(data.ModuleMap[module], file.Name)
		}
	}

	return data, nil
}
