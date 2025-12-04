package oscal

// NIST catalog URLs
const (
	// NIST80053Rev5CatalogURL is the official NIST 800-53 Revision 5 catalog URL
	NIST80053Rev5CatalogURL = "https://raw.githubusercontent.com/usnistgov/oscal-content/main/nist.gov/SP800-53/rev5/json/NIST_SP-800-53_rev5_catalog.json"
)

// OSCAL finding states
const (
	StateSatisfied    = "satisfied"
	StateNotSatisfied = "not-satisfied"
	StateOther        = "other"
)

// OSCAL observation methods
const (
	MethodTestAutomated    = "TEST-AUTOMATED"
	MethodTestManual       = "TEST-MANUAL"
	MethodExamineAutomated = "EXAMINE-AUTOMATED"
	MethodExamineManual    = "EXAMINE-MANUAL"
)
