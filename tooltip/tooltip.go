// Package tooltip renders CSS-only tooltips:
//
//	@tooltip.New(tooltip.Text("Delete forever")) {
//		@button.New(button.Label("Delete")) {
//			@icon.New(icon.Trash)
//		}
//	}
//
// The tooltip shows on hover and on keyboard focus (:hover /
// :focus-within — works in every browser). The wrapped element gets
// aria-describedby pointing at the tooltip text. Where the Interest
// Invokers API exists (Chromium), interestfor is added to a wrapped
// button/link as progressive enhancement (hover-intent delay, Esc to
// dismiss).
//
// Documented caveat of the zero-JS approach: the bubble is positioned
// absolutely within the wrapper, so an overflow:hidden ancestor clips it.
package tooltip

import (
	"context"
	"errors"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/ids"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
)

// ErrNoText is returned when the tooltip has no text.
var ErrNoText = errors.New("tooltip: tooltip.Text(...) is required")

// Config holds tooltip options.
type Config struct {
	opts.Common
	TipText string
}

// Option configures a tooltip.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Text sets the tooltip content.
func Text(text string) Option { return func(c *Config) { c.TipText = text } }

// New wraps its block with a tooltip.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the tooltip wrapper node.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{}
	for _, opt := range options {
		opt(&cfg)
	}
	if cfg.TipText == "" {
		return nil, ErrNoText
	}

	wrap := dom.El(atom.Span, dom.Marker("tooltip"))
	if err := render.Children(ctx, wrap); err != nil {
		return nil, err
	}

	tipID := ids.New(ctx, "tooltip")
	tip := dom.El(atom.Span, dom.Marker("tooltip-content"),
		dom.Attr("id", tipID), dom.Attr("role", "tooltip"))
	tip.AppendChild(dom.Text(cfg.TipText))
	dom.SetAttr(tip, "class", tipClasses())
	wrap.AppendChild(tip)

	// Describe the wrapped element; enhance buttons/links with interest
	// invokers where the platform has them.
	if target := dom.FindShallow(wrap, firstElement); target != nil && target != tip {
		dom.SetAttr(target, "aria-describedby", tipID)
		if target.DataAtom == atom.Button || target.DataAtom == atom.A {
			dom.SetAttr(target, "interestfor", tipID)
		}
	}

	cfg.Apply(wrap, classes(cfg))
	return wrap, nil
}

func firstElement(n *html.Node) bool {
	return n.Type == html.ElementNode
}
