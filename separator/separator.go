// Package separator renders horizontal and vertical dividers:
//
//	@separator.New()
//	@separator.New(separator.Vertical)
package separator

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

// Orientation of the separator.
type Orientation string

const (
	Horizontal Orientation = "horizontal" // default
	VerticalO  Orientation = "vertical"
)

// Config holds separator options.
type Config struct {
	opts.Common
	Orientation Orientation
}

// Option configures a separator.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// WithOrientation sets the separator orientation.
func WithOrientation(o Orientation) Option {
	return func(c *Config) { c.Orientation = o }
}

// Vertical renders a vertical divider.
var Vertical = WithOrientation(VerticalO)

// New renders a separator as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the separator node: <hr> horizontally, a
// role="separator" <div> vertically.
func Node(_ context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{Orientation: Horizontal}
	for _, opt := range options {
		opt(&cfg)
	}

	var n *html.Node
	if cfg.Orientation == VerticalO {
		n = dom.El(atom.Div, dom.Marker("separator"),
			dom.Attr("role", "separator"), dom.Attr("aria-orientation", "vertical"))
	} else {
		n = dom.El(atom.Hr, dom.Marker("separator"))
	}
	cfg.Apply(n, classes(cfg))
	return n, nil
}

func classes(c Config) string {
	var b styles.Builder
	if c.Orientation == VerticalO {
		b.Add("w-px self-stretch bg-base-200 dark:bg-base-600")
	} else {
		b.Add("border-0 h-px w-full bg-base-200 dark:bg-base-600")
	}
	return b.String()
}
