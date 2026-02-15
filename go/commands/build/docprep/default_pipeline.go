package docprep

import (
	"os"

	"github.com/ready-to-release/eac/go/commands/build/docprep/cleanup"
	"github.com/ready-to-release/eac/go/commands/build/docprep/content"
	"github.com/ready-to-release/eac/go/commands/build/docprep/diagrams"
	"github.com/ready-to-release/eac/go/commands/build/docprep/linking"
	"github.com/ready-to-release/eac/go/commands/build/docprep/navigation"
	"github.com/ready-to-release/eac/go/commands/build/docprep/staging"
	"github.com/ready-to-release/eac/go/core/paths"
	"golang.org/x/sync/errgroup"
)

// DefaultPipeline returns the standard preprocessing pipeline.
// It assembles all phases from the sub-packages into the correct order.
func DefaultPipeline() *Pipeline {
	return NewPipeline().
		// Phase 1: Scan source markdown for asset references (lazy copying optimization)
		AddFunc("scan-asset-refs", scanAssetRefsPhase).
		// Phase 2: Copy static files to staging
		AddFunc("copy-sources", copySourcesPhase).
		// Phase 3: Build file index for efficient file iteration
		AddFunc("build-file-index", buildFileIndexPhase).
		// Phase 4: Convert attr_list images to HTML (before link translation)
		AddFunc("convert-attr-images", convertAttrImagesPhase).
		// Phase 5: Build and apply link translations (source paths -> staging paths)
		AddFunc("link-translation", linkTranslationPhase).
		// Phase 6: Execute commands, capture outputs
		AddFunc("execute-commands", executeCommandsPhase).
		// Phase 7: Ensure root index and navigation structure
		AddFunc("navigation-structure", navigationStructurePhase).
		// Phase 8: Insert inline command outputs at markers
		AddFunc("inline-content", inlineContentPhase).
		// Phase 9: Process command help markers
		AddFunc("command-help", commandHelpPhase).
		// Phase 10: Output-mode-specific macros (strip for PDF, inject for site)
		AddFunc("navigation-macros", navigationMacrosPhase).
		// Phase 11: Diagram processing (mermaid + structurizr in parallel)
		AddFunc("diagram-processing", diagramProcessingPhase).
		// Phase 12: Image width constraints
		AddFunc("image-constraints", imageConstraintsPhase).
		// Phase 13: Fix broken internal links
		AddFunc("fix-broken-links", fixBrokenLinksPhase).
		// Phase 14: Clean up unreferenced assets
		AddFunc("cleanup-assets", cleanupAssetsPhase)
}

func scanAssetRefsPhase(pctx *PreprocessContext) error {
	pctx.ReferencedAssets = make(map[string]bool)

	copySources := pctx.Book.GetCopySources()
	for _, src := range copySources {
		srcDir := staging.ResolveSourceDir(src.From, pctx.WorkspaceRoot)
		if srcDir == "" {
			continue
		}

		refs, err := staging.ScanAssetReferences(srcDir)
		if err != nil {
			pctx.Warn("asset scan failed: %v", err)
			pctx.ReferencedAssets = nil
			return nil
		}

		for ref := range refs {
			pctx.ReferencedAssets[ref] = true
		}
	}

	if len(pctx.ReferencedAssets) > 0 {
		pctx.Log.Infof("    Found %d asset references", len(pctx.ReferencedAssets))
	}
	return nil
}

func copySourcesPhase(pctx *PreprocessContext) error {
	cfg := &staging.CopyConfig{
		Book:             pctx.Book,
		WorkspaceRoot:    pctx.WorkspaceRoot,
		StagingDir:       pctx.StagingDir,
		IsPDF:            pctx.Mode.IsPDF(),
		SkipDrawioFiles:  pctx.Mode.ShouldSkipDrawioFiles(),
		ReferencedAssets: pctx.ReferencedAssets,
		FileMapper:       pctx.FileMapper,
		Logf:             pctx.Log.Infof,
		Warnf:            pctx.Warn,
	}
	return staging.CopyAllSources(cfg)
}

func buildFileIndexPhase(pctx *PreprocessContext) error {
	idx, err := staging.NewFileIndex(pctx.StagingDir)
	if err != nil {
		return err
	}
	pctx.FileIndex = idx
	pctx.Log.Infof("    Indexed %d files (%d markdown)", idx.FileCount(), idx.MarkdownCount())
	return nil
}

func convertAttrImagesPhase(pctx *PreprocessContext) error {
	return content.ProcessAttrListImages(pctx.FileIndex, pctx.Log.Infof)
}

func linkTranslationPhase(pctx *PreprocessContext) error {
	tm := linking.NewTranslationMap(
		pctx.WorkspaceRoot,
		pctx.StagingDir,
		pctx.Mode.ShouldStripExternalLinks(),
		pctx.Log.Infof,
	)

	// Transfer file mappings from copy phase
	if sfm, ok := pctx.FileMapper.(*staging.SimpleFileMap); ok {
		for _, pair := range sfm.AllMappings() {
			tm.AddFileMapping(pair[0], pair[1])
		}
	}

	if err := tm.BuildTranslations(pctx.Book.SiteURL); err != nil {
		return err
	}

	return tm.ApplyAllTranslations()
}

