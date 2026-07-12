// Package breadcrumbs renders a breadcrumb trail:
//
//	@breadcrumbs.New() {
//		@breadcrumbs.Item("/") { Home }
//		@breadcrumbs.Item("/projects") { Projects }
//		@breadcrumbs.Item("", breadcrumbs.Current()) { Loom }
//	}
//
// Separators are drawn by CSS between items; the current (last) crumb is
// rendered as plain text with aria-current="page".
package breadcrumbs

import (
	"context"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
)

// Config holds breadcrumbs options.
type Config struct {
	opts.Common
	current bool
}

// Option configures breadcrumbs.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Current marks the item as the current page: rendered as text, not a link.
func Current() Option { return func(c *Config) { c.current = true } }

// New renders the <nav> trail.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		nav := dom.El(atom.Nav, dom.Marker("breadcrumbs"), dom.Attr("aria-label", "Breadcrumb"))
		list := dom.El(atom.Ol, dom.Attr("class", listClasses()))
		if err := render.Children(ctx, list); err != nil {
			return nil, err
		}
		nav.AppendChild(list)
		cfg.Apply(nav, "")
		return nav, nil
	})
}

// Item renders one crumb. With Current (or an empty href) it is plain
// text; otherwise a link.
func Item(href string, options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}

		li := dom.El(atom.Li, dom.Marker("breadcrumbs-item"), dom.Attr("class", itemClasses()))

		var content *html.Node
		if cfg.current || href == "" {
			content = dom.El(atom.Span, dom.Attr("aria-current", "page"), dom.Attr("class", currentClasses()))
		} else {
			content = dom.El(atom.A, dom.Attr("href", href), dom.Attr("class", linkClasses()))
		}
		if err := render.Children(ctx, content); err != nil {
			return nil, err
		}
		li.AppendChild(content)
		cfg.Apply(li, "")
		return li, nil
	})
}
