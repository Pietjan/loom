// Package description renders a term/detail list (<dl>) — the read-only
// counterpart to a form:
//
//	@description.New() {
//		@description.Term() { Plan }
//		@description.Detail() { Pro }
//		@description.Term() { Renews }
//		@description.Detail() { March 1, 2026 }
//	}
//
// A two-column grid aligns terms and details; it collapses to stacked rows
// on narrow viewports. Zero JS.
package description

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

// Config holds description options.
type Config struct {
	opts.Common
}

// Option configures a description list.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// New renders the <dl> container.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		dl := dom.El(atom.Dl, dom.Marker("description"))
		if err := render.Children(ctx, dl); err != nil {
			return nil, err
		}
		cfg.Apply(dl, listClasses())
		return dl, nil
	})
}

// Term renders a <dt> label.
func Term(options ...Option) templ.Component {
	return part(atom.Dt, "description-term", termClasses(), options)
}

// Detail renders a <dd> value.
func Detail(options ...Option) templ.Component {
	return part(atom.Dd, "description-detail", detailClasses(), options)
}

func part(a atom.Atom, marker, cls string, options []Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		n := dom.El(a, dom.Marker(marker))
		if err := render.Children(ctx, n); err != nil {
			return nil, err
		}
		cfg.Apply(n, cls)
		return n, nil
	})
}

func listClasses() string {
	var b styles.Builder
	b.Add("grid grid-cols-1 gap-x-6 gap-y-3 sm:grid-cols-[minmax(0,12rem)_1fr]")
	b.Add("border-t border-base-200 pt-4 dark:border-base-600")
	return b.String()
}

func termClasses() string {
	var b styles.Builder
	b.Add("text-sm font-medium text-base-500 dark:text-base-400")
	return b.String()
}

func detailClasses() string {
	var b styles.Builder
	b.Add("text-sm text-base-800 sm:mt-0 -mt-2 mb-1 dark:text-base-100")
	return b.String()
}
