package loom_test

// The accessibility contract harness: every composite registers a
// representative render here, and every render is checked against the
// invariants that pulseui silently violated (dead references between
// generated markup and behavior). New composites MUST add themselves.

import (
	"strings"
	"testing"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/checkbox"
	"github.com/pietjan/loom/field"
	"github.com/pietjan/loom/fieldset"
	"github.com/pietjan/loom/input"
	"github.com/pietjan/loom/inputgroup"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/testutil"
	"github.com/pietjan/loom/picker"
	"github.com/pietjan/loom/radio"
	"github.com/pietjan/loom/textarea"
	"github.com/pietjan/loom/toggle"
)

// composites lists one representative render per composite component.
var composites = map[string]func() templ.Component{
	"field-input": func() templ.Component {
		return testutil.WithChildren(field.Root(field.Error("bad"), field.Required()), testutil.Sequence(
			testutil.WithChildren(field.Label(), testutil.Text("Email")),
			input.New(input.Name("email")),
			testutil.WithChildren(field.Description(), testutil.Text("hint")),
		))
	},
	"field-textarea": func() templ.Component {
		return testutil.WithChildren(field.Root(), testutil.Sequence(
			testutil.WithChildren(field.Label(), testutil.Text("Bio")),
			textarea.New(textarea.Name("bio")),
		))
	},
	"field-inputgroup": func() templ.Component {
		return testutil.WithChildren(field.Root(field.Error("bad")), testutil.Sequence(
			testutil.WithChildren(field.Label(), testutil.Text("Website")),
			testutil.WithChildren(inputgroup.New(), testutil.Sequence(
				testutil.WithChildren(inputgroup.Addon(), testutil.Text("https://")),
				input.New(input.Name("site")),
			)),
		))
	},
	"field-picker": func() templ.Component {
		return testutil.WithChildren(field.Root(), testutil.Sequence(
			testutil.WithChildren(field.Label(), testutil.Text("Pet")),
			testutil.WithChildren(picker.New(picker.Name("pet"), picker.Placeholder("Choose")),
				testutil.WithChildren(picker.Item("cat"), testutil.Text("Cat"))),
		))
	},
	"radio-group": func() templ.Component {
		return testutil.WithChildren(radio.Group(radio.Name("plan"), radio.Legend("Plan")), testutil.Sequence(
			radio.New(radio.Value("free"), radio.Label("Free")),
			radio.New(radio.Value("pro"), radio.Label("Pro"), radio.Checked()),
		))
	},
	"fieldset": func() templ.Component {
		return testutil.WithChildren(fieldset.New(fieldset.Legend("Address"), fieldset.Disabled()), testutil.Sequence(
			testutil.WithChildren(field.Root(), testutil.Sequence(
				testutil.WithChildren(field.Label(), testutil.Text("Street")),
				input.New(input.Name("street")),
			)),
		))
	},
	"checkbox-labeled": func() templ.Component {
		return checkbox.New(checkbox.Name("terms"), checkbox.Label("I agree"))
	},
	"toggle-labeled": func() templ.Component {
		return toggle.New(toggle.Name("notify"), toggle.Label("Email me"))
	},
}

func TestContracts(t *testing.T) {
	for name, build := range composites {
		t.Run(name, func(t *testing.T) {
			tree := testutil.Parse(t, testutil.Render(t, build()))
			assertReferencesResolve(t, tree)
			assertUniqueIDs(t, tree)
			assertControlsAreLabelled(t, tree)
		})
	}
}

// assertReferencesResolve checks every ID-reference attribute points at an
// element that exists in the same document — the invariant whose violation
// made pulseui's navlist silently dead.
func assertReferencesResolve(t *testing.T, tree *html.Node) {
	t.Helper()
	ids := map[string]bool{}
	for _, n := range dom.FindAll(tree, dom.ByAttr("id")) {
		ids[dom.GetAttr(n, "id")] = true
	}

	refAttrs := []string{"for", "aria-describedby", "aria-labelledby", "aria-controls", "commandfor", "popovertarget", "interestfor", "list", "anchor"}
	for _, n := range dom.FindAll(tree, func(n *html.Node) bool { return n.Type == html.ElementNode }) {
		for _, attr := range refAttrs {
			val := dom.GetAttr(n, attr)
			if val == "" {
				continue
			}
			for _, ref := range strings.Fields(val) {
				if !ids[ref] {
					t.Errorf("<%s %s=%q>: id %q does not exist in the document", n.Data, attr, val, ref)
				}
			}
		}
	}
}

// assertUniqueIDs checks no id appears twice.
func assertUniqueIDs(t *testing.T, tree *html.Node) {
	t.Helper()
	seen := map[string]bool{}
	for _, n := range dom.FindAll(tree, dom.ByAttr("id")) {
		id := dom.GetAttr(n, "id")
		if seen[id] {
			t.Errorf("duplicate id %q", id)
		}
		seen[id] = true
	}
}

// assertControlsAreLabelled checks every form control has an accessible
// name: a label[for] pointing at it, a wrapping label, an aria-label, or
// an aria-labelledby.
func assertControlsAreLabelled(t *testing.T, tree *html.Node) {
	t.Helper()
	labelFor := map[string]bool{}
	for _, l := range dom.FindAll(tree, dom.ByTag(atom.Label)) {
		if f := dom.GetAttr(l, "for"); f != "" {
			labelFor[f] = true
		}
	}

	controls := dom.FindAll(tree, dom.ByMarker("input", "textarea", "picker", "checkbox", "radio", "toggle"))
	for _, c := range controls {
		if dom.GetAttr(c, "aria-label") != "" || dom.GetAttr(c, "aria-labelledby") != "" {
			continue
		}
		if labelFor[dom.GetAttr(c, "id")] {
			continue
		}
		if inside(c, dom.ByTag(atom.Label)) {
			continue
		}
		t.Errorf("<%s data-ui=%q name=%q> has no accessible name", c.Data, dom.MarkerName(c), dom.GetAttr(c, "name"))
	}
}

func inside(n *html.Node, m dom.Matcher) bool {
	for p := n.Parent; p != nil; p = p.Parent {
		if p.Type == html.ElementNode && m(p) {
			return true
		}
	}
	return false
}
