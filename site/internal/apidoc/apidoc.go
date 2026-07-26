// Package apidoc extracts a component package's public API from its source,
// so each documentation page can render a reference that cannot drift.
//
// Every loom component package follows the same shape - New/Root plus part
// functions returning templ.Component, option constructors returning Option,
// pre-baked option vars, typed string enums whose default carries a "default"
// line comment, and fail-loud error vars. That regularity means the AST alone
// is enough: declarations are bucketed by their written return type, with no
// type checking and no dependencies beyond the standard library.
package apidoc

import (
	"errors"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// ImportBase is the module path every component package hangs off.
const ImportBase = "github.com/pietjan/loom"

// optionSuffix names the option types: Option itself and the narrower ones a
// few packages add (SeriesOption, EdgeOption, NodeOption).
const optionSuffix = "Option"

// errNoPackages marks a root directory that holds no component packages,
// which means the site was pointed somewhere unexpected.
var errNoPackages = errors.New("apidoc: no component packages found under")

// API is one component package's public surface.
type API struct {
	Name       string // package name, e.g. "button"
	ImportPath string // e.g. "github.com/pietjan/loom/button"
	Doc        string // package doc synopsis

	Components []Decl  // funcs returning templ.Component, in source order
	Options    []Group // option constructors and presets, by option type
	Values     []Group // typed constant groups
	Errors     []Decl  // exported error vars
}

// Decl is a single documented declaration.
type Decl struct {
	Name      string
	Signature string // func signature, or a constant's value as written
	Value     string // a string constant's unquoted value, e.g. address-book
	Doc       string // first sentence of the doc comment
	Default   bool   // constant marked with a "default" line comment
}

// Group is a named run of declarations - the constructors for one option
// type, or the constants of one enum. Groups are complete: how many of a
// long one to show is the caller's decision.
type Group struct {
	Name  string // "Option", "SeriesOption", "Variant", …
	Doc   string
	Decls []Decl
}

// Load parses every component package under root, keyed by directory name.
// A directory is a component package when it declares an Option type, which
// is what distinguishes the ~46 components from cmd/, internal/ and testdata/.
func Load(root string) (map[string]*API, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("apidoc: read %s: %w", root, err)
	}
	apis := make(map[string]*API)
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		switch e.Name() {
		case "cmd", "internal", "site", "testdata":
			continue
		}
		api, err := parsePackage(filepath.Join(root, e.Name()))
		if err != nil {
			return nil, err
		}
		if api != nil {
			apis[e.Name()] = api
		}
	}
	if len(apis) == 0 {
		return nil, fmt.Errorf("%w: %s", errNoPackages, root)
	}
	return apis, nil
}

// parsePackage reads one directory, returning nil when it is not a component
// package rather than an error - Load walks every sibling directory.
func parsePackage(dir string) (*API, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("apidoc: read %s: %w", dir, err)
	}
	name := filepath.Base(dir)
	fset := token.NewFileSet()

	// Parsed a file at a time rather than with parser.ParseDir, which is
	// deprecated as of Go 1.25 - as is the ast.Package it returns.
	var files []*ast.File
	for _, e := range entries {
		fname := e.Name()
		if e.IsDir() || !strings.HasSuffix(fname, ".go") || strings.HasSuffix(fname, "_test.go") {
			continue
		}
		path := filepath.Join(dir, fname)
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("apidoc: parse %s: %w", path, err)
		}
		// A directory whose package is named something else is not the
		// component package this slug refers to.
		if f.Name.Name != name {
			continue
		}
		files = append(files, f)
	}
	if len(files) == 0 {
		return nil, nil
	}
	d, err := doc.NewFromFiles(fset, files, ImportBase+"/"+name, doc.PreserveAST)
	if err != nil {
		return nil, fmt.Errorf("apidoc: document %s: %w", dir, err)
	}

	optionTypes := optionTypeNames(d)
	if len(optionTypes) == 0 {
		return nil, nil
	}
	api := &API{
		Name:       name,
		ImportPath: ImportBase + "/" + name,
		Doc:        firstSentence(d.Doc),
	}
	collectFuncs(api, d, fset, optionTypes)
	collectVars(api, d, fset, optionTypes)
	collectValues(api, d, fset)
	sortGroups(api)
	return api, nil
}

