// Package timeline renders a sequence of events or steps along a
// connector line, Flux-style: each item pairs an indicator (dot, icon,
// number, avatar) with content, and the line runs through the indicator
// column:
//
//	@timeline.New() {
//		@timeline.Item(timeline.Complete) {
//			@timeline.Indicator() { @icon.New(icon.Check, icon.ExtraSmall) }
//			@timeline.Content() {
//				@heading.New() { Deployed v2 }
//				@text.New() { Rolled out to production. }
//			}
//		}
//		@timeline.Item() {
//			@timeline.Content() { @text.New() { Header layout landed. } }
//		}
//	}
//
// An Item without an explicit Indicator gets a plain dot. Statuses
// (Complete, Current, Incomplete) color the indicator and its connector
// segment; Horizontal() lays the whole timeline out inline-start to
// inline-end. Large() enlarges indicators for numbered steps. The dots,
// lines, and layout are all CSS. Zero JS.
package timeline

import (
	"context"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
	"github.com/pietjan/loom/internal/scope"
)

// Size selects the indicator size.
type Size string

const (
	SizeBase  Size = "base" // default
	SizeLarge Size = "lg"   // room for step numbers
)

// Status styles an item's indicator and connector segment.
type Status string

const (
	StatusNone       Status = ""           // neutral (default)
	StatusComplete   Status = "complete"   // filled indicator, accent connector
	StatusCurrent    Status = "current"    // highlighted indicator, neutral connector
	StatusIncomplete Status = "incomplete" // muted indicator
)

// Color tints an indicator from the standard palette.
type Color string

const (
	ColorZinc    Color = "zinc"
	ColorRed     Color = "red"
	ColorOrange  Color = "orange"
	ColorAmber   Color = "amber"
	ColorYellow  Color = "yellow"
	ColorLime    Color = "lime"
	ColorGreen   Color = "green"
	ColorEmerald Color = "emerald"
	ColorTeal    Color = "teal"
	ColorCyan    Color = "cyan"
	ColorSky     Color = "sky"
	ColorBlue    Color = "blue"
	ColorIndigo  Color = "indigo"
	ColorViolet  Color = "violet"
	ColorPurple  Color = "purple"
	ColorFuchsia Color = "fuchsia"
	ColorPink    Color = "pink"
	ColorRose    Color = "rose"
)

// Config holds timeline options.
type Config struct {
	opts.Common
	Size       Size
	Status     Status
	Color      Color
	horizontal bool
	bare       bool
}

// Option configures the timeline, an item, or an indicator.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Horizontal lays items out inline instead of stacked (on New).
func Horizontal() Option { return func(c *Config) { c.horizontal = true } }

// WithSize sets the indicator size (on New, Item, or Indicator; inherited
// downward).
func WithSize(s Size) Option { return func(c *Config) { c.Size = s } }

// WithStatus sets the progress state (on Item, inherited by its
// Indicator, or on Indicator directly).
func WithStatus(s Status) Option { return func(c *Config) { c.Status = s } }

// WithColor tints the indicator (on Indicator).
func WithColor(col Color) Option { return func(c *Config) { c.Color = col } }

// Bare strips the indicator's circle — no background, no fixed size — so
// a custom child like an avatar stands alone.
func Bare() Option { return func(c *Config) { c.bare = true } }

// Pre-baked size, status, and color options.
var (
	Large      = WithSize(SizeLarge)
	Complete   = WithStatus(StatusComplete)
	Current    = WithStatus(StatusCurrent)
	Incomplete = WithStatus(StatusIncomplete)
	Zinc       = WithColor(ColorZinc)
	Red        = WithColor(ColorRed)
	Orange     = WithColor(ColorOrange)
	Amber      = WithColor(ColorAmber)
	Yellow     = WithColor(ColorYellow)
	Lime       = WithColor(ColorLime)
	Green      = WithColor(ColorGreen)
	Emerald    = WithColor(ColorEmerald)
	Teal       = WithColor(ColorTeal)
	Cyan       = WithColor(ColorCyan)
	Sky        = WithColor(ColorSky)
	Blue       = WithColor(ColorBlue)
	Indigo     = WithColor(ColorIndigo)
	Violet     = WithColor(ColorViolet)
	Purple     = WithColor(ColorPurple)
	Fuchsia    = WithColor(ColorFuchsia)
	Pink       = WithColor(ColorPink)
	Rose       = WithColor(ColorRose)
)

