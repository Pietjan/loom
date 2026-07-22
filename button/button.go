// Package button provides pre-styled buttons in six variants and three
// sizes:
//
//	@button.New(button.Primary) { Save }
//	@button.New(button.Ghost, button.Label("Close")) {
//		@icon.New(icon.X)
//	}
//
// An icon-only button renders square and MUST be given an accessible name
// via button.Label (or an aria-label attribute); rendering fails otherwise.
// Group arranges buttons as a segmented control.
//
// button.Href renders an <a> instead — a link wearing a button's clothes,
// for navigation that should look like an action:
//
//	@button.New(button.Primary, button.Href("/signup")) { Get started }
//
// Reach for it whenever the control navigates; a <button> with an onclick
// that assigns location breaks middle-click, "open in new tab" and
// crawlers. For a text link, use the link package instead.
package button

import (
	"context"
	"errors"
	"strings"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/render"
)

// ErrNoAccessibleName is returned for icon-only buttons without a label.
var ErrNoAccessibleName = errors.New("button: icon-only button needs an accessible name — add button.Label(...) or an aria-label attribute")

// ErrHrefWithType is returned when a link button is also given a type. An
// anchor cannot submit or reset a form, so the combination would silently
// render a link that does nothing the caller asked for.
var ErrHrefWithType = errors.New("button: button.Href cannot be combined with button.WithType — an anchor cannot submit or reset a form")

// New renders a button as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the <button> node — or, with Href, the <a> (or, when also
// disabled, the inert <span>) that wears the same clothes.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{Variant: VariantOutline, Size: SizeBase}
	for _, opt := range options {
		opt(&cfg)
	}
	if cfg.Href != "" && cfg.Type != "" {
		return nil, ErrHrefWithType
	}

	btn := dom.El(element(cfg), dom.Marker("button"))
	if err := render.Children(ctx, btn); err != nil {
		return nil, err
	}

	cfg.square = iconOnly(btn)
	if cfg.square && cfg.Label == "" && dom.GetAttribute(cfg.Attrs, "aria-label") == "" {
		return nil, ErrNoAccessibleName
	}

	switch {
	case cfg.Href == "":
		t := cfg.Type
		if t == "" {
			t = TypeButton
		}
		dom.SetAttr(btn, "type", string(t))
		if cfg.Disabled {
			dom.SetAttr(btn, "disabled", "")
		}
	case cfg.Disabled:
		// No href: a dead link is worse than no link. The inert span
		// matches how pagination renders an unavailable page.
		dom.SetAttr(btn, "aria-disabled", "true")
	default:
		dom.SetAttr(btn, "href", cfg.Href)
		if cfg.external {
			dom.SetAttr(btn, "target", "_blank")
			dom.SetAttr(btn, "rel", "noopener noreferrer")
		}
	}
	if cfg.Label != "" {
		dom.SetAttr(btn, "aria-label", cfg.Label)
	}
	cfg.Apply(btn, classes(cfg))
	return btn, nil
}

// element picks the tag: a real button unless Href asked for a link, and
// a span rather than an href-less anchor when that link is disabled —
// anchors without href are not focusable, and span says so honestly.
func element(c Config) atom.Atom {
	switch {
	case c.Href == "":
		return atom.Button
	case c.Disabled:
		return atom.Span
	default:
		return atom.A
	}
}

// Group renders a segmented control wrapper for buttons.
func Group(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		g := dom.El(atom.Div, dom.Marker("button-group"), dom.Attr("role", "group"))
		if err := render.Children(ctx, g); err != nil {
			return nil, err
		}
		cfg.Apply(g, groupClasses())
		return g, nil
	})
}

// iconOnly reports whether the button's only content is icons: at least
// one icon-marked element, no other elements, no non-whitespace text.
func iconOnly(btn *html.Node) bool {
	icons := 0
	for c := btn.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			if strings.TrimSpace(c.Data) != "" {
				return false
			}
		case html.ElementNode:
			if dom.MarkerName(c) != "icon" {
				return false
			}
			icons++
		}
	}
	return icons > 0
}
