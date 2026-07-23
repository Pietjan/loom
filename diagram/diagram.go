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
// A node's body is real HTML rendered inside an SVG <foreignObject>, so it
// scales with the diagram and can hold anything — an icon and label, a badge,
// a card. By default each node gets loom's box chrome (rounded border, fill,
// tone); diagram.Bare() drops it for a body that brings its own.
//
// Linear pipelines and trees are DAG subsets handled for free; cycles lay out
// and keep their drawn direction; diagram.Dir(diagram.LeftRight) flips the
// flow axis.
//
// Sizing, honestly: the server can't measure rendered HTML, so a node's box is
// inferred from its content's text (widest line × a fixed glyph advance, line
// count for height) — good for labels and small chips, approximate for rich
// bodies. diagram.Size(w, h) overrides it, and content overflows the box
// rather than clipping, so an off estimate degrades gracefully.
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

func emit(cfg Config, nodes []collected, l laid) *html.Node {
	svg := dom.El(atom.Svg,
		dom.Marker("diagram"),
		dom.Attr("viewBox", fmt.Sprintf("0 0 %s %s", fmtCoord(l.W), fmtCoord(l.H))),
		dom.Attr("role", "img"))
	dom.SetAttr(svg, "aria-label", ariaLabel(cfg, nodes))

	for _, e := range l.edges {
		drawEdge(svg, e)
	}
	for i, n := range nodes {
		svg.AppendChild(drawNode(n, l.boxes[i]))
	}

	cfg.Apply(svg, rootClasses())
	return svg
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

func drawNode(n collected, b box) *html.Node {
	g := dom.CustomEl("g", dom.Marker("diagram-node"))

	if !n.bare {
		switch n.shape {
		case diamond:
			g.AppendChild(dom.CustomEl("polygon",
				dom.Attr("points", diamondPoints(b.x, b.y, b.w, b.h)),
				dom.Attr("class", nodeShapeClasses(n.tone))))
		default:
			rx := "8"
			if n.shape == stadium {
				rx = fmtCoord(b.h / 2)
			}
			g.AppendChild(dom.CustomEl("rect",
				dom.Attr("x", fmtCoord(b.x-b.w/2)),
				dom.Attr("y", fmtCoord(b.y-b.h/2)),
				dom.Attr("width", fmtCoord(b.w)),
				dom.Attr("height", fmtCoord(b.h)),
				dom.Attr("rx", rx),
				dom.Attr("class", nodeShapeClasses(n.tone))))
		}
	}

	// The body renders as HTML inside a foreignObject so it scales with the
	// SVG. overflow=visible lets a body larger than its inferred box spill
	// instead of clipping.
	fo := dom.CustomEl("foreignObject",
		dom.Attr("x", fmtCoord(b.x-b.w/2)),
		dom.Attr("y", fmtCoord(b.y-b.h/2)),
		dom.Attr("width", fmtCoord(b.w)),
		dom.Attr("height", fmtCoord(b.h)),
		dom.Attr("overflow", "visible"))
	wrap := dom.El(atom.Div, dom.Attr("class", contentClasses(n.bare)))
	moveChildren(n.body, wrap)
	fo.AppendChild(wrap)
	g.AppendChild(fo)

	return g
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
