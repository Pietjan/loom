// Package combobox renders a filter-as-you-type form control: a query
// field with a floating panel of choices under it.
//
//	@combobox.Root(combobox.Name("person"), combobox.Field("assignee")) {
//		@combobox.Input(combobox.Placeholder("Find a person…"))
//		@combobox.List() {
//			@combobox.Item(combobox.Value("ada")) { Ada Lovelace }
//			@combobox.Item(combobox.Value("alan"), combobox.Chosen()) { Alan Turing }
//		}
//		@combobox.Empty() { Nobody by that name. }
//	}
//
// This is the component launcher's doc points at as missing, and the two
// are deliberately not the same thing. A launcher is a command palette:
// its rows are links and buttons that do something, it holds no value, and
// its list is a panel in the flow. A combobox produces a value - Root
// renders the hidden input that carries it - and its list floats.
//
// # Floating, by the platform
//
// The list carries the popover attribute, so the top layer, light dismiss
// and Esc are the browser's rather than a script's, and it escapes any
// clipping ancestor. Root sets an anchor-name and the list a matching
// position-anchor; where CSS anchor positioning is missing the list falls
// back to absolute under the field (css/loom.css), which is why Root is
// also a positioning context.
//
// # What this component does not do
//
// Narrowing the list, deciding when it is open, and moving the cursor
// through it are the composer's, exactly as they are for launcher: this
// renders the state it is told. Nothing here opens the panel, because
// nothing here runs. A driver calls showPopover() when results arrive and
// hidePopover() when a choice is made; the platform handles every other
// way a popover closes.
//
// # No listbox roles, for launcher's reason
//
// Rows are buttons and a script layering arrow keys on them should move
// real focus, which keeps Enter native. role="option" on a focusable
// button describes a widget whose cursor is aria-activedescendant and
// whose focus never moves - announcing one interaction while implementing
// another. The field is marked aria-expanded and aria-controls so the
// relationship is stated; the rows say what they are.
package combobox

import (
	"context"
	"errors"
	"strconv"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/ids"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
	"github.com/pietjan/loom/internal/scope"
)

// ErrNoScope is returned when a part renders outside combobox.Root without
// being told which combobox it belongs to.
var ErrNoScope = errors.New("combobox: must be inside combobox.Root, or given combobox.For")

// Scope carries the generated ids from Root to the other parts.
type Scope struct {
	InputID    string
	ListID     string
	EmptyID    string
	AnchorName string
}

// Config holds combobox options.
type Config struct {
	opts.Common
	PairName    string
	FieldName   string
	Placeholder string
	ItemValue   string
	Query       string
	expanded    bool
	chosen      bool
	disabled    bool
}

// Option configures a combobox part.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Name gives the combobox a stable, user-chosen id stem, so every part is
// addressable as loom-combobox-<name>-{input,list,empty} and each row as
// -opt-<n>.
//
// It matters more here than anywhere else in loom. The list is a popover,
// and a popover's open state lives on the element: re-render the panel with
// a fresh id and the morph replaces it, which closes it. A combobox
// re-renders on every keystroke, so a generated id would close the list as
// the user typed.
func Name(name string) Option { return func(c *Config) { c.PairName = name } }

// For addresses a named combobox from a part rendered outside Root - a
// handler answering the filter request with just the list.
func For(name string) Option { return func(c *Config) { c.PairName = name } }

// Field sets the form-control name of the hidden input carrying the chosen
// value. Without it the combobox submits nothing and is a filter rather
// than a control.
func Field(name string) Option { return func(c *Config) { c.FieldName = name } }

// Placeholder sets the query field's placeholder text.
func Placeholder(p string) Option { return func(c *Config) { c.Placeholder = p } }

// Query sets the text in the query field. Rendering it explicitly is not
// optional for a component that is re-rendered: a morph compares an input's
// value attribute, so omitting it clears what was typed.
func Query(q string) Option { return func(c *Config) { c.Query = q } }

// Value sets a row's value, and on Root the currently chosen one - what
// the hidden input submits.
func Value(v string) Option { return func(c *Config) { c.ItemValue = v } }

// Expanded marks the query field as having its list open (aria-expanded).
// The composer sets it because the composer decides.
func Expanded() Option { return func(c *Config) { c.expanded = true } }

// Chosen marks the row matching the current value.
func Chosen() Option { return func(c *Config) { c.chosen = true } }

// Disabled renders a row as unavailable.
func Disabled() Option { return func(c *Config) { c.disabled = true } }

func build(options []Option) Config {
	cfg := Config{}
	for _, opt := range options {
		opt(&cfg)
	}
	return cfg
}

