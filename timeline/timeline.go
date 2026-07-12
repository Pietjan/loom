// Package timeline renders a vertical sequence of events with a connector
// line:
//
//	@timeline.New() {
//		@timeline.Item(timeline.Title("Deployed v2"), timeline.Time("2h ago")) {
//			@text.New() { Rolled out to production. }
//		}
//		@timeline.Item(timeline.Title("Merged #482")) {
//			@text.New() { Header layout landed. }
//		}
//	}
//
// The dot markers and connecting line are drawn with CSS. Zero JS.
package timeline

import (
	"context"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
)

// Config holds timeline options.
type Config struct {
	opts.Common
	TitleText string
	TimeText  string
}

// Option configures a timeline or item.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Title sets an item's heading.
func Title(text string) Option { return func(c *Config) { c.TitleText = text } }

// Time sets an item's timestamp label.
func Time(text string) Option { return func(c *Config) { c.TimeText = text } }

// New renders the timeline container (an ordered list).
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		ol := dom.El(atom.Ol, dom.Marker("timeline"))
		if err := render.Children(ctx, ol); err != nil {
			return nil, err
		}
		cfg.Apply(ol, listClasses())
		return ol, nil
	})
}

// Item renders one event. Its heading (Title) and optional Time sit above
// the block's children.
func Item(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}

		li := dom.El(atom.Li, dom.Marker("timeline-item"), dom.Attr("class", itemClasses()))
		// The dot marker; the connector line is the item's inline-start border.
		li.AppendChild(dom.El(atom.Span, dom.Attr("aria-hidden", "true"), dom.Attr("class", dotClasses())))

		if cfg.TitleText != "" {
			head := dom.El(atom.Div, dom.Marker("timeline-title"), dom.Attr("class", titleClasses()))
			head.AppendChild(dom.Text(cfg.TitleText))
			if cfg.TimeText != "" {
				t := dom.El(atom.Span, dom.Attr("class", timeClasses()))
				t.AppendChild(dom.Text(cfg.TimeText))
				head.AppendChild(t)
			}
			li.AppendChild(head)
		}

		body := dom.El(atom.Div, dom.Attr("class", "text-sm"))
		if err := render.Children(ctx, body); err != nil {
			return nil, err
		}
		li.AppendChild(body)

		cfg.Apply(li, "")
		return li, nil
	})
}
