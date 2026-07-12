// Package input renders text inputs. Standalone or inside a field, where
// it adopts the field's id, aria, and state wiring automatically:
//
//	@input.New(input.Type("email"), input.Name("email"), input.Placeholder("you@example.com"))
package input

import (
	"context"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/field"
	"github.com/pietjan/loom/inputgroup"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
	"github.com/pietjan/loom/internal/scope"
)

// Config holds input options.
type Config struct {
	opts.Common
	invalid bool
	grouped bool // inside an input group: render without its own shell
}

// Option configures an input.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Type sets the input type (default "text").
func Type(t string) Option { return Attr("type", t) }

// Name sets the form field name.
func Name(name string) Option { return Attr("name", name) }

// Value sets the current value.
func Value(v string) Option { return Attr("value", v) }

// Placeholder sets the placeholder text.
func Placeholder(p string) Option { return Attr("placeholder", p) }

// Disabled disables the input.
func Disabled() Option { return Attr("disabled", "") }

// New renders an input as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the <input> node, adopting a surrounding field.Scope.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{}
	for _, opt := range options {
		opt(&cfg)
	}

	n := dom.El(atom.Input, dom.Marker("input"))
	if dom.GetAttribute(cfg.Attrs, "type") == "" {
		dom.SetAttr(n, "type", "text")
	}

	if sc, ok := scope.From[field.Scope](ctx); ok {
		cfg.invalid = sc.Invalid
		if sc.Invalid {
			dom.SetAttr(n, "aria-invalid", "true")
		}
		if sc.Required {
			dom.SetAttr(n, "required", "")
		}
		if sc.Disabled {
			dom.SetAttr(n, "disabled", "")
		}
	}
	// Inside an input group the wrapper owns the border/background/ring,
	// so the input drops its own shell and just fills the row.
	if _, ok := scope.From[inputgroup.Scope](ctx); ok {
		cfg.grouped = true
	}

	cfg.Apply(n, classes(cfg))
	return n, nil
}
