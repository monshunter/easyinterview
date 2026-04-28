// Command codegen-openapi renders the easyinterview HTTP contract
// (openapi/openapi.yaml) into Go and TypeScript artefacts under
// `backend/internal/api/generated/` and `frontend/src/api/generated/`.
//
// Inputs:
//   - openapi/openapi.yaml  (hand-authored contract; B1-AUTO block synced from B1)
//   - shared/conventions.yaml  (B1 truth source for enums / errors / structures)
//   - openapi/templates/{go,ts}/*.tmpl  (text/template renderers)
//
// Outputs (idempotent — drift is detected by `make codegen-check`):
//   - backend/internal/api/generated/{types,server,spec,openapi}.gen.go
//   - frontend/src/api/generated/{types,client,spec}.ts
//
// Run via the root Makefile target: `make codegen-openapi`.
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	openapiPath := flag.String("openapi", "openapi/openapi.yaml", "path to openapi.yaml")
	conventionsPath := flag.String("conventions", "shared/conventions.yaml", "path to shared/conventions.yaml (B1 truth source)")
	templatesDir := flag.String("templates", "openapi/templates", "path to template directory")
	repoRoot := flag.String("repo-root", ".", "repository root for output paths")
	verbose := flag.Bool("v", false, "verbose output")
	flag.Parse()

	if err := Run(*openapiPath, *conventionsPath, *templatesDir, *repoRoot, *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "codegen-openapi: %v\n", err)
		os.Exit(1)
	}
	if *verbose {
		fmt.Fprintln(os.Stderr, "codegen-openapi: ok")
	}
}
