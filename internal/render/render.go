// Package render bridges *html.Node component trees and the templ runtime.
//
// It fixes two latent bugs from the predecessor library (pulseui):
//
//  1. templ.ClearChildren returns a NEW context; discarding the return
//     value leaves the children installed, so a child block rendering a
//     component invoked without its own block would re-render the
//     parent's children inside it (worst case: infinite recursion).
//  2. html.ParseFragment must be given the REAL parent node as parsing
//     context — parsing against a generic <div> silently drops
//     table-structure elements like <tr>.
package render

import (
	"bytes"
	"context"
	"io"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
)

// Component wraps a Node-building function as a templ.Component.
func Component(fn func(ctx context.Context) (*html.Node, error)) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		n, err := fn(ctx)
		if err != nil {
			return err
		}
		return html.Render(w, n)
	})
}

// Coordinator wraps a component that renders no element of its own: it
// enriches the context (typically installing a scope) and renders its
// children directly, with no node building or re-parsing cost. Used by
// pairing roots like modal.Root and dropdown coordination.
func Coordinator(fn func(ctx context.Context) (context.Context, error)) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		children := templ.GetChildren(ctx)
		ctx = templ.ClearChildren(ctx)
		ctx, err := fn(ctx)
		if err != nil {
			return err
		}
		return children.Render(ctx, w)
	})
}

// Children renders the templ child block into parent, appending the
// resulting nodes. Each enrich function transforms the context the
// children render with — the standard way to install scopes
// (scope.With(...)).
func Children(ctx context.Context, parent *html.Node, enrich ...func(context.Context) context.Context) error {
	return ChildrenAs(ctx, parent, parent, enrich...)
}

// ChildrenAs is Children with an explicit parsing context, for the rare
// parent whose HTML insertion mode predates its modern content model:
// x/net/html's "in select" mode strips the rich option content the
// customizable-select spec now allows, so picker parses its children
// against a neutral <div> and appends them to the real <select>. Prefer
// Children — parsing against the real parent is what keeps table
// fragments intact.
func ChildrenAs(ctx context.Context, parent, parseCtx *html.Node, enrich ...func(context.Context) context.Context) error {
	children := templ.GetChildren(ctx)
	ctx = templ.ClearChildren(ctx)
	for _, e := range enrich {
		ctx = e(ctx)
	}
	var buf bytes.Buffer
	if err := children.Render(ctx, &buf); err != nil {
		return err
	}
	return appendParsed(&buf, parent, parseCtx)
}

// Fragment renders any templ.Component and appends the resulting nodes to
// parent, parsed in parent's context. This is the sanctioned way to embed
// a rendered component into a node tree when direct Node() access is not
// available.
func Fragment(ctx context.Context, c templ.Component, parent *html.Node) error {
	var buf bytes.Buffer
	if err := c.Render(ctx, &buf); err != nil {
		return err
	}
	return appendParsed(&buf, parent, parent)
}

func appendParsed(buf *bytes.Buffer, parent, parseCtx *html.Node) error {
	if buf.Len() == 0 {
		return nil
	}
	nodes, err := html.ParseFragment(buf, parseCtx)
	if err != nil {
		return err
	}
	for _, n := range nodes {
		parent.AppendChild(n)
	}
	return nil
}
