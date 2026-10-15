// Package sankey renders a flow diagram as SVG on the server: nodes drawn
// as bars, the flows between them as ribbons whose thickness is the value
// they carry.
//
//	@sankey.New(
//		sankey.Title("Signup funnel"),
//		sankey.Flow("Search", "Landing", 1200),
//		sankey.Flow("Direct", "Landing", 600),
//		sankey.Flow("Landing", "Signup", 900, sankey.Emerald),
//		sankey.Flow("Landing", "Bounce", 900, sankey.Rose),
//	)
//
// Naming an endpoint in a flow declares it, so a drawing needs nothing but
// its flows. sankey.Node(id, ...) relabels or recolors one; it does not
// declare it, and an id no flow mentions is a typo rather than a lone node,
// so it fails loudly.
//
// Layout is d3-sankey's: a node's column is its longest path from a source,
// sinks are pushed out to the last column, and nodes are placed within their
// column by repeated sweeps that pull each toward the weighted center of what
// it connects to, resolving overlaps after every sweep. One value-to-pixel
// scale serves every column, so a ribbon's thickness means the same thing
// wherever it is read. Where a node's inflow and outflow disagree, its bar is
// the larger of the two and the shorter side stacks from the bar's top.
//
// Each node takes a hue from the palette in the order it first appears and
// its bar is painted in it, at the palette's mark strength. A ribbon runs
// between the two hues it connects - a gradient from the node it leaves to
// the node it arrives at - at the wash strength chart fills the area under
// a line with, so a drawing full of crossing ribbons stays a drawing rather
// than a wall of color, and every ribbon meets each bar in that bar's own
// hue. sankey.Node("x", sankey.Teal) recolors a node and everything leaving
// it; sankey.Flow(..., sankey.Rose) overrides one ribbon, which then wears
// that hue end to end rather than blending.
//
// Hovering a ribbon or a bar reveals its value in a regular loom tooltip,
// the HTML overlay chart uses: the SVG scales proportionally, so viewBox
// coordinates map 1:1 to percentage positions over it. A ribbon is a curve
// and a tooltip's wrapper is a box, so a ribbon's hit area is a spot at its
// midpoint rather than its whole length - the same trade chart makes for a
// data point.
//
// The first column's labels sit in a gutter to its left and the last
// column's in one to its right, both sized from the text they hold. An
// intermediate column's labels have nowhere outside to go and are drawn over
// the ribbons leaving it, which a washed ribbon is pale enough to allow.
// Label widths are estimated from a per-character advance - the server
// cannot measure rendered text - which the gutters absorb.
//
// The flows must form a DAG: a cycle has no column order, and is a render
// error rather than a guess.
//
// Naming note: like diagram, this package deviates from loom's convention of
// exporting Node(ctx, ...) as the raw-node render entry. Here Node(id, ...)
// is the option that configures a graph node, which is what the API wants
// the name for; the render function is unexported (build).
package sankey

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/ids"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
	"github.com/pietjan/loom/tooltip"
)

// ErrNoFlows is returned when the diagram has no data.
var ErrNoFlows = errors.New("sankey: at least one sankey.Flow(...) is required")

// ErrSelfLoop is returned for a flow whose ends are the same node.
var ErrSelfLoop = errors.New("sankey: self-loop is not supported")

// ErrNonPositiveValue is returned for a flow carrying zero or less: a
// ribbon's thickness is its value, so there is nothing to draw.
var ErrNonPositiveValue = errors.New("sankey: flow value must be positive")

// ErrCycle is returned when the flows do not form a DAG.
var ErrCycle = errors.New("sankey: flows must not form a cycle")

// ErrUnknownNode is returned when sankey.Node names an id no flow mentions.
var ErrUnknownNode = errors.New("sankey: no flow mentions node")

// mark is the part of a node or a flow that options configure.
type mark struct {
	label    string
	color    Color
	colorSet bool
}

