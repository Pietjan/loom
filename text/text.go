// Package text renders body copy with consistent tones:
//
//	@text.New() { Regular paragraph text. }
//	@text.New(text.Subtle) { Secondary information. }
//	@text.Strong() { Inline emphasis }
package text

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

// Tone selects the text emphasis.
type Tone string

const (
	ToneDefault Tone = "default"
	ToneStrong  Tone = "strong"
	ToneSubtle  Tone = "subtle"
)

// Size selects the text size.
type Size string

const (
	SizeBase  Size = "base" // default
	SizeSmall Size = "sm"
	SizeLarge Size = "lg"
)

// Config holds text options.
type Config struct {
	opts.Common
	Tone Tone
	Size Size
}

// Option configures text.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// WithTone sets the emphasis tone.
func WithTone(t Tone) Option { return func(c *Config) { c.Tone = t } }

// WithSize sets the text size.
func WithSize(s Size) Option { return func(c *Config) { c.Size = s } }

// Pre-baked options.
var (
	Subtle = WithTone(ToneSubtle)
	Small  = WithSize(SizeSmall)
	Large  = WithSize(SizeLarge)
)

// New renders a <p> as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Strong renders an inline <strong> with strong tone.
func Strong(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := build(append(options, WithTone(ToneStrong)))
		n := dom.El(atom.Strong, dom.Marker("text"))
		if err := render.Children(ctx, n); err != nil {
			return nil, err
		}
		cfg.Apply(n, classes(cfg))
		return n, nil
	})
}

// Node builds the <p> node.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := build(options)
	p := dom.El(atom.P, dom.Marker("text"))
	if err := render.Children(ctx, p); err != nil {
		return nil, err
	}
	cfg.Apply(p, classes(cfg))
	return p, nil
}

func build(options []Option) Config {
	cfg := Config{Tone: ToneDefault, Size: SizeBase}
	for _, opt := range options {
		opt(&cfg)
	}
	return cfg
}

func classes(c Config) string {
	var b styles.Builder
	styles.Match(&b, c.Size, map[Size]string{
		SizeSmall: "text-xs",
		SizeBase:  "text-sm",
		SizeLarge: "text-base",
	})
	styles.Match(&b, c.Tone, map[Tone]string{
		ToneDefault: "text-base-500 dark:text-base-300",
		ToneStrong:  "font-medium text-base-800 dark:text-white",
		ToneSubtle:  "text-base-400 dark:text-base-400",
	})
	return b.String()
}
