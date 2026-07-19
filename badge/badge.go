// Package badge renders small status labels:
//
//	@badge.New(badge.Green) { Active }
//	@badge.New(badge.Red, badge.Pill()) {
//		@icon.New(icon.XCircle, icon.ExtraSmall)
//		Failed
//	}
//
// Icons inside a badge are sized down via CSS — no tree surgery.
package badge

import (
	"context"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
)

// Color selects the badge palette.
type Color string

const (
	ColorZinc    Color = "zinc" // default
	ColorRed     Color = "red"
	ColorOrange  Color = "orange"
	ColorAmber   Color = "amber"
	ColorYellow  Color = "yellow"
	ColorLime    Color = "lime"
	ColorGreen   Color = "green"
	ColorEmerald Color = "emerald"
	ColorTeal    Color = "teal"
	ColorCyan    Color = "cyan"
	ColorSky     Color = "sky"
	ColorBlue    Color = "blue"
	ColorIndigo  Color = "indigo"
	ColorViolet  Color = "violet"
	ColorPurple  Color = "purple"
	ColorFuchsia Color = "fuchsia"
	ColorPink    Color = "pink"
	ColorRose    Color = "rose"
)

// Size selects the badge size.
type Size string

const (
	SizeBase  Size = "base" // default
	SizeSmall Size = "sm"
	SizeLarge Size = "lg"
)

// Config holds badge options.
type Config struct {
	opts.Common
	Color Color
	Size  Size
	pill  bool
}

// Option configures a badge.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// WithColor sets the badge palette.
func WithColor(c Color) Option { return func(cfg *Config) { cfg.Color = c } }

// WithSize sets the badge size.
func WithSize(s Size) Option { return func(cfg *Config) { cfg.Size = s } }

// Pill rounds the badge fully.
func Pill() Option { return func(cfg *Config) { cfg.pill = true } }

// Pre-baked color and size options.
var (
	Zinc    = WithColor(ColorZinc)
	Red     = WithColor(ColorRed)
	Orange  = WithColor(ColorOrange)
	Amber   = WithColor(ColorAmber)
	Yellow  = WithColor(ColorYellow)
	Lime    = WithColor(ColorLime)
	Green   = WithColor(ColorGreen)
	Emerald = WithColor(ColorEmerald)
	Teal    = WithColor(ColorTeal)
	Cyan    = WithColor(ColorCyan)
	Sky     = WithColor(ColorSky)
	Blue    = WithColor(ColorBlue)
	Indigo  = WithColor(ColorIndigo)
	Violet  = WithColor(ColorViolet)
	Purple  = WithColor(ColorPurple)
	Fuchsia = WithColor(ColorFuchsia)
	Pink    = WithColor(ColorPink)
	Rose    = WithColor(ColorRose)
	Small   = WithSize(SizeSmall)
	Large   = WithSize(SizeLarge)
)

// New renders a badge as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the badge <span> node.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{Color: ColorZinc, Size: SizeBase}
	for _, opt := range options {
		opt(&cfg)
	}

	span := dom.El(atom.Span, dom.Marker("badge"))
	if err := render.Children(ctx, span); err != nil {
		return nil, err
	}
	cfg.Apply(span, classes(cfg))
	return span, nil
}
