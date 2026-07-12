// Package stat renders a KPI tile — a label, a big value, and optional
// supporting content (a delta badge, a sparkline) passed as children:
//
//	@stat.New(stat.Label("Revenue"), stat.Value("$48.2k")) {
//		@badge.New(badge.Green) { +12% }
//	}
//
// It is a composition container: drop any loom components into the block.
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
	LabelText string
	ValueText string
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

	if cfg.LabelText != "" {
		label := dom.El(atom.Div, dom.Marker("stat-label"), dom.Attr("class", labelClasses()))
		label.AppendChild(dom.Text(cfg.LabelText))
		root.AppendChild(label)
	}

	row := dom.El(atom.Div, dom.Attr("class", "flex items-center gap-2"))
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

	cfg.Apply(root, classes())
	return root, nil
}
