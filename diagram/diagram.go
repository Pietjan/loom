// Package diagram renders a directed-graph flowchart as SVG on the server,
// zero JavaScript. Nodes are declared as children — each carrying any templ
// component as its body — and edges wire them by id; the component lays them
// out automatically with a layered (Sugiyama-style) algorithm and draws the
// boxes, connectors, and arrowheads:
//
//	@diagram.New(
//		diagram.Edge("start", "ok"),
//		diagram.Edge("ok", "ship", diagram.Label("yes")),
//	) {
//		@diagram.Node("start") { Start }
//		@diagram.Node("ok", diagram.Diamond()) { Tests pass? }
//		@diagram.Node("ship") { @card.New() { Ship it 🚀 } }
//	}
//
// The result has two layers inside a stage of the diagram's natural size: an
// SVG holding the connectors and each node's chrome, and the node bodies as
// plain HTML positioned over it. Bodies stay real HTML — not <foreignObject>,
// whose overflow handling is inconsistent across browsers — so a body can hold
// anything (an icon and label, a badge, a card) and is never clipped. By
// default each node gets loom's box chrome; diagram.Bare() drops it for a body
// that brings its own.
//
// Linear pipelines and trees are DAG subsets handled for free; cycles lay out
// and keep their drawn direction; diagram.Dir(diagram.LeftRight) flips the
// flow axis.
//
// Sizing, honestly: the server can't measure rendered HTML, so a node's box is
// inferred from its content's text (widest line × a fixed glyph advance, line
// count for height) — good for labels and small chips, approximate for rich
// bodies. diagram.Size(w, h) overrides it. The box is what layout reserves and
// what edges attach to; a body bigger than its box overflows visibly rather
// than clipping, so an off estimate degrades gracefully instead of losing
// content. The stage is a fixed pixel size (the SVG and HTML must not scale
// apart) — wrap it in an overflow-x-auto container on narrow screens.
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
var ErrNoNodes = errors.New("diagram: at least one diagram.Node(...) child is required")

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

// helper attributes a Node stashes on its rendered element for New's post-pass
// to read back. They are stripped before final render.
const (
	attrID    = "data-node-id"
	attrShape = "data-node-shape"
	attrTone  = "data-node-tone"
	attrBare  = "data-node-bare"
	attrW     = "data-node-w"
	attrH     = "data-node-h"
)

type edge struct {
	from, to string
	label    string
}

// Config holds diagram-level options (edges, direction, title). Nodes are
// supplied as children, not options.
type Config struct {
	opts.Common
	dir   Direction
	title string
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

// Title sets the accessible name (aria-label). Without one, the node bodies'
// text is joined. Like chart, no <title> is emitted (it would show as a native
// hover tooltip).
func Title(s string) Option { return func(c *Config) { c.title = s } }

// Label puts a caption on an edge.
func Label(text string) EdgeOption { return func(e *edge) { e.label = text } }

// EdgeOption configures an edge.
type EdgeOption = func(*edge)

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

// nodeConfig collects a Node's options.
type nodeConfig struct {
	shape shape
	tone  Tone
	bare  bool
	w, h  float64
}

// NodeOption configures a node.
type NodeOption = func(*nodeConfig)

// Diamond renders a node as a decision diamond.
func Diamond() NodeOption { return func(n *nodeConfig) { n.shape = diamond } }

// Stadium renders a node as a pill.
func Stadium() NodeOption { return func(n *nodeConfig) { n.shape = stadium } }

// Bare drops the default box chrome so a body with its own styling (e.g. a
// card) isn't double-framed.
func Bare() NodeOption { return func(n *nodeConfig) { n.bare = true } }

// WithTone accents a node's outline.
func WithTone(t Tone) NodeOption { return func(n *nodeConfig) { n.tone = t } }

// Size fixes a node's box, overriding the inferred size.
func Size(w, h int) NodeOption {
	return func(n *nodeConfig) { n.w, n.h = float64(w), float64(h) }
}

// Node declares a node with the given id; its templ children are the body.
// Placed inside a @diagram.New(...) block.
func Node(id string, options ...NodeOption) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		var nc nodeConfig
		for _, opt := range options {
			opt(&nc)
		}
		div := dom.El(atom.Div, dom.Marker("diagram-node"))
		dom.SetAttr(div, attrID, id)
		if nc.shape != rounded {
			dom.SetAttr(div, attrShape, shapeName(nc.shape))
		}
		if nc.tone != ToneDefault {
			dom.SetAttr(div, attrTone, toneName(nc.tone))
		}
		if nc.bare {
			dom.SetAttr(div, attrBare, "")
		}
		if nc.w > 0 {
			dom.SetAttr(div, attrW, fmtCoord(nc.w))
			dom.SetAttr(div, attrH, fmtCoord(nc.h))
		}
		if err := render.Children(ctx, div); err != nil {
			return nil, err
		}
		return div, nil
	})
}

