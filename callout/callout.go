// Package callout renders attention boxes with an optional leading icon
// and Heading/Text parts:
//
//	@callout.New(callout.Warning) {
//		@icon.New(icon.Warning)
//		@callout.Heading() { Subscription expiring }
//		@callout.Text() { Renew before Friday to keep access. }
//	}
package callout

import (
	"context"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
	"github.com/pietjan/loom/internal/styles"
)

// Tone selects the callout intent.
type Tone string

const (
	ToneNeutral Tone = "neutral" // default
	ToneInfo    Tone = "info"
	ToneSuccess Tone = "success"
	ToneWarning Tone = "warning"
	ToneDanger  Tone = "danger"
)

// Config holds callout options.
type Config struct {
	opts.Common
	Tone Tone
}

// Option configures a callout.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// WithTone sets the callout intent.
func WithTone(t Tone) Option { return func(c *Config) { c.Tone = t } }

// Pre-baked tone options.
var (
	Info    = WithTone(ToneInfo)
	Success = WithTone(ToneSuccess)
	Warning = WithTone(ToneWarning)
	Danger  = WithTone(ToneDanger)
)

// New renders a callout as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the callout node.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{Tone: ToneNeutral}
	for _, opt := range options {
		opt(&cfg)
	}

	div := dom.El(atom.Div, dom.Marker("callout"))
	if err := render.Children(ctx, div); err != nil {
		return nil, err
	}
	cfg.Apply(div, classes(cfg))
	return div, nil
}

// Heading renders the callout title part.
func Heading(options ...Option) templ.Component {
	return part("callout-heading", "font-medium text-sm", options)
}

// Text renders the callout body part.
func Text(options ...Option) templ.Component {
	return part("callout-text", "text-sm opacity-80", options)
}

func part(marker, base string, options []Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		div := dom.El(atom.Div, dom.Marker(marker))
		if err := render.Children(ctx, div); err != nil {
			return nil, err
		}
		cfg.Apply(div, base)
		return div, nil
	})
}

func classes(c Config) string {
	var b styles.Builder
	// Grid: icon column + content column; heading and text stack in the
	// second column, the icon spans rows.
	b.Add("grid grid-cols-[auto_1fr] gap-x-3 gap-y-1 rounded-lg border p-4")
	b.Add("[&>[data-ui=icon]]:row-span-2 [&>[data-ui=icon]]:mt-0.5 [&>[data-ui=icon]]:size-5")
	b.Add("[&>[data-ui=callout-heading]]:col-start-2 [&>[data-ui=callout-text]]:col-start-2")
	styles.Match(&b, c.Tone, map[Tone]string{
		ToneNeutral: "border-base-200 bg-base-50 text-base-800 dark:border-base-600 dark:bg-base-700 dark:text-base-100",
		ToneInfo:    "border-blue-200 bg-blue-50 text-blue-900 dark:border-blue-400/30 dark:bg-blue-400/10 dark:text-blue-200",
		ToneSuccess: "border-green-200 bg-green-50 text-green-900 dark:border-green-400/30 dark:bg-green-400/10 dark:text-green-200",
		ToneWarning: "border-amber-200 bg-amber-50 text-amber-900 dark:border-amber-400/30 dark:bg-amber-400/10 dark:text-amber-200",
		ToneDanger:  "border-red-200 bg-red-50 text-red-900 dark:border-red-400/30 dark:bg-red-400/10 dark:text-red-200",
	})
	return b.String()
}
