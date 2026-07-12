// Package fieldset groups related fields under a legend:
//
//	@fieldset.New(fieldset.Legend("Shipping address"), fieldset.Disabled()) {
//		@field.Root() { ... }
//		@field.Root() { ... }
//	}
//
// Disabling the fieldset disables every control inside it — that cascade
// is the platform's own (<fieldset disabled>), not library code.
package fieldset

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

// Config holds fieldset options.
type Config struct {
	opts.Common
	LegendText string
}

// Option configures a fieldset.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Legend sets the fieldset's visible legend.
func Legend(text string) Option { return func(c *Config) { c.LegendText = text } }

// Disabled disables every control inside the fieldset (native cascade).
func Disabled() Option { return Attr("disabled", "") }

// New renders a fieldset as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the <fieldset> node.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{}
	for _, opt := range options {
		opt(&cfg)
	}

	fs := dom.El(atom.Fieldset, dom.Marker("fieldset"))
	if cfg.LegendText != "" {
		legend := dom.El(atom.Legend, dom.Attr("class", legendClasses()))
		legend.AppendChild(dom.Text(cfg.LegendText))
		fs.AppendChild(legend)
	}
	if err := render.Children(ctx, fs); err != nil {
		return nil, err
	}
	cfg.Apply(fs, classes())
	return fs, nil
}

func classes() string {
	var b styles.Builder
	b.Add("grid gap-4 border-0 p-0 m-0 min-w-0")
	b.Add("disabled:opacity-75")
	return b.String()
}

func legendClasses() string {
	var b styles.Builder
	b.Add("mb-2 p-0 text-base font-semibold text-base-800 dark:text-white")
	return b.String()
}
