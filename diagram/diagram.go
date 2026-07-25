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
// flow axis. Each box face carries at most two edge ports, one per drawn
// direction, so an outgoing shaft is never laid over an incoming arrowhead;
// same-direction edges share a port and fan or join from one point.
// diagram.Ports(diagram.PortsSpread) instead gives every edge its own port.
// Spacing is tunable with diagram.Gap(layer, node) and
// diagram.Margin(n); a labelled edge gets a dot on the line whose label a
// CSS-only tooltip reveals on hover or focus. diagram.Dashed() and
// diagram.Dotted() break an edge's line while leaving its arrowhead solid;
// diagram.NoArrow() drops the arrowhead for a plain connector and
// diagram.BiDirectional() adds one at the source end too. The arrowhead's
// shape is diagram.HollowArrow(), OpenArrow(), DiamondArrow(), or DotArrow(),
// defaulting to a filled triangle.
//
// Naming note: this package deliberately deviates from loom's convention of
// exporting Node(ctx, ...) as the raw-node render entry. Here Node(id, ...) is
// the node constructor the builder API needs — it returns a templ.Component
// whose children are the body — so the render function is unexported (build).
// diagram is an L1, non-embeddable component, and tests and the site only use
// New(), so nothing needs a diagram.Node(ctx, ...).
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
	"github.com/pietjan/loom/tooltip"
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

// PortMode chooses how edges sharing a box face attach to it.
type PortMode int

const (
	// PortsShared gives each face at most two ports, one per drawn direction:
	// same-direction edges share, fanning or joining from one point (default).
	PortsShared PortMode = iota
	// PortsSpread gives every edge its own port, spread along the face and
	// ordered by its neighbour — busier, but each connection is distinct.
	PortsSpread
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

// lineStyle is an edge's stroke pattern.
type lineStyle int

const (
	lineSolid  lineStyle = iota // continuous (default)
	lineDashed                  // evenly broken
	lineDotted                  // round dots
)

// arrowEnds says which ends of an edge carry an arrowhead.
type arrowEnds int

const (
	arrowTo   arrowEnds = iota // head at the target only (default)
	arrowNone                  // a plain connector, no heads
	arrowBoth                  // heads at both ends
)

// arrowShape is an arrowhead's silhouette.
type arrowShape int

const (
	arrowSolid   arrowShape = iota // filled triangle (default)
	arrowHollow                    // outlined triangle
	arrowOpen                      // barb (two strokes, no fill)
	arrowDiamond                   // filled diamond
	arrowRound                     // filled dot
)

type edge struct {
	from, to string
	label    string
	style    lineStyle
	arrows   arrowEnds
	head     arrowShape
}

// Config holds diagram-level options (edges, direction, title). Nodes are
// supplied as children, not options.
type Config struct {
	opts.Common
	dir       Direction
	title     string
	direct    bool
	gapLayer  float64
	gapNode   float64
	margin    float64
	marginSet bool
	ports     PortMode
	edges     []edge
}

// gaps resolves the configured spacing, falling back to the defaults.
func (c Config) gaps() gaps {
	sp := gaps{layer: defaultLayerGap, node: defaultNodeGap, margin: defaultMargin}
	if c.gapLayer > 0 {
		sp.layer = c.gapLayer
	}
	if c.gapNode > 0 {
		sp.node = c.gapNode
	}
	if c.marginSet {
		sp.margin = c.margin
	}
	return sp
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

// Direct draws edges as straight point-to-point lines instead of the default
// right-angled routing.
func Direct() Option { return func(c *Config) { c.direct = true } }

// Ports selects how edges sharing a box face attach to it (default
// PortsShared). PortsSpread gives every edge its own attachment point.
func Ports(m PortMode) Option { return func(c *Config) { c.ports = m } }

// Gap sets the spacing between layers (along the flow axis) and between
// sibling nodes within a layer. Zero or negative keeps the default.
func Gap(layer, node int) Option {
	return func(c *Config) {
		if layer > 0 {
			c.gapLayer = float64(layer)
		}
		if node > 0 {
			c.gapNode = float64(node)
		}
	}
}

// Margin sets the padding around the whole drawing. Negative keeps the default.
func Margin(n int) Option {
	return func(c *Config) {
		if n >= 0 {
			c.margin = float64(n)
			c.marginSet = true
		}
	}
}

// Title sets the accessible name (aria-label). Without one, the node bodies'
// text is joined. Like chart, no <title> is emitted (it would show as a native
// hover tooltip).
func Title(s string) Option { return func(c *Config) { c.title = s } }

// Label puts a caption on an edge.
func Label(text string) EdgeOption { return func(e *edge) { e.label = text } }

// Dashed draws an edge as a dashed line instead of a solid one.
func Dashed() EdgeOption { return func(e *edge) { e.style = lineDashed } }

// Dotted draws an edge as a dotted line instead of a solid one.
func Dotted() EdgeOption { return func(e *edge) { e.style = lineDotted } }

// NoArrow draws an edge as a plain connector with no arrowhead — an
// undirected association rather than a flow.
func NoArrow() EdgeOption { return func(e *edge) { e.arrows = arrowNone } }

// BiDirectional puts an arrowhead on both ends of an edge, for a two-way
// relationship.
func BiDirectional() EdgeOption { return func(e *edge) { e.arrows = arrowBoth } }

// HollowArrow draws the arrowhead as an outlined (unfilled) triangle.
func HollowArrow() EdgeOption { return func(e *edge) { e.head = arrowHollow } }

// OpenArrow draws the arrowhead as an open barb — two strokes meeting at the
// tip, with no fill.
func OpenArrow() EdgeOption { return func(e *edge) { e.head = arrowOpen } }

// DiamondArrow draws the arrowhead as a filled diamond.
func DiamondArrow() EdgeOption { return func(e *edge) { e.head = arrowDiamond } }

// DotArrow draws the arrowhead as a filled dot.
func DotArrow() EdgeOption { return func(e *edge) { e.head = arrowRound } }

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
			c.w, c.h = inferSize(el, c.bare)
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
		layoutNodes[i] = layoutNode{id: c.id, w: c.w, h: c.h, shape: c.shape}
	}

	l, err := layout(layoutNodes, cfg.edges, cfg.dir, cfg.direct, cfg.ports, cfg.gaps())
	if err != nil {
		return nil, err
	}
	return emit(ctx, cfg, nodes, l)
}

