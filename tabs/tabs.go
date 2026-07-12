// Package tabs renders client-side tabbed content on a <details name>
// disclosure group — all panels ship in the page and the platform
// switches them, zero JavaScript:
//
//	@tabs.New(tabs.Label("Account sections")) {
//		@tabs.Section(tabs.Title("Profile"), tabs.Open()) { ... }
//		@tabs.Section(tabs.Title("Billing")) { ... }
//	}
//
// The tab layout relies on ::details-content (Baseline 2025); older
// browsers fall back to a vertical accordion via @supports — content is
// never lost. Screen readers announce disclosure semantics ("expanded/
// collapsed"), which is what this honestly is. Known caveat: keyboard
// users can close the open section (Enter on its summary), leaving all
// panels hidden; the mouse path is guarded with pointer-events.
//
// Honest limitation, by design: the full ARIA tabs pattern (role=tab,
// arrow-key roving) cannot be met without JavaScript, and loom ships
// none.
package tabs

import (
	"context"
	"errors"
	"fmt"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/ids"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
	"github.com/pietjan/loom/internal/scope"
)

// ErrNoGroup is returned when a Section renders outside tabs.New.
var ErrNoGroup = errors.New("tabs: Section must be inside tabs.New")

// groupScope shares the exclusive <details name> with sections.
type groupScope struct {
	Name string
}

// Config holds tabs options.
type Config struct {
	opts.Common
	LabelText string
	TitleText string
	open      bool
}

// Option configures tabs.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Label names the tab group for assistive tech.
func Label(label string) Option { return func(c *Config) { c.LabelText = label } }

// Title sets a Section's tab handle text.
func Title(title string) Option { return func(c *Config) { c.TitleText = title } }

// Open marks the Section that starts active. Without it, the group opens
// its first section.
func Open() Option { return func(c *Config) { c.open = true } }

// New renders the tab group; see the package doc for semantics and
// fallback behavior. Layout rules live in css/loom.css
// ([data-ui=tabs-group]).
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}

		g := dom.El(atom.Div, dom.Marker("tabs-group"))
		if cfg.LabelText != "" {
			dom.SetAttr(g, "aria-label", cfg.LabelText)
		}
		name := ids.New(ctx, "tabs")
		if err := render.Children(ctx, g, scope.With(groupScope{Name: name})); err != nil {
			return nil, err
		}

		// The grid needs one explicit column per tab handle (the panel
		// spans 1/-1, which only covers the explicit grid), plus a filler
		// so handles don't stretch. Known only now that children exist.
		sections := dom.FindAllShallow(g, dom.ByMarker("tabs-section"))
		dom.AddAttr(g, "style", fmt.Sprintf("grid-template-columns: repeat(%d, max-content) 1fr", len(sections)))

		// Tabs always show a panel: default the first section open.
		open := false
		for _, s := range sections {
			open = open || dom.HasAttr(s, "open")
		}
		if !open && len(sections) > 0 {
			dom.SetAttr(sections[0], "open", "")
		}

		cfg.Apply(g, groupClasses())
		return g, nil
	})
}

// Section renders one tab (handle + panel) inside tabs.New.
func Section(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		gs, ok := scope.From[groupScope](ctx)
		if !ok {
			return nil, ErrNoGroup
		}
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}

		d := dom.El(atom.Details, dom.Marker("tabs-section"), dom.Attr("name", gs.Name))
		if cfg.open {
			dom.SetAttr(d, "open", "")
		}

		summary := dom.El(atom.Summary, dom.Attr("class", sectionTabClasses()))
		summary.AppendChild(dom.Text(cfg.TitleText))
		d.AppendChild(summary)

		panel := dom.El(atom.Div, dom.Marker("tabs-section-panel"))
		if err := render.Children(ctx, panel); err != nil {
			return nil, err
		}
		dom.SetAttr(panel, "class", sectionPanelClasses())
		d.AppendChild(panel)

		cfg.Apply(d, sectionClasses())
		return d, nil
	})
}
