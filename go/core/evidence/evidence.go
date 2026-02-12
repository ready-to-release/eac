package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/core/config"
)

// writeFile is the shared internal writer used by both module-level and component-level
// evidence functions. It marshals findings, computes a SHA256 integrity hash, and writes
// the resulting evidence JSON to outputDir/<scanner>.json.
func writeFile(outputDir, module string, scanner ScannerType, findings interface{}) (string, error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	timestamp := GetTimestamp()

	findingsJSON, err := json.Marshal(findings)
	if err != nil {
		return "", fmt.Errorf("failed to marshal findings: %w", err)
	}

	hash := sha256.Sum256(findingsJSON)
	hashHex := hex.EncodeToString(hash[:])

	evidence := File{
		Module:    module,
		Scanner:   string(scanner),
		Timestamp: timestamp,
		SHA256:    hashHex,
		Findings:  findingsJSON,
	}

	evidenceJSON, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal evidence: %w", err)
	}

	filename := fmt.Sprintf("%s.json", scanner)
	outputPath := filepath.Join(outputDir, filename)

	if err := os.WriteFile(outputPath, evidenceJSON, 0o644); err != nil {
		return "", fmt.Errorf("failed to write evidence file: %w", err)
	}

	return outputPath, nil
}

// resolveOutputDir loads the repository config and returns the absolute scan output
// directory built from the given path segments (e.g. module, or module + component).
func resolveOutputDir(workspaceRoot string, pathSegments ...string) (string, error) {
	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot})
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	parts := []string{cfg.Repository.Paths.Out.Scan}
	parts = append(parts, pathSegments...)

	var outputDir string
	if filepath.IsAbs(cfg.Repository.Paths.Out.Scan) {
		outputDir = filepath.Join(parts...)
	} else {
		outputDir = filepath.Join(append([]string{workspaceRoot}, parts...)...)
	}

	return outputDir, nil
}

// writeEvidence is the shared internal writer that resolves the output directory
// from the given path segments, then delegates to writeFile. Both module-level and
// component-level evidence functions use this single code path.
func writeEvidence(workspaceRoot, module string, scanner ScannerType, findings interface{}, pathSegments ...string) (string, error) {
	outputDir, err := resolveOutputDir(workspaceRoot, pathSegments...)
	if err != nil {
		return "", err
	}
	return writeFile(outputDir, module, scanner, findings)
}

// writeErrorEvidence is the shared internal writer for error evidence. It constructs
// the standard error findings payload and delegates to writeEvidence.
func writeErrorEvidence(workspaceRoot, module string, scanner ScannerType, errorMsg string, pathSegments ...string) (string, error) {
	errorFindings := map[string]interface{}{
		"error":  errorMsg,
		"status": "failed",
	}
	return writeEvidence(workspaceRoot, module, scanner, errorFindings, pathSegments...)
}

// WriteEvidence writes a security evidence file with SHA256 integrity hash.
//
// Output structure: out/scan/<module>/<scanner>.json
func WriteEvidence(workspaceRoot, module string, scanner ScannerType, findings interface{}) (string, error) {
	return writeEvidence(workspaceRoot, module, scanner, findings, module)
}

// WriteErrorEvidence writes an evidence file when a scan fails.
// This maintains the audit trail even when scans encounter errors.
func WriteErrorEvidence(workspaceRoot, module string, scanner ScannerType, errorMsg string) (string, error) {
	return writeErrorEvidence(workspaceRoot, module, scanner, errorMsg, module)
}

// WriteComponentEvidence writes a security evidence file for a specific component.
//
// Output structure: out/scan/<module>/<component>/<scanner>.json
func WriteComponentEvidence(workspaceRoot, module, component string, scanner ScannerType, findings interface{}) (string, error) {
	return writeEvidence(workspaceRoot, module, scanner, findings, module, component)
}

// WriteComponentErrorEvidence writes an evidence file when a component scan fails.
// This maintains the audit trail even when scans encounter errors.
func WriteComponentErrorEvidence(workspaceRoot, module, component string, scanner ScannerType, errorMsg string) (string, error) {
	return writeErrorEvidence(workspaceRoot, module, scanner, errorMsg, module, component)
}

// ReadEvidence reads and parses an evidence file.
func ReadEvidence(evidencePath string) (*File, error) {
	data, err := os.ReadFile(evidencePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read evidence file: %w", err)
	}

	var ev File
	if err := json.Unmarshal(data, &ev); err != nil {
		return nil, fmt.Errorf("failed to parse evidence file: %w", err)
	}

	return &ev, nil
}

// VerifyEvidence checks the SHA256 integrity of an evidence file.
// Returns true if the hash matches, false otherwise.
func VerifyEvidence(ev *File) bool {
	hash := sha256.Sum256(ev.Findings)
	calculatedHash := hex.EncodeToString(hash[:])
	return calculatedHash == ev.SHA256
}
