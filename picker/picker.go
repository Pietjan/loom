// Package picker renders a styled <select> built on the customizable
// select API (appearance: base-select, <selectedcontent>, ::picker):
//
//	@picker.New(picker.Name("pet"), picker.Placeholder("Choose a pet")) {
//		@picker.Item("cat") { Cat }
//		@picker.Item("dog", picker.Selected()) { Dog }
//	}
//
// In Chromium the picker is fully styled and options may hold rich content
// (icons). Everywhere else it degrades — by the spec's own design — to a
// classic select: the internal <button><selectedcontent> is ignored and
// option content collapses to text. Submission semantics are identical in
// both worlds. The base-select rules live in css/loom.css.
package picker

import (
	"context"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/field"
	"github.com/pietjan/loom/icon"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
	"github.com/pietjan/loom/internal/scope"
)

// Config holds picker options.
type Config struct {
	opts.Common
	PlaceholderText string
	invalid         bool
}

// Option configures a picker or one of its options.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Name sets the form field name.
func Name(name string) Option { return Attr("name", name) }

// Placeholder renders a disabled, hidden first option shown until a real
// selection is made.
func Placeholder(text string) Option { return func(c *Config) { c.PlaceholderText = text } }

// Selected pre-selects an option (Option only).
func Selected() Option { return Attr("selected", "") }

// Disabled disables the picker or a single option.
func Disabled() Option { return Attr("disabled", "") }

// New renders a picker as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the <select> node.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{}
	for _, opt := range options {
		opt(&cfg)
	}

	sel := dom.El(atom.Select, dom.Marker("picker"))

	// The customizable-select internal button: ignored (harmless) by
	// non-supporting browsers, the styled closed state in Chromium. The
	// arrow is a real icon — the browser's ::picker-icon is hidden in
	// css/loom.css, since pseudo-element content can't use currentColor or
	// our transitions.
	btn := dom.El(atom.Button, dom.Attr("class", buttonClasses()))
	btn.AppendChild(dom.CustomEl("selectedcontent"))
	arrow, err := icon.Node(ctx, icon.CaretDown, icon.ExtraSmall, icon.Class(arrowClasses()))
	if err != nil {
		return nil, err
	}
	btn.AppendChild(arrow)
	sel.AppendChild(btn)

	var placeholder *html.Node
	if cfg.PlaceholderText != "" {
		placeholder = dom.El(atom.Option, dom.Attr("value", ""),
			dom.Attr("disabled", ""), dom.Attr("hidden", ""))
		placeholder.AppendChild(dom.Text(cfg.PlaceholderText))
		sel.AppendChild(placeholder)
	}

	// Parse children against a neutral context: x/net/html's in-select
	// insertion mode predates rich option content and would strip it.
	if err := render.ChildrenAs(ctx, sel, dom.El(atom.Div)); err != nil {
		return nil, err
	}

	// The placeholder is only the selected default when no option claimed
	// selection (decidable only after children exist).
	if placeholder != nil && dom.FindShallow(sel, dom.ByAttr("selected")) == nil {
		dom.SetAttr(placeholder, "selected", "")
	}

	if sc, ok := scope.From[field.Scope](ctx); ok {
		cfg.invalid = sc.Invalid
		if sc.Invalid {
			dom.SetAttr(sel, "aria-invalid", "true")
		}
		if sc.Required {
			dom.SetAttr(sel, "required", "")
		}
		if sc.Disabled {
			dom.SetAttr(sel, "disabled", "")
		}
	}

	cfg.Apply(sel, classes(cfg))
	return sel, nil
}

// Item renders one option. Rich content (icons, secondary text) is
// allowed; non-supporting browsers collapse it to its text.
//
// Each option carries a trailing icon checkmark, visible only while
// checked — the browser's ::checkmark pseudo is hidden in css/loom.css.
// The data-picker-check attribute lets loom.css hide the checkmark's
// clone inside <selectedcontent> (the browser copies the checked option's
// children verbatim).
func Item(value string, options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		o := dom.El(atom.Option, dom.Marker("picker-option"), dom.Attr("value", value))
		if err := render.Children(ctx, o); err != nil {
			return nil, err
		}
		check, err := icon.Node(ctx, icon.Check, icon.ExtraSmall, icon.Class(checkClasses()))
		if err != nil {
			return nil, err
		}
		dom.SetAttr(check, "data-picker-check", "")
		o.AppendChild(check)
		cfg.Apply(o, optionClasses())
		return o, nil
	})
}
