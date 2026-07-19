// Package inputgroup joins an input with leading/trailing addons — an
// icon, a text label, or a small button — inside one bordered control:
//
//	@inputgroup.New() {
//		@inputgroup.Addon() { @icon.New(icon.MagnifyingGlass, icon.Small) }
//		@input.New(input.Name("q"), input.Placeholder("Search"))
//	}
//	@inputgroup.New() {
//		@inputgroup.Addon() { https:// }
//		@input.New(input.Name("site"))
//		@inputgroup.Addon() { .com }
//	}
//
// Addons before the input are leading, after are trailing (source order).
// The group owns the border, background, and focus ring; the inner input
// goes borderless (cmd/css/loom.css). Works inside a field — the field
// wires its label and aria to the input through the group.
package inputgroup

import (
	"context"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
	"github.com/pietjan/loom/internal/scope"
	"github.com/pietjan/loom/internal/styles"
)

// Scope tells a control it is inside an input group, so it renders
// without its own border/background/rounding (the group owns those). A
// marker value — its presence is the signal.
type Scope struct{}

// Config holds input-group options.
type Config struct {
	opts.Common
}

// Option configures an input group.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// New renders the group wrapper.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		g := dom.El(atom.Div, dom.Marker("input-group"))
		if err := render.Children(ctx, g, scope.With(Scope{})); err != nil {
			return nil, err
		}
		cfg.Apply(g, groupClasses())
		return g, nil
	})
}

// Addon renders a non-interactive leading/trailing addon (icon or text).
func Addon(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		span := dom.El(atom.Span, dom.Marker("input-addon"))
		if err := render.Children(ctx, span); err != nil {
			return nil, err
		}
		cfg.Apply(span, addonClasses())
		return span, nil
	})
}

func groupClasses() string {
	var b styles.Builder
	b.Add("flex items-center w-full h-10 rounded-lg overflow-hidden")
	b.Add("bg-white text-sm shadow-xs")
	b.Add("border border-base-200 border-b-base-300/80 dark:border-base-600 dark:bg-base-700")
	// The group shows the focus ring for whichever control it wraps.
	b.Add("focus-within:outline focus-within:outline-2 focus-within:outline-accent focus-within:outline-offset-2")
	b.Add("has-[:disabled]:opacity-75")
	return b.String()
}

func addonClasses() string {
	var b styles.Builder
	b.Add("inline-flex items-center gap-1 px-3 h-full shrink-0 whitespace-nowrap")
	b.Add("text-base-500 dark:text-base-400")
	b.Add("**:data-[ui=icon]:size-5 **:data-[ui=icon]:text-base-400")
	// Separator between an addon and the field, drawn on the inward edge.
	b.Add("[&:first-child]:border-e [&:last-child]:border-s border-base-200 dark:border-base-600")
	return b.String()
}
