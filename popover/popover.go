// Package popover renders a floating panel anchored to a trigger — richer
// than a tooltip, more free-form than a dropdown menu. Zero JS:
//
//	@popover.Root() {
//		@popover.Trigger() {
//			@button.New(button.Ghost) { Details }
//		}
//		@popover.Content() {
//			@heading.New() { Storage }
//			@text.New() { 8.2 GB of 10 GB used. }
//		}
//	}
//
// Root generates the pair: the trigger button gets
// command="toggle-popover" and an anchor-name; the content is a [popover]
// panel with the matching position-anchor. Positioning uses CSS anchor
// positioning where supported, with a flow-position fallback (see
// cmd/css/loom.css). Light dismiss and Esc come from the Popover API.
package popover

import (
	"context"
	"errors"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/ids"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
	"github.com/pietjan/loom/internal/scope"
	"github.com/pietjan/loom/internal/styles"
)

// ErrNoScope is returned when Trigger or Content render outside popover.Root.
var ErrNoScope = errors.New("popover: must be inside popover.Root")

// ErrNoButton is returned when a Trigger block contains no button.
var ErrNoButton = errors.New("popover: trigger needs a <button> in its block")

// Scope carries the generated pair from Root to Trigger and Content.
type Scope struct {
	PanelID    string
	AnchorName string
}

// Config holds popover options.
type Config struct {
	opts.Common
}

// Option configures a popover part.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Root renders the positioning wrapper and generates the trigger/content pair.
func Root(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		id := ids.New(ctx, "popover")
		sc := Scope{PanelID: id, AnchorName: "--" + id}

		root := dom.El(atom.Div, dom.Marker("popover"))
		if err := render.Children(ctx, root, scope.With(sc)); err != nil {
			return nil, err
		}
		cfg.Apply(root, "relative inline-flex")
		return root, nil
	})
}

// Trigger wires the first button in its block to toggle the panel.
func Trigger(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		sc, ok := scope.From[Scope](ctx)
		if !ok {
			return nil, ErrNoScope
		}
		wrap := dom.El(atom.Span, dom.Attr("class", "contents"))
		if err := render.Children(ctx, wrap); err != nil {
			return nil, err
		}
		btn := dom.FindShallow(wrap, dom.Any(dom.ByTag(atom.Button), dom.ByMarker("button")))
		if btn == nil {
			return nil, ErrNoButton
		}
		dom.SetAttr(btn, "command", "toggle-popover")
		dom.SetAttr(btn, "commandfor", sc.PanelID)
		dom.AddAttr(btn, "style", "anchor-name: "+sc.AnchorName)
		return wrap, nil
	})
}

// Content renders the [popover] panel.
func Content(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		sc, ok := scope.From[Scope](ctx)
		if !ok {
			return nil, ErrNoScope
		}
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		panel := dom.El(atom.Div, dom.Marker("popover-content"),
			dom.Attr("id", sc.PanelID), dom.Attr("popover", ""))
		dom.AddAttr(panel, "style", "position-anchor: "+sc.AnchorName)
		if err := render.Children(ctx, panel); err != nil {
			return nil, err
		}
		cfg.Apply(panel, contentClasses())
		return panel, nil
	})
}

func contentClasses() string {
	var b styles.Builder
	b.Add("w-72 max-w-[calc(100vw-2rem)] space-y-2 rounded-lg border border-base-200 bg-white p-4 shadow-lg")
	b.Add("dark:border-base-600 dark:bg-base-700")
	return b.String()
}
