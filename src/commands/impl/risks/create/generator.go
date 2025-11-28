package create

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// generateControl generates a risk control specification using specs create
func generateControl(config *Config, risk RiskData) error {
	// Determine output path
	outputPath := filepath.Join(config.OutputDir, risk.Domain, fmt.Sprintf("%s.feature", risk.ControlName))

	// Check if file exists
	fullPath := outputPath
	if !filepath.IsAbs(fullPath) {
		fullPath = filepath.Join(config.WorkspaceRoot, outputPath)
	}

	fileExists := false
	if _, err := os.Stat(fullPath); err == nil {
		fileExists = true
	}

	// If file exists and force is not set, skip
	if fileExists && !config.Force {
		fmt.Printf("  Skipped: %s (already exists, use --force to overwrite)\n", outputPath)
		return nil
	}

	// If file exists and force is set, check for orphaned tags
	if fileExists && config.Force {
		// Generate new control content first (temporarily)
		newContent, err := generateControlContent(config, risk)
		if err != nil {
			return fmt.Errorf("failed to generate control content: %w", err)
		}

		// Analyze tag orphans
		specsDir := filepath.Join(config.WorkspaceRoot, "specs")
		analysis, err := analyzeTagOrphans(fullPath, newContent, specsDir)
		if err != nil {
			return fmt.Errorf("failed to analyze tags: %w", err)
		}

		// Check if orphaned tags would be created
		if len(analysis.OrphanedTags) > 0 {
			if !config.AllowOrphans {
				// Prevent overwrite
				return fmt.Errorf(formatOrphanedTagsError(analysis, fullPath))
			} else {
				// Warn but allow
				fmt.Fprintf(os.Stderr, "⚠️  Warning: Overwriting with orphaned tags:\n")
				for _, tag := range analysis.OrphanedTags {
					fmt.Fprintf(os.Stderr, "  - %s\n", tag)
				}
				fmt.Fprintf(os.Stderr, "\n")
			}
		}
	}

	// Create control using specs create
	if err := createControlWithSpecsCreate(config, risk, outputPath); err != nil {
		return fmt.Errorf("failed to create control: %w", err)
	}

	if fileExists {
		fmt.Printf("  Overwritten: %s\n", outputPath)
	} else {
		fmt.Printf("  Created: %s\n", outputPath)
	}

	return nil
}

// buildControlDescription builds a description for specs create
func buildControlDescription(risk RiskData) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Risk Control for %s\n\n", risk.RiskID))
	builder.WriteString(fmt.Sprintf("Description: %s\n\n", risk.Description))
	builder.WriteString(fmt.Sprintf("Severity: %s\n", risk.Severity))
	builder.WriteString(fmt.Sprintf("Impact: %s\n\n", risk.Impact))

	if len(risk.AffectedFiles) > 0 {
		builder.WriteString(fmt.Sprintf("Affected Files: %s\n", strings.Join(risk.AffectedFiles, ", ")))
	}

	if len(risk.RelatedSpecs) > 0 {
		builder.WriteString(fmt.Sprintf("Related Specs: %s\n", strings.Join(risk.RelatedSpecs, ", ")))
	}

	return builder.String()
}

// createControlWithSpecsCreate calls specs create to generate the control
func createControlWithSpecsCreate(config *Config, risk RiskData, outputPath string) error {
	// Ensure output directory exists
	fullPath := outputPath
	if !filepath.IsAbs(fullPath) {
		fullPath = filepath.Join(config.WorkspaceRoot, outputPath)
	}

	outputDir := filepath.Dir(fullPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// For simplicity, we'll generate a basic Gherkin file directly
	// In a real implementation, you'd call specs create via command
	content, err := generateControlContent(config, risk)
	if err != nil {
		return fmt.Errorf("failed to generate control content: %w", err)
	}

	// Write the file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write control file: %w", err)
	}

	return nil
}

// generateControlContent generates the Gherkin content for a risk control
func generateControlContent(config *Config, risk RiskData) (string, error) {
	// Generate tag
	tagBase := strings.ReplaceAll(risk.ControlName, "-", "-")
	tag := fmt.Sprintf("@risk-control:%s", tagBase)

	// Build Gherkin content
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("%s\n", tag))
	builder.WriteString(fmt.Sprintf("Feature: Risk Control - %s\n\n", risk.Description))
	builder.WriteString(fmt.Sprintf("  Risk ID: %s\n", risk.RiskID))
	builder.WriteString(fmt.Sprintf("  Severity: %s\n", risk.Severity))
	builder.WriteString(fmt.Sprintf("  Impact: %s\n\n", risk.Impact))

	builder.WriteString("  As a security officer\n")
	builder.WriteString(fmt.Sprintf("  I want to implement controls for %s\n", risk.Description))
	builder.WriteString("  So that the identified risk is mitigated\n\n")

	builder.WriteString("  Rule: Control is implemented\n\n")

	builder.WriteString(fmt.Sprintf("    %s-01\n", tag))
	builder.WriteString("    Scenario: Risk control verification\n")
	builder.WriteString(fmt.Sprintf("      Given the risk \"%s\" has been identified\n", risk.RiskID))
	builder.WriteString("      When the control measures are implemented\n")
	builder.WriteString("      Then the risk is mitigated\n")

	if len(risk.AffectedFiles) > 0 {
		builder.WriteString(fmt.Sprintf("      And affected files are protected: %s\n", strings.Join(risk.AffectedFiles, ", ")))
	}

	return builder.String(), nil
}

// callSpecsCreate calls the specs create command (for reference, not currently used)
func callSpecsCreate(workspaceRoot string, description string, outputPath string) error {
	cmd := exec.Command("go", "run", "./src/commands", "specs", "create", description, "--output", outputPath)
	cmd.Dir = workspaceRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