// New renders a diagram from Edge/Dir/Title options and Node children.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		var cfg Config
		for _, opt := range options {
			opt(&cfg)
		}
		return build(ctx, cfg)
	})
}

// collected is a node gathered from the children block: its options plus the
// rendered body and the text lines used for sizing and the accessible name.
type collected struct {
	id    string
	shape shape
	tone  Tone
	bare  bool
	w, h  float64
	body  *html.Node // the rendered node element, holding the body
	lines []string
}

func build(ctx context.Context, cfg Config) (*html.Node, error) {
	scratch := dom.El(atom.Div)
	if err := render.Children(ctx, scratch); err != nil {
		return nil, err
	}

	els := dom.FindAllShallow(scratch, dom.ByMarker("diagram-node"))
	if len(els) == 0 {
		return nil, ErrNoNodes
	}

	nodes := make([]collected, len(els))
	layoutNodes := make([]layoutNode, len(els))
	for i, el := range els {
		c := collected{
			id:    dom.GetAttr(el, attrID),
			shape: parseShape(dom.GetAttr(el, attrShape)),
			tone:  parseTone(dom.GetAttr(el, attrTone)),
			bare:  dom.HasAttr(el, attrBare),
			body:  el,
			lines: contentLines(el),
		}
		if w := dom.GetAttr(el, attrW); w != "" {
			c.w, c.h = atof(w), atof(dom.GetAttr(el, attrH))
		} else {
			c.w, c.h = inferSize(c.lines, el)
			if c.shape == diamond {
				// A rhombus needs a bigger box to keep the body clear of its
				// slanted edges — only when we inferred the size.
				c.w, c.h = c.w*1.5, c.h*1.7
			}
		}
		for _, a := range []string{attrID, attrShape, attrTone, attrBare, attrW, attrH} {
			dom.DelAttr(el, a)
		}
		nodes[i] = c
		layoutNodes[i] = layoutNode{id: c.id, w: c.w, h: c.h}
	}

	l, err := layout(layoutNodes, cfg.edges, cfg.dir)
	if err != nil {
		return nil, err
	}
	return emit(cfg, nodes, l), nil
}

