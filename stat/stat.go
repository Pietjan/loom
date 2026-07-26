// Package stat renders a KPI tile - a label, a big value, and optional
// supporting content (a delta badge, a sparkline) passed as children:
//
//	@stat.New(stat.Label("Revenue"), stat.Value("$48.2k")) {
//		@badge.New(badge.Green) { +12% }
//	}
//
// It is a composition container: drop any loom components into the block.
// Background takes a second, decorative component that fills the bottom of
// the tile behind the text - a sparkline is what it is for:
//
//	@stat.New(
//		stat.Label("Revenue"), stat.Value("$48.2k"),
//		stat.Background(chart.New(
//			chart.Sparkline(), chart.Area(), chart.Smooth(), chart.Inset(0),
//			chart.Series("Revenue", revenue),
//		)),
//	) {
//		@badge.New(badge.Green) { +12% }
//	}
package stat

import (
	"context"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
)

// Config holds stat options.
type Config struct {
	opts.Common
	LabelText  string
	ValueText  string
	background templ.Component
}

// Option configures a stat.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Label sets the metric name.
func Label(text string) Option { return func(c *Config) { c.LabelText = text } }

// Value sets the metric value.
func Value(text string) Option { return func(c *Config) { c.ValueText = text } }

// Background renders a component as a decorative layer across the bottom
// of the tile, behind the label and value - intended for a sparkline
// (chart.Sparkline(), with chart.Inset(0) so it meets the tile's border
// rather than floating off it). The layer is muted and does not take
// pointer events; the component keeps whatever accessible name it
// carries.
func Background(c templ.Component) Option { return func(cfg *Config) { cfg.background = c } }

// New renders a stat tile as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the stat node.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{}
	for _, opt := range options {
		opt(&cfg)
	}

	root := dom.El(atom.Div, dom.Marker("stat"))

	// First in source order so it paints under everything; the content
	// below carries its own positioning to stay on top.
	if cfg.background != nil {
		layer := dom.El(atom.Div, dom.Marker("stat-background"), dom.Attr("class", backgroundClasses()))
		// Cleared children: the tile's block belongs to the content row,
		// not to whatever component was handed in as the background.
		if err := render.Fragment(templ.ClearChildren(ctx), cfg.background, layer); err != nil {
			return nil, err
		}
		root.AppendChild(layer)
	}

	if cfg.LabelText != "" {
		label := dom.El(atom.Div, dom.Marker("stat-label"),
			dom.Attr("class", labelClasses(cfg.background != nil)))
		label.AppendChild(dom.Text(cfg.LabelText))
		root.AppendChild(label)
	}

	row := dom.El(atom.Div, dom.Attr("class", rowClasses(cfg.background != nil)))
	if cfg.ValueText != "" {
		value := dom.El(atom.Div, dom.Marker("stat-value"), dom.Attr("class", valueClasses()))
		value.AppendChild(dom.Text(cfg.ValueText))
		row.AppendChild(value)
	}
	// Children (delta badge, sparkline, ...) sit beside the value.
	if err := render.Children(ctx, row); err != nil {
		return nil, err
	}
	root.AppendChild(row)

	cfg.Apply(root, classes(cfg.background != nil))
	return root, nil
}
