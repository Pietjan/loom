package apidoc_test

import (
	"strings"
	"testing"

	"github.com/pietjan/loom/site/internal/apidoc"
	"github.com/pietjan/loom/site/pages"
)

// root is the library tree, relative to this package's directory.
const root = "../../.."

func load(t *testing.T) map[string]*apidoc.API {
	t.Helper()
	apis, err := apidoc.Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return apis
}

// TestEveryPageHasAReference is the load-bearing check: every documented
// component must resolve to a package whose API could be extracted. It fails
// when a page is added without a package, when a package is renamed, and —
// most usefully — when a package stops following the New/Option shape the
// extraction relies on, which nothing else in the repo enforces.
func TestEveryPageHasAReference(t *testing.T) {
	apis := load(t)
	for _, p := range pages.All() {
		api, ok := apis[p.Slug]
		if !ok {
			t.Errorf("%s: no package found for documented page", p.Slug)
			continue
		}
		if len(api.Components) == 0 {
			t.Errorf("%s: no exported func returns templ.Component", p.Slug)
		}
		// Options are deliberately not required. A handful of packages —
		// table, kbd, description, inputgroup, dropdown, popover — have a
		// Config of nothing but opts.Common, so the shared Class/ID/Attr
		// triple is their entire option set and the group renders empty.
		if want := apidoc.ImportBase + "/" + p.Slug; api.ImportPath != want {
			t.Errorf("%s: import path = %q, want %q", p.Slug, api.ImportPath, want)
		}
	}
}

// TestClassification pins the buckets that are easy to get wrong: the shared
// Class/ID/Attr triple must stay out (it is documented once, in the rendered
// common-options note), presets must be listed alongside the constructors
// that build them, and enum defaults must be detected.
func TestClassification(t *testing.T) {
	api := load(t)["button"]

	var names []string
	for _, g := range api.Options {
		for _, d := range g.Decls {
			names = append(names, d.Name)
		}
	}
	joined := strings.Join(names, " ")
	for _, want := range []string{"Primary", "Ghost", "Tiny", "WithVariant", "Disabled", "Href"} {
		if !strings.Contains(joined, want) {
			t.Errorf("button options missing %q; got %v", want, names)
		}
	}
	for _, skip := range []string{"Class", "ID", "Attr"} {
		for _, name := range names {
			if name == skip {
				t.Errorf("button options should not include the shared %q", skip)
			}
		}
	}

	var defaults []string
	for _, g := range api.Values {
		for _, d := range g.Decls {
			if d.Default {
				defaults = append(defaults, d.Name)
			}
		}
	}
	want := []string{"SizeBase", "TypeButton", "VariantOutline"}
	if strings.Join(defaults, ",") != strings.Join(want, ",") {
		t.Errorf("button defaults = %v, want %v", defaults, want)
	}

	if len(api.Errors) != 2 {
		t.Errorf("button errors = %d, want 2 (%v)", len(api.Errors), api.Errors)
	}
}

// TestComponentOrder checks that parts are listed in composition order rather
// than alphabetically — a table's reference should read the way its markup is
// written, not start at Body.
func TestComponentOrder(t *testing.T) {
	var got []string
	for _, c := range load(t)["table"].Components {
		got = append(got, c.Name)
	}
	want := "New Header Body Row Column Cell"
	if strings.Join(got, " ") != want {
		t.Errorf("table components = %q, want %q", strings.Join(got, " "), want)
	}
}

// TestNodeBuilderExcluded covers the one genuine name collision: diagram.Node
// declares a component part, while 32 other packages export a Node(ctx, ...)
// builder that is an implementation detail. Bucketing by return type has to
// keep them apart.
func TestNodeBuilderExcluded(t *testing.T) {
	apis := load(t)

	var found bool
	for _, c := range apis["diagram"].Components {
		if c.Name == "Node" {
			found = true
		}
	}
	if !found {
		t.Error("diagram.Node should be listed as a component")
	}

	for _, c := range apis["button"].Components {
		if c.Name == "Node" {
			t.Error("button.Node(ctx, ...) is a builder, not a component")
		}
	}
}

// TestValuesAreComplete checks that groups are returned whole. Truncating is
// the renderer's job — the icon browser needs every name, and gets them from
// here rather than from a second source that could disagree.
func TestValuesAreComplete(t *testing.T) {
	for _, g := range load(t)["icon"].Values {
		if g.Name != "Name" {
			continue
		}
		if len(g.Decls) < 1000 {
			t.Errorf("icon.Name returned %d values, want the full set", len(g.Decls))
		}
		// The browser renders icon.New(icon.Name(Value)), so the value has to
		// be the unquoted constant rather than the Go literal.
		for _, d := range g.Decls {
			if d.Name != "AddressBook" {
				continue
			}
			if d.Value != "address-book" {
				t.Errorf("AddressBook value = %q, want %q", d.Value, "address-book")
			}
			if d.Signature != `"address-book"` {
				t.Errorf("AddressBook signature = %q, want the quoted literal", d.Signature)
			}
			return
		}
		t.Error("icon.AddressBook not found")
		return
	}
	t.Error("icon.Name value group not found")
}
