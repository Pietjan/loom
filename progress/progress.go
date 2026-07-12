// Package progress renders a progress bar:
//
//	@progress.New(progress.Value(70))
//	@progress.New(progress.Value(3, progress.Of(10)), progress.Emerald)
//	@progress.New(progress.Indeterminate())
//
// Determinate bars expose value/max via ARIA; the indeterminate variant
// animates with CSS (see cmd/css/loom.css). Zero JS.
package progress

import (
	"context"
	"strconv"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
)

// Color selects the bar color.
type Color string

const (
	Accent  Color = "accent" // default
	Emerald Color = "emerald"
	Amber   Color = "amber"
	Red     Color = "red"
)

// Config holds progress options.
type Config struct {
	opts.Common
	Val           float64
	Max           float64
	Color         Color
	indeterminate bool
}

// Option configures a progress bar.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Value sets the current progress (0..max). Pass progress.Of to change max.
func Value(v float64, mods ...Option) Option {
	return func(c *Config) {
		c.Val = v
		for _, m := range mods {
			m(c)
		}
	}
}

// Of sets the maximum (default 100).
func Of(max float64) Option { return func(c *Config) { c.Max = max } }

// Indeterminate renders an animated bar with no known value.
func Indeterminate() Option { return func(c *Config) { c.indeterminate = true } }

// WithColor sets the bar color.
func WithColor(col Color) Option { return func(c *Config) { c.Color = col } }

// Pre-baked color options.
var (
	WithEmerald = WithColor(Emerald)
	WithAmber   = WithColor(Amber)
	WithRed     = WithColor(Red)
)

// New renders a progress bar as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the progress node.
func Node(_ context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{Max: 100, Color: Accent}
	for _, opt := range options {
		opt(&cfg)
	}

	track := dom.El(atom.Div, dom.Marker("progress"),
		dom.Attr("role", "progressbar"))
	bar := dom.El(atom.Div, dom.Marker("progress-bar"), dom.Attr("class", barClasses(cfg)))

	if cfg.indeterminate {
		dom.SetAttr(track, "aria-label", "Loading")
		dom.SetAttr(track, "data-indeterminate", "")
	} else {
		pct := clamp(cfg.Val/cfg.Max) * 100
		dom.SetAttr(track, "aria-valuemin", "0")
		dom.SetAttr(track, "aria-valuemax", trim(cfg.Max))
		dom.SetAttr(track, "aria-valuenow", trim(cfg.Val))
		dom.SetAttr(bar, "style", "width: "+trim(pct)+"%")
	}

	track.AppendChild(bar)
	cfg.Apply(track, trackClasses())
	return track, nil
}

func clamp(f float64) float64 {
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}

func trim(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
