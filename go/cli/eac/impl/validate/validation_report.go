package validate

import "fmt"

// validationSection represents one category of issues in a validation report.
type validationSection struct {
	Icon        string   // e.g. "❌" or "⚠️ "
	Label       string   // e.g. "Circular Dependencies"
	Items       []string // bullet-point items
	Description string   // optional description line shown before items
	FixHint     string   // optional fix hint shown after items
}

// validationReport describes a complete validation report to be printed.
type validationReport struct {
	Title          string              // e.g. "Module Hierarchy Validation Report"
	Sections       []validationSection // categorized issue sections
	SuccessMessage string              // shown when all sections are empty, e.g. "All module hierarchy checks passed!"
}

// printValidationReport prints a validation report using the shared format.
//
// Output format:
//
//	=== <Title> ===
//
//	<icon> <Label> (<count>):
//	   <Description>
//	  - <item>
//
//	   <FixHint>
//
// If no sections have items, the success message is printed instead.
func printValidationReport(r validationReport) {
	log.Infof("=== %s ===", r.Title)
	log.Info("")

	hasIssues := false

	for _, section := range r.Sections {
		if len(section.Items) == 0 {
			continue
		}
		hasIssues = true

		log.Infof("%s %s (%d):", section.Icon, section.Label, len(section.Items))

		if section.Description != "" {
			log.Infof("   %s", section.Description)
		}

		for _, item := range section.Items {
			log.Infof("  %s", item)
		}
		log.Info("")

		if section.FixHint != "" {
			log.Infof("   %s", section.FixHint)
			log.Info("")
		}
	}

	if !hasIssues {
		log.Infof("✅ %s", r.SuccessMessage)
		log.Info("")
	}
}

// printValidationReportWithSummary prints a validation report that includes
// numeric summary lines before the issue sections.
//
// Output format:
//
//	=== <Title> ===
//
//	<summaryLines>
//
//	<sections or success message>
func printValidationReportWithSummary(r validationReport, summaryLines []string) {
	log.Infof("=== %s ===", r.Title)
	log.Info("")

	for _, line := range summaryLines {
		log.Info(line)
	}
	log.Info("")

	hasIssues := false

	for _, section := range r.Sections {
		if len(section.Items) == 0 {
			continue
		}
		hasIssues = true

		log.Infof("%s %s", section.Icon, section.Label)

		for _, item := range section.Items {
			log.Info(item)
		}
		log.Info("")

		if section.FixHint != "" {
			log.Info(section.FixHint)
			log.Info("")
		}
	}

	if !hasIssues {
		log.Infof("✅ %s", r.SuccessMessage)
		log.Info("")
	}
}

// formatBullet formats an item with the bullet prefix used throughout validation reports.
func formatBullet(item string) string {
	return fmt.Sprintf("• %s", item)
}

// formatBulletWithDetail formats a bullet item with an indented detail line.
func formatBulletWithDetail(item, detail string) string {
	return fmt.Sprintf("• %s\n    %s", item, detail)
}
