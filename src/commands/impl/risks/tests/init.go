// Package tests provides BDD step definitions for the risks commands.
package tests

import (
	"github.com/cucumber/godog"
)

// InitializeScenario registers all step definitions for risks commands.
func InitializeScenario(sc *godog.ScenarioContext) {
	registerAssessmentSteps(sc)
	registerCreateSteps(sc)
	registerListSteps(sc)
}