// NodeOption configures a node: its bar and the text beside it.
type NodeOption = func(*mark)

// FlowOption configures a flow. A node and a flow take the same two
// options, so this is NodeOption under the name each signature wants -
// every color below, and Label, serves either one.
type FlowOption = NodeOption

// Label overrides a node's drawn text, or a flow's tooltip text.
func Label(text string) NodeOption { return func(m *mark) { m.label = text } }

// Colored paints a node's bar in a palette hue, or paints one ribbon that
// hue end to end instead of blending between the nodes it joins.
func Colored(c Color) NodeOption {
	return func(m *mark) { m.color, m.colorSet = c, true }
}

// Pre-baked color options.
var (
	Accent  = Colored(ColorAccent)
	Zinc    = Colored(ColorZinc)
	Red     = Colored(ColorRed)
	Orange  = Colored(ColorOrange)
	Amber   = Colored(ColorAmber)
	Yellow  = Colored(ColorYellow)
	Lime    = Colored(ColorLime)
	Green   = Colored(ColorGreen)
	Emerald = Colored(ColorEmerald)
	Teal    = Colored(ColorTeal)
	Cyan    = Colored(ColorCyan)
	Sky     = Colored(ColorSky)
	Blue    = Colored(ColorBlue)
	Indigo  = Colored(ColorIndigo)
	Violet  = Colored(ColorViolet)
	Purple  = Colored(ColorPurple)
	Fuchsia = Colored(ColorFuchsia)
	Pink    = Colored(ColorPink)
	Rose    = Colored(ColorRose)
)

// override is a Node option set aside until its node is interned.
type override struct {
	id string
	mark
}

// Config holds sankey options.
type Config struct {
	opts.Common
	TitleText string
	W, H      float64
	flows     []flow
	overrides []override
	nodeWidth float64
	gap       float64
	totals    bool
	format    func(float64) string
}

// Option configures a sankey.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Title sets the accessible name (aria-label). Without one, the node labels
// are joined. Like chart, no <title> is emitted - browsers surface it as a
// native hover tooltip, over our own.
func Title(title string) Option { return func(c *Config) { c.TitleText = title } }

// Size sets the SVG viewBox (drawing coordinates; the rendered diagram
// scales to its container). Default 640×320.
func Size(w, h int) Option {
	return func(c *Config) { c.W, c.H = float64(w), float64(h) }
}

// Flow adds a flow of value from one node to another, declaring both. The
// ribbon runs from the source node's hue into the target's unless given a
// hue of its own, which it then wears end to end.
func Flow(from, to string, value float64, options ...FlowOption) Option {
	return func(c *Config) {
		f := flow{from: from, to: to, value: value}
		for _, opt := range options {
			opt(&f.mark)
		}
		c.flows = append(c.flows, f)
	}
}

// Node relabels or recolors a node some flow already declares. It does not
// declare one: a node nothing flows through has no place in the drawing.
func Node(id string, options ...NodeOption) Option {
	return func(c *Config) {
		o := override{id: id}
		for _, opt := range options {
			opt(&o.mark)
		}
		c.overrides = append(c.overrides, o)
	}
}

// NodeWidth sets the thickness of a node's bar in viewBox units. The
// default 7 is a chart data point across, which is what makes a bar read as
// the same kind of mark.
func NodeWidth(units int) Option {
	return func(c *Config) {
		if units > 0 {
			c.nodeWidth = float64(units)
		}
	}
}

// Gap sets the vertical space between two nodes in a column, in viewBox
// units (default 14). A column too crowded to hold it gets less, so the
// drawing keeps its height rather than overflowing.
func Gap(units int) Option {
	return func(c *Config) {
		if units >= 0 {
			c.gap = float64(units)
		}
	}
}

// Totals appends each node's total to its label.
func Totals() Option { return func(c *Config) { c.totals = true } }

// Format overrides value formatting for labels and tooltips.
func Format(fn func(float64) string) Option { return func(c *Config) { c.format = fn } }

