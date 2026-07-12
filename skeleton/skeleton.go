// Package skeleton renders loading placeholders that pulse:
//
//	@skeleton.New()                         // a text line
//	@skeleton.New(skeleton.Circle(), skeleton.Class("size-10"))
//	@skeleton.New(skeleton.Class("h-24"))   // a block
//
// Shape it with utility classes; the pulse is Tailwind's animate-pulse.
// Mark the surrounding region aria-busy while it loads. Zero JS.
package skeleton

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

// Shape selects the placeholder outline.
type Shape string

const (
	ShapeLine   Shape = "line" // default: a rounded text line
	ShapeCircle Shape = "circle"
	ShapeRect   Shape = "rect"
)

// Config holds skeleton options.
type Config struct {
	opts.Common
	Shape Shape
}

// Option configures a skeleton.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// WithShape sets the placeholder shape.
func WithShape(s Shape) Option { return func(c *Config) { c.Shape = s } }

// Pre-baked shapes.
var (
	Circle = WithShape(ShapeCircle)
	Rect   = WithShape(ShapeRect)
)

// New renders a skeleton placeholder.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the skeleton node.
func Node(_ context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{Shape: ShapeLine}
	for _, opt := range options {
		opt(&cfg)
	}
	n := dom.El(atom.Div, dom.Marker("skeleton"), dom.Attr("aria-hidden", "true"))
	cfg.Apply(n, classes(cfg))
	return n, nil
}

func classes(c Config) string {
	var b styles.Builder
	b.Add("animate-pulse bg-base-200 dark:bg-base-700")
	styles.Match(&b, c.Shape, map[Shape]string{
		ShapeLine:   "h-4 w-full rounded",
		ShapeCircle: "size-10 rounded-full",
		ShapeRect:   "h-24 w-full rounded-lg",
	})
	return b.String()
}
