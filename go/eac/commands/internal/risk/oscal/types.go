// Package oscal provides OSCAL 1.1.2 data structures for risk assessment.
//
// OSCAL (Open Security Controls Assessment Language) is a NIST standard
// for expressing security controls and assessment results in machine-readable formats.
package oscal

import (
	"time"
)

// Profile represents an OSCAL profile document (control selection/tailoring).
// Profiles select and tailor controls from catalogs for specific use cases.
//
// OSCAL 1.1.2 Reference: https://pages.nist.gov/OSCAL/reference/latest/profile/json-outline/
type Profile struct {
	Profile ProfileContent `json:"profile"`
}

// ProfileContent is the inner structure of an OSCAL profile.
type ProfileContent struct {
	UUID     string   `json:"uuid"`
	Metadata Metadata `json:"metadata"`
	Imports  []Import `json:"imports"`
	Modify   *Modify  `json:"modify,omitempty"`
}

// Import defines catalog imports with control selection.
type Import struct {
	Href            string           `json:"href"`
	IncludeControls []ControlInclude `json:"include-controls,omitempty"`
	ExcludeControls []ControlExclude `json:"exclude-controls,omitempty"`
}

// ControlInclude specifies controls to include from a catalog.
type ControlInclude struct {
	WithID string `json:"with-id"`
}

// ControlExclude specifies controls to exclude from a catalog.
type ControlExclude struct {
	WithID string `json:"with-id"`
}

// Modify defines modifications to imported controls (parameter settings, etc.).
type Modify struct {
	SetParameters []SetParameter `json:"set-parameters,omitempty"`
	Alters        []Alter        `json:"alters,omitempty"`
}

// SetParameter defines a parameter value assignment.
type SetParameter struct {
	ParamID string   `json:"param-id"`
	Values  []string `json:"values,omitempty"`
}

// Alter defines alterations to a control.
type Alter struct {
	ControlID string       `json:"control-id"`
	Adds      []Addition   `json:"adds,omitempty"`
	Removes   []RemoveItem `json:"removes,omitempty"`
}

// Addition represents content to add to a control.
type Addition struct {
	Position string  `json:"position,omitempty"` // before, after, starting, ending
	Parts    []Part  `json:"parts,omitempty"`
	Props    []Prop  `json:"props,omitempty"`
	Links    []Link  `json:"links,omitempty"`
}

// RemoveItem represents content to remove from a control.
type RemoveItem struct {
	ByName  string `json:"by-name,omitempty"`
	ByClass string `json:"by-class,omitempty"`
	ByID    string `json:"by-id,omitempty"`
}

// AssessmentResults represents an OSCAL assessment-results document.
// Contains findings from security assessments with evidence linkage.
//
// OSCAL 1.1.2 Reference: https://pages.nist.gov/OSCAL/reference/latest/assessment-results/json-outline/
type AssessmentResults struct {
	AssessmentResults AssessmentResultsContent `json:"assessment-results"`
}

// AssessmentResultsContent is the inner structure of assessment-results.
type AssessmentResultsContent struct {
	UUID     string   `json:"uuid"`
	Metadata Metadata `json:"metadata"`
	ImportAP ImportAP `json:"import-ap"`
	Results  []Result `json:"results"`
}

// ImportAP references the assessment plan (or profile) used for this assessment.
type ImportAP struct {
	Href string `json:"href"`
}

