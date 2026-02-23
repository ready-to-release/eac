package dependabot

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DiscoverOptions controls which ecosystems to scan for.
type DiscoverOptions struct {
	IncludeGomod         bool
	IncludeNpm           bool
	IncludePip           bool
	IncludeDocker        bool
	IncludeGithubActions bool
	ExcludeDirs          []string // Directory names to skip during filesystem walks
}

// DefaultDiscoverOptions returns options that scan all ecosystems with standard exclusions.
func DefaultDiscoverOptions() DiscoverOptions {
	return DiscoverOptions{
		IncludeGomod:         true,
		IncludeNpm:           true,
		IncludePip:           true,
		IncludeDocker:        true,
		IncludeGithubActions: true,
		ExcludeDirs:          []string{"node_modules", "out", ".cache", "vendor", ".git", "templates", ".vscode"},
	}
}

// DiscoverAll scans the repository and returns all ecosystem entries.
func DiscoverAll(repoRoot string, opts DiscoverOptions) ([]EcosystemEntry, error) {
	var entries []EcosystemEntry

	if opts.IncludeGithubActions {
		actionsDir := filepath.Join(repoRoot, ".github", "workflows")
		if dirExists(actionsDir) {
			entries = append(entries, EcosystemEntry{
				Ecosystem: "github-actions",
				Directory: "/",
			})
		}
	}

	if opts.IncludeGomod {
		goEntries, err := discoverGoModules(repoRoot)
		if err != nil {
			return nil, fmt.Errorf("discovering Go modules: %w", err)
		}
		entries = append(entries, goEntries...)
	}

	if opts.IncludeNpm {
		npmEntries, err := discoverByFile(repoRoot, "package.json", "npm", opts.ExcludeDirs)
		if err != nil {
			return nil, fmt.Errorf("discovering npm packages: %w", err)
		}
		entries = append(entries, npmEntries...)
	}

	if opts.IncludePip {
		pipEntries, err := discoverByGlob(repoRoot, "requirements*.txt", "pip", opts.ExcludeDirs)
		if err != nil {
			return nil, fmt.Errorf("discovering pip packages: %w", err)
		}
		entries = append(entries, pipEntries...)
	}

	if opts.IncludeDocker {
		dockerEntries, err := discoverByFile(repoRoot, "Dockerfile", "docker", opts.ExcludeDirs)
		if err != nil {
			return nil, fmt.Errorf("discovering Dockerfiles: %w", err)
		}
		entries = append(entries, dockerEntries...)
	}

	return entries, nil
}

// discoverGoModules parses go.work to find all Go module directories.
func discoverGoModules(repoRoot string) ([]EcosystemEntry, error) {
	goWorkPath := filepath.Join(repoRoot, "go.work")
	f, err := os.Open(goWorkPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("opening go.work: %w", err)
	}
	defer f.Close()

	var entries []EcosystemEntry
	inUseBlock := false
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "use (" {
			inUseBlock = true
			continue
		}
		if inUseBlock && line == ")" {
			inUseBlock = false
			continue
		}
		if inUseBlock && line != "" && !strings.HasPrefix(line, "//") {
			modPath := strings.TrimPrefix(line, "./")
			dir := "/" + filepath.ToSlash(modPath)
			entries = append(entries, EcosystemEntry{
				Ecosystem: "gomod",
				Directory: dir,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading go.work: %w", err)
	}

	return entries, nil
}

// discoverByFile walks the repo looking for directories containing the given filename.
func discoverByFile(repoRoot, filename, ecosystem string, excludeDirs []string) ([]EcosystemEntry, error) {
	var entries []EcosystemEntry
	excludeSet := toSet(excludeDirs)

	err := filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if d.IsDir() {
			if _, skip := excludeSet[filepath.Base(path)]; skip {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == filename {
			dir := filepath.Dir(path)
			relDir, relErr := filepath.Rel(repoRoot, dir)
			if relErr != nil {
				return nil
			}
			dirSlash := "/" + filepath.ToSlash(relDir)
			if dirSlash == "/." {
				dirSlash = "/"
			}
			entries = append(entries, EcosystemEntry{
				Ecosystem: ecosystem,
				Directory: dirSlash,
			})
		}
		return nil
	})

	return entries, err
}

// discoverByGlob walks the repo looking for directories containing files matching the glob.
// Deduplicates by directory so multiple matching files in the same directory produce one entry.
func discoverByGlob(repoRoot, glob, ecosystem string, excludeDirs []string) ([]EcosystemEntry, error) {
	var entries []EcosystemEntry
	excludeSet := toSet(excludeDirs)
	seen := make(map[string]bool)

	err := filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if _, skip := excludeSet[filepath.Base(path)]; skip {
				return filepath.SkipDir
			}
			return nil
		}
		matched, _ := filepath.Match(glob, d.Name())
		if matched {
			dir := filepath.Dir(path)
			relDir, relErr := filepath.Rel(repoRoot, dir)
			if relErr != nil {
				return nil
			}
			dirSlash := "/" + filepath.ToSlash(relDir)
			if !seen[dirSlash] {
				seen[dirSlash] = true
				entries = append(entries, EcosystemEntry{
					Ecosystem: ecosystem,
					Directory: dirSlash,
				})
			}
		}
		return nil
	})

	return entries, err
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func toSet(items []string) map[string]struct{} {
	s := make(map[string]struct{}, len(items))
	for _, item := range items {
		s[item] = struct{}{}
	}
	return s
}