// New renders a sankey diagram as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return build(ctx, options...)
	})
}

// build resolves the options into a graph, lays it out, and returns the
// two-layer result: a <figure> holding the SVG and its tooltip overlay.
func build(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{W: 640, H: 320, nodeWidth: defaultNodeWidth, gap: defaultGap, format: formatValue}
	for _, opt := range options {
		opt(&cfg)
	}

	g, err := buildGraph(cfg)
	if err != nil {
		return nil, err
	}
	g.layout(cfg)

	// The overlay must cover exactly the SVG's box, so they share a
	// positioning wrapper.
	plot := dom.El(atom.Div, dom.Attr("class", "relative"))
	plot.AppendChild(draw(ctx, cfg, g))
	if err := appendTooltips(ctx, plot, cfg, g); err != nil {
		return nil, err
	}

	fig := dom.El(atom.Figure, dom.Marker("sankey"))
	fig.AppendChild(plot)
	cfg.Apply(fig, rootClasses())
	return fig, nil
}

func draw(ctx context.Context, cfg Config, g *graph) *html.Node {
	svg := dom.El(atom.Svg,
		dom.Marker("sankey-svg"),
		dom.Attr("viewBox", fmt.Sprintf("0 0 %s %s", fmtCoord(cfg.W), fmtCoord(cfg.H))),
		dom.Attr("role", "img"),
		dom.Attr("class", svgClasses()))
	dom.SetAttr(svg, "aria-label", accessibleName(cfg, g))

	defs, gradient := gradients(ctx, g)
	svg.AppendChild(defs)

	// Ribbons first: the bars and the labels are drawn over them.
	for i, f := range g.flows {
		src, dst := g.nodes[f.src], g.nodes[f.dst]
		svg.AppendChild(dom.CustomEl("path", dom.Marker("sankey-link"),
			dom.Attr("d", ribbonPath(src.x1, dst.x0, f.sy, f.ty, f.w)),
			dom.Attr("fill", "url(#"+gradient[i]+")"),
			dom.Attr("fill-opacity", washAlpha)))
	}
	for _, n := range g.nodes {
		svg.AppendChild(dom.CustomEl("rect", dom.Marker("sankey-node"),
			dom.Attr("x", fmtCoord(n.x0)),
			dom.Attr("y", fmtCoord(n.y0)),
			dom.Attr("width", fmtCoord(n.x1-n.x0)),
			dom.Attr("height", fmtCoord(n.y1-n.y0)),
			// Rounded to a capsule, which is what every other bar in loom
			// is: progress rounds both its track and its fill full.
			dom.Attr("rx", fmtCoord((n.x1-n.x0)/2)),
			// chart's point stroke, straddling the bar's edge the same way.
			dom.Attr("stroke-width", "1.5"),
			dom.Attr("class", nodeClasses(n.color))))
	}
	for _, n := range g.nodes {
		svg.AppendChild(label(g, n))
	}
	return svg
}

// gradients builds one <linearGradient> per hue pair in use and returns
// the <defs> holding them plus, for each flow, the id of the one it takes.
// Flows sharing a pair share a definition - a drawing has far more ribbons
// than it has pairs. The gradient spans each ribbon's own bounding box
// (the default objectBoundingBox units), so one definition serves ribbons
// of any length or position.
func gradients(ctx context.Context, g *graph) (*html.Node, []string) {
	defs := dom.CustomEl("defs")
	byFlow := make([]string, len(g.flows))
	seen := map[[2]Color]string{}

	for i, f := range g.flows {
		pair := [2]Color{f.color, g.nodes[f.dst].color}
		if f.colorSet {
			// A ribbon given a color of its own wears it end to end: the
			// option named one hue, not a blend into another.
			pair[1] = f.color
		}
		id, ok := seen[pair]
		if !ok {
			id = ids.New(ctx, "sankey-flow")
			seen[pair] = id

			grad := dom.CustomEl("linearGradient", dom.Attr("id", id))
			for _, stop := range []struct {
				offset string
				color  Color
			}{{"0%", pair[0]}, {"100%", pair[1]}} {
				grad.AppendChild(dom.CustomEl("stop",
					dom.Attr("offset", stop.offset),
					dom.Attr("class", stopClasses(stop.color))))
			}
			defs.AppendChild(grad)
		}
		byFlow[i] = id
	}
	return defs, byFlow
}

