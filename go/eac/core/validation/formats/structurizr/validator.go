package structurizr

import (
	"github.com/ready-to-release/eac/go/eac/core/validation"
)

// ValidatorMode defines the validation mode.
type ValidatorMode int

const (
	ModeQuick     ValidatorMode = iota // Fast, no Docker (AI generation)
	ModeFull                           // Docker validation (validate command)
	ModeComposite                      // Quick then full (create command)
)

// Validator is the base interface for Structurizr validators.
type Validator interface {
	validation.Validator
}
