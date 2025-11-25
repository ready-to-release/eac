package commitmessage

// Line length and formatting constants
const (
	// MaxLineLength is the maximum line length for commit message body text
	MaxLineLength = 72

	// MaxHeaderLength is the maximum length for the top-level commit header
	MaxHeaderLength = 72

	// MaxSubjectLength is the maximum length for module subject lines
	MaxSubjectLength = 72

	// MaxModuleNameLength is the maximum length for a module name
	MaxModuleNameLength = 50

	// MinDashesLength is the minimum length for a dashes separator line
	MinDashesLength = 3

	// MaxDiffSize is the maximum size for a git diff (10 MB)
	MaxDiffSize = 10 * 1024 * 1024

	// MaxPromptDiffSize is the maximum size for git diff to include in AI prompt (100 KB)
	// Larger diffs will be truncated to avoid exceeding Claude CLI prompt limits
	MaxPromptDiffSize = 100 * 1024
)

// Contract version constants
const (
	// ContractVersion is the expected version of the commit message contract
	ContractVersion = "0.1.0"
)

// Standard semantic commit types as defined in the contract
// Note: This list should be loaded from the contract at runtime
// See: contracts/commit-message/0.1.0/structure.yml
var StandardCommitTypes = []string{
	"feat",
	"fix",
	"refactor",
	"docs",
	"chore",
	"test",
	"perf",
	"style",
}
