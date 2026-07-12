// Package pagination renders page navigation as links — the server
// renders the target page, so it works with zero JS:
//
//	@pagination.New() {
//		@pagination.Prev("?page=2")
//		@pagination.Item("?page=1") { 1 }
//		@pagination.Item("?page=2", pagination.Current()) { 2 }
//		@pagination.Gap()
//		@pagination.Item("?page=9") { 9 }
//		@pagination.Next("?page=3")
//	}
//
// Build the item list server-side from your page count. The current page
// carries aria-current="page"; a disabled Prev/Next (empty href) renders
// inert.
package pagination

import (
	"context"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/icon"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
)

// Config holds pagination options.
type Config struct {
	opts.Common
	current bool
}

// Option configures pagination.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Current marks the item as the active page.
func Current() Option { return func(c *Config) { c.current = true } }

// New renders the <nav> container.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		nav := dom.El(atom.Nav, dom.Marker("pagination"), dom.Attr("aria-label", "Pagination"))
		if err := render.Children(ctx, nav); err != nil {
			return nil, err
		}
		cfg.Apply(nav, navClasses())
		return nav, nil
	})
}

// Item renders one page-number link.
func Item(href string, options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		a := dom.El(atom.A, dom.Marker("pagination-item"), dom.Attr("href", href))
		if cfg.current {
			dom.SetAttr(a, "aria-current", "page")
		}
		if err := render.Children(ctx, a); err != nil {
			return nil, err
		}
		cfg.Apply(a, itemClasses())
		return a, nil
	})
}

// Prev renders the previous-page control. An empty href renders it
// disabled.
func Prev(href string, options ...Option) templ.Component {
	return arrow(href, "Previous page", icon.ChevronLeft, options)
}

// Next renders the next-page control. An empty href renders it disabled.
func Next(href string, options ...Option) templ.Component {
	return arrow(href, "Next page", icon.ChevronRight, options)
}

func arrow(href, label string, name icon.Name, options []Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		glyph, err := icon.Node(ctx, name, icon.Micro)
		if err != nil {
			return nil, err
		}

		var el *html.Node
		if href == "" {
			el = dom.El(atom.Span, dom.Attr("aria-disabled", "true"))
		} else {
			el = dom.El(atom.A, dom.Attr("href", href))
		}
		dom.SetAttr(el, dom.MarkerAttr, "pagination-item")
		dom.SetAttr(el, "aria-label", label)
		el.AppendChild(glyph)
		cfg.Apply(el, itemClasses())
		return el, nil
	})
}

// Gap renders a non-interactive ellipsis between page ranges.
func Gap() templ.Component {
	return render.Component(func(_ context.Context) (*html.Node, error) {
		span := dom.El(atom.Span, dom.Marker("pagination-gap"),
			dom.Attr("aria-hidden", "true"), dom.Attr("class", gapClasses()))
		span.AppendChild(dom.Text("…"))
		return span, nil
	})
}
