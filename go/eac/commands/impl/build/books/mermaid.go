package books

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/build/buildutil"
	"github.com/ready-to-release/eac/go/eac/core/paths"
)

// Size presets for mermaid diagrams
var mermaidSizePresets = map[string]string{
	"small":  "33%",
	"medium": "50%",
	"large":  "66%",
	"full":   "100%",
}

// mermaidBlockPattern matches mermaid code blocks with optional size directive
// Captures: (1) size directive value, (2) mermaid content
var mermaidBlockPattern = regexp.MustCompile("(?s)```mermaid\\s*\n%%\\{(?:size|width):([^}]+)\\}%%\\s*\n(.*?)```")

// mermaidBlockPlain matches plain mermaid blocks without size directive
var mermaidBlockPlain = regexp.MustCompile("(?s)```mermaid\\s*\n(.*?)```")

// processMermaidSizing wraps mermaid blocks with size directives in container divs
// This enables CSS-based sizing for both web (mermaid2) and PDF (mermaid-to-svg)
//
// Syntax in markdown:
//
//	```mermaid
//	%%{size:medium}%%
//	flowchart TD
//	    A --> B
//	```
//
// Or with explicit width:
//
//	```mermaid
//	%%{width:40%}%%
//	flowchart TD
//	    A --> B
//	```
//
// Size presets: small (33%), medium (50%), large (66%), full (100%)
func (p *Preprocessor) processMermaidSizing() error {
	p.log("    Processing mermaid diagram sizing...")

	processed := 0
	wrapped := 0

	err := filepath.WalkDir(p.stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		original := string(content)
		modified := wrapMermaidBlocks(original)

		if modified != original {
			if err := os.WriteFile(path, []byte(modified), 0644); err != nil {
				return err
			}
			wrapped++
		}
		processed++
		return nil
	})

	if err != nil {
		return err
	}

	p.log("    Processed %d files, wrapped %d mermaid blocks with sizing", processed, wrapped)
	return nil
}

// wrapMermaidBlocks finds mermaid blocks with size directives and wraps them
func wrapMermaidBlocks(content string) string {
	// Process blocks with size directives
	result := mermaidBlockPattern.ReplaceAllStringFunc(content, func(match string) string {
		// Extract size value and content
		submatches := mermaidBlockPattern.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}

		sizeValue := strings.TrimSpace(submatches[1])
		mermaidContent := submatches[2]

		// Resolve preset name to percentage
		width := sizeValue
		if preset, ok := mermaidSizePresets[strings.ToLower(sizeValue)]; ok {
			width = preset
		}

		// Ensure width ends with %
		if !strings.HasSuffix(width, "%") && !strings.HasSuffix(width, "px") {
			width = width + "%"
		}

		// Build wrapped block
		// Use both data-size attribute (for CSS) and inline style (for PDF)
		var wrapper strings.Builder
		wrapper.WriteString("<div class=\"mermaid-wrapper\" data-size=\"")
		wrapper.WriteString(strings.ToLower(sizeValue))
		wrapper.WriteString("\" style=\"max-width:")
		wrapper.WriteString(width)
		wrapper.WriteString("; margin: 0 auto;\">\n\n")
		wrapper.WriteString("```mermaid\n")
		wrapper.WriteString(mermaidContent)
		wrapper.WriteString("```\n\n</div>")

		return wrapper.String()
	})

	return result
}

// countMermaidBlocks returns the number of mermaid blocks in content (for logging)
func countMermaidBlocks(content string) int {
	return len(mermaidBlockPlain.FindAllString(content, -1))
}

// mermaidBlock represents a mermaid diagram found in markdown
// Used for caching and pre-rendering during preprocessing
type mermaidBlock struct {
	content      string // The mermaid diagram code
	hash         string // SHA256 hash of content (first 8 chars for filename)
	sourceFile   string // Absolute path to the .md file
	relPath      string // Relative path from staging dir (for logging)
	blockIndex   int    // Index of block in file (0, 1, 2, ...)
	filename     string // Generated SVG filename: {source}_mermaid_{idx}_{hash}.svg
	startPos     int    // Start position in file (for replacement later)
	endPos       int    // End position in file (for replacement later)
}