// emit builds the two-layer result: an SVG holding the edges and node chrome,
// and the node bodies as real HTML positioned over it. Bodies stay HTML (not
// <foreignObject>) so oversized content overflows reliably instead of being
// clipped — foreignObject's overflow handling is inconsistent across browsers.
// The stage is a fixed pixel size so the SVG and the HTML never scale apart.
func emit(cfg Config, nodes []collected, l laid) *html.Node {
	stage := dom.El(atom.Div,
		dom.Marker("diagram"),
		dom.Attr("role", "img"))
	dom.SetAttr(stage, "aria-label", ariaLabel(cfg, nodes))

	svg := dom.El(atom.Svg,
		dom.Marker("diagram-canvas"),
		dom.Attr("viewBox", fmt.Sprintf("0 0 %s %s", fmtCoord(l.W), fmtCoord(l.H))),
		dom.Attr("width", fmtCoord(l.W)),
		dom.Attr("height", fmtCoord(l.H)),
		dom.Attr("aria-hidden", "true"),
		dom.Attr("class", canvasClasses()))
	for _, e := range l.edges {
		drawEdge(svg, e)
	}
	for i, n := range nodes {
		if shape := drawShape(n, l.boxes[i]); shape != nil {
			svg.AppendChild(shape)
		}
	}
	stage.AppendChild(svg)

	for i, n := range nodes {
		stage.AppendChild(nodeBody(n, l.boxes[i]))
	}

	cfg.Apply(stage, rootClasses())
	// Structural sizing goes on last so it can't be dropped by user attrs.
	dom.SetAttr(stage, "style", fmt.Sprintf("width: %spx; height: %spx", fmtCoord(l.W), fmtCoord(l.H)))
	return stage
}

func ariaLabel(cfg Config, nodes []collected) string {
	if cfg.title != "" {
		return cfg.title
	}
	var b strings.Builder
	for i, n := range nodes {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(n.lines[0])
	}
	return b.String()
}

// drawShape returns the node's chrome for the SVG layer, or nil when the node
// is Bare (its body brings its own).
func drawShape(n collected, b box) *html.Node {
	if n.bare {
		return nil
	}
	if n.shape == diamond {
		return dom.CustomEl("polygon",
			dom.Marker("diagram-shape"),
			dom.Attr("points", diamondPoints(b.x, b.y, b.w, b.h)),
			dom.Attr("class", nodeShapeClasses(n.tone)))
	}
	rx := "8"
	if n.shape == stadium {
		rx = fmtCoord(b.h / 2)
	}
	return dom.CustomEl("rect",
		dom.Marker("diagram-shape"),
		dom.Attr("x", fmtCoord(b.x-b.w/2)),
		dom.Attr("y", fmtCoord(b.y-b.h/2)),
		dom.Attr("width", fmtCoord(b.w)),
		dom.Attr("height", fmtCoord(b.h)),
		dom.Attr("rx", rx),
		dom.Attr("class", nodeShapeClasses(n.tone)))
}

// nodeBody positions the node's rendered body over its box as plain HTML.
// The box is the layout size, but nothing clips: content larger than the box
// simply overflows it.
func nodeBody(n collected, b box) *html.Node {
	div := dom.El(atom.Div,
		dom.Marker("diagram-node"),
		dom.Attr("class", contentClasses(n.bare)))
	dom.SetAttr(div, "style", fmt.Sprintf("left: %spx; top: %spx; width: %spx; height: %spx",
		fmtCoord(b.x-b.w/2), fmtCoord(b.y-b.h/2), fmtCoord(b.w), fmtCoord(b.h)))
	moveChildren(n.body, div)
	return div
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

// moveChildren reparents src's children onto dst.
func moveChildren(src, dst *html.Node) {
	var kids []*html.Node
	for c := src.FirstChild; c != nil; c = c.NextSibling {
		kids = append(kids, c)
	}
	for _, c := range kids {
		src.RemoveChild(c)
		dst.AppendChild(c)
	}
}

func shapeName(s shape) string {
	switch s {
	case diamond:
		return "diamond"
	case stadium:
		return "stadium"
	default:
		return "rounded"
	}
}

func parseShape(s string) shape {
	switch s {
	case "diamond":
		return diamond
	case "stadium":
		return stadium
	default:
		return rounded
	}
}

var toneNames = map[Tone]string{
	ToneAccent: "accent", ToneIndigo: "indigo", ToneEmerald: "emerald",
	ToneAmber: "amber", ToneRose: "rose",
}

func toneName(t Tone) string { return toneNames[t] }

func parseTone(s string) Tone {
	for t, name := range toneNames {
		if name == s {
			return t
		}
	}
	return ToneDefault
}

func atof(s string) float64 {
	var v float64
	fmt.Sscanf(s, "%g", &v)
	return v
}