// listScope flows orientation and size from New to its Items.
type listScope struct {
	horizontal bool
	size       Size
}

// itemScope flows orientation, size, and status from an Item to its
// Indicator and Content.
type itemScope struct {
	horizontal bool
	size       Size
	status     Status
}

// New renders the timeline container (an ordered list).
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{Size: SizeBase}
		for _, opt := range options {
			opt(&cfg)
		}
		ol := dom.El(atom.Ol, dom.Marker("timeline"))
		sc := listScope{horizontal: cfg.horizontal, size: cfg.Size}
		if err := render.Children(ctx, ol, scope.With(sc)); err != nil {
			return nil, err
		}
		cfg.Apply(ol, listClasses(cfg))
		return ol, nil
	})
}

// Item renders one event: an Indicator beside (or above, when
// horizontal) a Content block. An item without an explicit Indicator
// child gets the default dot.
func Item(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		ls, _ := scope.From[listScope](ctx)
		cfg := Config{Size: ls.size, horizontal: ls.horizontal}
		if cfg.Size == "" {
			cfg.Size = SizeBase
		}
		for _, opt := range options {
			opt(&cfg)
		}

		li := dom.El(atom.Li, dom.Marker("timeline-item"))
		sc := itemScope{horizontal: cfg.horizontal, size: cfg.Size, status: cfg.Status}
		if err := render.Children(ctx, li, scope.With(sc)); err != nil {
			return nil, err
		}

		if dom.FindShallow(li, dom.ByMarker("timeline-indicator")) == nil {
			rail, circle := indicatorNode(Config{Size: sc.size, Status: sc.status, horizontal: sc.horizontal})
			dom.SetAttr(rail, "class", railClasses(cfg))
			circle.AppendChild(dotNode())
			li.InsertBefore(rail, li.FirstChild)
		}

		cfg.Apply(li, itemClasses(cfg))
		return li, nil
	})
}

// Indicator renders the item's marker — children (an icon, a step
// number, an avatar) centered in a circle, or a plain dot when empty —
// plus the connector segment to the next item.
func Indicator(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		is, _ := scope.From[itemScope](ctx)
		cfg := Config{Size: is.size, Status: is.status, horizontal: is.horizontal}
		if cfg.Size == "" {
			cfg.Size = SizeBase
		}
		for _, opt := range options {
			opt(&cfg)
		}

		root, circle := indicatorNode(cfg)
		if err := render.Children(ctx, circle); err != nil {
			return nil, err
		}
		if circle.FirstChild == nil && !cfg.bare {
			circle.AppendChild(dotNode())
		}
		cfg.Apply(root, railClasses(cfg))
		return root, nil
	})
}

// Content renders the item's body cell.
func Content(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		is, _ := scope.From[itemScope](ctx)
		cfg := Config{Size: is.size, horizontal: is.horizontal}
		if cfg.Size == "" {
			cfg.Size = SizeBase
		}
		for _, opt := range options {
			opt(&cfg)
		}
		div := dom.El(atom.Div, dom.Marker("timeline-content"))
		if err := render.Children(ctx, div); err != nil {
			return nil, err
		}
		cfg.Apply(div, contentClasses(cfg))
		return div, nil
	})
}

// indicatorNode builds the indicator column: the marker circle and the
// connector segment that runs to the next item.
func indicatorNode(cfg Config) (root, circle *html.Node) {
	root = dom.El(atom.Div, dom.Marker("timeline-indicator"))
	circle = dom.El(atom.Span, dom.Attr("class", circleClasses(cfg)))
	root.AppendChild(circle)
	line := dom.El(atom.Span, dom.Attr("aria-hidden", "true"), dom.Attr("class", lineClasses(cfg)))
	root.AppendChild(line)
	return root, circle
}

// dotNode is the default marker for an indicator with no children.
func dotNode() *html.Node {
	return dom.El(atom.Span, dom.Attr("aria-hidden", "true"), dom.Attr("class", dotClasses()))
}
