// Package kanban renders a board of workflow columns with stacked cards,
// Flux-style:
//
//	@kanban.New() {
//		@kanban.Column() {
//			@kanban.Header(kanban.Heading("In Progress"), kanban.Count(3)) {
//				@button.New(button.Subtle, button.Small, button.Square) { @icon.New(icon.Plus, icon.Small) }
//			}
//			@kanban.Cards() {
//				@kanban.Card(kanban.Heading("Update privacy policy")) {
//					@kanban.CardFooter() { @avatar.New(avatar.Initials("CP")) }
//				}
//			}
//			@kanban.Footer() {
//				@button.New(button.Subtle, button.Small) { New card }
//			}
//		}
//	}
//
// Header children are trailing actions (buttons, a dropdown); Card
// children are the card body, with CardHeader (badges above the heading)
// and CardFooter (metadata below it) as optional slots. To make a card
// navigable, compose a link inside it. The board is static markup + CSS —
// reordering is your application's concern. Zero JS.
package kanban

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

// Config holds kanban options.
type Config struct {
	opts.Common
	HeadingText    string
	SubheadingText string
	CountValue     int
	hasCount       bool
}

// Option configures the board, a column part, or a card.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Heading sets the title (on Header or Card).
func Heading(text string) Option { return func(c *Config) { c.HeadingText = text } }

// Subheading sets the secondary line under a Header's title.
func Subheading(text string) Option { return func(c *Config) { c.SubheadingText = text } }

// Count shows the card count next to a Header's title.
func Count(n int) Option {
	return func(c *Config) {
		c.CountValue = n
		c.hasCount = true
	}
}

// New renders the board container.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		div := dom.El(atom.Div, dom.Marker("kanban"))
		if err := render.Children(ctx, div); err != nil {
			return nil, err
		}
		cfg.Apply(div, boardClasses())
		return div, nil
	})
}

// Column renders one workflow stage: a Header, a Cards stack, and an
// optional Footer.
func Column(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		div := dom.El(atom.Div, dom.Marker("kanban-column"))
		if err := render.Children(ctx, div); err != nil {
			return nil, err
		}
		cfg.Apply(div, columnClasses())
		return div, nil
	})
}

// Header renders a column's title row. Children render after the title
// as trailing actions.
func Header(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}

		div := dom.El(atom.Div, dom.Marker("kanban-column-header"))

		title := dom.El(atom.Div, dom.Attr("class", headerTitleClasses()))
		row := dom.El(atom.Div, dom.Attr("class", headerRowClasses()))
		h := dom.El(atom.Div, dom.Attr("class", headingClasses()))
		h.AppendChild(dom.Text(cfg.HeadingText))
		row.AppendChild(h)
		if cfg.hasCount {
			count := dom.El(atom.Span, dom.Attr("class", countClasses()))
			count.AppendChild(dom.Text(strconv.Itoa(cfg.CountValue)))
			row.AppendChild(count)
		}
		title.AppendChild(row)
		if cfg.SubheadingText != "" {
			sub := dom.El(atom.Div, dom.Attr("class", subheadingClasses()))
			sub.AppendChild(dom.Text(cfg.SubheadingText))
			title.AppendChild(sub)
		}
		div.AppendChild(title)

		if err := render.Children(ctx, div); err != nil {
			return nil, err
		}
		cfg.Apply(div, headerClasses())
		return div, nil
	})
}

// Cards renders the scrollable stack of cards in a column.
func Cards(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		div := dom.El(atom.Div, dom.Marker("kanban-column-cards"))
		if err := render.Children(ctx, div); err != nil {
			return nil, err
		}
		cfg.Apply(div, cardsClasses())
		return div, nil
	})
}

// Footer renders a column's bottom section (add-card forms, actions).
func Footer(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		div := dom.El(atom.Div, dom.Marker("kanban-column-footer"))
		if err := render.Children(ctx, div); err != nil {
			return nil, err
		}
		cfg.Apply(div, footerClasses())
		return div, nil
	})
}

// Card renders one item. The Heading sits above the card's children —
// or, when the block starts with a CardHeader, directly after it.
func Card(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}

		root := dom.El(atom.Div, dom.Marker("kanban-card"))
		if err := render.Children(ctx, root); err != nil {
			return nil, err
		}

		if cfg.HeadingText != "" {
			h := dom.El(atom.Div, dom.Marker("kanban-card-heading"), dom.Attr("class", cardHeadingClasses()))
			h.AppendChild(dom.Text(cfg.HeadingText))
			if slot := dom.FindShallow(root, dom.ByMarker("kanban-card-header")); slot != nil {
				slot.Parent.InsertBefore(h, slot.NextSibling)
			} else {
				root.InsertBefore(h, root.FirstChild)
			}
		}

		cfg.Apply(root, cardClasses())
		return root, nil
	})
}

// CardHeader is the slot above a card's heading — badges, labels.
func CardHeader(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		div := dom.El(atom.Div, dom.Marker("kanban-card-header"))
		if err := render.Children(ctx, div); err != nil {
			return nil, err
		}
		cfg.Apply(div, cardHeaderClasses())
		return div, nil
	})
}

// CardFooter is the slot below a card's body — avatars, metadata.
func CardFooter(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		div := dom.El(atom.Div, dom.Marker("kanban-card-footer"))
		if err := render.Children(ctx, div); err != nil {
			return nil, err
		}
		cfg.Apply(div, cardFooterClasses())
		return div, nil
	})
}