// Result represents a single assessment result.
type Result struct {
	UUID             string            `json:"uuid"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Start            string            `json:"start"`
	End              string            `json:"end,omitempty"`
	Props            []Prop            `json:"props,omitempty"`
	ReviewedControls *ReviewedControls `json:"reviewed-controls,omitempty"`
	Observations     []Observation     `json:"observations,omitempty"`
	Findings         []Finding         `json:"findings"`
}

// ReviewedControls specifies which controls were assessed.
type ReviewedControls struct {
	ControlSelections []ControlSelection `json:"control-selections,omitempty"`
}

// ControlSelection defines a set of controls for assessment.
type ControlSelection struct {
	IncludeControls []ControlRef `json:"include-controls,omitempty"`
	ExcludeControls []ControlRef `json:"exclude-controls,omitempty"`
}

// ControlRef is a reference to a control.
type ControlRef struct {
	ControlID string `json:"control-id"`
}

// Observation represents evidence collected during assessment.
type Observation struct {
	UUID             string            `json:"uuid"`
	Title            string            `json:"title,omitempty"`
	Description      string            `json:"description"`
	Methods          []string          `json:"methods"`
	Collected        string            `json:"collected"`
	Props            []Prop            `json:"props,omitempty"`
	RelevantEvidence []RelevantEvidence `json:"relevant-evidence,omitempty"`
}

// RelevantEvidence links to evidence files.
type RelevantEvidence struct {
	Href        string `json:"href"`
	Description string `json:"description,omitempty"`
}

// Finding represents an assessment finding for a specific control objective.
type Finding struct {
	UUID                string               `json:"uuid"`
	Title               string               `json:"title"`
	Description         string               `json:"description"`
	Props               []Prop               `json:"props,omitempty"`
	Target              Target               `json:"target"`
	RelatedObservations []RelatedObservation `json:"related-observations,omitempty"`
	Remarks             string               `json:"remarks,omitempty"`
}

// Target identifies what the finding is about.
type Target struct {
	Type     string `json:"type"`      // "objective-id", "statement-id", etc.
	TargetID string `json:"target-id"` // Control ID (e.g., "ac-1")
	Status   Status `json:"status"`
}

// Status indicates the assessment status of a target.
type Status struct {
	State  string `json:"state"`            // "satisfied", "not-satisfied"
	Reason string `json:"reason,omitempty"` // Optional explanation
}

// RelatedObservation links a finding to observations.
type RelatedObservation struct {
	ObservationUUID string `json:"observation-uuid"`
}

// Metadata contains document metadata common to all OSCAL documents.
type Metadata struct {
	Title        string   `json:"title"`
	LastModified string   `json:"last-modified"`
	Version      string   `json:"version"`
	OSCALVersion string   `json:"oscal-version"`
	Remarks      string   `json:"remarks,omitempty"`
	Props        []Prop   `json:"props,omitempty"`
	Links        []Link   `json:"links,omitempty"`
	Roles        []Role   `json:"roles,omitempty"`
	Parties      []Party  `json:"parties,omitempty"`
}

// Prop is a named property with value.
type Prop struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Class   string `json:"class,omitempty"`
	Remarks string `json:"remarks,omitempty"`
}

// Link represents a hyperlink reference.
type Link struct {
	Href      string `json:"href"`
	Rel       string `json:"rel,omitempty"`
	MediaType string `json:"media-type,omitempty"`
	Text      string `json:"text,omitempty"`
}

// Part represents a control part (section of control narrative).
type Part struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name"`
	Prose string `json:"prose,omitempty"`
	Props []Prop `json:"props,omitempty"`
	Parts []Part `json:"parts,omitempty"`
}

// Role defines a function assumed by a party.
type Role struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// Party represents an organization or person.
type Party struct {
	UUID  string `json:"uuid"`
	Type  string `json:"type"` // "person", "organization"
	Name  string `json:"name"`
	Props []Prop `json:"props,omitempty"`
}

// Constants for OSCAL versions and values.
const (
	// OSCALVersion is the OSCAL version used by this implementation.
	OSCALVersion = "1.1.2"

	// Finding states
	StateSatisfied    = "satisfied"
	StateNotSatisfied = "not-satisfied"

	// Target types
	TargetTypeObjective  = "objective-id"
	TargetTypeStatement  = "statement-id"
	TargetTypeControlID  = "control-id"

	// Observation methods
	MethodTestAutomated = "TEST-AUTOMATED"
	MethodTestManual    = "TEST-MANUAL"
	MethodInterview     = "INTERVIEW"
	MethodExamine       = "EXAMINE"
)

// Common catalog URLs for reference.
const (
	NIST80053Rev5CatalogURL = "https://raw.githubusercontent.com/usnistgov/oscal-content/main/nist.gov/SP800-53/rev5/json/NIST_SP-800-53_rev5_catalog.json"
)

// NewProfile creates a new OSCAL profile with required metadata.
func NewProfile(uuid, title, catalogURL string, controlIDs []string) *Profile {
	now := time.Now().UTC().Format(time.RFC3339)

	includeControls := make([]ControlInclude, len(controlIDs))
	for i, id := range controlIDs {
		includeControls[i] = ControlInclude{WithID: id}
	}

	return &Profile{
		Profile: ProfileContent{
			UUID: uuid,
			Metadata: Metadata{
				Title:        title,
				LastModified: now,
				Version:      "1.0.0",
				OSCALVersion: OSCALVersion,
			},
			Imports: []Import{
				{
					Href:            catalogURL,
					IncludeControls: includeControls,
				},
			},
		},
	}
}

// NewAssessmentResults creates a new OSCAL assessment-results with required metadata.
func NewAssessmentResults(uuid, title, profileHref string) *AssessmentResults {
	now := time.Now().UTC().Format(time.RFC3339)

	return &AssessmentResults{
		AssessmentResults: AssessmentResultsContent{
			UUID: uuid,
			Metadata: Metadata{
				Title:        title,
				LastModified: now,
				Version:      "1.0",
				OSCALVersion: OSCALVersion,
			},
			ImportAP: ImportAP{
				Href: profileHref,
			},
			Results: []Result{},
		},
	}
}

// AddResult adds a new result to the assessment-results.
func (ar *AssessmentResults) AddResult(result Result) {
	ar.AssessmentResults.Results = append(ar.AssessmentResults.Results, result)
}

// GetControlIDs extracts all control IDs from a profile.
func (p *Profile) GetControlIDs() []string {
	var controlIDs []string
	for _, imp := range p.Profile.Imports {
		for _, ctrl := range imp.IncludeControls {
			controlIDs = append(controlIDs, ctrl.WithID)
		}
	}
	return controlIDs
}

// GetCatalogURL returns the catalog URL from the first import.
func (p *Profile) GetCatalogURL() string {
	if len(p.Profile.Imports) > 0 {
		return p.Profile.Imports[0].Href
	}
	return ""
}
