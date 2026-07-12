// Package toggle renders a switch (role="switch" checkbox):
//
//	@toggle.New(toggle.Name("notifications"), toggle.Label("Email me"))
//
// Structure: an unmarked wrapper span holds the appearance-none input (the
// track) and a sibling thumb that slides via peer-checked — no
// pseudo-elements on the input, which Firefox doesn't support.
package toggle

import (
	"context"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/field"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
	"github.com/pietjan/loom/internal/scope"
)

// Config holds toggle options.
type Config struct {
	opts.Common
	Label   string
	Checked bool
}

// Option configures a toggle.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Name sets the form field name.
func Name(name string) Option { return Attr("name", name) }

// Label renders an inline label next to the switch.
func Label(label string) Option { return func(c *Config) { c.Label = label } }

// Checked pre-checks the switch.
func Checked() Option { return func(c *Config) { c.Checked = true } }

// Disabled disables the switch.
func Disabled() Option { return Attr("disabled", "") }

// New renders a toggle as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the toggle node.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{}
	for _, opt := range options {
		opt(&cfg)
	}

	track := dom.El(atom.Input, dom.Marker("toggle"),
		dom.Attr("type", "checkbox"), dom.Attr("role", "switch"))
	if cfg.Checked {
		dom.SetAttr(track, "checked", "")
	}
	if sc, ok := scope.From[field.Scope](ctx); ok {
		if sc.Required {
			dom.SetAttr(track, "required", "")
		}
		if sc.Disabled {
			dom.SetAttr(track, "disabled", "")
		}
	}
	cfg.Apply(track, trackClasses())

	thumb := dom.El(atom.Span, dom.Attr("aria-hidden", "true"), dom.Attr("class", thumbClasses()))

	wrap := dom.El(atom.Span, dom.Attr("class", "relative inline-flex"))
	wrap.AppendChild(track)
	wrap.AppendChild(thumb)

	if cfg.Label == "" {
		return wrap, nil
	}

	label := dom.El(atom.Label, dom.Attr("class", wrapperClasses()))
	label.AppendChild(wrap)
	labelText := dom.El(atom.Span)
	labelText.AppendChild(dom.Text(cfg.Label))
	label.AppendChild(labelText)
	return label, nil
}
