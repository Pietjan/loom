// Package kbd renders keyboard keys:
//
//	Press @kbd.New() { ⌘ } @kbd.New() { K } to search.
package kbd

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

// Config holds kbd options.
type Config struct {
	opts.Common
}

// Option configures a kbd.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// New renders a <kbd> key.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		k := dom.El(atom.Kbd, dom.Marker("kbd"))
		if err := render.Children(ctx, k); err != nil {
			return nil, err
		}
		cfg.Apply(k, classes())
		return k, nil
	})
}

func classes() string {
	var b styles.Builder
	b.Add("inline-flex items-center justify-center min-w-5 h-5 px-1.5 rounded")
	b.Add("border border-base-200 border-b-2 bg-white text-xs font-medium text-base-600")
	b.Add("dark:border-base-600 dark:bg-base-700 dark:text-base-200")
	return b.String()
}
