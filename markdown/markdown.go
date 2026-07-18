// Package markdown renders a Markdown string as styled loom markup:
//
//	@markdown.New(readme)
//	@markdown.New(comment, markdown.Unsafe())
//
// The source is parsed with goldmark (GFM: tables, strikethrough, task
// lists, autolinks) and the AST is rendered onto loom components —
// headings, text, links, and separators are the real loom primitives, so
// markdown output matches the rest of the page. Code fences are
// highlighted server-side with chroma; token colors ship in loom.css.
// Raw HTML in the source is dropped unless Unsafe() is given.
package markdown

import (
	"context"

	"github.com/a-h/templ"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	gmtext "github.com/yuin/goldmark/text"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
)

// Config holds markdown options.
type Config struct {
	opts.Common
	unsafe bool
}

// Option configures markdown rendering.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Unsafe passes raw HTML blocks and inline HTML in the source through to
// the output. The default drops them — goldmark's own safe default.
func Unsafe() Option { return func(c *Config) { c.unsafe = true } }

// parser is shared across renders; a goldmark parser is safe for
// concurrent use once constructed.
var parser = goldmark.New(goldmark.WithExtensions(extension.GFM)).Parser()

// New renders a markdown document as a templ component.
func New(source string, options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, source, options...)
	})
}

// Node builds the markdown container node.
func Node(ctx context.Context, source string, options ...Option) (*html.Node, error) {
	// The composed heading/text/link nodes each render the templ child
	// block; markdown's content comes from source alone, so the block
	// must be cleared here or it would be injected into every one of
	// them.
	ctx = templ.ClearChildren(ctx)

	var cfg Config
	for _, opt := range options {
		opt(&cfg)
	}

	root := dom.El(atom.Div, dom.Marker("markdown"))
	src := []byte(source)
	w := walker{ctx: ctx, source: src, cfg: cfg}
	if err := w.blocks(parser.Parse(gmtext.NewReader(src)), root); err != nil {
		return nil, err
	}
	cfg.Apply(root, classes(cfg))
	return root, nil
}
