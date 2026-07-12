// Package navbar renders horizontal navigation — a <nav> of inline links,
// typically inside a header:
//
//	@navbar.New() {
//		@navbar.Item("/", navbar.Current()) {
//			@icon.New(icon.Home, icon.Mini)
//			Home
//		}
//		@navbar.Item("/inbox", navbar.Badge("12")) {
//			@icon.New(icon.Inbox, icon.Mini)
//			Inbox
//		}
//	}
//
// The current page is marked with aria-current="page" and styled with an
// accent underline. For vertical navigation (sidebars) use navlist; for
// the top-bar shell use header.
package navbar

import (
	"context"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
)

// Config holds navbar options.
type Config struct {
	opts.Common
	LabelText string
	BadgeText string
	current   bool
}

// Option configures a navbar or item.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Label names the navigation region for assistive tech.
func Label(label string) Option { return func(c *Config) { c.LabelText = label } }

// Current marks the item as the current page (aria-current="page").
func Current() Option { return func(c *Config) { c.current = true } }

// Badge adds a count badge to an item.
func Badge(text string) Option { return func(c *Config) { c.BadgeText = text } }

// New renders the <nav> container.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		nav := dom.El(atom.Nav, dom.Marker("navbar"))
		if cfg.LabelText != "" {
			dom.SetAttr(nav, "aria-label", cfg.LabelText)
		}
		if err := render.Children(ctx, nav); err != nil {
			return nil, err
		}
		cfg.Apply(nav, navClasses())
		return nav, nil
	})
}

// Item renders one horizontal nav link.
func Item(href string, options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		a := dom.El(atom.A, dom.Marker("navbar-item"), dom.Attr("href", href))
		if cfg.current {
			dom.SetAttr(a, "aria-current", "page")
		}
		if err := render.Children(ctx, a); err != nil {
			return nil, err
		}
		if cfg.BadgeText != "" {
			badge := dom.El(atom.Span, dom.Attr("class", badgeClasses()))
			badge.AppendChild(dom.Text(cfg.BadgeText))
			a.AppendChild(badge)
		}
		cfg.Apply(a, itemClasses())
		return a, nil
	})
}

// Spacer pushes following items to the end of the bar. Handy inside a
// header when navbar is used directly as the flex row.
func Spacer() templ.Component {
	return render.Component(func(_ context.Context) (*html.Node, error) {
		return dom.El(atom.Div, dom.Attr("class", "flex-1"), dom.Attr("aria-hidden", "true")), nil
	})
}
