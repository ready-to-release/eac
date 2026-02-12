package tool

// builtinTestTypeMapping provides fallback mappings when tool config and
// adapter registry are unavailable (e.g., in unit tests without adapter init).
// This map is immutable and safe for concurrent read access.
var builtinTestTypeMapping = map[string]string{
	"gotest":     "go",
	"godog":      "gherkin",
	"mocha":      "typescript",
	"tscucumber": "gherkin",
}