// optionTypeNames returns the package's option types - Option, plus the
// narrower ones a few packages add (chart.SeriesOption, diagram.EdgeOption
// and NodeOption). Matching on the suffix picks those up without a list.
func optionTypeNames(d *doc.Package) map[string]string {
	types := make(map[string]string)
	for _, t := range d.Types {
		if strings.HasSuffix(t.Name, optionSuffix) {
			types[t.Name] = firstSentence(t.Doc)
		}
	}
	return types
}

// collectFuncs buckets top-level funcs by their written result type: a
// templ.Component is a component, an option type is an option constructor.
//
// This is what keeps the two awkward cases honest without special-casing
// either. diagram.Node(id string, ...NodeOption) templ.Component is a
// component part and lands with the components, while the
// Node(ctx, ...) (*html.Node, error) builder that 32 packages export is
// neither and drops out.
func collectFuncs(api *API, d *doc.Package, fset *token.FileSet, optionTypes map[string]string) {
	byType := make(map[string][]Decl)
	for _, f := range allFuncs(d) {
		if f.Decl == nil || !ast.IsExported(f.Name) {
			continue
		}
		result := soleResultType(f.Decl.Type)
		decl := Decl{
			Name:      f.Name,
			Signature: signature(fset, f.Decl),
			Doc:       firstSentence(f.Doc),
		}
		switch {
		case result == "templ.Component":
			api.Components = append(api.Components, decl)
		case isOptionType(result, optionTypes):
			byType[result] = append(byType[result], decl)
		}
	}
	for name, decls := range byType {
		api.Options = append(api.Options, Group{
			Name:  name,
			Doc:   optionTypes[name],
			Decls: decls,
		})
	}
}

// collectVars picks up the pre-baked options (button.Primary = WithVariant(...))
// and the exported errors. The shared Class/ID/Attr triple is deliberately
// skipped: it is byte-identical in all 46 packages and is documented once, in
// the rendered "Common options" note, rather than repeated on every page.
func collectVars(api *API, d *doc.Package, fset *token.FileSet, optionTypes map[string]string) {
	presets := make(map[string][]Decl)
	for _, v := range d.Vars {
		for _, spec := range valueSpecs(v) {
			for i, name := range spec.Names {
				if !ast.IsExported(name.Name) || i >= len(spec.Values) {
					continue
				}
				value := spec.Values[i]
				if isCommonOption(value) {
					// Class/ID/Attr: rendered from the shared blurb instead.
					continue
				}
				call, ok := value.(*ast.CallExpr)
				if !ok {
					continue
				}
				if isErrorsNew(call) {
					api.Errors = append(api.Errors, Decl{
						Name:      name.Name,
						Signature: name.Name,
						Doc:       declDoc(v, spec),
					})
					continue
				}
				// A preset is a call to one of this package's own option
				// constructors, so its type is that constructor's result.
				fn, ok := call.Fun.(*ast.Ident)
				if !ok {
					continue
				}
				typ := optionResultOf(d, fn.Name, optionTypes)
				if typ == "" {
					continue
				}
				presets[typ] = append(presets[typ], Decl{
					Name:      name.Name,
					Signature: name.Name,
					Doc:       presetDoc(fset, declDoc(v, spec), call),
				})
			}
		}
	}
	for typ, decls := range presets {
		api.Options = appendToGroup(api.Options, typ, optionTypes[typ], decls)
	}
}

// collectValues gathers each typed constant group - Variant, Size, Color -
// marking the value whose line comment says it is the default.
func collectValues(api *API, d *doc.Package, fset *token.FileSet) {
	for _, t := range d.Types {
		if strings.HasSuffix(t.Name, optionSuffix) || t.Name == "Config" {
			continue
		}
		var decls []Decl
		for _, c := range t.Consts {
			for _, spec := range valueSpecs(c) {
				for i, name := range spec.Names {
					if !ast.IsExported(name.Name) {
						continue
					}
					decl := Decl{
						Name: name.Name,
						// Only the per-line comment: a const block's own doc
						// describes the group, and would repeat on every row.
						Doc:     firstSentence(spec.Doc.Text()),
						Default: isDefaultMarked(spec),
					}
					if i < len(spec.Values) {
						// Only string enums have a value worth showing; an
						// iota block's numbering is an implementation detail,
						// and its later lines carry no expression at all.
						if v := exprString(fset, spec.Values[i]); v != "iota" {
							decl.Signature = v
							decl.Value, _ = strconv.Unquote(v)
						}
					}
					decls = append(decls, decl)
				}
			}
		}
		if len(decls) == 0 {
			continue
		}
		api.Values = append(api.Values, Group{
			Name:  t.Name,
			Doc:   firstSentence(t.Doc),
			Decls: decls,
		})
	}
}