// label draws a node's text: outside the drawing for the two end columns,
// which have a gutter for it, and over the outgoing ribbons otherwise - a
// washed ribbon is pale enough to read a tick label through.
func label(g *graph, n *node) *html.Node {
	x, anchor := n.x1+labelGap, "start"
	if n.depth == 0 {
		x, anchor = n.x0-labelGap, "end"
	}
	t := dom.CustomEl("text", dom.Marker("sankey-label"),
		dom.Attr("x", fmtCoord(x)),
		dom.Attr("y", fmtCoord(n.center())),
		dom.Attr("text-anchor", anchor),
		dom.Attr("dominant-baseline", "middle"),
		dom.Attr("class", labelClasses()))
	t.AppendChild(dom.Text(n.text))
	return t
}

// appendTooltips adds the HTML hover layer: one loom tooltip per ribbon and
// per bar, anchored over that mark's viewBox geometry expressed as
// percentages of the plot box. A bar is a rectangle and is covered exactly;
// a ribbon is a curve, so it gets a hit spot at its midpoint, which is where
// both of its cubics evaluate to the mean of their ends.
func appendTooltips(ctx context.Context, plot *html.Node, cfg Config, g *graph) error {
	// The hoverable surface fills whatever box its tooltip wrapper defines.
	fill := render.Component(func(context.Context) (*html.Node, error) {
		return dom.El(atom.Span, dom.Attr("class", "block size-full")), nil
	})
	add := func(text, class, style string) error {
		tip, err := tooltip.Node(templ.WithChildren(ctx, fill),
			tooltip.Text(text),
			tooltip.Class(class),
			tooltip.Attr("style", style))
		if err != nil {
			return err
		}
		plot.AppendChild(tip)
		return nil
	}

	for _, f := range g.flows {
		src, dst := g.nodes[f.src], g.nodes[f.dst]
		text := f.label
		if text == "" {
			text = fmt.Sprintf("%s → %s: %s", src.label, dst.label, cfg.format(f.value))
		}
		x := (src.x1 + dst.x0) / 2
		y := (f.sy+f.ty)/2 + f.w/2
		style := fmt.Sprintf("left: %s%%; top: %s%%",
			fmtCoord(x/cfg.W*100), fmtCoord(y/cfg.H*100))
		if err := add(text, "absolute size-5 -translate-x-1/2 -translate-y-1/2", style); err != nil {
			return err
		}
	}

	for _, n := range g.nodes {
		// Insets rather than a height: a percentage height would resolve
		// against a stage whose own height comes from the SVG. The bar is
		// only a chart point across, too narrow to aim at, so the target is
		// widened around it - the same thing chart does for its points.
		style := fmt.Sprintf("left: %s%%; top: %s%%; bottom: %s%%",
			fmtCoord((n.x0+n.x1)/2/cfg.W*100), fmtCoord(n.y0/cfg.H*100),
			fmtCoord((cfg.H-n.y1)/cfg.H*100))
		if err := add(n.label+": "+cfg.format(n.value), "absolute w-5 -translate-x-1/2", style); err != nil {
			return err
		}
	}
	return nil
}

// accessibleName is the title, or the node labels joined.
func accessibleName(cfg Config, g *graph) string {
	if cfg.TitleText != "" {
		return cfg.TitleText
	}
	names := make([]string, 0, len(g.nodes))
	for _, n := range g.nodes {
		names = append(names, n.label)
	}
	return strings.Join(names, ", ")
}
