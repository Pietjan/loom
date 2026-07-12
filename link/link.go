// Package link renders styled anchors:
//
//	@link.New("/settings") { Settings }
//	@link.New("https://example.com", link.External, link.Subtle) { Docs }
package link

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

// Tone selects the link styling.
type Tone string

const (
	ToneDefault Tone = "default" // accent colored, underline on hover
	ToneGhost   Tone = "ghost"   // text colored, underline on hover
	ToneSubtle  Tone = "subtle"  // muted, no underline
)

// Config holds link options.
type Config struct {
	opts.Common
	Tone     Tone
	external bool
}

// Option configures a link.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// WithTone sets the link tone.
func WithTone(t Tone) Option { return func(c *Config) { c.Tone = t } }

// External opens the link in a new tab with rel="noopener noreferrer".
func External() Option { return func(c *Config) { c.external = true } }

// Pre-baked tone options.
var (
	Ghost  = WithTone(ToneGhost)
	Subtle = WithTone(ToneSubtle)
)

// New renders an anchor as a templ component.
func New(href string, options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, href, options...)
	})
}

// Node builds the <a> node.
func Node(ctx context.Context, href string, options ...Option) (*html.Node, error) {
	cfg := Config{Tone: ToneDefault}
	for _, opt := range options {
		opt(&cfg)
	}

	a := dom.El(atom.A, dom.Marker("link"), dom.Attr("href", href))
	if cfg.external {
		dom.SetAttr(a, "target", "_blank")
		dom.SetAttr(a, "rel", "noopener noreferrer")
	}
	if err := render.Children(ctx, a); err != nil {
		return nil, err
	}
	cfg.Apply(a, classes(cfg))
	return a, nil
}

func classes(c Config) string {
	var b styles.Builder
	b.Add("font-medium")
	styles.Match(&b, c.Tone, map[Tone]string{
		ToneDefault: "text-accent hover:underline",
		ToneGhost:   "text-base-800 hover:underline dark:text-white",
		ToneSubtle:  "text-base-400 hover:text-base-800 dark:hover:text-white",
	})
	return b.String()
}
