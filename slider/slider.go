// Package slider renders a range slider on a native <input type=range>:
//
//	@slider.New(slider.Name("volume"), slider.Value(50))
//	@slider.New(slider.Name("zoom"), slider.Min(1), slider.Max(10), slider.Step(0.5), slider.Value(3))
//
// Inside a field it adopts the field's id and disabled state. The track
// and thumb are styled in cmd/css/loom.css (pseudo-elements utilities
// can't reach). Zero JS.
package slider

import (
	"context"
	"strconv"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/field"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
	"github.com/pietjan/loom/internal/scope"
)

// Config holds slider options.
type Config struct {
	opts.Common
}

// Option configures a slider.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Name sets the form field name.
func Name(name string) Option { return Attr("name", name) }

// Value sets the current value.
func Value(v float64) Option { return Attr("value", num(v)) }

// Min sets the minimum (default 0).
func Min(v float64) Option { return Attr("min", num(v)) }

// Max sets the maximum (default 100).
func Max(v float64) Option { return Attr("max", num(v)) }

// Step sets the step increment.
func Step(v float64) Option { return Attr("step", num(v)) }

// Disabled disables the slider.
func Disabled() Option { return Attr("disabled", "") }

// New renders a slider as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the <input type=range> node.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{}
	for _, opt := range options {
		opt(&cfg)
	}

	n := dom.El(atom.Input, dom.Marker("slider"), dom.Attr("type", "range"))
	if dom.GetAttribute(cfg.Attrs, "min") == "" {
		dom.SetAttr(n, "min", "0")
	}
	if dom.GetAttribute(cfg.Attrs, "max") == "" {
		dom.SetAttr(n, "max", "100")
	}

	if sc, ok := scope.From[field.Scope](ctx); ok {
		if sc.Disabled {
			dom.SetAttr(n, "disabled", "")
		}
	}

	cfg.Apply(n, classes())
	return n, nil
}

func num(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