// resolve finds the part's id stem: the scope installed by Root when the
// part renders inside it, or the name it was handed directly when it does
// not.
func resolve(ctx context.Context, cfg Config) (Scope, error) {
	if sc, ok := scope.From[Scope](ctx); ok {
		return sc, nil
	}
	if cfg.PairName == "" {
		return Scope{}, ErrNoScope
	}
	return stem("loom-combobox-" + cfg.PairName), nil
}

func stem(s string) Scope {
	return Scope{
		InputID:    s + "-input",
		ListID:     s + "-list",
		EmptyID:    s + "-empty",
		AnchorName: "--" + s,
	}
}

// Root renders the wrapper, establishes the id stem, and carries the
// chosen value in a hidden input.
func Root(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := build(options)

		s := "loom-combobox-" + cfg.PairName
		if cfg.PairName == "" {
			s = ids.New(ctx, "combobox")
		}
		sc := stem(s)

		root := dom.El(atom.Div, dom.Marker("combobox"))
		// The anchor for the panel. Inline because it names this instance:
		// a class cannot carry a per-instance custom ident.
		dom.AddAttr(root, "style", "anchor-name: "+sc.AnchorName)

		if err := render.Children(ctx, root, scope.With(sc)); err != nil {
			return nil, err
		}

		// The value the control submits. Rendered here rather than by the
		// caller so that a combobox with a Field is a form control without
		// anything else being remembered.
		if cfg.FieldName != "" {
			hidden := dom.El(atom.Input, dom.Marker("combobox-value"),
				dom.Attr("type", "hidden"),
				dom.Attr("name", cfg.FieldName),
				dom.Attr("value", cfg.ItemValue))
			root.AppendChild(hidden)
		}

		cfg.Apply(root, rootClasses())
		return root, nil
	})
}

// Input renders the query field.
func Input(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := build(options)
		sc, err := resolve(ctx, cfg)
		if err != nil {
			return nil, err
		}

		in := dom.El(atom.Input, dom.Marker("combobox-input"),
			dom.Attr("id", sc.InputID),
			dom.Attr("type", "text"),
			// Off, not because the native list is unwanted, but because it
			// would render a second dropdown over this one.
			dom.Attr("autocomplete", "off"),
			dom.Attr("aria-controls", sc.ListID),
			dom.Attr("aria-expanded", strconv.FormatBool(cfg.expanded)))
		dom.SetAttr(in, "value", cfg.Query)
		if cfg.Placeholder != "" {
			dom.SetAttr(in, "placeholder", cfg.Placeholder)
		}
		cfg.Apply(in, inputClasses())
		return in, nil
	})
}

// List renders the floating panel and stamps a stable id on every row.
//
// The ids are derived from the list's own id and the row's position, so a
// server re-rendering just this fragment reproduces them exactly - which is
// what lets a morph match rows instead of replacing them, and keep the
// focus that is the cursor.
func List(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := build(options)
		sc, err := resolve(ctx, cfg)
		if err != nil {
			return nil, err
		}

		list := dom.El(atom.Div, dom.Marker("combobox-list"),
			dom.Attr("id", sc.ListID),
			dom.Attr("popover", ""))
		dom.AddAttr(list, "style", "position-anchor: "+sc.AnchorName)

		if err := render.Children(ctx, list); err != nil {
			return nil, err
		}
		// Post-pass: attributes only, never structure.
		for i, row := range dom.FindAllShallow(list, dom.ByMarker("combobox-item")) {
			dom.SetAttr(row, "id", sc.ListID+"-opt-"+strconv.Itoa(i))
		}
		cfg.Apply(list, listClasses())
		return list, nil
	})
}

// Item renders one choice. Rows are buttons: activating one chooses a
// value, and Enter on a focused button needs no help.
func Item(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := build(options)

		n := dom.El(atom.Button, dom.Marker("combobox-item"),
			dom.Attr("type", "button"),
			dom.Attr("data-value", cfg.ItemValue))
		if cfg.chosen {
			// The chosen row, styled off the attribute - not aria-selected,
			// which belongs to a listbox this deliberately is not.
			dom.SetAttr(n, "data-chosen", "")
		}
		if cfg.disabled {
			dom.SetAttr(n, "disabled", "")
		}
		if err := render.Children(ctx, n); err != nil {
			return nil, err
		}
		cfg.Apply(n, itemClasses())
		return n, nil
	})
}

// Empty renders the no-results message inside the panel.
func Empty(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := build(options)
		sc, err := resolve(ctx, cfg)
		if err != nil {
			return nil, err
		}
		p := dom.El(atom.P, dom.Marker("combobox-empty"), dom.Attr("id", sc.EmptyID))
		if err := render.Children(ctx, p); err != nil {
			return nil, err
		}
		cfg.Apply(p, emptyClasses())
		return p, nil
	})
}
