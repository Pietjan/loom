// Package radio renders radio buttons, grouped by a Group that shares the
// form name via scope — items never repeat it:
//
//	@radio.Group(radio.Name("plan"), radio.Legend("Plan")) {
//		@radio.New(radio.Value("free"), radio.Label("Free"))
//		@radio.New(radio.Value("pro"), radio.Label("Pro"), radio.Checked())
//	}
//
// Group renders a <fieldset> with a <legend>, the native way to label a
// radio group.
package radio

import (
	"context"
	"errors"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/field"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
	"github.com/pietjan/loom/internal/scope"
)

// ErrNoName is returned when a radio renders without a name — neither its
// own Name option nor a surrounding Group provides one.
var ErrNoName = errors.New("radio: no name — set radio.Name(...) on the Group (or the radio itself)")

// groupScope shares the group's form name with its items.
type groupScope struct {
	Name string
}

// Config holds radio options.
type Config struct {
	opts.Common
	Label   string
	Legend  string
	Checked bool
	invalid bool
}

// Option configures a radio or group.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Name sets the shared form field name (put it on the Group).
func Name(name string) Option { return Attr("name", name) }

// Value sets this radio's submitted value.
func Value(v string) Option { return Attr("value", v) }

// Label renders an inline label next to the radio.
func Label(label string) Option { return func(c *Config) { c.Label = label } }

// Legend sets the group's visible legend (Group only).
func Legend(legend string) Option { return func(c *Config) { c.Legend = legend } }

// Checked pre-selects this radio.
func Checked() Option { return func(c *Config) { c.Checked = true } }

// Disabled disables the radio (or the whole Group).
func Disabled() Option { return Attr("disabled", "") }

// Group renders a fieldset that names and labels its radios.
func Group(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}

		fs := dom.El(atom.Fieldset, dom.Marker("radio-group"))
		if cfg.Legend != "" {
			legend := dom.El(atom.Legend, dom.Attr("class", legendClasses()))
			legend.AppendChild(dom.Text(cfg.Legend))
			fs.AppendChild(legend)
		}

		name := dom.GetAttribute(cfg.Attrs, "name")
		cfg.Attrs = dom.DeleteAttribute(cfg.Attrs, "name") // name belongs to the radios, not the fieldset
		if err := render.Children(ctx, fs, scope.With(groupScope{Name: name})); err != nil {
			return nil, err
		}
		cfg.Apply(fs, groupClasses())
		return fs, nil
	})
}

// New renders a single radio as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the radio, wrapped in a <label> when an inline label is
// configured.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{}
	for _, opt := range options {
		opt(&cfg)
	}

	r := dom.El(atom.Input, dom.Marker("radio"), dom.Attr("type", "radio"))
	if cfg.Checked {
		dom.SetAttr(r, "checked", "")
	}

	if dom.GetAttribute(cfg.Attrs, "name") == "" {
		gs, ok := scope.From[groupScope](ctx)
		if !ok || gs.Name == "" {
			return nil, ErrNoName
		}
		dom.SetAttr(r, "name", gs.Name)
	}

	if sc, ok := scope.From[field.Scope](ctx); ok {
		cfg.invalid = sc.Invalid
		if sc.Invalid {
			dom.SetAttr(r, "aria-invalid", "true")
		}
		if sc.Required {
			dom.SetAttr(r, "required", "")
		}
		if sc.Disabled {
			dom.SetAttr(r, "disabled", "")
		}
	}
	cfg.Apply(r, classes(cfg))

	if cfg.Label == "" {
		return r, nil
	}

	wrapper := dom.El(atom.Label, dom.Attr("class", wrapperClasses()))
	wrapper.AppendChild(r)
	labelText := dom.El(atom.Span)
	labelText.AppendChild(dom.Text(cfg.Label))
	wrapper.AppendChild(labelText)
	return wrapper, nil
}