// extractMermaidBlocks scans a markdown file for mermaid code blocks
// Returns all blocks with metadata for caching and rendering
func extractMermaidBlocks(content string, absSourcePath string, stagingDir string) []mermaidBlock {
	blocks := []mermaidBlock{}

	// Get relative path for logging
	relPath, _ := filepath.Rel(stagingDir, absSourcePath)
	if relPath == "" {
		relPath = filepath.Base(absSourcePath)
	}

	// Get base filename for SVG naming
	basename := filepath.Base(absSourcePath)
	basename = strings.TrimSuffix(basename, filepath.Ext(basename))

	// Find all mermaid blocks (both with and without size directives)
	// Use mermaidBlockPlain to match all ```mermaid...``` blocks
	matches := mermaidBlockPlain.FindAllStringSubmatchIndex(content, -1)

	for idx, match := range matches {
		// match is an index slice: [fullStart, fullEnd, group1Start, group1End]
		if len(match) < 4 {
			continue
		}

		// Extract the diagram content (group 1)
		diagramContent := strings.TrimSpace(content[match[2]:match[3]])

		// Skip empty blocks
		if diagramContent == "" {
			continue
		}

		// Remove size directives from content before hashing
		// This ensures the hash is based on actual diagram code, not formatting
		diagramForHash := stripSizeDirective(diagramContent)

		// Hash the content for cache key (8 chars like the plugin does)
		hash := hashContent(diagramForHash)

		// Generate filename: {basename}_mermaid_{idx}_{hash}.svg
		filename := fmt.Sprintf("%s_mermaid_%d_%s.svg", basename, idx, hash)

		blocks = append(blocks, mermaidBlock{
			content:    diagramContent,
			hash:       hash,
			sourceFile: absSourcePath,
			relPath:    relPath,
			blockIndex: idx,
			filename:   filename,
			startPos:   match[0],
			endPos:     match[1],
		})
	}

	return blocks
}

// stripSizeDirective removes size directive lines from diagram content
// Example: %%{size:medium}%% is removed before hashing
func stripSizeDirective(content string) string {
	// Remove lines starting with %%{size: or %%{width:
	lines := strings.Split(content, "\n")
	filtered := []string{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "%%{size:") || strings.HasPrefix(trimmed, "%%{width:") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
}

// hashContent returns first 8 chars of SHA256 hash
// This matches the naming convention used by the mermaid-to-svg plugin
func hashContent(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h)[:8]
}

// cacheStatus represents the cache state for a mermaid block
type cacheStatus struct {
	block     mermaidBlock
	cached    bool   // true if SVG exists in cache
	cachePath string // absolute path to cached SVG (if exists)
}

// mermaidImageName is the Docker image used for mermaid-cli rendering
const mermaidImageName = "cli-mermaid-cli:latest"

// ensureMermaidImage ensures the mermaid-cli Docker image exists.
// Similar to ensureMkDocsImage, this checks if image exists first; only builds if missing.
func ensureMermaidImage(workspaceRoot string, logWriter io.Writer) error {
	// Check if image already exists
	cmd := exec.Command("docker", "image", "inspect", mermaidImageName)
	if err := cmd.Run(); err == nil {
		fmt.Fprintf(logWriter, "    Docker image exists: %s (using pre-built)\n", mermaidImageName)
		return nil
	}

	// Image doesn't exist, build it
	fmt.Fprintf(logWriter, "    Building Docker image: %s\n", mermaidImageName)

	// Detect Docker-in-Docker mode for path handling
	isDinD := buildutil.IsDockerInDocker()
	hostRepoRoot := workspaceRoot
	if isDinD {
		if hostRoot := os.Getenv("R2R_HOST_REPOROOT"); hostRoot != "" {
			hostRepoRoot = hostRoot
		}
	}

	// Build paths for Dockerfile and context
	var dockerfilePath, contextPath string
	if isDinD {
		// Host is Windows, construct Windows paths manually
		dockerfilePath = hostRepoRoot + "\\containers\\mermaid-cli\\Dockerfile"
		contextPath = hostRepoRoot + "\\containers\\mermaid-cli"
	} else {
		dockerfilePath = filepath.Join(workspaceRoot, "containers", "mermaid-cli", "Dockerfile")
		contextPath = filepath.Join(workspaceRoot, "containers", "mermaid-cli")
	}

	fmt.Fprintf(logWriter, "    Dockerfile: %s\n", dockerfilePath)
	fmt.Fprintf(logWriter, "    Context: %s\n", contextPath)

	cmd = exec.Command("docker", "build",
		"-t", mermaidImageName,
		"-f", dockerfilePath,
		contextPath)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	return nil
}

