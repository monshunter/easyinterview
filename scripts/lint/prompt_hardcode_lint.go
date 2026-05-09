// Hardcoded-prompt boundary lint per F3 prompt-rubric-registry plan §1.6.
//
// Walks the named backend/internal/* roots and rejects assignments of the
// form `prompt := "..."`, `Prompt = "..."`, `systemMessage := "..."`, etc.
// where the right-hand side is a raw string literal (backticks) or a
// long quoted string. Real prompt bodies belong in `config/prompts/`, not
// in business packages.
//
// Run via the Python wrapper `scripts/lint/prompt_hardcode_lint.py` so the
// CLI shape stays consistent across the lint suite. The Python wrapper
// also feeds tests their negative fixtures.
//
// Allowlist:
//   - *_test.go files
//   - any path containing /testdata/ or /fixtures/
package main

import (
	"encoding/json"
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

var defaultRoots = []string{
	"backend/internal/practice",
	"backend/internal/report",
	"backend/internal/resume",
	"backend/internal/debrief",
	"backend/internal/targetjob",
}

type violation struct {
	Path    string `json:"path"`
	Line    int    `json:"line"`
	VarName string `json:"var"`
	Snippet string `json:"snippet"`
}

func main() {
	var (
		rootsFlag  string
		outputJSON bool
	)
	flag.StringVar(&rootsFlag, "roots", strings.Join(defaultRoots, ","), "comma-separated roots to scan")
	flag.BoolVar(&outputJSON, "json", false, "emit JSON instead of text")
	flag.Parse()

	var all []violation
	for _, root := range strings.Split(rootsFlag, ",") {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		violations, err := scanRoot(root)
		if err != nil {
			fmt.Fprintf(os.Stderr, "prompt_hardcode_lint: %s: %v\n", root, err)
			os.Exit(2)
		}
		all = append(all, violations...)
	}

	if len(all) == 0 {
		if outputJSON {
			fmt.Println("[]")
		}
		return
	}

	sort.Slice(all, func(i, j int) bool {
		if all[i].Path == all[j].Path {
			return all[i].Line < all[j].Line
		}
		return all[i].Path < all[j].Path
	})

	if outputJSON {
		_ = json.NewEncoder(os.Stdout).Encode(all)
	} else {
		fmt.Fprintln(os.Stderr, "FAIL: hardcoded prompt assignments detected; move them to config/prompts/")
		for _, v := range all {
			fmt.Fprintf(os.Stderr, "  %s:%d  %s = %s\n", v.Path, v.Line, v.VarName, v.Snippet)
		}
	}
	os.Exit(1)
}

func scanRoot(root string) ([]violation, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}
	repoRoot, _ := filepath.Abs(".")
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
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		rel, _ := filepath.Rel(repoRoot, path)
		rel = filepath.ToSlash(rel)
		if isAllowlisted(rel) {
			return nil
		}
		out = append(out, scanFile(rel, path)...)
		return nil
	})
	return out, err
}

func isAllowlisted(rel string) bool {
	if strings.HasSuffix(rel, "_test.go") {
		return true
	}
	if strings.Contains(rel, "/testdata/") || strings.Contains(rel, "/fixtures/") {
		return true
	}
	return false
}

func scanFile(rel, abs string) []violation {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, abs, nil, 0)
	if err != nil {
		return nil
	}
	var out []violation
	ast.Inspect(file, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.AssignStmt:
			out = append(out, checkAssignStmt(rel, fset, s)...)
		case *ast.GenDecl:
			out = append(out, checkGenDecl(rel, fset, s)...)
		}
		return true
	})
	return out
}

func checkAssignStmt(rel string, fset *token.FileSet, as *ast.AssignStmt) []violation {
	if len(as.Lhs) != 1 || len(as.Rhs) != 1 {
		return nil
	}
	ident, ok := as.Lhs[0].(*ast.Ident)
	if !ok {
		return nil
	}
	if !isPromptIdentifier(ident.Name) {
		return nil
	}
	lit, ok := as.Rhs[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return nil
	}
	if !isPromptBodyLiteral(lit.Value) {
		return nil
	}
	pos := fset.Position(as.Pos())
	return []violation{{Path: rel, Line: pos.Line, VarName: ident.Name, Snippet: shortSnippet(lit.Value)}}
}

func checkGenDecl(rel string, fset *token.FileSet, gd *ast.GenDecl) []violation {
	if gd.Tok != token.VAR && gd.Tok != token.CONST {
		return nil
	}
	var out []violation
	for _, spec := range gd.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for i, name := range vs.Names {
			if !isPromptIdentifier(name.Name) {
				continue
			}
			if i >= len(vs.Values) {
				continue
			}
			lit, ok := vs.Values[i].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				continue
			}
			if !isPromptBodyLiteral(lit.Value) {
				continue
			}
			pos := fset.Position(name.Pos())
			out = append(out, violation{Path: rel, Line: pos.Line, VarName: name.Name, Snippet: shortSnippet(lit.Value)})
		}
	}
	return out
}

func isPromptIdentifier(name string) bool {
	if name == "systemMessage" {
		return true
	}
	if strings.HasPrefix(name, "Prompt") {
		return true
	}
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, "prompt")
}

func isPromptBodyLiteral(value string) bool {
	if value == "" {
		return false
	}
	// Raw string literals are always treated as prompt bodies.
	if strings.HasPrefix(value, "`") {
		return true
	}
	// Long quoted strings (>= 60 chars including quotes) or strings with
	// embedded newlines are flagged. Short version-string-style assignments
	// (for example PromptVersion = "v0.1.0") stay below the threshold.
	if len(value) >= 60 {
		return true
	}
	if strings.Contains(value, "\\n") {
		return true
	}
	return false
}

func shortSnippet(value string) string {
	if len(value) > 60 {
		return value[:57] + "..."
	}
	return value
}
