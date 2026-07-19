// Package navlist renders vertical navigation:
//
//	@navlist.New(navlist.Label("Main")) {
//		@navlist.Item("/", navlist.Current()) {
//			@icon.New(icon.House, icon.Small)
//			Dashboard
//		}
//		@navlist.Group(navlist.Title("Settings"), navlist.Open()) {
//			@navlist.Item("/settings/profile") { Profile }
//			@navlist.Item("/settings/billing") { Billing }
//		}
//	}
//
// Expandable groups are plain <details><summary> — self-contained
// platform disclosure, zero JS and zero cross-component coordination (the
// predecessor's group toggling was JS keyed to markup that was never
// emitted; nothing here can drift like that). The current page is marked
// with aria-current="page" and styled off that attribute.
package navlist

import (
	"context"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
)

// Config holds navlist options.
type Config struct {
	opts.Common
	LabelText string
	TitleText string
	current   bool
	open      bool
}

// Option configures a navlist part.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Label names the navigation region for assistive tech.
func Label(label string) Option { return func(c *Config) { c.LabelText = label } }

// Title sets a group's summary text.
func Title(title string) Option { return func(c *Config) { c.TitleText = title } }

// Current marks the item as the current page (aria-current="page").
func Current() Option { return func(c *Config) { c.current = true } }

// Open renders a group expanded.
func Open() Option { return func(c *Config) { c.open = true } }

// New renders the <nav> container.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		nav := dom.El(atom.Nav, dom.Marker("navlist"))
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

// Item renders one navigation link.
func Item(href string, options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		a := dom.El(atom.A, dom.Marker("navlist-item"), dom.Attr("href", href))
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

// Group renders an expandable section of items (<details>).
func Group(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}

		d := dom.El(atom.Details, dom.Marker("navlist-group"))
		if cfg.open {
			dom.SetAttr(d, "open", "")
		}

		summary := dom.El(atom.Summary, dom.Attr("class", summaryClasses()))
		summary.AppendChild(dom.Text(cfg.TitleText))
		summary.AppendChild(chevron())
		d.AppendChild(summary)

		panel := dom.El(atom.Div, dom.Attr("class", panelClasses()))
		if err := render.Children(ctx, panel); err != nil {
			return nil, err
		}
		d.AppendChild(panel)

		cfg.Apply(d, groupClasses())
		return d, nil
	})
}

// Heading renders a non-interactive section heading between items.
func Heading(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		h := dom.El(atom.Div, dom.Marker("navlist-heading"))
		if err := render.Children(ctx, h); err != nil {
			return nil, err
		}
		cfg.Apply(h, headingClasses())
		return h, nil
	})
}

func chevron() *html.Node {
	svg := dom.El(atom.Svg,
		dom.Attr("viewBox", "0 0 16 16"),
		dom.Attr("fill", "none"),
		dom.Attr("aria-hidden", "true"),
		dom.Attr("class", chevronClasses()))
	path := dom.CustomEl("path",
		dom.Attr("d", "M4 6l4 4 4-4"),
		dom.Attr("stroke", "currentColor"),
		dom.Attr("stroke-width", "1.5"),
		dom.Attr("stroke-linecap", "round"),
		dom.Attr("stroke-linejoin", "round"))
	svg.AppendChild(path)
	return svg
}