// renderSingleDiagram renders a single mermaid diagram to SVG using mermaid-cli
// Returns error if rendering fails
func renderSingleDiagram(block mermaidBlock, outputPath string, workspaceRoot string, logWriter io.Writer) error {
	// Detect Docker-in-Docker mode
	isDinD := buildutil.IsDockerInDocker()
	hostRepoRoot := workspaceRoot
	if isDinD {
		if hostRoot := os.Getenv("R2R_HOST_REPOROOT"); hostRoot != "" {
			hostRepoRoot = hostRoot
		}
	}

	// Create temp file for mermaid content
	tmpDir := filepath.Dir(outputPath)
	tmpFile := filepath.Join(tmpDir, block.filename+".mmd")

	// Write diagram content to temp file
	if err := os.WriteFile(tmpFile, []byte(block.content), 0644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}
	defer os.Remove(tmpFile) // Clean up temp file

	// Calculate Docker paths (relative to workspace root)
	relTmpFile, err := filepath.Rel(workspaceRoot, tmpFile)
	if err != nil {
		return fmt.Errorf("calculating relative tmp path: %w", err)
	}
	relOutputPath, err := filepath.Rel(workspaceRoot, outputPath)
	if err != nil {
		return fmt.Errorf("calculating relative output path: %w", err)
	}

	// Convert to Docker paths (forward slashes)
	dockerTmpFile := "/docs/" + strings.ReplaceAll(relTmpFile, "\\", "/")
	dockerOutputPath := "/docs/" + strings.ReplaceAll(relOutputPath, "\\", "/")

	// Format Docker volume path using host paths for DinD
	dockerVolume := buildutil.FormatDockerVolumePath(hostRepoRoot)

	// Build Docker command
	// Use dedicated cli-mkdocs-mermaid container for diagram rendering
	// The container has mermaid-cli, chromium, and embedded configs at /etc/mermaid/
	args := []string{
		"run", "--rm",
		"-v", dockerVolume + ":/docs",
		"-w", "/docs",
		"--shm-size=1gb", // Increase shared memory for Chromium (prevents crashes)
		"--security-opt", "seccomp=unconfined", // Allow Chromium to run without sandboxing restrictions
		mermaidImageName,
		"mmdc",
		"-i", dockerTmpFile,
		"-o", dockerOutputPath,
		"-t", "dark",        // Theme (dark for PDF)
		"-b", "transparent", // Background
		"--configFile", "/etc/mermaid/mermaid-config.json", // Disable htmlLabels for PDF compatibility
		"-p", "/etc/mermaid/puppeteer-config.json", // Puppeteer config for container environment
	}

	// Add user spec in DinD mode to avoid permission issues
	if isDinD {
		uid := os.Getuid()
		gid := os.Getgid()
		userSpec := fmt.Sprintf("%d:%d", uid, gid)
		// Insert --user after "run" and "--rm"
		args = append([]string{"run", "--rm", "--user", userSpec}, args[2:]...)
	}

	// Run docker command
	cmd := exec.Command("docker", args...)
	cmd.Dir = workspaceRoot

	// Capture output for debugging
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mmdc failed for %s: %w (stderr: %s)",
			block.filename, err, stderr.String())
	}

	// Verify SVG was created
	if _, err := os.Stat(outputPath); err != nil {
		return fmt.Errorf("SVG not created: %w", err)
	}

	return nil
}

