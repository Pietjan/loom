// Package testutil provides the shared rendering, parsing, and golden-file
// helpers used by every component test.
package testutil

import (
	"bytes"
	"context"

	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/ids"
)

// updating reports whether golden files should be rewritten. An env var
// rather than a flag, so `go test ./...` doesn't fail on packages that
// don't import testutil.
func updating() bool { return os.Getenv("LOOM_UPDATE") != "" }

// WithChildren attaches a child block to a component, as templ's generated
// code would for @component() { ... }.
func WithChildren(c templ.Component, children templ.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return c.Render(templ.WithChildren(ctx, children), w)
	})
}

// Text is a child block containing plain text.
func Text(s string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, s)
		return err
	})
}

// Sequence renders components one after another as a single child block.
func Sequence(cs ...templ.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		for _, c := range cs {
			if err := c.Render(ctx, w); err != nil {
				return err
			}
		}
		return nil
	})
}

// Context returns a render context with a fresh seeded ID counter, so
// generated IDs are deterministic (loom-x-1, loom-x-2, ...).
func Context() context.Context {
	return ids.WithCounter(context.Background())
}

// Render renders a component with a seeded context and returns the HTML.
func Render(t *testing.T, c templ.Component) string {
	t.Helper()
	var buf bytes.Buffer
	if err := c.Render(Context(), &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	return buf.String()
}

// RenderErr renders a component and returns the error, for asserting
// fail-loud contracts.
func RenderErr(c templ.Component) error {
	var buf bytes.Buffer
	return c.Render(Context(), &buf)
}

// Parse parses rendered HTML into a container node so tests can run dom
// queries against it.
func Parse(t *testing.T, s string) *html.Node {
	t.Helper()
	container := dom.El(atom.Body)
	nodes, err := html.ParseFragment(strings.NewReader(s), container)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	for _, n := range nodes {
		container.AppendChild(n)
	}
	return container
}

// Tree wraps a parsed render for marker-based assertions.
type Tree struct {
	Root *html.Node
	t    *testing.T
}

// NewTree parses rendered HTML for querying.
func NewTree(t *testing.T, s string) *Tree {
	t.Helper()
	return &Tree{Root: Parse(t, s), t: t}
}

// One returns the single node with the given data-ui marker, failing the
// test when it is absent or ambiguous.
func (tr *Tree) One(marker string) *html.Node {
	tr.t.Helper()
	all := dom.FindAll(tr.Root, dom.ByMarker(marker))
	if len(all) != 1 {
		tr.t.Fatalf("expected exactly one data-ui=%q, found %d", marker, len(all))
	}
	return all[0]
}

// Maybe returns the node with the given marker, or nil.
func (tr *Tree) Maybe(marker string) *html.Node {
	return dom.Find(tr.Root, dom.ByMarker(marker))
}

// Golden renders the component and compares the formatted tree against
// testdata/<name>.golden.html. Run with LOOM_UPDATE=1 to rewrite.
func Golden(t *testing.T, name string, c templ.Component) {
	t.Helper()
	tree := Parse(t, Render(t, c))

	var sb strings.Builder
	for n := tree.FirstChild; n != nil; n = n.NextSibling {
		sb.WriteString(dom.Format(n))
	}
	got := sb.String()

	path := filepath.Join("testdata", name+".golden.html")
	if updating() {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("missing golden file %s (run: LOOM_UPDATE=1 go test ./...): %v", path, err)
	}
	if got != string(want) {
		t.Errorf("golden mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", path, got, want)
	}
}
