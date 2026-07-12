// Package field composes a form control with its label, description, and
// error message — all wiring (for/id, aria-describedby, aria-invalid,
// required, disabled) is automatic:
//
//	@field.Root(field.Error(msg), field.Required()) {
//		@field.Label() { Email }
//		@input.New(input.Type("email"), input.Name("email"))
//		@field.Description() { We never share it. }
//	}
//
// How it works: Root installs a Scope (generated IDs + state flags) into
// the context; controls (input, textarea, picker, ...) read it while
// rendering themselves. After the children exist, Root runs a post-pass
// that wires label for= and the control's aria-describedby from the parts
// that actually rendered.
package field

import (
	"context"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/ids"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
	"github.com/pietjan/loom/internal/scope"
)

// Scope is installed by Root for the components inside its block.
// Controls adopt ControlID as their id (unless the user set one), mirror
// Invalid as aria-invalid plus error styling, and apply Required/Disabled.
type Scope struct {
	ControlID     string
	DescriptionID string
	ErrorID       string
	Invalid       bool
	Required      bool
	Disabled      bool
}

// controlMarkers are the data-ui values recognized as form controls for
// wiring. Control packages must keep this list in sync (asserted by the
// contract tests).
var controlMarkers = []string{"input", "textarea", "picker", "checkbox", "radio", "toggle"}

// Config holds field options.
type Config struct {
	opts.Common
	ErrorMsg string
	Required bool
	Disabled bool
}

// Option configures a field.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Error marks the field invalid and renders msg as the error message
// (after the children). An empty msg is a no-op, so validation results can
// be passed through unconditionally.
func Error(msg string) Option { return func(c *Config) { c.ErrorMsg = msg } }

// Required marks the control required.
func Required() Option { return func(c *Config) { c.Required = true } }

// Disabled disables the control.
func Disabled() Option { return func(c *Config) { c.Disabled = true } }

// Root renders the field wrapper and coordinates its parts.
func Root(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the field and wires its parts.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{}
	for _, opt := range options {
		opt(&cfg)
	}

	sc := Scope{
		ControlID:     ids.New(ctx, "field"),
		DescriptionID: ids.New(ctx, "field-desc"),
		ErrorID:       ids.New(ctx, "field-err"),
		Invalid:       cfg.ErrorMsg != "",
		Required:      cfg.Required,
		Disabled:      cfg.Disabled,
	}

	root := dom.El(atom.Div, dom.Marker("field"))
	if err := render.Children(ctx, root, scope.With(sc)); err != nil {
		return nil, err
	}

	if cfg.ErrorMsg != "" {
		errEl := dom.El(atom.P, dom.Marker("field-error"))
		errEl.AppendChild(dom.Text(cfg.ErrorMsg))
		dom.SetAttr(errEl, "class", errorClasses())
		root.AppendChild(errEl)
	}

	wire(root, sc)
	cfg.Apply(root, classes(cfg))
	return root, nil
}

// Label renders the field's label; for= is wired by Root's post-pass.
func Label(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		l := dom.El(atom.Label, dom.Marker("field-label"))
		if err := render.Children(ctx, l); err != nil {
			return nil, err
		}
		if sc, ok := scope.From[Scope](ctx); ok && sc.Required {
			mark := dom.El(atom.Span, dom.Attr("class", "text-red-500"), dom.Attr("aria-hidden", "true"))
			mark.AppendChild(dom.Text("*"))
			l.AppendChild(mark)
		}
		cfg.Apply(l, labelClasses())
		return l, nil
	})
}

// Description renders supporting text; its id joins the control's
// aria-describedby via Root's post-pass.
func Description(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		d := dom.El(atom.P, dom.Marker("field-description"))
		if err := render.Children(ctx, d); err != nil {
			return nil, err
		}
		cfg.Apply(d, descriptionClasses())
		return d, nil
	})
}

// wire connects the parts that actually rendered: label for= the control's
// real id, description/error ids, and the control's aria-describedby.
// Post-pass rules apply: attributes only, on marker-identified nodes.
func wire(root *html.Node, sc Scope) {
	control := findControl(root)

	controlID := sc.ControlID
	if control != nil {
		if id := dom.GetAttr(control, "id"); id != "" {
			controlID = id // user-supplied id wins; follow it everywhere
		} else {
			dom.SetAttr(control, "id", controlID)
		}
	}

	if label := dom.FindShallow(root, dom.ByMarker("field-label")); label != nil && control != nil {
		if dom.GetAttr(label, "for") == "" {
			dom.SetAttr(label, "for", controlID)
		}
	}

	var describedBy []string
	if desc := dom.FindShallow(root, dom.ByMarker("field-description")); desc != nil {
		if dom.GetAttr(desc, "id") == "" {
			dom.SetAttr(desc, "id", sc.DescriptionID)
		}
		describedBy = append(describedBy, dom.GetAttr(desc, "id"))
	}
	if errEl := dom.FindShallow(root, dom.ByMarker("field-error")); errEl != nil {
		if dom.GetAttr(errEl, "id") == "" {
			dom.SetAttr(errEl, "id", sc.ErrorID)
		}
		describedBy = append(describedBy, dom.GetAttr(errEl, "id"))
	}

	if control != nil && len(describedBy) > 0 {
		dom.SetAttr(control, "aria-describedby", join(describedBy))
	}
}

// findControl locates the field's form control. It looks shallowly first,
// then descends into an input-group (which is part of a control's
// presentation, so it's transparent to field wiring) if the control is
// wrapped in one.
func findControl(root *html.Node) *html.Node {
	if c := dom.FindShallow(root, dom.ByMarker(controlMarkers...)); c != nil {
		return c
	}
	if grp := dom.FindShallow(root, dom.ByMarker("input-group")); grp != nil {
		return dom.Find(grp, dom.ByMarker(controlMarkers...))
	}
	return nil
}

func join(ss []string) string {
	out := ss[0]
	for _, s := range ss[1:] {
		out += " " + s
	}
	return out
}
