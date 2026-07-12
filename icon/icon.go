// Package icon renders heroicons in four variants: outline (24px, the
// default), solid (24px), mini (20px), and micro (16px).
//
//	@icon.New(icon.Bell)
//	@icon.New(icon.Bell, icon.Solid, icon.Class("text-red-500"))
//
// Icons are decorative by default (aria-hidden). Every icon name has a
// generated constant in generated.go (see cmd/icons).
package icon

import (
	"context"

	"github.com/a-h/templ"
	"golang.org/x/net/html"

	"github.com/pietjan/loom/internal/assets"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
)

// Name identifies a heroicon, e.g. icon.Bell ("bell").
type Name string

// Variant selects the heroicons style set.
type Variant string

const (
	VariantOutline Variant = "outline" // 24px stroked (default)
	VariantSolid   Variant = "solid"   // 24px filled
	VariantMini    Variant = "mini"    // 20px filled
	VariantMicro   Variant = "micro"   // 16px filled
)

// Config holds icon options.
type Config struct {
	opts.Common
	Variant Variant
}

// Option configures an icon.
type Option = func(*Config)

// Common options, instantiated from the shared generics.
var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// WithVariant sets the heroicons style set.
func WithVariant(v Variant) Option {
	return func(c *Config) { c.Variant = v }
}

// Pre-baked variant options.
var (
	Outline = WithVariant(VariantOutline)
	Solid   = WithVariant(VariantSolid)
	Mini    = WithVariant(VariantMini)
	Micro   = WithVariant(VariantMicro)
)

// New renders the icon as a templ component.
func New(name Name, options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, name, options...)
	})
}

// Node builds the icon's <svg> node.
func Node(_ context.Context, name Name, options ...Option) (*html.Node, error) {
	cfg := Config{Variant: VariantOutline}
	for _, opt := range options {
		opt(&cfg)
	}

	svg, err := assets.LoadIcon(string(name), string(cfg.Variant))
	if err != nil {
		return nil, err
	}

	dom.SetAttr(svg, dom.MarkerAttr, "icon")
	dom.SetAttr(svg, "data-variant", string(cfg.Variant))
	dom.SetAttr(svg, "aria-hidden", "true")
	cfg.Apply(svg, classes(cfg))
	return svg, nil
}
