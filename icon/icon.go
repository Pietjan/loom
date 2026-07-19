// Package icon renders Phosphor icons in two variants — regular (stroked,
// the default) and fill — at three sizes: base (24px), small (20px), and
// extra small (16px).
//
//	@icon.New(icon.Bell)
//	@icon.New(icon.Bell, icon.Fill, icon.Class("text-red-500"))
//	@icon.New(icon.Bell, icon.Small)
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

// Name identifies an icon, e.g. icon.Bell ("bell").
type Name string

// Variant selects the Phosphor weight.
type Variant string

const (
	VariantRegular Variant = "regular" // stroked (default)
	VariantFill    Variant = "fill"    // filled
)

// Size is the rendered icon size.
type Size string

const (
	SizeBase       Size = "base" // 24px (default)
	SizeSmall      Size = "sm"   // 20px
	SizeExtraSmall Size = "xs"   // 16px
)

// Config holds icon options.
type Config struct {
	opts.Common
	Variant Variant
	Size    Size
}

// Option configures an icon.
type Option = func(*Config)

// Common options, instantiated from the shared generics.
var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// WithVariant sets the Phosphor weight.
func WithVariant(v Variant) Option {
	return func(c *Config) { c.Variant = v }
}

// WithSize sets the rendered size.
func WithSize(s Size) Option {
	return func(c *Config) { c.Size = s }
}

// Pre-baked variant and size options.
var (
	Regular    = WithVariant(VariantRegular)
	Fill       = WithVariant(VariantFill)
	Small      = WithSize(SizeSmall)
	ExtraSmall = WithSize(SizeExtraSmall)
)

// New renders the icon as a templ component.
func New(name Name, options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, name, options...)
	})
}

// Node builds the icon's <svg> node.
func Node(_ context.Context, name Name, options ...Option) (*html.Node, error) {
	cfg := Config{Variant: VariantRegular, Size: SizeBase}
	for _, opt := range options {
		opt(&cfg)
	}

	svg, err := assets.LoadIcon(string(name), string(cfg.Variant))
	if err != nil {
		return nil, err
	}

	dom.SetAttr(svg, dom.MarkerAttr, "icon")
	dom.SetAttr(svg, "data-variant", string(cfg.Variant))
	dom.SetAttr(svg, "data-size", string(cfg.Size))
	dom.SetAttr(svg, "aria-hidden", "true")
	cfg.Apply(svg, classes(cfg))
	return svg, nil
}
