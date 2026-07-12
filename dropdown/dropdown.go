// Package dropdown renders popover menus anchored to their trigger — no
// JavaScript:
//
//	@dropdown.Root() {
//		@dropdown.Trigger() {
//			@button.New() { Options }
//		}
//		@dropdown.Menu() {
//			@dropdown.Item("/profile") { Profile }
//			@dropdown.ItemButton(dropdown.Attr("name", "logout")) { Log out }
//			@dropdown.Divider()
//			@dropdown.Item("/settings") { Settings }
//		}
//	}
//
// Root renders a positioning wrapper and generates the pair: the trigger's
// button gets command="toggle-popover" + an anchor-name; the menu is a
// [popover] panel with the matching position-anchor. Positioning uses CSS
// anchor positioning where supported (Chromium), with a plain
// flow-position fallback (absolute under the wrapper) everywhere else —
// rules in css/loom.css. Light dismiss and Esc come from the popover.
//
// The panel is deliberately NOT role="menu": the ARIA menu pattern
// requires JS roving tabindex. Plain focusable links and buttons,
// Tab-navigable, are honest and correct.
package dropdown

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
)

// ErrNoScope is returned when Trigger or Menu render outside dropdown.Root.
var ErrNoScope = errors.New("dropdown: must be inside dropdown.Root")

// ErrNoButton is returned when a Trigger block contains no button.
var ErrNoButton = errors.New("dropdown: trigger needs a <button> in its block (e.g. @button.New())")

// Scope carries the generated pair from Root to Trigger and Menu.
type Scope struct {
	MenuID     string
	AnchorName string
}

// Config holds dropdown options.
type Config struct {
	opts.Common
}

// Option configures a dropdown part.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Root renders the positioning wrapper and generates the trigger/menu pair.
func Root(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}

		id := ids.New(ctx, "dropdown")
		sc := Scope{MenuID: id, AnchorName: "--" + id}

		// The wrapper is the containing block for the non-anchor fallback
		// (position: relative + absolute menu).
		root := dom.El(atom.Div, dom.Marker("dropdown"))
		if err := render.Children(ctx, root, scope.With(sc)); err != nil {
			return nil, err
		}
		cfg.Apply(root, rootClasses())
		return root, nil
	})
}

// Trigger wires the first button in its block to toggle the menu.
func Trigger(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		sc, ok := scope.From[Scope](ctx)
		if !ok {
			return nil, ErrNoScope
		}
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
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
		dom.SetAttr(btn, "commandfor", sc.MenuID)
		dom.AddAttr(btn, "style", "anchor-name: "+sc.AnchorName)
		return wrap, nil
	})
}

// Menu renders the [popover] panel.
func Menu(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		sc, ok := scope.From[Scope](ctx)
		if !ok {
			return nil, ErrNoScope
		}
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}

		menu := dom.El(atom.Div, dom.Marker("dropdown-menu"),
			dom.Attr("id", sc.MenuID), dom.Attr("popover", ""))
		dom.AddAttr(menu, "style", "position-anchor: "+sc.AnchorName)
		if err := render.Children(ctx, menu); err != nil {
			return nil, err
		}
		cfg.Apply(menu, menuClasses())
		return menu, nil
	})
}

// Item renders a link row.
func Item(href string, options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		a := dom.El(atom.A, dom.Marker("dropdown-item"), dom.Attr("href", href))
		if err := render.Children(ctx, a); err != nil {
			return nil, err
		}
		cfg.Apply(a, itemClasses())
		return a, nil
	})
}

// ItemButton renders a button row (for form submits or custom commands).
func ItemButton(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		b := dom.El(atom.Button, dom.Marker("dropdown-item"), dom.Attr("type", "button"))
		if err := render.Children(ctx, b); err != nil {
			return nil, err
		}
		cfg.Apply(b, itemClasses())
		return b, nil
	})
}

// Divider renders a separator row.
func Divider(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		hr := dom.El(atom.Hr, dom.Marker("dropdown-divider"))
		cfg.Apply(hr, dividerClasses())
		return hr, nil
	})
}
