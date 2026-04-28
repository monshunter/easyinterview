package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Run is the top-level orchestrator: it re-renders the B1-AUTO block in
// openapi.yaml, then walks the resulting OpenAPI document into Go and TS
// artefacts. Output is byte-stable so `make codegen-check` (which runs Run
// and then `git diff --exit-code`) catches any drift.
func Run(openapiPath, conventionsPath, templatesDir, repoRoot string, verbose bool) error {
	conventions, err := loadConventions(conventionsPath)
	if err != nil {
		return fmt.Errorf("load conventions.yaml: %w", err)
	}

	if err := syncB1AutoBlock(openapiPath, conventions); err != nil {
		return fmt.Errorf("sync B1-AUTO block: %w", err)
	}

	doc, err := loadOpenAPI(openapiPath)
	if err != nil {
		return fmt.Errorf("load openapi.yaml: %w", err)
	}

	goOutDir := filepath.Join(repoRoot, "backend", "internal", "api", "generated")
	tsOutDir := filepath.Join(repoRoot, "frontend", "src", "api", "generated")
	if err := os.MkdirAll(goOutDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(tsOutDir, 0o755); err != nil {
		return err
	}

	if err := renderGo(doc, conventions, templatesDir, goOutDir); err != nil {
		return fmt.Errorf("render Go: %w", err)
	}
	if err := renderTS(doc, conventions, templatesDir, tsOutDir); err != nil {
		return fmt.Errorf("render TS: %w", err)
	}

	if verbose {
		fmt.Fprintln(os.Stderr, "codegen-openapi: rendered Go + TS artefacts")
	}
	return nil
}
