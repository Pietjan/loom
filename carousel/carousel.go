// Package carousel renders a horizontally scroll-snapping slider — CSS
// scroll-snap plus in-page anchor links for the dots, zero JS:
//
//	@carousel.New() {
//		@carousel.Slide() { ...slide one... }
//		@carousel.Slide() { ...slide two... }
//		@carousel.Slide() { ...slide three... }
//	}
//
// Each slide snaps into view; the dots below are <a> links that scroll
// their slide into place. Swipe and trackpad scroll work natively.
// Autoplay would need JS and is out of scope.
package carousel

import (
	"context"
	"strconv"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/ids"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
)

// Config holds carousel options.
type Config struct {
	opts.Common
	dots bool
}

// Option configures a carousel.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// NoDots hides the dot navigation (dots are on by default).
func NoDots() Option { return func(c *Config) { c.dots = false } }

// New renders the carousel.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the carousel node: a scroll-snap track plus a dot nav.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{dots: true}
	for _, opt := range options {
		opt(&cfg)
	}

	root := dom.El(atom.Div, dom.Marker("carousel"))
	track := dom.El(atom.Div, dom.Marker("carousel-track"), dom.Attr("class", trackClasses()))
	if err := render.Children(ctx, track); err != nil {
		return nil, err
	}

	// Give each slide a stable id so the dots can target it.
	slides := dom.FindAllShallow(track, dom.ByMarker("carousel-slide"))
	base := ids.New(ctx, "carousel")
	for i, s := range slides {
		if dom.GetAttr(s, "id") == "" {
			dom.SetAttr(s, "id", base+"-"+strconv.Itoa(i+1))
		}
	}
	root.AppendChild(track)

	if cfg.dots && len(slides) > 1 {
		nav := dom.El(atom.Div, dom.Marker("carousel-dots"),
			dom.Attr("role", "tablist"), dom.Attr("aria-label", "Slides"),
			dom.Attr("class", dotsClasses()))
		for i, s := range slides {
			dot := dom.El(atom.A,
				dom.Attr("href", "#"+dom.GetAttr(s, "id")),
				dom.Attr("aria-label", "Go to slide "+strconv.Itoa(i+1)),
				dom.Attr("class", dotClasses()))
			nav.AppendChild(dot)
		}
		root.AppendChild(nav)
	}

	cfg.Apply(root, rootClasses())
	return root, nil
}

// Slide renders one carousel slide.
func Slide(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		s := dom.El(atom.Div, dom.Marker("carousel-slide"), dom.Attr("class", slideClasses()))
		if err := render.Children(ctx, s); err != nil {
			return nil, err
		}
		cfg.Apply(s, "")
		return s, nil
	})
}