// sortGroups puts every list in reading order. Components keep source order,
// which is composition order - table reads New, Header, Body, Row, Column,
// Cell rather than alphabetically scrambled. Options lead with the presets,
// since those are the idiomatic call form.
func sortGroups(api *API) {
	for i := range api.Options {
		decls := api.Options[i].Decls
		sort.SliceStable(decls, func(a, b int) bool {
			pa, pb := isPreset(decls[a]), isPreset(decls[b])
			if pa != pb {
				return pa
			}
			return decls[a].Name < decls[b].Name
		})
	}
	sort.Slice(api.Options, func(a, b int) bool {
		// The plain "Option" type first, then any narrower ones.
		if (api.Options[a].Name == optionSuffix) != (api.Options[b].Name == optionSuffix) {
			return api.Options[a].Name == optionSuffix
		}
		return api.Options[a].Name < api.Options[b].Name
	})
	sort.Slice(api.Values, func(a, b int) bool { return api.Values[a].Name < api.Values[b].Name })
	sort.Slice(api.Errors, func(a, b int) bool { return api.Errors[a].Name < api.Errors[b].Name })
}

// isPreset reports whether a decl is a pre-baked option value rather than a
// constructor - presets render as a bare name, constructors carry parens.
func isPreset(d Decl) bool { return !strings.Contains(d.Signature, "(") }

func appendToGroup(groups []Group, name, doc string, decls []Decl) []Group {
	for i := range groups {
		if groups[i].Name == name {
			groups[i].Decls = append(groups[i].Decls, decls...)
			return groups
		}
	}
	return append(groups, Group{Name: name, Doc: doc, Decls: decls})
}

// allFuncs returns every exported top-level func in declaration order.
//
// go/doc files a function under the type it returns, so the option
// constructors live in the Option type's Funcs rather than in Package.Funcs;
// only those returning a type from another package - templ.Component - stay
// at the top level. Merging both and re-sorting by position undoes that
// grouping and restores source order, which for a multi-part package is the
// composition order: table reads New, Header, Body, Row, Column, Cell.
func allFuncs(d *doc.Package) []*doc.Func {
	funcs := append([]*doc.Func(nil), d.Funcs...)
	for _, t := range d.Types {
		funcs = append(funcs, t.Funcs...)
	}
	sort.Slice(funcs, func(a, b int) bool {
		if funcs[a].Decl == nil || funcs[b].Decl == nil {
			return false
		}
		return funcs[a].Decl.Pos() < funcs[b].Decl.Pos()
	})
	return funcs
}

// soleResultType returns the written type of a func's single result, or ""
// when it returns nothing or several values.
func soleResultType(ft *ast.FuncType) string {
	if ft.Results == nil || len(ft.Results.List) != 1 {
		return ""
	}
	if len(ft.Results.List[0].Names) > 1 {
		return ""
	}
	return typeString(ft.Results.List[0].Type)
}

func isOptionType(name string, optionTypes map[string]string) bool {
	_, ok := optionTypes[name]
	return ok
}

// optionResultOf reports the option type a named constructor returns, so a
// preset can be filed under the same group as the constructor that built it.
func optionResultOf(d *doc.Package, fn string, optionTypes map[string]string) string {
	for _, f := range allFuncs(d) {
		if f.Name != fn || f.Decl == nil {
			continue
		}
		if result := soleResultType(f.Decl.Type); isOptionType(result, optionTypes) {
			return result
		}
	}
	return ""
}

// presetDoc describes a preset. Most carry no comment of their own - they sit
// in a shared "Pre-baked options" var block - so fall back to the call that
// defines them, which says exactly what the preset does: WithVariant(VariantPrimary).
func presetDoc(fset *token.FileSet, docText string, call *ast.CallExpr) string {
	if s := firstSentence(docText); s != "" {
		return s
	}
	return "Shorthand for " + exprString(fset, call) + "."
}