// defaultExecutor is the shared command executor for all pipeline phases.
var defaultExecutor content.CommandExecutor = content.ToolCommandExecutor{}

func executeCommandsPhase(pctx *PreprocessContext) error {
	// Resolve pre-built fragments directory if module moniker is available
	fragmentsDir := resolveFragmentsDir(pctx.WorkspaceRoot, pctx.Moniker)

	outputs, err := content.ExecuteCommands(
		pctx.Ctx,
		pctx.Book,
		pctx.WorkspaceRoot,
		pctx.StagingDir,
		defaultExecutor,
		pctx.Log.Infof,
		fragmentsDir,
	)
	if err != nil {
		return err
	}
	pctx.CmdOutputs = outputs
	return nil
}

// resolveFragmentsDir finds pre-built markdown command fragments.
// Returns empty string if no fragments directory exists.
func resolveFragmentsDir(workspaceRoot, moniker string) string {
	if moniker == "" {
		return ""
	}
	dir := paths.MarkdownCommandsBuildOutputPath(workspaceRoot, moniker)
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return dir
	}
	return ""
}

func navigationStructurePhase(pctx *PreprocessContext) error {
	if err := navigation.EnsureRootIndex(pctx.Book, pctx.StagingDir, pctx.Log.Infof); err != nil {
		return err
	}
	return navigation.EnsureNavigationStructure(pctx.Book, pctx.StagingDir, pctx.Log.Infof)
}

func inlineContentPhase(pctx *PreprocessContext) error {
	return content.InsertInlineContent(
		pctx.Book,
		pctx.StagingDir,
		pctx.CmdOutputs,
		pctx.Log.Infof,
		pctx.Warn,
	)
}

func commandHelpPhase(pctx *PreprocessContext) error {
	return content.ProcessCommandMarkers(
		pctx.Ctx,
		pctx.FileIndex,
		pctx.StagingDir,
		pctx.WorkspaceRoot,
		defaultExecutor,
		pctx.Log.Infof,
		pctx.Warn,
	)
}

func navigationMacrosPhase(pctx *PreprocessContext) error {
	if pctx.Mode.ShouldStripMacros() {
		if err := navigation.StripNavTitles(pctx.FileIndex, pctx.Log.Infof); err != nil {
			return err
		}
		return navigation.StripMacros(pctx.FileIndex, pctx.Log.Infof)
	}

	if pctx.Mode.ShouldInjectMacros() {
		return navigation.InjectMacros(pctx.FileIndex, pctx.StagingDir, pctx.Log.Infof)
	}

	return nil
}

func diagramProcessingPhase(pctx *PreprocessContext) error {
	var g errgroup.Group

	// Mermaid: sizing + scan + replace
	g.Go(func() error {
		if err := diagrams.ProcessMermaidSizing(pctx.FileIndex, pctx.Log.Infof); err != nil {
			return err
		}

		blocksByFile, statuses, err := diagrams.ScanForMermaidDiagrams(
			pctx.FileIndex,
			pctx.StagingDir,
			pctx.WorkspaceRoot,
			pctx.Log.Infof,
			nil, // debugf
		)
		if err != nil {
			return err
		}

		return diagrams.ReplaceMermaidBlocksWithImages(
			blocksByFile,
			statuses,
			pctx.Mode.ExtraPathPrefix(),
			pctx.Log.Infof,
		)
	})

	// Structurizr: independent of mermaid
	g.Go(func() error {
		return diagrams.ProcessStructurizrDiagrams(
			pctx.FileIndex,
			pctx.StagingDir,
			pctx.WorkspaceRoot,
			pctx.Mode.ExtraPathPrefix(),
			pctx.Log.Infof,
			pctx.Warn,
			nil, // debugf
		)
	})

	// PlantUML: independent of mermaid and structurizr
	g.Go(func() error {
		blocksByFile, statuses, err := diagrams.ScanForPlantUMLDiagrams(
			pctx.FileIndex,
			pctx.StagingDir,
			pctx.WorkspaceRoot,
			pctx.Log.Infof,
			nil, // debugf
		)
		if err != nil {
			return err
		}

		return diagrams.ReplacePlantUMLBlocksWithImages(
			blocksByFile,
			statuses,
			pctx.Mode.ExtraPathPrefix(),
			pctx.Log.Infof,
		)
	})

	return g.Wait()
}

func imageConstraintsPhase(pctx *PreprocessContext) error {
	return content.ProcessImageWidthConstraints(pctx.FileIndex, pctx.Log.Infof)
}

func fixBrokenLinksPhase(pctx *PreprocessContext) error {
	return cleanup.FixBrokenInternalLinks(
		pctx.FileIndex,
		pctx.StagingDir,
		pctx.Book.SiteURL,
		pctx.Log.Infof,
	)
}

func cleanupAssetsPhase(pctx *PreprocessContext) error {
	return cleanup.CleanupUnreferencedAssets(
		pctx.StagingDir,
		pctx.Log.Infof,
		pctx.Warn,
	)
}
