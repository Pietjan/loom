// Package launcher renders a command palette - a search field above a list
// of actions, filtered as you type:
//
//	@launcher.Root(launcher.Name("cmdk")) {
//		@launcher.Input(launcher.Field("q"), launcher.Placeholder("Search…"),
//			launcher.Attr("hx-post", "/commands"),
//			launcher.Attr("hx-trigger", "keyup changed delay:150ms"),
//			launcher.Attr("hx-target", "#loom-launcher-cmdk-list"))
//		@launcher.List() {
//			for _, c := range results {
//				@launcher.Item(c.Href) { { c.Title } }
//			}
//		}
//		@launcher.Empty() { No commands match. }
//	}
//
// Narrowing the list is the caller's job, server-side: the input carries
// whatever request attributes you put on it, the server re-renders List,
// and the swap target is the list id - which Name makes predictable. That
// is the whole component, and it needs no script.
//
// Rows are links and buttons: activating one navigates or submits. This is
// a launcher, not a form control - it produces no value, holds no
// selection, and nothing here is submitted but the query. For choosing a
// value from a set, see picker; the filter-as-you-type form control is a
// different component and does not exist yet.
//
// That is also why there is no listbox ARIA. role="option" on a link
// overrides the link role, so it would announce "option" while activating
// it navigates - the wrong interaction, described confidently. A search
// field above a list of links is what this is, and is what it says. A
// script layering arrow keys on top should move focus between the rows,
// which keeps Enter native and needs no roles at all.
//
// Compose it inside modal for the Cmd+K case, or use it inline as a
// filtered nav list.
package launcher

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

// ErrNoScope is returned when a part renders outside launcher.Root without
// being told which launcher it belongs to.
var ErrNoScope = errors.New("launcher: must be inside launcher.Root, or given launcher.For")

// Scope carries the generated ids from Root to the other parts.
type Scope struct {
	InputID string
	ListID  string
	EmptyID string
}

// Config holds launcher options.
type Config struct {
	opts.Common
	PairName    string
	FieldName   string
	Placeholder string
}

// Option configures a launcher part.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Name gives the launcher a stable, user-chosen id stem, so every part is
// addressable as loom-launcher-<name>-{input,list,empty} and each row as
// -opt-<n>. Without it the stem is generated, which is fine for a page
// rendered once and wrong for anything swapped: a fragment re-rendered on
// its own gets fresh ids, so a morph sees every row as new and throws away
// the focus and scroll position it was supposed to preserve.
func Name(name string) Option { return func(c *Config) { c.PairName = name } }

// For addresses a named launcher from a part rendered outside Root - the
// case that matters here, because the swap target is the list: a handler
// answering the filter request renders launcher.List(launcher.For("cmdk"))
// on its own and reproduces exactly the ids the full page emitted.
func For(name string) Option { return func(c *Config) { c.PairName = name } }

// Field sets the query input's form-control name. Spelled out rather than
// reusing Name, which is taken by the id pairing above.
func Field(name string) Option { return func(c *Config) { c.FieldName = name } }

// Placeholder sets the query input's placeholder text.
func Placeholder(p string) Option { return func(c *Config) { c.Placeholder = p } }

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
	return stem("loom-launcher-" + cfg.PairName), nil
}

func stem(s string) Scope {
	return Scope{InputID: s + "-input", ListID: s + "-list", EmptyID: s + "-empty"}
}

// Root renders the wrapper and establishes the id stem for its parts.
func Root(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}

		s := "loom-launcher-" + cfg.PairName
		if cfg.PairName == "" {
			s = ids.New(ctx, "launcher")
		}

		root := dom.El(atom.Div, dom.Marker("launcher"))
		if err := render.Children(ctx, root, scope.With(stem(s))); err != nil {
			return nil, err
		}
		cfg.Apply(root, rootClasses())
		return root, nil
	})
}

// Input renders the query field above the list.
func Input(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		sc, err := resolve(ctx, cfg)
		if err != nil {
			return nil, err
		}

		in := dom.El(atom.Input, dom.Marker("launcher-input"),
			dom.Attr("id", sc.InputID),
			dom.Attr("type", "search"),
			// Off, not because the native list is unwanted, but because it
			// would render a second dropdown over this one.
			dom.Attr("autocomplete", "off"))
		if cfg.FieldName != "" {
			dom.SetAttr(in, "name", cfg.FieldName)
		}
		if cfg.Placeholder != "" {
			dom.SetAttr(in, "placeholder", cfg.Placeholder)
		}
		cfg.Apply(in, inputClasses())
		return in, nil
	})
}

// List renders the results panel and stamps a stable id on every row.
//
// The ids are derived from the list's own id and the row's position, so a
// server re-rendering just this fragment reproduces them exactly - which is
// what lets a morph match rows instead of replacing them. They are not ARIA
// wiring; nothing here points at them, and a script does not need them to
// move focus. They are there so the row a swap lands on is the row that was
// already in the page.
func List(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		sc, err := resolve(ctx, cfg)
		if err != nil {
			return nil, err
		}

		list := dom.El(atom.Div, dom.Marker("launcher-list"), dom.Attr("id", sc.ListID))
		if err := render.Children(ctx, list); err != nil {
			return nil, err
		}
		// Post-pass: attributes only, never structure.
		for i, row := range dom.FindAllShallow(list, dom.ByMarker("launcher-item")) {
			dom.SetAttr(row, "id", sc.ListID+"-opt-"+strconv.Itoa(i))
		}
		cfg.Apply(list, listClasses())
		return list, nil
	})
}

// Item renders a link row.
func Item(href string, options ...Option) templ.Component {
	return row(atom.A, dom.Attr("href", href), options)
}

// ItemButton renders a button row, for form submits or custom commands.
func ItemButton(options ...Option) templ.Component {
	return row(atom.Button, dom.Attr("type", "button"), options)
}

func row(tag atom.Atom, extra html.Attribute, options []Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		n := dom.El(tag, dom.Marker("launcher-item"), extra)
		if err := render.Children(ctx, n); err != nil {
			return nil, err
		}
		cfg.Apply(n, itemClasses())
		return n, nil
	})
}

// Empty renders the no-results message. Nothing toggles it: css/loom.css
// shows it exactly when the launcher has focus and no row is left visible,
// so it is correct whether the rows were narrowed by the server or hidden
// by a script.
func Empty(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		sc, err := resolve(ctx, cfg)
		if err != nil {
			return nil, err
		}
		p := dom.El(atom.P, dom.Marker("launcher-empty"), dom.Attr("id", sc.EmptyID))
		if err := render.Children(ctx, p); err != nil {
			return nil, err
		}
		cfg.Apply(p, emptyClasses())
		return p, nil
	})
}
