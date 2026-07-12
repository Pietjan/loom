// Package header renders a full-width top application bar — the header
// layout, as an alternative to a left sidebar:
//
//	@header.New(header.Sticky()) {
//		@sidebar.Toggle()
//		@link.New("/") { Acme Inc. }
//		@navbar.New() {
//			@navbar.Item("/", navbar.Current()) { Home }
//			@navbar.Item("/inbox") { Inbox }
//		}
//		@header.Spacer()
//		@button.New(button.Ghost) { Sign out }
//	}
//	@header.Main() {
//		...page content...
//	}
//
// It is a plain <header> flex row — brand, horizontal navbar, a spacer,
// and actions. Sticky pins it to the top on scroll; Container constrains
// its content to a centered max width (pair with header.Main(Container())).
package header

import (
	"context"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
	"github.com/pietjan/loom/internal/styles"
)

// Config holds header options.
type Config struct {
	opts.Common
	sticky    bool
	container bool
}

// Option configures a header.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Sticky pins the header to the top of the viewport on scroll.
func Sticky() Option { return func(c *Config) { c.sticky = true } }

// Container constrains the header content to a centered max width.
func Container() Option { return func(c *Config) { c.container = true } }

// New renders the <header> bar.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}

		h := dom.El(atom.Header, dom.Marker("header"))

		// With Container, children live in a centered inner row so the bar
		// background still spans full width.
		host := h
		if cfg.container {
			inner := dom.El(atom.Div, dom.Attr("class", rowClasses()+" mx-auto w-full max-w-7xl"))
			h.AppendChild(inner)
			host = inner
		}
		if err := render.Children(ctx, host); err != nil {
			return nil, err
		}
		cfg.Apply(h, classes(cfg))
		return h, nil
	})
}

// Spacer pushes following items to the end of the bar.
func Spacer() templ.Component {
	return render.Component(func(_ context.Context) (*html.Node, error) {
		return dom.El(atom.Div, dom.Attr("class", "flex-1"), dom.Attr("aria-hidden", "true")), nil
	})
}

// Main renders the page content region below the header.
func Main(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		m := dom.El(atom.Main, dom.Marker("header-main"))
		if err := render.Children(ctx, m); err != nil {
			return nil, err
		}
		cls := "p-6"
		if cfg.container {
			cls = "mx-auto w-full max-w-7xl p-6"
		}
		cfg.Apply(m, cls)
		return m, nil
	})
}

// rowClasses are the flex-row layout shared by the header and, under
// Container, its inner wrapper.
func rowClasses() string {
	var b styles.Builder
	b.Add("flex items-center gap-3 h-14")
	return b.String()
}

func classes(c Config) string {
	var b styles.Builder
	b.Add("bg-white border-b border-base-200 px-4")
	b.Add("dark:bg-base-800 dark:border-base-600")
	b.If(c.sticky, "sticky top-0 z-40")
	// The flex row lives on the header itself unless Container moved it to
	// an inner wrapper.
	b.If(!c.container, rowClasses())
	return b.String()
}
