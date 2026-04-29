// Boundary lint for os.Getenv (and friends) per secrets-and-config spec §4.1.
//
// Allowlisted packages may legitimately read from process env / flags; every
// other Go file under backend/ must consume configuration through
// internal/platform/config. Run via `make lint-getenv-boundary` (which also
// participates in `make lint-config`).
//
// Allowlist:
//   backend/internal/platform/config/...
//   backend/internal/platform/secrets/...
//   backend/cmd/api/...
//   backend/cmd/worker/...
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type violation struct {
	path string
	line int
	call string
}

var defaultAllowlist = []string{
	"backend/internal/platform/config/",
	"backend/internal/platform/secrets/",
	"backend/cmd/api/",
	"backend/cmd/worker/",
}

// defaultBannedCalls intentionally focuses on the os.Getenv family per
// spec §6 C-7 (the actual AC). The spec text in §4.1 also mentions
// flag.String, but every CLI binary parses its own flags by nature, so
// catching it here would generate false positives for codegen and lint
// tools without raising real risk. Runtime configuration flows through
// internal/platform/config either way.
var defaultBannedCalls = map[string][]string{
	"os": {"Getenv", "LookupEnv", "Environ", "ExpandEnv"},
}

func main() {
	var (
		root      string
		allowExtra string
	)
	flag.StringVar(&root, "root", "backend", "directory to scan (relative to repo root)")
	flag.StringVar(&allowExtra, "allow-extra", "", "comma-separated extra allowlisted prefixes")
	flag.Parse()

	allowlist := append([]string(nil), defaultAllowlist...)
	if allowExtra != "" {
		for _, p := range strings.Split(allowExtra, ",") {
			if p = strings.TrimSpace(p); p != "" {
				allowlist = append(allowlist, p)
			}
		}
	}

	violations, err := scan(root, allowlist)
	if err != nil {
		fmt.Fprintf(os.Stderr, "getenv_boundary: %v\n", err)
		os.Exit(2)
	}
	if len(violations) == 0 {
		return
	}
	sort.Slice(violations, func(i, j int) bool {
		if violations[i].path == violations[j].path {
			return violations[i].line < violations[j].line
		}
		return violations[i].path < violations[j].path
	})
	fmt.Fprintln(os.Stderr, "FAIL: os.Getenv outside platform/config allowlist (spec §4.1).")
	fmt.Fprintln(os.Stderr, "Allowed prefixes:")
	for _, p := range allowlist {
		fmt.Fprintf(os.Stderr, "  - %s\n", p)
	}
	fmt.Fprintln(os.Stderr, "Violations:")
	for _, v := range violations {
		fmt.Fprintf(os.Stderr, "  %s:%d  %s\n", v.path, v.line, v.call)
	}
	os.Exit(1)
}

func scan(root string, allowlist []string) ([]violation, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	repoRoot, err := filepath.Abs(".")
	if err != nil {
		return nil, err
	}
	var out []violation
	err = filepath.WalkDir(abs, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == "node_modules" || name == "generated" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			rel = path
		}
		rel = filepath.ToSlash(rel)
		for _, prefix := range allowlist {
			if strings.HasPrefix(rel, prefix) {
				return nil
			}
		}
		out = append(out, scanFile(rel, path)...)
		return nil
	})
	return out, err
}

func scanFile(rel, abs string) []violation {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, abs, nil, parser.ParseComments)
	if err != nil {
		return []violation{{path: rel, line: 0, call: "parse error: " + err.Error()}}
	}
	var out []violation
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		banned, hasBanned := defaultBannedCalls[ident.Name]
		if !hasBanned {
			return true
		}
		for _, fn := range banned {
			if sel.Sel.Name == fn {
				pos := fset.Position(call.Pos())
				out = append(out, violation{path: rel, line: pos.Line, call: ident.Name + "." + sel.Sel.Name})
				break
			}
		}
		return true
	})
	return out
}
