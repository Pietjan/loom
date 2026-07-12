// Package card renders a bordered content container:
//
//	@card.New() {
//		@heading.New() { Revenue }
//		@text.New() { Up 12% this month. }
//	}
package card

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

// Size selects the card padding.
type Size string

const (
	SizeBase  Size = "base" // default
	SizeSmall Size = "sm"
)

// Config holds card options.
type Config struct {
	opts.Common
	Size Size
}

// Option configures a card.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// WithSize sets the card padding scale.
func WithSize(s Size) Option { return func(c *Config) { c.Size = s } }

// Small is the compact padding option.
var Small = WithSize(SizeSmall)

// New renders a card as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the card <div> node.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{Size: SizeBase}
	for _, opt := range options {
		opt(&cfg)
	}

	div := dom.El(atom.Div, dom.Marker("card"))
	if err := render.Children(ctx, div); err != nil {
		return nil, err
	}
	cfg.Apply(div, classes(cfg))
	return div, nil
}

func classes(c Config) string {
	var b styles.Builder
	b.Add("rounded-lg border border-base-200 bg-white shadow-xs")
	b.Add("dark:border-base-600 dark:bg-base-700")
	styles.Match(&b, c.Size, map[Size]string{
		SizeBase:  "p-6",
		SizeSmall: "p-4",
	})
	return b.String()
}
