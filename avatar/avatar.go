// Package avatar renders a user avatar — a photo, or initials as a
// fallback — with optional size, status dot, and stacking:
//
//	@avatar.New(avatar.Src("/u/olivia.jpg"), avatar.Alt("Olivia Martin"))
//	@avatar.New(avatar.Initials("OM"), avatar.Alt("Olivia Martin"), avatar.Status(avatar.Online))
//	@avatar.Group() {
//		@avatar.New(avatar.Src("/u/a.jpg"), avatar.Alt("Ada"))
//		@avatar.New(avatar.Src("/u/grace.jpg"), avatar.Alt("Grace"))
//	}
package avatar

import (
	"context"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
)

// Size selects the avatar diameter.
type Size string

const (
	SizeBase  Size = "base" // default
	SizeSmall Size = "sm"
	SizeLarge Size = "lg"
)

// Status shows a presence dot in the corner. None (default) shows nothing.
type Status string

const (
	StatusNone    Status = ""
	StatusOnline  Status = "online"
	StatusOffline Status = "offline"
	StatusBusy    Status = "busy"
	StatusAway    Status = "away"
)

// Config holds avatar options.
type Config struct {
	opts.Common
	SrcURL      string
	AltText     string
	InitialText string
	Size        Size
	Status      Status
}

// Option configures an avatar.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Src sets the avatar image URL.
func Src(url string) Option { return func(c *Config) { c.SrcURL = url } }

// Alt sets the accessible name (image alt, or label for initials).
func Alt(text string) Option { return func(c *Config) { c.AltText = text } }

// Initials sets the fallback text shown when there is no image.
func Initials(text string) Option { return func(c *Config) { c.InitialText = text } }

// WithSize sets the avatar size.
func WithSize(s Size) Option { return func(c *Config) { c.Size = s } }

// WithStatus sets the presence dot.
func WithStatus(s Status) Option { return func(c *Config) { c.Status = s } }

// Pre-baked options.
var (
	Small   = WithSize(SizeSmall)
	Large   = WithSize(SizeLarge)
	Online  = WithStatus(StatusOnline)
	Offline = WithStatus(StatusOffline)
	Busy    = WithStatus(StatusBusy)
	Away    = WithStatus(StatusAway)
)

// New renders an avatar as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the avatar node.
func Node(_ context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{Size: SizeBase}
	for _, opt := range options {
		opt(&cfg)
	}

	root := dom.El(atom.Span, dom.Marker("avatar"))

	if cfg.SrcURL != "" {
		img := dom.El(atom.Img,
			dom.Attr("src", cfg.SrcURL),
			dom.Attr("alt", cfg.AltText),
			dom.Attr("class", imageClasses()))
		root.AppendChild(img)
	} else {
		text := dom.El(atom.Span, dom.Attr("class", "select-none"))
		text.AppendChild(dom.Text(cfg.InitialText))
		root.AppendChild(text)
		if cfg.AltText != "" {
			dom.SetAttr(root, "role", "img")
			dom.SetAttr(root, "aria-label", cfg.AltText)
		}
	}

	if cfg.Status != StatusNone {
		dot := dom.El(atom.Span,
			dom.Attr("class", statusClasses(cfg.Status)),
			dom.Attr("aria-hidden", "true"))
		root.AppendChild(dot)
	}

	cfg.Apply(root, classes(cfg))
	return root, nil
}

// Group stacks avatars with an overlapping ring.
func Group(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		g := dom.El(atom.Div, dom.Marker("avatar-group"))
		if err := render.Children(ctx, g); err != nil {
			return nil, err
		}
		cfg.Apply(g, groupClasses())
		return g, nil
	})
}
