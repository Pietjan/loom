// Package flash renders a server-rendered flash message — the alert you
// show after a redirect (post/redirect/get). Dismissal and auto-hide are
// pure CSS, so it needs no JavaScript:
//
//	@flash.New(flash.Success(), flash.Dismissible()) {
//		Your changes were saved.
//	}
//	@flash.New(flash.Danger(), flash.Autohide()) {
//		Could not reach the server.
//	}
//
// Dismissible adds a close control backed by a hidden checkbox: checking
// it hides the flash via has-[:checked] (no JS). Autohide fades the flash
// out after a few seconds via a CSS animation (see cmd/css/loom.css).
// Place it wherever you want it to appear, or inside a fixed container for
// toast-style stacking.
package flash

import (
	"context"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/icon"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
)

// Tone selects the flash intent.
type Tone string

const (
	ToneNeutral Tone = "neutral" // default
	ToneInfo    Tone = "info"
	ToneSuccess Tone = "success"
	ToneWarning Tone = "warning"
	ToneDanger  Tone = "danger"
)

// Config holds flash options.
type Config struct {
	opts.Common
	Tone        Tone
	dismissible bool
	autohide    bool
}

// Option configures a flash.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// WithTone sets the flash intent.
func WithTone(t Tone) Option { return func(c *Config) { c.Tone = t } }

// Dismissible adds a close button (hidden-checkbox dismiss, no JS).
func Dismissible() Option { return func(c *Config) { c.dismissible = true } }

// Autohide fades the flash out after a few seconds (CSS animation).
func Autohide() Option { return func(c *Config) { c.autohide = true } }

// Pre-baked tone options.
var (
	Info    = WithTone(ToneInfo)
	Success = WithTone(ToneSuccess)
	Warning = WithTone(ToneWarning)
	Danger  = WithTone(ToneDanger)
)

// New renders a flash message as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the flash node.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{Tone: ToneNeutral}
	for _, opt := range options {
		opt(&cfg)
	}

	root := dom.El(atom.Div, dom.Marker("flash"))
	// Assertive for problems, polite otherwise.
	if cfg.Tone == ToneDanger || cfg.Tone == ToneWarning {
		dom.SetAttr(root, "role", "alert")
	} else {
		dom.SetAttr(root, "role", "status")
	}
	if cfg.autohide {
		dom.SetAttr(root, "data-autohide", "")
	}

	body := dom.El(atom.Div, dom.Attr("class", "flex-1 text-sm"))
	if err := render.Children(ctx, body); err != nil {
		return nil, err
	}
	root.AppendChild(body)

	if cfg.dismissible {
		root.AppendChild(dismissControl(ctx))
	}

	cfg.Apply(root, classes(cfg))
	return root, nil
}

// dismissControl is a hidden checkbox (the dismissed state) wrapped in a
// label showing an ✕. The flash's has-[:checked]:hidden does the hiding.
func dismissControl(ctx context.Context) *html.Node {
	label := dom.El(atom.Label, dom.Attr("class", closeClasses()))
	check := dom.El(atom.Input, dom.Attr("type", "checkbox"),
		dom.Attr("class", "sr-only"), dom.Attr("aria-label", "Dismiss"))
	label.AppendChild(check)
	if x, err := icon.Node(ctx, icon.X, icon.ExtraSmall, icon.Attr("aria-hidden", "true")); err == nil {
		label.AppendChild(x)
	}
	return label
}
