// Package textarea renders multi-line text inputs. Inside a field it
// adopts the field's id, aria, and state wiring automatically.
package textarea

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

// Config holds textarea options.
type Config struct {
	opts.Common
	Value   string
	invalid bool
}

// Option configures a textarea.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Name sets the form field name.
func Name(name string) Option { return Attr("name", name) }

// Rows sets the visible row count (default 4).
func Rows(rows int) Option { return Attr("rows", strconv.Itoa(rows)) }

// Placeholder sets the placeholder text.
func Placeholder(p string) Option { return Attr("placeholder", p) }

// Value sets the initial content.
func Value(v string) Option { return func(c *Config) { c.Value = v } }

// Disabled disables the textarea.
func Disabled() Option { return Attr("disabled", "") }

// New renders a textarea as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the <textarea> node, adopting a surrounding field.Scope.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{}
	for _, opt := range options {
		opt(&cfg)
	}

	n := dom.El(atom.Textarea, dom.Marker("textarea"))
	if dom.GetAttribute(cfg.Attrs, "rows") == "" {
		dom.SetAttr(n, "rows", "4")
	}
	if cfg.Value != "" {
		n.AppendChild(dom.Text(cfg.Value))
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

	cfg.Apply(n, classes(cfg))
	return n, nil
}
