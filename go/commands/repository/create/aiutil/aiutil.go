// Package aiutil provides shared utilities for AI-powered commit message generation commands
// (get commit-message, get squash-message).
package aiutil

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/clibase/aiproviders"
	coreai "github.com/ready-to-release/eac/go/core/ai"
	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
)

// GenerationParams configures a single AI generation call.
type GenerationParams struct {
	WorkspaceRoot  string
	Prompt         string // Full prompt (template + user input)
	TypeName       string // e.g. coreai.TypeCommitMessage, coreai.TypeSquashMessage
	SchemaFilename string // e.g. "commit-message.schema.json"
	Model          string // Optional model override from agent frontmatter
	Debug          bool
	TestExecutor   *aiproviders.Executor // Optional: use this executor instead of creating one (for testing)
}

// ExecuteGeneration performs the shared AI executor pipeline:
// create executor → register providers → load AI config → build schema validator →
// build retry config → generate with retry.
func ExecuteGeneration(params GenerationParams) (*coreai.RetryResult, error) {
	var executor *aiproviders.Executor
	if params.TestExecutor != nil {
		executor = params.TestExecutor
	} else {
		executor = aiproviders.NewExecutor(params.WorkspaceRoot, nil)
	}

	var genExecutor coreai.GenerationExecutor
	if params.Model != "" {
		genExecutor = aiproviders.NewWithModelExecutor(executor, params.Model)
	} else {
		genExecutor = executor
	}

	aiConfig, err := coreai.LoadAIConfig(params.WorkspaceRoot)
	if err != nil {
		logging.C().Warnf("Could not load AI config, using default retry strategy: %v", err)
		aiConfig = nil
	}

	schemaPath := filepath.Join(
		paths.ContractsVersionPath(params.WorkspaceRoot, paths.EACCoreModule, paths.DefaultsVersion),
		paths.SchemasDir,
		params.SchemaFilename,
	)
	validator, err := domain.NewJSONSchemaValidator(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create JSON schema validator: %w", err)
	}

	opts := []coreai.RetryConfigOption{
		coreai.WithLogger(logging.C().Zap()),
	}
	if params.Debug {
		opts = append(opts, coreai.WithDebug(true))
	}

	retryConfig, err := coreai.BuildRetryConfig(
		params.TypeName,
		coreai.FormatJSON,
		genExecutor,
		validator,
		params.WorkspaceRoot,
		aiConfig,
		opts...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build retry config: %w", err)
	}

	result, err := coreai.GenerateWithRetry(context.Background(), retryConfig, params.Prompt)
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	return result, nil
}

// LogDebugArtifact logs debug content with labeled sections to the log file.
// Used by AI generation commands to record intermediate outputs for troubleshooting.
func LogDebugArtifact(log *logging.ComponentLogger, label, content string) {
	log.Debug(fmt.Sprintf("=== %s START ===", label))
	log.Debug(content)
	log.Debug(fmt.Sprintf("=== %s END ===", label))
}

// ExtractFirstSentence extracts the first sentence from text.
// Returns the first sentence terminated by a period, question mark, or exclamation mark.
// If no sentence delimiter is found, returns the first line with a trailing period.
// The fallback parameter is used when text is empty.
func ExtractFirstSentence(text, fallback string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		if fallback != "" {
			return ensurePeriod(fallback)
		}
		return "Summary of changes."
	}

	// Find first sentence ending
	for _, delim := range []string{". ", ".\n", "? ", "?\n", "! ", "!\n"} {
		if idx := strings.Index(text, delim); idx != -1 {
			return strings.TrimSpace(text[:idx+1])
		}
	}

	// If text ends with punctuation, return first line
	if strings.HasSuffix(text, ".") || strings.HasSuffix(text, "?") || strings.HasSuffix(text, "!") {
		lines := strings.Split(text, "\n")
		return strings.TrimSpace(lines[0])
	}

	// Fallback: first line with period
	lines := strings.Split(text, "\n")
	firstLine := strings.TrimSpace(lines[0])
	if firstLine == "" {
		return ensurePeriod(fallback)
	}
	return ensurePeriod(firstLine)
}

func ensurePeriod(s string) string {
	if s == "" {
		return "."
	}
	if !strings.HasSuffix(s, ".") {
		return s + "."
	}
	return s
}

// shuffled returns a shuffled copy of src without mutating the original slice.
func shuffled(src []string) []string {
	dst := make([]string, len(src))
	copy(dst, src)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(dst), func(i, j int) {
		dst[i], dst[j] = dst[j], dst[i]
	})

	return dst
}

// WhimsicalStatusLines are fun status messages shown during AI generation.
var WhimsicalStatusLines = []string{
	"Discombobulating the git diffs...",
	"Reticulating splines...",
	"Consulting the commit oracle...",
	"Parsing semantic tea leaves...",
	"Harmonizing module boundaries...",
	"Calibrating imperative mood detector...",
	"Summoning the contract guardian...",
	"Extracting essence from code changes...",
	"Wrapping lines at 72 characters...",
	"Polishing commit message prose...",
	"Validating YAML path globs...",
	"Generating subject line haikus...",
	"Contemplating the WHY not the WHAT...",
	"Assembling markdown tables...",
	"Invoking the anti-corruption layer...",
	"Measuring semantic distance...",
	"Calculating commit entropy...",
	"Negotiating with git demons...",
}

// AngryStatusLines are slightly frustrated messages for retry/auto-fix phases.
var AngryStatusLines = []string{
	"Ugh, fixing validation errors...",
	"*Sigh* Correcting the mistakes...",
	"This would've been easier if...",
	"Fine, let me fix that...",
	"Okay okay, adjusting things...",
	"Not my finest work, but fixing it...",
	"Patience wearing thin... fixing...",
	"Let's try this again, properly...",
	"Tweaking the problematic bits...",
	"Making it contract-compliant...",
	"Ironing out the wrinkles...",
	"Second time's the charm...",
}

// startProgressWithLines begins showing progress updates with custom status lines.
// Returns a cancel function to stop the progress ticker.
func startProgressWithLines(initialMessage string, statusLines []string, interval time.Duration) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())

	logging.C().Info(initialMessage)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		statusIndex := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				logging.C().Info(statusLines[statusIndex%len(statusLines)])
				statusIndex++
			}
		}
	}()

	return cancel
}

// WithProgress wraps fn with periodic whimsical progress log lines.
// The initial message is logged immediately; subsequent lines appear every 10s.
func WithProgress(stage string, fn func() error) error {
	statusLines := shuffled(WhimsicalStatusLines)

	stopProgress := startProgressWithLines(stage, statusLines, 10*time.Second)
	defer stopProgress()

	return fn()
}

// WithAngryProgress wraps fn with "angry" progress log lines (for retry phases).
// Updates appear every 8 seconds (faster cadence than normal).
func WithAngryProgress(stage string, fn func() error) error {
	statusLines := shuffled(AngryStatusLines)

	stopProgress := startProgressWithLines(stage, statusLines, 8*time.Second)
	defer stopProgress()

	return fn()
}
