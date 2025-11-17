package commitmessage

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// WhimsicalStatusLines are fun status messages shown during generation
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

// startProgressWithLines begins showing progress updates with custom status lines
// Returns a cancel function to stop the progress ticker
func startProgressWithLines(initialMessage string, statusLines []string) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())

	// Show initial message immediately
	fmt.Println(initialMessage)

	// Start ticker for updates every 10 seconds
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		statusIndex := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Show next status
				fmt.Println(statusLines[statusIndex%len(statusLines)])
				statusIndex++
			}
		}
	}()

	return cancel
}

// AngryStatusLines are slightly frustrated messages for auto-fix phase
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

// WithProgress wraps a function with progress updates
func WithProgress(stage string, fn func() error) error {
	// Create local copy to avoid mutating global state
	statusLines := make([]string, len(WhimsicalStatusLines))
	copy(statusLines, WhimsicalStatusLines)

	// Shuffle local copy for variety
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(statusLines), func(i, j int) {
		statusLines[i], statusLines[j] = statusLines[j], statusLines[i]
	})

	stopProgress := startProgressWithLines(stage, statusLines)
	defer stopProgress()

	return fn()
}

// WithAngryProgress wraps a function with "angry" progress updates (for auto-fix)
func WithAngryProgress(stage string, fn func() error) error {
	// Create local copy to avoid mutating global state
	statusLines := make([]string, len(AngryStatusLines))
	copy(statusLines, AngryStatusLines)

	// Shuffle local copy
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(statusLines), func(i, j int) {
		statusLines[i], statusLines[j] = statusLines[j], statusLines[i]
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Show initial message immediately
	fmt.Println(stage)

	// Start ticker for angry updates every 8 seconds (faster than normal)
	go func() {
		ticker := time.NewTicker(8 * time.Second)
		defer ticker.Stop()

		statusIndex := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Show next status
				fmt.Println(statusLines[statusIndex%len(statusLines)])
				statusIndex++
			}
		}
	}()

	return fn()
}
