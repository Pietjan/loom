// Package button provides pre-styled buttons in six variants and three
// sizes:
//
//	@button.New(button.Primary) { Save }
//	@button.New(button.Ghost, button.Label("Close")) {
//		@icon.New(icon.XMark)
//	}
//
// An icon-only button renders square and MUST be given an accessible name
// via button.Label (or an aria-label attribute); rendering fails otherwise.
// Group arranges buttons as a segmented control.
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

// New renders a button as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the <button> node.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{Type: TypeButton, Variant: VariantOutline, Size: SizeBase}
	for _, opt := range options {
		opt(&cfg)
	}

	btn := dom.El(atom.Button, dom.Marker("button"))
	if err := render.Children(ctx, btn); err != nil {
		return nil, err
	}

	cfg.square = iconOnly(btn)
	if cfg.square && cfg.Label == "" && dom.GetAttribute(cfg.Attrs, "aria-label") == "" {
		return nil, ErrNoAccessibleName
	}

	dom.SetAttr(btn, "type", string(cfg.Type))
	if cfg.Disabled {
		dom.SetAttr(btn, "disabled", "")
	}
	if cfg.Label != "" {
		dom.SetAttr(btn, "aria-label", cfg.Label)
	}
	cfg.Apply(btn, classes(cfg))
	return btn, nil
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
