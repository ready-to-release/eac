// Package internal provides evidence file generation with integrity verification.
package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/config"
)

// WriteEvidence writes a security evidence file with SHA256 integrity hash.
//
// Output structure: out/scan/<module>/<scanner>/
//
// Parameters:
//   - workspaceRoot: Repository root directory
//   - module: Module moniker
//   - scanner: Scanner type (sbom, vuln, etc.)
//   - findings: Scanner-specific findings (will be marshaled to JSON)
//
// Returns:
//   - outputPath: Full path to created evidence file
//   - error: Any error during file creation
//
// The evidence file includes:
//   - Module and scanner identification
//   - RFC3339 timestamp for traceability
//   - SHA256 hash of findings for integrity verification
//   - Scanner-specific findings in JSON format
func WriteEvidence(workspaceRoot, module string, scanner ScannerType, findings interface{}) (string, error) {
	// Load config for path resolution
	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot})
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	// Build output directory path
	// Scan path already includes "out" prefix (e.g., "out/scan")
	// Join directly with workspaceRoot like other commands do (build, test, etc.)
	var outputDir string
	if filepath.IsAbs(cfg.Repository.Paths.Out.Scan) {
		// If Scan is absolute, use it directly
		outputDir = filepath.Join(cfg.Repository.Paths.Out.Scan, module)
	} else {
		// If Scan is relative, join with workspaceRoot
		outputDir = filepath.Join(workspaceRoot, cfg.Repository.Paths.Out.Scan, module)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate timestamp
	timestamp := GetTimestamp()

	// Marshal findings to JSON
	findingsJSON, err := json.Marshal(findings)
	if err != nil {
		return "", fmt.Errorf("failed to marshal findings: %w", err)
	}

	// Calculate SHA256 hash of findings for integrity verification
	hash := sha256.Sum256(findingsJSON)
	hashHex := hex.EncodeToString(hash[:])

	// Create evidence file structure
	evidence := EvidenceFile{
		Module:    module,
		Scanner:   string(scanner),
		Timestamp: timestamp,
		SHA256:    hashHex,
		Findings:  findingsJSON,
	}

	// Marshal evidence to JSON with indentation for human readability
	evidenceJSON, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal evidence: %w", err)
	}

	// Generate filename based on scanner type (e.g., vuln.json, sbom.json)
	filename := fmt.Sprintf("%s.json", scanner)
	outputPath := filepath.Join(outputDir, filename)

	// Write evidence file
	if err := os.WriteFile(outputPath, evidenceJSON, 0644); err != nil {
		return "", fmt.Errorf("failed to write evidence file: %w", err)
	}

	return outputPath, nil
}

// WriteErrorEvidence writes an evidence file when a scan fails.
// This maintains the audit trail even when scans encounter errors.
func WriteErrorEvidence(workspaceRoot, module string, scanner ScannerType, errorMsg string) (string, error) {
	errorFindings := map[string]interface{}{
		"error":  errorMsg,
		"status": "failed",
	}
	return WriteEvidence(workspaceRoot, module, scanner, errorFindings)
}

// ReadEvidence reads and parses an evidence file.
// Returns the evidence structure and any error encountered.
func ReadEvidence(evidencePath string) (*EvidenceFile, error) {
	data, err := os.ReadFile(evidencePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read evidence file: %w", err)
	}

	var evidence EvidenceFile
	if err := json.Unmarshal(data, &evidence); err != nil {
		return nil, fmt.Errorf("failed to parse evidence file: %w", err)
	}

	return &evidence, nil
}

// VerifyEvidence checks the SHA256 integrity of an evidence file.
// Returns true if the hash matches, false otherwise.
func VerifyEvidence(evidence *EvidenceFile) bool {
	// Recalculate hash of findings
	hash := sha256.Sum256(evidence.Findings)
	calculatedHash := hex.EncodeToString(hash[:])

	// Compare with stored hash
	return calculatedHash == evidence.SHA256
}
