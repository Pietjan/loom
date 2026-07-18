package loom_test

// Enforces the layering rule from the architecture:
//
//   - L1 primitives import no other component package.
//   - L2 composites import component packages only along explicitly
//     allowed one-directional edges (to read a Scope type), plus icon.
//   - Cross-package cooperation otherwise goes through ctx scopes and
//     post-passes, never by calling another component's Node().
//
// When adding a component, add it to l1 or l2 below; a new scope
// dependency must be added to allowedEdges deliberately.

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const module = "github.com/pietjan/loom"

var l1 = []string{
	"icon", "button", "badge", "heading", "text", "separator", "link", "card", "callout", "chart", "tooltip",
	"avatar", "breadcrumbs", "progress", "skeleton", "pagination", "stat", "timeline", "description", "kbd", "kanban",
	"carousel", "flash", "markdown",
}

var l2 = []string{
	"field", "input", "textarea", "checkbox", "radio", "toggle", "picker", "fieldset",
	"slider", "fileupload", "inputgroup",
	"modal", "dropdown", "popover", "accordion", "tabs", "navlist", "sidebar", "navbar", "header", "table",
}

// embeddable primitives may be composed by any component (their Node
// called directly): self-contained leaves with no scope coupling.
var embeddable = map[string]bool{
	"icon": true, "tooltip": true,
	"heading": true, "text": true, "link": true, "separator": true,
}

// allowedEdges lists the only permitted component→component imports,
// besides "anything may import icon".
var allowedEdges = map[string][]string{
	"input":      {"field", "inputgroup"},
	"textarea":   {"field"},
	"checkbox":   {"field"},
	"radio":      {"field"},
	"toggle":     {"field"},
	"picker":     {"field"},
	"slider":     {"field"},
	"fileupload": {"field"},
}

func TestComponentImportGraph(t *testing.T) {
	components := map[string]bool{}
	for _, p := range append(append([]string{}, l1...), l2...) {
		components[p] = true
	}
	isL1 := map[string]bool{}
	for _, p := range l1 {
		isL1[p] = true
	}

	for pkg := range components {
		if _, err := os.Stat(pkg); os.IsNotExist(err) {
			continue // not built yet
		}
		imports, err := packageImports(pkg)
		if err != nil {
			t.Fatalf("%s: %v", pkg, err)
		}
		for _, imp := range imports {
			if !strings.HasPrefix(imp, module+"/") {
				continue
			}
			rel := strings.TrimPrefix(imp, module+"/")
			if strings.HasPrefix(rel, "internal/") || rel == "theme" {
				continue
			}
			if !components[rel] {
				continue
			}
			if embeddable[rel] {
				continue // whitelisted for everyone
			}
			if isL1[pkg] {
				t.Errorf("L1 package %q imports component package %q — primitives import only internal/*", pkg, rel)
				continue
			}
			if !contains(allowedEdges[pkg], rel) {
				t.Errorf("package %q imports %q — not an allowed scope edge; components cooperate via ctx scopes and post-passes, not cross-package Node() calls", pkg, rel)
			}
		}
	}
}

func packageImports(dir string) ([]string, error) {
	fset := token.NewFileSet()
	var out []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, e.Name()), nil, parser.ImportsOnly)
		if err != nil {
			return nil, err
		}
		for _, imp := range f.Imports {
			out = append(out, strings.Trim(imp.Path.Value, `"`))
		}
	}
	return out, nil
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
