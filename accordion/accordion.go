// Package accordion renders disclosure groups on native <details>:
//
//	@accordion.Root(accordion.Exclusive()) {
//		@accordion.Item(accordion.Title("Shipping")) {
//			@text.New() { 3-5 business days. }
//		}
//		@accordion.Item(accordion.Title("Returns"), accordion.Open()) {
//			@text.New() { Within 30 days. }
//		}
//	}
//
// Exclusive() makes opening one item close the others — the platform's
// own <details name> behavior, zero JS. The chevron rotation is pure CSS.
package accordion

import (
	"context"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/ids"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
	"github.com/pietjan/loom/internal/scope"
)

// groupScope shares the exclusive-group name with items.
type groupScope struct {
	Name string
}

// Config holds accordion options.
type Config struct {
	opts.Common
	TitleText string
	exclusive bool
	open      bool
}

// Option configures an accordion or item.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Exclusive makes items close each other (shared <details name>).
func Exclusive() Option { return func(c *Config) { c.exclusive = true } }

// Title sets an item's summary text.
func Title(title string) Option { return func(c *Config) { c.TitleText = title } }

// Open renders an item expanded.
func Open() Option { return func(c *Config) { c.open = true } }

// Root renders the accordion container.
func Root(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}

		root := dom.El(atom.Div, dom.Marker("accordion"))
		var enrich []func(context.Context) context.Context
		if cfg.exclusive {
			enrich = append(enrich, scope.With(groupScope{Name: ids.New(ctx, "accordion")}))
		}
		if err := render.Children(ctx, root, enrich...); err != nil {
			return nil, err
		}
		cfg.Apply(root, rootClasses())
		return root, nil
	})
}

// Item renders one <details> disclosure.
func Item(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}

		d := dom.El(atom.Details, dom.Marker("accordion-item"))
		if gs, ok := scope.From[groupScope](ctx); ok {
			dom.SetAttr(d, "name", gs.Name)
		}
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

		cfg.Apply(d, itemClasses())
		return d, nil
	})
}

// chevron builds the rotating indicator (CSS-only via group-open).
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
