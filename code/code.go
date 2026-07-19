// Package code renders a standalone syntax-highlighted code block:
//
//	@code.New(src, code.Language("go"))
//	@code.New(patch, code.Diff())
//	@code.New(patch, code.Language("go"), code.Diff())
package code

import (
	"context"
	"strings"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/highlight"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
	"github.com/pietjan/loom/internal/styles"
)

// Config holds code options.
type Config struct {
	opts.Common
	language string
	diff     bool
}

// Option configures a code block.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Language sets the chroma lexer name ("go", "python", ...). An unknown
// language degrades to plain text.
func Language(lang string) Option { return func(c *Config) { c.language = lang } }

// Diff renders the source as a unified diff: added, removed, and hunk
// lines get full-width backgrounds. Combined with Language, each line's
// content is also syntax-highlighted; lines are lexed individually, so
// tokens spanning lines (a split block comment) may color slightly off.
func Diff() Option { return func(c *Config) { c.diff = true } }

// New renders a code block as a templ component.
func New(src string, options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, src, options...)
	})
}

// Node builds the code block node.
func Node(_ context.Context, src string, options ...Option) (*html.Node, error) {
	var cfg Config
	for _, opt := range options {
		opt(&cfg)
	}

	pre := dom.El(atom.Pre, dom.Marker("code"))
	if cfg.diff {
		pre.AppendChild(diffCode(src, cfg.language))
	} else {
		code := dom.El(atom.Code)
		highlight.Into(code, src, cfg.language)
		pre.AppendChild(code)
	}
	cfg.Apply(pre, classes(cfg))
	return pre, nil
}

type lineKind int

const (
	lineContext lineKind = iota
	lineAdd
	lineDel
	lineHunk
	lineMeta
)

// classify buckets a unified-diff line by its prefix. File headers are
// checked before add/del so "+++"/"---" don't read as changes.
func classify(line string) lineKind {
	switch {
	case strings.HasPrefix(line, "+++"), strings.HasPrefix(line, "---"),
		strings.HasPrefix(line, "diff "), strings.HasPrefix(line, "index "):
		return lineMeta
	case strings.HasPrefix(line, "+"):
		return lineAdd
	case strings.HasPrefix(line, "-"):
		return lineDel
	case strings.HasPrefix(line, "@@"):
		return lineHunk
	default:
		return lineContext
	}
}

// diffCode renders each diff line as a block span so backgrounds span the
// full width; block layout supplies the line breaks, so copied text keeps
// its newlines.
func diffCode(src, lang string) *html.Node {
	code := dom.El(atom.Code, dom.Attr("class", "block w-fit min-w-full"))
	for _, line := range strings.Split(strings.TrimSuffix(src, "\n"), "\n") {
		kind := classify(line)
		span := dom.El(atom.Span, dom.Attr("class", diffLineClass(kind)))
		if lang != "" && line != "" && (kind == lineAdd || kind == lineDel || kind == lineContext) {
			span.AppendChild(dom.Text(line[:1]))
			highlight.Into(span, line[1:], lang)
		} else {
			span.AppendChild(dom.Text(line))
		}
		code.AppendChild(span)
	}
	return code
}

func classes(c Config) string {
	var b styles.Builder
	b.Add("overflow-x-auto rounded-2xl border")
	b.Add("font-mono text-sm font-medium leading-loose tracking-wide")
	b.Add("border-base-200 bg-base-50 dark:border-white/10 dark:bg-base-800")
	b.Add("text-[#424258] dark:text-[#EEFFFF]")
	// Diff mode moves horizontal padding onto the lines so their
	// backgrounds bleed to the border.
	if c.diff {
		b.Add("py-5")
	} else {
		b.Add("p-5")
	}
	return b.String()
}

func diffLineClass(k lineKind) string {
	var b styles.Builder
	// min-h-lh keeps empty lines one line tall — an empty block would
	// otherwise collapse.
	b.Add("block min-h-lh px-5")
	styles.Match(&b, k, map[lineKind]string{
		lineAdd:  "bg-green-100/50 text-green-900 dark:bg-green-400/15 dark:text-green-200",
		lineDel:  "bg-red-100/40 text-red-700 dark:bg-red-400/15 dark:text-red-200",
		lineHunk: "bg-blue-50 text-blue-700 dark:bg-blue-400/10 dark:text-blue-300",
		lineMeta: "font-semibold text-base-500 dark:text-base-400",
	})
	return b.String()
}