// isCommonOption matches the Class/ID/Attr triple, whose right-hand side
// instantiates a shared generic: opts.Class[*Config].
func isCommonOption(expr ast.Expr) bool {
	idx, ok := expr.(*ast.IndexExpr)
	if !ok {
		return false
	}
	sel, ok := idx.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "opts"
}

func isErrorsNew(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "errors" && (sel.Sel.Name == "New" || sel.Sel.Name == "Errorf")
}

// isDefaultMarked reports whether a const spec carries the "// default"
// trailing comment the library uses to mark an enum's zero value.
func isDefaultMarked(spec *ast.ValueSpec) bool {
	if spec.Comment == nil {
		return false
	}
	for _, c := range spec.Comment.List {
		if strings.Contains(strings.ToLower(c.Text), "default") {
			return true
		}
	}
	return false
}

// valueSpecs returns every ValueSpec of a var or const declaration. A grouped
// declaration - var ( Submit = …; Reset = … ) - is a single doc.Value holding
// one spec per line, so all of them have to be walked.
func valueSpecs(v *doc.Value) []*ast.ValueSpec {
	if v.Decl == nil {
		return nil
	}
	specs := make([]*ast.ValueSpec, 0, len(v.Decl.Specs))
	for _, s := range v.Decl.Specs {
		if spec, ok := s.(*ast.ValueSpec); ok {
			specs = append(specs, spec)
		}
	}
	return specs
}

// declDoc prefers a declaration's own comment, falling back to the enclosing
// block's only when that block declares this one thing - otherwise the block
// comment ("Pre-baked options.") would be repeated on every row.
func declDoc(v *doc.Value, spec *ast.ValueSpec) string {
	if s := firstSentence(spec.Doc.Text()); s != "" {
		return s
	}
	if len(v.Decl.Specs) == 1 {
		return firstSentence(v.Doc)
	}
	return ""
}

// signature renders a func declaration as its name and parameters. The result
// type is dropped: it is what the declaration was bucketed by, so the group
// heading above the table already states it.
func signature(fset *token.FileSet, fn *ast.FuncDecl) string {
	params := *fn.Type
	params.Results = nil
	return fn.Name.Name + strings.TrimPrefix(exprString(fset, &params), "func")
}

func exprString(fset *token.FileSet, expr ast.Expr) string {
	var sb strings.Builder
	if err := printer.Fprint(&sb, fset, expr); err != nil {
		return ""
	}
	return sb.String()
}

// typeString renders a type expression, e.g. "templ.Component" or "Option".
func typeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		if pkg, ok := t.X.(*ast.Ident); ok {
			return pkg.Name + "." + t.Sel.Name
		}
	case *ast.StarExpr:
		return "*" + typeString(t.X)
	}
	return ""
}

// firstSentence trims a doc comment down to its opening sentence, dropping
// the leading "Name " that godoc convention puts there.
func firstSentence(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	// Stop at a blank line: package docs continue into usage examples.
	if i := strings.Index(text, "\n\n"); i >= 0 {
		text = text[:i]
	}
	text = strings.Join(strings.Fields(text), " ")
	if s := sentenceEnd(text); s > 0 {
		text = text[:s]
	}
	return text
}

// sentenceEnd finds the end of the first sentence, ignoring the periods in
// identifiers like button.Label and in abbreviations like "e.g.".
func sentenceEnd(text string) int {
	for i, r := range text {
		if r != '.' {
			continue
		}
		if i+1 >= len(text) {
			return i + 1
		}
		// A period inside an identifier is followed by a letter rather than
		// a space, so anything else is still mid-sentence.
		if text[i+1] != ' ' {
			continue
		}
		if isAbbreviation(text[:i+1]) {
			continue
		}
		// Real sentences resume with a capital; "5 in. of rain" does not.
		if next := strings.TrimLeft(text[i+1:], " "); next != "" && !isUpper(rune(next[0])) {
			continue
		}
		return i + 1
	}
	return 0
}

// isAbbreviation reports whether text ends in one of the short forms that
// appear in these doc comments, where the trailing period is not a full stop.
func isAbbreviation(text string) bool {
	for _, abbr := range []string{"e.g.", "i.e.", "etc.", "cf.", "vs."} {
		if strings.HasSuffix(strings.ToLower(text), abbr) {
			return true
		}
	}
	return false
}

func isUpper(r rune) bool { return r >= 'A' && r <= 'Z' }