// emit builds the two-layer result: an SVG holding the edges and node chrome,
// and the node bodies as real HTML positioned over it. Bodies stay HTML (not
// <foreignObject>) so oversized content overflows reliably instead of being
// clipped — foreignObject's overflow handling is inconsistent across browsers.
// The stage is a fixed pixel size so the SVG and the HTML never scale apart.
func emit(ctx context.Context, cfg Config, nodes []collected, l laid) (*html.Node, error) {
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
		drawEdge(svg, e, l.radius)
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

	// A labelled edge gets a dot on the line rather than always-on text: the
	// dot is the affordance, and loom's CSS-only tooltip reveals the label on
	// hover or keyboard focus. Still no JavaScript.
	for _, e := range l.edges {
		if e.label == "" {
			continue
		}
		tip, err := edgeTip(ctx, e)
		if err != nil {
			return nil, err
		}
		stage.AppendChild(tip)
	}

	cfg.Apply(stage, rootClasses())
	// Structural sizing goes on last so it can't be dropped by user attrs.
	dom.SetAttr(stage, "style", fmt.Sprintf("width: %spx; height: %spx", fmtCoord(l.W), fmtCoord(l.H)))
	return stage, nil
}

// edgeTip builds the hoverable dot that reveals an edge's label, placed at the
// midpoint of the drawn shaft.
//
// That is the trimmed polyline, not the routed one: the shaft stops at the
// arrowhead's base, so measuring the midpoint on the untrimmed points puts the
// dot half the arrowhead's length past the middle of the line anyone can see —
// a quarter of the way along a short edge, and plainly off centre against a
// 5px dot.
func edgeTip(ctx context.Context, e routed) (*html.Node, error) {
	// tabindex makes the dot focusable so the tooltip's :focus-within path can
	// fire — otherwise the label would be unreachable by keyboard.
	dot := render.Component(func(context.Context) (*html.Node, error) {
		return dom.El(atom.Span, dom.Marker("diagram-edge-dot"),
			dom.Attr("tabindex", "0"),
			dom.Attr("class", dotClasses())), nil
	})
	mid := midpoint(trimShaft(e.pts, e.arrows, e.head))
	return tooltip.Node(templ.WithChildren(ctx, dot),
		tooltip.Text(e.label),
		tooltip.Class("absolute -translate-x-1/2 -translate-y-1/2 p-1"),
		tooltip.Attr("style", fmt.Sprintf("left: %spx; top: %spx", fmtCoord(mid.x), fmtCoord(mid.y))))
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
	rx := "6" // rounded-md, loom's radius for compact elements
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

func drawEdge(svg *html.Node, e routed, radius float64) {
	// The shaft stops at the arrowhead base (trimShaft), while the heads below
	// are built from the untrimmed points so their tips stay on the borders.
	path := dom.CustomEl("path",
		dom.Marker("diagram-edge"),
		dom.Attr("d", roundedPath(trimShaft(e.pts, e.arrows, e.head), radius)),
		dom.Attr("class", edgeClasses()))
	// The dash pattern is an SVG presentation attribute, not a Tailwind
	// utility: there is no stroke-dasharray class, and the arrowhead below
	// stays solid so only the shaft is broken.
	if dash, cap := dashPattern(e.style); dash != "" {
		dom.SetAttr(path, "stroke-dasharray", dash)
		if cap != "" {
			dom.SetAttr(path, "stroke-linecap", cap)
		}
	}
	svg.AppendChild(path)

	// The head sits at whichever end(s) the edge asks for. An arrowhead points
	// along a polyline's final segment, so the tail head is drawn from the
	// reversed points; a plain connector draws neither.
	if e.arrows == arrowTo || e.arrows == arrowBoth {
		svg.AppendChild(arrowHead(e.pts, e.head))
	}
	if e.arrows == arrowBoth {
		svg.AppendChild(arrowHead(reversePts(e.pts), e.head))
	}
}

// arrowHead builds the arrowhead element at the end of a polyline in the
// requested shape. The solid, diamond, and dot heads are filled in the line's
// colour and so cover the shaft passing under them; the hollow head is filled
// with the surface colour to the same effect; the open barb leaves the shaft
// showing, which is the look it's meant to have.
func arrowHead(pts []xy, shape arrowShape) *html.Node {
	switch shape {
	case arrowHollow:
		return dom.CustomEl("polygon",
			dom.Marker("diagram-arrow"),
			dom.Attr("points", arrowTriangle(pts)),
			dom.Attr("class", arrowHollowClasses()))
	case arrowOpen:
		return dom.CustomEl("polyline",
			dom.Marker("diagram-arrow"),
			dom.Attr("points", arrowBarb(pts)),
			dom.Attr("stroke-linecap", "round"),
			dom.Attr("stroke-linejoin", "round"),
			dom.Attr("class", arrowOpenClasses()))
	case arrowDiamond:
		return dom.CustomEl("polygon",
			dom.Marker("diagram-arrow"),
			dom.Attr("points", arrowDiamondPts(pts)),
			dom.Attr("class", arrowClasses()))
	case arrowRound:
		cx, cy, r := arrowDot(pts)
		return dom.CustomEl("circle",
			dom.Marker("diagram-arrow"),
			dom.Attr("cx", fmtCoord(cx)),
			dom.Attr("cy", fmtCoord(cy)),
			dom.Attr("r", fmtCoord(r)),
			dom.Attr("class", arrowClasses()))
	default: // arrowSolid
		return dom.CustomEl("polygon",
			dom.Marker("diagram-arrow"),
			dom.Attr("points", arrowTriangle(pts)),
			dom.Attr("class", arrowClasses()))
	}
}

// reversePts returns pts in reverse order, so an end-of-line helper lands on
// the start instead.
func reversePts(pts []xy) []xy {
	out := make([]xy, len(pts))
	for i, p := range pts {
		out[len(pts)-1-i] = p
	}
	return out
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
