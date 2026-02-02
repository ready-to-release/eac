package contracts

import "embed"

// Contract version constants.
const (
	// EACCoreVersion is the current version of core contracts.
	EACCoreVersion = "0.1.0"

	// EACDocsVersion is the current version of docs contracts.
	EACDocsVersion = "0.1.0"

	// R2RCLIVersion is the current version of r2r-cli contracts.
	R2RCLIVersion = "0.1.0"
)

// FS is the embedded filesystem containing all contract files.
// Use the helper functions (EACCorePath, R2RCLIPath, etc.) to construct
// paths for ReadFile operations.
//
//go:embed core/0.1.0/*.schema.json
//go:embed core/0.1.0/*.json
//go:embed core/0.1.0/defaults/*.yml
//go:embed docs/0.1.0/*.schema.json
//go:embed r2r-cli/0.1.0/*
var FS embed.FS

// EACCorePath returns the full path for an core contract file.
// Example: EACCorePath("repository.schema.json") returns "core/0.1.0/repository.schema.json".
func EACCorePath(filename string) string {
	return "core/" + EACCoreVersion + "/" + filename
}

// EACCoreDefaultPath returns the full path for an core default file.
// Example: EACCoreDefaultPath("repository.yml") returns "core/0.1.0/defaults/repository.yml".
func EACCoreDefaultPath(filename string) string {
	return "core/" + EACCoreVersion + "/defaults/" + filename
}

// EACDocsPath returns the full path for an docs contract file.
// Example: EACDocsPath("manifest.schema.json") returns "docs/0.1.0/manifest.schema.json".
func EACDocsPath(filename string) string {
	return "docs/" + EACDocsVersion + "/" + filename
}

// R2RCLIPath returns the full path for an r2r-cli contract file.
// Example: R2RCLIPath("command.ebnf") returns "r2r-cli/0.1.0/command.ebnf".
func R2RCLIPath(filename string) string {
	return "r2r-cli/" + R2RCLIVersion + "/" + filename
}
