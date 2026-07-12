// Package sidebar renders a responsive app sidebar with zero JavaScript:
//
//	@sidebar.New() {
//		@navlist.New(navlist.Label("Main")) { ... }
//	}
//	// elsewhere, e.g. in the navbar:
//	@sidebar.Toggle()
//
// The trick: the sidebar element carries the popover attribute. On narrow
// viewports it behaves as a popover — hidden until the Toggle button
// (command="toggle-popover") opens it as a slide-in overlay with light
// dismiss and Esc, courtesy of the platform. On wide viewports structural
// CSS (css/loom.css) overrides the popover's hidden state so the sidebar
// is statically visible, and the Toggle hides itself.
//
// A page has one sidebar, so trigger and target pair on a stable default
// id; use Name/For only for multiple sidebars.
package sidebar

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

// DefaultID pairs New and Toggle without any coordination.
const DefaultID = "loom-sidebar"

// Config holds sidebar options.
type Config struct {
	opts.Common
	PairName string
}

// Option configures the sidebar or its toggle.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Name gives the sidebar a non-default id (only needed with multiple
// sidebars).
func Name(name string) Option { return func(c *Config) { c.PairName = name } }

// For points a Toggle at a named sidebar.
func For(name string) Option { return func(c *Config) { c.PairName = name } }

// New renders the sidebar <aside>.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the sidebar node.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{}
	for _, opt := range options {
		opt(&cfg)
	}
	id := cfg.PairName
	if id == "" {
		id = DefaultID
	}

	aside := dom.El(atom.Aside, dom.Marker("sidebar"),
		dom.Attr("id", id), dom.Attr("popover", ""))
	if err := render.Children(ctx, aside); err != nil {
		return nil, err
	}
	cfg.Apply(aside, classes())
	return aside, nil
}

// Toggle renders the mobile menu button (hidden on wide viewports).
func Toggle(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		id := cfg.PairName
		if id == "" {
			id = DefaultID
		}

		btn := dom.El(atom.Button, dom.Marker("sidebar-toggle"),
			dom.Attr("type", "button"),
			dom.Attr("command", "toggle-popover"),
			dom.Attr("commandfor", id),
			dom.Attr("aria-label", "Toggle sidebar"))
		bars, err := icon.Node(ctx, icon.Bars3, icon.WithVariant(icon.VariantMini))
		if err != nil {
			return nil, err
		}
		btn.AppendChild(bars)
		cfg.Apply(btn, toggleClasses())
		return btn, nil
	})
}