// renderMermaidDiagrams renders multiple mermaid diagrams in parallel
// Only renders cache misses (cached diagrams are skipped)
// Returns number of diagrams rendered and any error
func (p *Preprocessor) renderMermaidDiagrams(statuses []cacheStatus) (int, error) {
	// Check Docker availability first - fail fast if unavailable
	if !buildutil.IsDockerAvailable() {
		errorMsg := "Docker is not available but required for mermaid diagram rendering"
		if buildutil.IsDockerInDocker() {
			errorMsg += "\nRunning in container: mount Docker socket with -v /var/run/docker.sock:/var/run/docker.sock"
		} else {
			errorMsg += "\nEnsure Docker is installed and the daemon is running"
		}
		return 0, fmt.Errorf("%s", errorMsg)
	}

	// Ensure mermaid-cli image exists (build if needed)
	if err := ensureMermaidImage(p.workspaceRoot, p.logWriter); err != nil {
		return 0, fmt.Errorf("failed to ensure mermaid image: %w", err)
	}

	// Filter for cache misses
	toRender := []cacheStatus{}
	for _, status := range statuses {
		if !status.cached {
			toRender = append(toRender, status)
		}
	}

	if len(toRender) == 0 {
		p.log("    All diagrams cached, nothing to render")
		return 0, nil
	}

	// Determine worker count (max 4 parallel renders to avoid overwhelming Chromium)
	// Each worker spawns a Docker container with Chromium, so we keep this conservative
	maxWorkers := 4
	if maxWorkers > len(toRender) {
		maxWorkers = len(toRender)
	}

	p.log("    Rendering %d diagram(s) in parallel (using %d workers)...", len(toRender), maxWorkers)

	// Create channels for work distribution
	jobs := make(chan cacheStatus, len(toRender))
	type result struct {
		status cacheStatus
		err    error
	}
	results := make(chan result, len(toRender))

	// Start worker goroutines
	for w := 0; w < maxWorkers; w++ {
		go func(workerID int) {
			for status := range jobs {
				block := status.block
				err := renderSingleDiagram(block, status.cachePath, p.workspaceRoot, p.logWriter)
				results <- result{status: status, err: err}
			}
		}(w)
	}

	// Send jobs to workers
	for _, status := range toRender {
		jobs <- status
	}
	close(jobs)

	// Collect results
	rendered := 0
	failed := 0
	var errors []string

	for i := 0; i < len(toRender); i++ {
		res := <-results
		if res.err != nil {
			p.log("      ❌ Failed to render %s: %v", res.status.block.filename, res.err)
			failed++
			errors = append(errors, fmt.Sprintf("%s: %v", res.status.block.filename, res.err))
		} else {
			rendered++

			// Save to persistent cache after successful rendering
			block := res.status.block
			// Use stripped content for cache key to ensure consistency
			cleanContent := stripSizeDirective(block.content)
			if err := p.assetCache.PutMermaid(res.status.cachePath, MermaidCacheKey{
				Code: cleanContent,
			}); err != nil{
				p.log("      ⚠️  Failed to cache %s: %v", block.filename, err)
				// Non-fatal - rendering succeeded even if caching failed
			}

			// Log progress every 10 diagrams
			if rendered%10 == 0 || rendered == len(toRender) {
				p.log("      ✓ Progress: %d/%d diagrams rendered", rendered, len(toRender))
			}
		}
	}

	if failed > 0 {
		return rendered, fmt.Errorf("%d diagram(s) failed to render", failed)
	}

	p.log("    ✓ Rendered %d diagram(s) successfully", rendered)
	return rendered, nil
}

// checkMermaidCache checks which diagrams are already cached
// Checks both persistent cache (out/cache/mermaid/) and local staging cache
// Returns all blocks with their cache status
func (p *Preprocessor) checkMermaidCache(blocks []mermaidBlock) ([]cacheStatus, error) {
	// Cache directory: staging/assets/rendered/mermaid/
	cacheDir := paths.RenderedAssetsPath(p.stagingDir, "mermaid")

	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}

	statuses := make([]cacheStatus, 0, len(blocks))

	for _, block := range blocks {
		svgPath := filepath.Join(cacheDir, block.filename)

		// Check if file exists in local staging cache
		_, err := os.Stat(svgPath)
		cached := err == nil

		// If not in local cache, check persistent cache
		if !cached {
			// Use stripped content for cache key to ensure consistency
			// (size directives are removed by processMermaidSizing step)
			cleanContent := stripSizeDirective(block.content)
			persistentPath, persistentHit := p.assetCache.GetMermaid(MermaidCacheKey{
				Code: cleanContent,
			})

			if persistentHit {
				// Copy from persistent cache to local staging cache
				if err := copyFile(persistentPath, svgPath); err == nil {
					cached = true
					// No need to increment stats here - GetMermaid already did
				}
			}
		}

		statuses = append(statuses, cacheStatus{
			block:     block,
			cached:    cached,
			cachePath: svgPath,
		})
	}

	return statuses, nil
}

