package render_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/render"
	"github.com/pietjan/loom/internal/scope"
)

type testScope struct{ Value string }

// TestChildrenReceiveEnrichedContext pins the templ runtime behavior the
// whole composition system depends on: a child block is rendered with the
// context passed at render time, so scopes installed by the parent are
// visible to children declared in the user's template. If a templ upgrade
// breaks this, this test must fail loudly.
func TestChildrenReceiveEnrichedContext(t *testing.T) {
	child := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		sc, ok := scope.From[testScope](ctx)
		if !ok {
			_, err := io.WriteString(w, "<span>missing</span>")
			return err
		}
		_, err := io.WriteString(w, "<span>"+sc.Value+"</span>")
		return err
	})

	parent := render.Component(func(ctx context.Context) (*html.Node, error) {
		root := dom.El(atom.Div, dom.Marker("parent"))
		err := render.Children(ctx, root, scope.With(testScope{Value: "from-parent"}))
		return root, err
	})

	var buf bytes.Buffer
	ctx := templ.WithChildren(context.Background(), child)
	if err := parent.Render(ctx, &buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "<span>from-parent</span>") {
		t.Fatalf("scope did not reach child block; got: %s", buf.String())
	}
}

// TestChildrenAreCleared guards against the pulseui bug where
// templ.ClearChildren's return value was discarded: a component invoked
// inside a child block without its own block must NOT inherit the outer
// component's children (worst case that is infinite recursion).
func TestChildrenAreCleared(t *testing.T) {
	// inner renders its (absent) children into a <p>.
	inner := render.Component(func(ctx context.Context) (*html.Node, error) {
		p := dom.El(atom.P, dom.Marker("inner"))
		err := render.Children(ctx, p)
		return p, err
	})

	outer := render.Component(func(ctx context.Context) (*html.Node, error) {
		root := dom.El(atom.Div, dom.Marker("outer"))
		err := render.Children(ctx, root)
		return root, err
	})

	var buf bytes.Buffer
	// outer's child block renders inner (which has no block of its own).
	ctx := templ.WithChildren(context.Background(), inner)
	if err := outer.Render(ctx, &buf); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if strings.Count(got, "data-ui=\"inner\"") != 1 {
		t.Fatalf("inner rendered wrong number of times (children not cleared?): %s", got)
	}
	if strings.Contains(got, "<p data-ui=\"inner\"></p>") == false {
		t.Fatalf("inner should be empty; got: %s", got)
	}
}

// TestChildrenParseAgainstRealParent guards against the pulseui bug where
// fragments were parsed against a generic <div>, which silently drops
// table-structure elements.
func TestChildrenParseAgainstRealParent(t *testing.T) {
	rows := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, "<tr><td>cell</td></tr>")
		return err
	})

	tbody := dom.El(atom.Tbody)
	ctx := templ.WithChildren(context.Background(), rows)
	if err := render.Children(ctx, tbody); err != nil {
		t.Fatal(err)
	}

	if dom.Find(tbody, dom.ByTag(atom.Tr)) == nil {
		t.Fatalf("tr was dropped; tree: %s", dom.Format(tbody))
	}
	if dom.Find(tbody, dom.ByTag(atom.Td)) == nil {
		t.Fatalf("td was dropped; tree: %s", dom.Format(tbody))
	}
}

// TestCoordinator verifies a coordination-only component: no element of
// its own, children rendered with the enriched context.
func TestCoordinator(t *testing.T) {
	root := render.Coordinator(func(ctx context.Context) (context.Context, error) {
		return scope.Set(ctx, testScope{Value: "paired"}), nil
	})

	child := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		sc, _ := scope.From[testScope](ctx)
		_, err := io.WriteString(w, "<b>"+sc.Value+"</b>")
		return err
	})

	var buf bytes.Buffer
	ctx := templ.WithChildren(context.Background(), child)
	if err := root.Render(ctx, &buf); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "<b>paired</b>" {
		t.Fatalf("coordinator must render children only; got: %s", buf.String())
	}
}

// TestFragment embeds a rendered component into an existing tree.
func TestFragment(t *testing.T) {
	c := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, `<span data-ui="icon">*</span>`)
		return err
	})
	parent := dom.El(atom.Button)
	if err := render.Fragment(context.Background(), c, parent); err != nil {
		t.Fatal(err)
	}
	if dom.Find(parent, dom.ByMarker("icon")) == nil {
		t.Fatalf("fragment not appended; tree: %s", dom.Format(parent))
	}
}
