// Package diagram renders a directed-graph flowchart as SVG on the server,
// zero JavaScript. Nodes and edges are described with the usual loom
// functional-options builder; the component lays them out automatically with a
// layered (Sugiyama-style) algorithm and draws boxes, connectors, and
// arrowheads:
//
//	@diagram.New(
//		diagram.Node("start", "Start"),
//		diagram.Node("check", "OK?", diagram.Diamond()),
//		diagram.Node("done", "Done"),
//		diagram.Edge("start", "check"),
//		diagram.Edge("check", "done", diagram.Label("yes")),
//	)
//
// Linear pipelines and single-parent trees are just DAG subsets and lay out
// for free; cycles are handled by breaking back edges for layering and
// restoring their drawn direction. Left-to-right flow is a one-liner:
//
//	@diagram.New(diagram.Dir(diagram.LeftRight), ...)
//
// Naming note: this package deliberately deviates from loom's convention of
// exporting Node(ctx, ...) as the raw-node render entry. Here Node(id, label)
// is the node constructor the builder API needs; the render function is
// unexported (build). diagram is an L1, non-embeddable component, so nothing
// calls a diagram.Node(ctx, ...).
//
// Honest limitation, by design: label text is sized by a deterministic glyph
// estimate (no server-side font metrics), so box widths are approximate and
// long labels are not wrapped — the tradeoff buys reproducible output.
package diagram

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
)

// ErrNoNodes is returned when the diagram has no nodes.
var ErrNoNodes = errors.New("diagram: at least one diagram.Node(...) is required")

// Direction is the flow axis: the direction edges point and layers advance.
type Direction int

const (
	// TopBottom flows downward (default).
	TopBottom Direction = iota
	// LeftRight flows rightward.
	LeftRight
)

// Tone accents a node's outline.
type Tone int

const (
	ToneDefault Tone = iota
	ToneAccent
	ToneIndigo
	ToneEmerald
	ToneAmber
	ToneRose
)

// shape is a node's silhouette.
type shape int

const (
	rounded shape = iota // rounded rectangle (default)
	stadium              // pill
	diamond              // decision
)

type node struct {
	id, label string
	shape     shape
	tone      Tone
}

type edge struct {
	from, to string
	label    string
}

// Config holds diagram options.
type Config struct {
	opts.Common
	dir   Direction
	title string
	nodes []node
	edges []edge
}

// Option configures a diagram.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Dir sets the flow direction (default TopBottom).
func Dir(d Direction) Option { return func(c *Config) { c.dir = d } }

// Title sets the accessible name (aria-label). Without one, the node labels
// are joined. Like chart, no <title> is emitted (it would show as a native
// hover tooltip).
func Title(s string) Option { return func(c *Config) { c.title = s } }

// NodeOption configures a node.
type NodeOption = func(*node)

// Diamond renders a node as a decision diamond.
func Diamond() NodeOption { return func(n *node) { n.shape = diamond } }

// Stadium renders a node as a pill.
func Stadium() NodeOption { return func(n *node) { n.shape = stadium } }

// WithTone accents a node's outline.
func WithTone(t Tone) NodeOption { return func(n *node) { n.tone = t } }

// Node adds a node. id is the stable handle edges reference; label is the
// visible text.
func Node(id, label string, options ...NodeOption) Option {
	return func(c *Config) {
		n := node{id: id, label: label}
		for _, opt := range options {
			opt(&n)
		}
		c.nodes = append(c.nodes, n)
	}
}

// EdgeOption configures an edge.
type EdgeOption = func(*edge)

// Label puts a caption on an edge.
func Label(text string) EdgeOption { return func(e *edge) { e.label = text } }

// Edge connects two nodes by id; the arrow points from → to.
func Edge(from, to string, options ...EdgeOption) Option {
	return func(c *Config) {
		e := edge{from: from, to: to}
		for _, opt := range options {
			opt(&e)
		}
		c.edges = append(c.edges, e)
	}
}

// New renders a diagram as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return build(ctx, options...)
	})
}

func build(_ context.Context, options ...Option) (*html.Node, error) {
	var cfg Config
	for _, opt := range options {
		opt(&cfg)
	}
	l, err := layout(cfg)
	if err != nil {
		return nil, err
	}
	return emit(cfg, l), nil
}

func emit(cfg Config, l laid) *html.Node {
	svg := dom.El(atom.Svg,
		dom.Marker("diagram"),
		dom.Attr("viewBox", fmt.Sprintf("0 0 %s %s", fmtCoord(l.W), fmtCoord(l.H))),
		dom.Attr("role", "img"))
	dom.SetAttr(svg, "aria-label", ariaLabel(cfg))

	// Edges first so nodes paint over the connector ends.
	for _, e := range l.edges {
		drawEdge(svg, e)
	}
	for _, b := range l.boxes {
		drawNode(svg, b)
	}

	cfg.Apply(svg, rootClasses())
	return svg
}

func ariaLabel(cfg Config) string {
	if cfg.title != "" {
		return cfg.title
	}
	var b strings.Builder
	for i, n := range cfg.nodes {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(n.label)
	}
	return b.String()
}

func drawNode(svg *html.Node, b box) {
	g := dom.CustomEl("g", dom.Marker("diagram-node"))

	switch b.shape {
	case diamond:
		g.AppendChild(dom.CustomEl("polygon",
			dom.Attr("points", diamondPoints(b.x, b.y, b.w, b.h)),
			dom.Attr("class", nodeShapeClasses(b.tone))))
	default:
		rx := "8"
		if b.shape == stadium {
			rx = fmtCoord(b.h / 2)
		}
		g.AppendChild(dom.CustomEl("rect",
			dom.Attr("x", fmtCoord(b.x-b.w/2)),
			dom.Attr("y", fmtCoord(b.y-b.h/2)),
			dom.Attr("width", fmtCoord(b.w)),
			dom.Attr("height", fmtCoord(b.h)),
			dom.Attr("rx", rx),
			dom.Attr("class", nodeShapeClasses(b.tone))))
	}

	text := dom.CustomEl("text",
		dom.Attr("x", fmtCoord(b.x)),
		dom.Attr("y", fmtCoord(b.y)),
		dom.Attr("text-anchor", "middle"),
		dom.Attr("dominant-baseline", "central"),
		dom.Attr("class", nodeLabelClasses()))
	text.AppendChild(dom.Text(b.label))
	g.AppendChild(text)

	svg.AppendChild(g)
}

func drawEdge(svg *html.Node, e routed) {
	svg.AppendChild(dom.CustomEl("path",
		dom.Marker("diagram-edge"),
		dom.Attr("d", polyline(e.pts)),
		dom.Attr("class", edgeClasses())))
	svg.AppendChild(dom.CustomEl("polygon",
		dom.Marker("diagram-arrow"),
		dom.Attr("points", arrowhead(e.pts)),
		dom.Attr("class", arrowClasses())))

	if e.label != "" {
		mid := midpoint(e.pts)
		text := dom.CustomEl("text",
			dom.Marker("diagram-edge-label"),
			dom.Attr("x", fmtCoord(mid.x)),
			dom.Attr("y", fmtCoord(mid.y)),
			dom.Attr("text-anchor", "middle"),
			dom.Attr("dominant-baseline", "central"),
			dom.Attr("class", edgeLabelClasses()))
		text.AppendChild(dom.Text(e.label))
		svg.AppendChild(text)
	}
}