// replaceMermaidBlocksWithImages replaces mermaid code blocks with img references
// This is done ONLY in staging directory, source markdown stays pure
func (p *Preprocessor) replaceMermaidBlocksWithImages(blocksByFile map[string][]mermaidBlock) error {
	// Cache directory (absolute path)
	cacheDir := paths.RenderedAssetsPath(p.stagingDir, "mermaid")

	for filePath, blocks := range blocksByFile {
		if len(blocks) == 0 {
			continue
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("reading file %s: %w", filePath, err)
		}

		modified := string(content)

		// Replace blocks in reverse order (to preserve positions)
		for i := len(blocks) - 1; i >= 0; i-- {
			block := blocks[i]

			// Calculate relative path from markdown file to SVG using link translator
			// This ensures consistency with all other path calculations
			svgAbsPath := filepath.Join(cacheDir, block.filename)
			relPath, err := p.linkTranslator.CalculateRelativePath(filePath, svgAbsPath)
			if err != nil {
				return fmt.Errorf("calculating relative path for %s: %w", block.filename, err)
			}

			// Build img tag with relative path to SVG
			imgTag := fmt.Sprintf(
				"<img src=\"%s\" alt=\"Mermaid diagram\" style=\"max-width: 100%%;\">",
				relPath,
			)

			// Replace mermaid block with img tag
			modified = modified[:block.startPos] + imgTag + modified[block.endPos:]
		}

		// Write back to staging file
		if err := os.WriteFile(filePath, []byte(modified), 0644); err != nil {
			return fmt.Errorf("writing file %s: %w", filePath, err)
		}

		p.log("      ✓ Replaced %d mermaid block(s) in %s", len(blocks), blocks[0].relPath)
	}

	return nil
}

// scanForMermaidDiagrams scans all markdown files in staging directory
// Returns all mermaid blocks found, grouped by file
// Now includes cache checking and statistics
func (p *Preprocessor) scanForMermaidDiagrams() (map[string][]mermaidBlock, error) {
	blocksByFile := make(map[string][]mermaidBlock)
	allBlocks := []mermaidBlock{}

	// Step 1: Extract all mermaid blocks
	err := filepath.WalkDir(p.stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		blocks := extractMermaidBlocks(string(content), path, p.stagingDir)
		if len(blocks) > 0 {
			blocksByFile[path] = blocks
			allBlocks = append(allBlocks, blocks...)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(allBlocks) == 0 {
		p.log("    No mermaid diagrams found")
		return blocksByFile, nil
	}

	// Step 2: Check cache for all blocks
	statuses, err := p.checkMermaidCache(allBlocks)
	if err != nil {
		return nil, fmt.Errorf("checking cache: %w", err)
	}

	// Step 3: Analyze cache statistics
	cacheHits := 0
	cacheMisses := 0
	for _, status := range statuses {
		if status.cached {
			cacheHits++
		} else {
			cacheMisses++
		}
	}

	// Step 4: Render cache misses
	rendered := 0
	var renderErr error
	if cacheMisses > 0 {
		rendered, renderErr = p.renderMermaidDiagrams(statuses)
		// Note: renderErr may be non-nil if some renders failed,
		// but we continue to show statistics
	}

	// Step 5: Log summary
	totalDiagrams := len(allBlocks)
	cacheHitRate := 0.0
	if totalDiagrams > 0 {
		cacheHitRate = float64(cacheHits) / float64(totalDiagrams) * 100
	}

	p.log("    Found %d diagrams in %d files", totalDiagrams, len(blocksByFile))
	p.log("    Cache: %d hits, %d misses (%.1f%% hit rate)",
		cacheHits, cacheMisses, cacheHitRate)
	if rendered > 0 {
		p.log("    Rendered: %d diagrams", rendered)
	}

	// Step 6: Log detailed breakdown by file
	for _, blocks := range blocksByFile {
		relPath := blocks[0].relPath
		fileHits := 0
		fileMisses := 0

		for _, block := range blocks {
			// Find this block's cache status
			for _, status := range statuses {
				if status.block.filename == block.filename {
					if status.cached {
						fileHits++
					} else {
						fileMisses++
					}
					break
				}
			}
		}

		p.log("      %s: %d diagram(s) (%d cached, %d to render)",
			relPath, len(blocks), fileHits, fileMisses)

		// Show details for each diagram
		for _, block := range blocks {
			// Find cache status
			cached := false
			for _, status := range statuses {
				if status.block.filename == block.filename {
					cached = status.cached
					break
				}
			}

			cacheMarker := "❌"
			if cached {
				cacheMarker = "✓"
			}

			// Show first line of content
			firstLine := strings.Split(block.content, "\n")[0]
			if len(firstLine) > 45 {
				firstLine = firstLine[:45] + "..."
			}

			p.log("        [%d] %s %s %s",
				block.blockIndex, cacheMarker, firstLine, block.filename)
		}
	}

	// Return rendering error if any diagrams failed
	// (but still return the blocksByFile so caller can see what was found)
	if renderErr != nil {
		return blocksByFile, fmt.Errorf("rendering: %w", renderErr)
	}

	return blocksByFile, nil
}
