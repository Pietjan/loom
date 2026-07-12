package dom_test

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
)

func tree() *html.Node {
	// <div data-ui="field">
	//   <label data-ui="field-label">
	//   <div>                          <- plain wrapper, transparent
	//     <input data-ui="input">
	//   <div data-ui="card">           <- another component, opaque
	//     <button>
	//   <button>                       <- the button FindShallow should see
	root := dom.El(atom.Div, dom.Marker("field"))
	label := dom.El(atom.Label, dom.Marker("field-label"))
	wrapper := dom.El(atom.Div)
	input := dom.El(atom.Input, dom.Marker("input"))
	card := dom.El(atom.Div, dom.Marker("card"))
	cardButton := dom.El(atom.Button)
	outerButton := dom.El(atom.Button)

	root.AppendChild(label)
	wrapper.AppendChild(input)
	root.AppendChild(wrapper)
	card.AppendChild(cardButton)
	root.AppendChild(card)
	root.AppendChild(outerButton)
	return root
}

func TestFindDescendsEverywhere(t *testing.T) {
	root := tree()
	if got := dom.Find(root, dom.ByTag(atom.Button)); got == nil {
		t.Fatal("expected a button")
	}
	if got := len(dom.FindAll(root, dom.ByTag(atom.Button))); got != 2 {
		t.Fatalf("expected 2 buttons, got %d", got)
	}
}

func TestFindShallowSkipsOtherComponents(t *testing.T) {
	root := tree()

	// The card's button is invisible to shallow queries...
	got := dom.FindShallow(root, dom.ByTag(atom.Button))
	if got == nil {
		t.Fatal("expected the outer button")
	}
	if got.Parent != root {
		t.Fatal("shallow query descended into another component")
	}

	// ...but marker-identified nodes are still found through plain wrappers.
	if dom.FindShallow(root, dom.ByMarker("input")) == nil {
		t.Fatal("expected to find input through the plain wrapper")
	}

	// A matching component root is returned, not skipped.
	if dom.FindShallow(root, dom.ByMarker("card")) == nil {
		t.Fatal("expected to find the card itself")
	}
}

func TestFindAllShallow(t *testing.T) {
	root := tree()
	buttons := dom.FindAllShallow(root, dom.ByTag(atom.Button))
	if len(buttons) != 1 {
		t.Fatalf("expected 1 shallow button, got %d", len(buttons))
	}
}

func TestAttrHelpers(t *testing.T) {
	n := dom.El(atom.Input, dom.Attr("type", "text"))

	dom.SetAttr(n, "type", "email")
	if got := dom.GetAttr(n, "type"); got != "email" {
		t.Fatalf("SetAttr: got %q", got)
	}

	dom.AddAttr(n, "class", "a")
	dom.AddAttr(n, "class", "b")
	if got := dom.GetAttr(n, "class"); got != "a b" {
		t.Fatalf("AddAttr: got %q", got)
	}

	if !dom.HasAttr(n, "class") {
		t.Fatal("HasAttr: expected true")
	}
	dom.DelAttr(n, "class")
	if dom.HasAttr(n, "class") {
		t.Fatal("DelAttr: expected removed")
	}
}

func TestFormat(t *testing.T) {
	root := dom.El(atom.Div, dom.Marker("card"))
	h := dom.El(atom.H2)
	h.AppendChild(dom.Text("Title"))
	root.AppendChild(h)
	root.AppendChild(dom.El(atom.Input, dom.Attr("type", "text")))

	got := dom.Format(root)
	want := "<div data-ui=\"card\">\n  <h2>Title</h2>\n  <input type=\"text\">\n</div>\n"
	if got != want {
		t.Fatalf("Format mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
	if !strings.HasSuffix(got, "</div>\n") {
		t.Fatal("expected closing tag")
	}
}
