// Package oscal provides compatibility type aliases for go-oscal types.
//
// All OSCAL types (Profile, AssessmentResults, etc.) have been migrated to use
// go-oscal library types. This file provides convenient type aliases to simplify
// code that works with these types.
package oscal

import oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"

// Type aliases for go-oscal types to maintain compatibility
type (
	AssessmentResults = oscalTypes.AssessmentResults
	Result            = oscalTypes.Result
	Observation       = oscalTypes.Observation
	Finding           = oscalTypes.Finding
	Prop              = oscalTypes.Property
	RelevantEvidence  = oscalTypes.RelevantEvidence
	ReviewedControls  = oscalTypes.ReviewedControls
	ControlSelection  = oscalTypes.AssessedControls
	ControlRef        = oscalTypes.AssessedControlsSelectControlById
	Target            = oscalTypes.FindingTarget
	RelatedObservation = oscalTypes.RelatedObservation
)

// Type constants for FindingTarget
const (
	TargetTypeControlID = "objective-id"
)
