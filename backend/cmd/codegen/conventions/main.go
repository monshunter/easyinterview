// Command codegen-conventions renders the cross-language truth source
// (shared/conventions.yaml) into Go and TypeScript files under
// backend/internal/shared/ and frontend/src/lib/.
//
// Run via the root Makefile target: `make codegen-conventions`.
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	yamlPath := flag.String("yaml", "shared/conventions.yaml", "path to conventions.yaml")
	repoRoot := flag.String("repo-root", ".", "repository root for output paths")
	verbose := flag.Bool("v", false, "verbose output")
	flag.Parse()

	if err := Run(*yamlPath, *repoRoot, *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "codegen-conventions: %v\n", err)
		os.Exit(1)
	}
	if *verbose {
		fmt.Fprintln(os.Stderr, "codegen-conventions: ok")
	}
}
