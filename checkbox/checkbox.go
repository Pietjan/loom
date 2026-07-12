// Package checkbox renders styled checkboxes with an optional inline
// label:
//
//	@checkbox.New(checkbox.Name("terms"), checkbox.Label("I agree"))
//
// Inside a field, state (invalid/required/disabled) is adopted from the
// field scope. The wrapping <label> intentionally carries no data-ui
// marker: it is an implementation detail, and field's shallow queries must
// see through it to reach the control.
package checkbox

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

// Config holds checkbox options.
type Config struct {
	opts.Common
	Label   string
	Checked bool
	invalid bool
}

// Option configures a checkbox.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Name sets the form field name.
func Name(name string) Option { return Attr("name", name) }

// Value sets the submitted value (default "on").
func Value(v string) Option { return Attr("value", v) }

// Label renders an inline label to the right of the box.
func Label(label string) Option { return func(c *Config) { c.Label = label } }

// Checked pre-checks the box.
func Checked() Option { return func(c *Config) { c.Checked = true } }

// Disabled disables the checkbox.
func Disabled() Option { return Attr("disabled", "") }

// New renders a checkbox as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the checkbox, wrapped in a <label> when an inline label is
// configured.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{}
	for _, opt := range options {
		opt(&cfg)
	}

	box := dom.El(atom.Input, dom.Marker("checkbox"), dom.Attr("type", "checkbox"))
	if cfg.Checked {
		dom.SetAttr(box, "checked", "")
	}
	if sc, ok := scope.From[field.Scope](ctx); ok {
		cfg.invalid = sc.Invalid
		if sc.Invalid {
			dom.SetAttr(box, "aria-invalid", "true")
		}
		if sc.Required {
			dom.SetAttr(box, "required", "")
		}
		if sc.Disabled {
			dom.SetAttr(box, "disabled", "")
		}
	}
	cfg.Apply(box, classes(cfg))

	if cfg.Label == "" {
		return box, nil
	}

	wrapper := dom.El(atom.Label, dom.Attr("class", wrapperClasses()))
	wrapper.AppendChild(box)
	labelText := dom.El(atom.Span)
	labelText.AppendChild(dom.Text(cfg.Label))
	wrapper.AppendChild(labelText)
	return wrapper, nil
}
