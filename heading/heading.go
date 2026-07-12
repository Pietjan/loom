// Package heading renders section headings:
//
//	@heading.New(heading.Level(1), heading.XL) { Dashboard }
//	@heading.New() { Section title }
//
// Without a Level, the heading renders as a styled <div> so it doesn't
// disturb the page's document outline.
package heading

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

// Size selects the heading size.
type Size string

const (
	SizeBase       Size = "base" // default
	SizeLarge      Size = "lg"
	SizeExtraLarge Size = "xl"
)

// Config holds heading options.
type Config struct {
	opts.Common
	Level int // 0 renders a <div>; 1-6 render <h1>-<h6>
	Size  Size
}

// Option configures a heading.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Level renders a real <h1>-<h6> for the document outline.
func Level(level int) Option { return func(c *Config) { c.Level = level } }

// WithSize sets the visual size, independent of Level.
func WithSize(s Size) Option { return func(c *Config) { c.Size = s } }

// Pre-baked size options.
var (
	Large = WithSize(SizeLarge)
	XL    = WithSize(SizeExtraLarge)
)

var levelAtoms = map[int]atom.Atom{
	1: atom.H1, 2: atom.H2, 3: atom.H3, 4: atom.H4, 5: atom.H5, 6: atom.H6,
}

// New renders a heading as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the heading node.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{Size: SizeBase}
	for _, opt := range options {
		opt(&cfg)
	}

	tag := atom.Div
	if a, ok := levelAtoms[cfg.Level]; ok {
		tag = a
	}
	h := dom.El(tag, dom.Marker("heading"))
	if err := render.Children(ctx, h); err != nil {
		return nil, err
	}
	cfg.Apply(h, classes(cfg))
	return h, nil
}

func classes(c Config) string {
	var b styles.Builder
	b.Add("font-medium text-base-800 dark:text-white")
	styles.Match(&b, c.Size, map[Size]string{
		SizeBase:       "text-sm",
		SizeLarge:      "text-base",
		SizeExtraLarge: "text-xl",
	})
	return b.String()
}
