// Package modal renders native <dialog> modals driven entirely by invoker
// commands — no JavaScript:
//
//	@modal.Root() {
//		@modal.Trigger() {
//			@button.New() { Delete account }
//		}
//		@modal.Content() {
//			@modal.Title() { Are you sure? }
//			@text.New() { This cannot be undone. }
//			@modal.Close() {
//				@button.New() { Cancel }
//			}
//		}
//	}
//
// Root renders no element: it generates the dialog id and shares it via
// scope, so Trigger (command="show-modal") and Close (command="close")
// wire themselves to Content's <dialog>. A trigger far away from the
// dialog skips Root and pairs by name:
//
//	@modal.Trigger(modal.For("confirm")) { ... }
//	@modal.Content(modal.Name("confirm")) { ... }
//
// The browser provides focus trapping, Esc, ::backdrop, and focus return.
// closedby="any" enables light dismiss (click outside).
package modal

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

// ErrNoTarget is returned when a Trigger or Close cannot resolve its
// dialog: not inside modal.Root and no modal.For given. Failing loudly
// beats emitting a dead button.
var ErrNoTarget = errors.New("modal: trigger has no target — wrap it in modal.Root or pass modal.For(...)")

// ErrNoButton is returned when a Trigger or Close block contains no button
// to wire.
var ErrNoButton = errors.New("modal: trigger needs a <button> in its block (e.g. @button.New())")

// Scope carries the dialog id from Root to Trigger/Content/Close.
type Scope struct {
	DialogID string
}

// Config holds modal options.
type Config struct {
	opts.Common
	PairName string
}

// Option configures a modal part.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Name gives the dialog a stable, user-chosen id (on Root or Content).
func Name(name string) Option { return func(c *Config) { c.PairName = name } }

// For points a distant Trigger or Close at a named dialog.
func For(name string) Option { return func(c *Config) { c.PairName = name } }

// Root coordinates a trigger/content pair. It renders no element.
func Root(options ...Option) templ.Component {
	return render.Coordinator(func(ctx context.Context) (context.Context, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		id := cfg.PairName
		if id == "" {
			id = ids.New(ctx, "modal")
		}
		return scope.Set(ctx, Scope{DialogID: id}), nil
	})
}

// Trigger wires the first button in its block to open the dialog.
func Trigger(options ...Option) templ.Component {
	return command("show-modal", options)
}

// Close wires the first button in its block to close the dialog.
func Close(options ...Option) templ.Component {
	return command("close", options)
}

// command renders a pass-through block whose first shallow button gets
// command/commandfor. No wrapper element is emitted around a single
// button; multiple children keep their natural structure inside a
// contents-display span.
func command(cmd string, options []Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}

		id, err := target(ctx, cfg)
		if err != nil {
			return nil, err
		}

		wrap := dom.El(atom.Span, dom.Attr("class", "contents"))
		if err := render.Children(ctx, wrap); err != nil {
			return nil, err
		}

		btn := dom.FindShallow(wrap, dom.Any(dom.ByTag(atom.Button), dom.ByMarker("button")))
		if btn == nil {
			return nil, ErrNoButton
		}
		dom.SetAttr(btn, "command", cmd)
		dom.SetAttr(btn, "commandfor", id)
		return wrap, nil
	})
}

// Content renders the <dialog>.
func Content(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the <dialog> node.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{}
	for _, opt := range options {
		opt(&cfg)
	}

	id, err := target(ctx, cfg)
	if err != nil {
		return nil, errors.New("modal: content has no id — wrap it in modal.Root or pass modal.Name(...)")
	}

	d := dom.El(atom.Dialog, dom.Marker("modal"), dom.Attr("id", id), dom.Attr("closedby", "any"))
	if err := render.Children(ctx, d); err != nil {
		return nil, err
	}

	// A titled dialog labels itself.
	if title := dom.FindShallow(d, dom.ByMarker("modal-title")); title != nil {
		if dom.GetAttr(title, "id") == "" {
			dom.SetAttr(title, "id", id+"-title")
		}
		dom.SetAttr(d, "aria-labelledby", dom.GetAttr(title, "id"))
	}

	cfg.Apply(d, classes(cfg))
	return d, nil
}

// Title renders the dialog heading; Content wires aria-labelledby to it.
func Title(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		h := dom.El(atom.H2, dom.Marker("modal-title"))
		if err := render.Children(ctx, h); err != nil {
			return nil, err
		}
		cfg.Apply(h, titleClasses())
		return h, nil
	})
}

// target resolves the dialog id: explicit For/Name first, then Root scope.
func target(ctx context.Context, cfg Config) (string, error) {
	if cfg.PairName != "" {
		return cfg.PairName, nil
	}
	if sc, ok := scope.From[Scope](ctx); ok {
		return sc.DialogID, nil
	}
	return "", ErrNoTarget
}
